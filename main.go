package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/db"
	"gileade/gileade_backend/gateway"
	"gileade/gileade_backend/repository"
	"gileade/gileade_backend/service"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// main inicializa o servidor HTTP e dependencias do app.
func main() {
	// Em desenvolvimento, carrega variáveis a partir do arquivo .env.
	// Em produção, as variáveis devem vir do ambiente do processo.
	_ = godotenv.Load()

	cfg, err := db.NewPostgresConfigFromEnv()
	if err != nil {
		log.Fatalf("config DB inválida: %v", err)
	}

	dbConn, err := db.OpenPostgres(cfg)
	if err != nil {
		log.Fatalf("falha ao conectar no Postgres: %v", err)
	}

	if err := model.AutoMigrate(dbConn); err != nil {
		log.Fatalf("falha no AutoMigrate: %v", err)
	}

	if err := seedAdminUsers(dbConn); err != nil {
		log.Fatalf("falha ao semear usuarios admin: %v", err)
	}

	gw, err := gateway.NewMercadoPagoGatewayFromEnv()
	if err != nil {
		log.Fatalf("config Mercado Pago inválida: %v", err)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("APP_PORT inválida: %v", err)
	}

	r := NewRouter(AppDeps{DB: dbConn, MP: gw})
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("falha ao iniciar servidor: %v", err)
	}
}

// seedAdminUsers cria os usuarios administradores se ainda nao existirem.
// As senhas sao lidas das variaveis de ambiente ADMIN_PASSWORD e BACKUP_PASSWORD.
func seedAdminUsers(db *gorm.DB) error {
	authService := service.NewAuthService(db)
	pessoaRepo := repository.NewPessoaRepository(db)

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Printf("ADMIN_PASSWORD nao definida, pulando criacao do admin")
	} else {
		if err := seedAdmin(db, authService, pessoaRepo, "00000000191", "Administrador", adminPassword); err != nil {
			return err
		}
	}

	backupPassword := os.Getenv("BACKUP_PASSWORD")
	if backupPassword == "" {
		log.Printf("BACKUP_PASSWORD nao definida, pulando criacao do backup")
	} else {
		if err := seedAdmin(db, authService, pessoaRepo, "00000000291", "Backup Admin", backupPassword); err != nil {
			return err
		}
	}

	return nil
}

func seedAdmin(db *gorm.DB, authService *service.AuthService, pessoaRepo *repository.PessoaRepository, cpf, nome, senha string) error {
	ctx := context.Background()

	_, err := pessoaRepo.GetByCPF(ctx, cpf)
	if err == nil {
		log.Printf("usuario admin '%s' (cpf=%s) ja existe", nome, cpf)
		return nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	senhaHash, err := authService.HashPassword(senha)
	if err != nil {
		return err
	}

	pessoa := model.Pessoa{
		Nome:         nome,
		TipoUsuario:  model.TipoUsuarioAdmin,
		Senha:        senhaHash,
		CPF:          cpf,
		Idade:        0,
		Celular:      "",
		Igreja:       "",
		PapelIgreja:  model.PapelIgrejaMembro,
		EstadoCivil:  model.EstadoCivilSolteiro,
		Email:        "",
		Sexo:         model.SexoMasculino,
		Cidade:       "",
		EstadoUF:     model.EstadoUFSaoPaulo,
		Escolaridade: model.EscolaridadeEnsinoSuperiorCompleto,
	}

	if err := pessoaRepo.Create(ctx, &pessoa); err != nil {
		return err
	}

	log.Printf("usuario admin '%s' criado com sucesso (id=%d)", nome, pessoa.ID)
	return nil
}
