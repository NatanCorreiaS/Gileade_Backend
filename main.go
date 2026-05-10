package main

import (
	"log"
	"os"
	"strconv"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/db"
	"gileade/gileade_backend/gateway"

	"github.com/joho/godotenv"
)

// Inicializa o servidor HTTP com rotas de integração Mercado Pago (Checkout Pro).
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
