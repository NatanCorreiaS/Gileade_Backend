package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/db"
	"gileade/gileade_backend/repository"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

// Smoke test de operações básicas.
// Executa um fluxo completo (CRUD + transações) contra o Postgres configurado em variáveis de ambiente.
// Observação: este arquivo não inicializa servidor HTTP ainda.
func main() {
	// Em desenvolvimento, carrega variáveis a partir do arquivo .env.
	// Em produção, as variáveis devem vir do ambiente do processo.
	_ = godotenv.Load()

	portStr := os.Getenv("DB_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("DB_PORT inválida: %v", err)
	}

	cfg := db.PostgresConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		TimeZone: os.Getenv("DB_TIMEZONE"),
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = "UTC"
	}

	if cfg.Host == "" || cfg.User == "" || cfg.Password == "" || cfg.DBName == "" || portStr == "" {
		log.Fatalf("config DB inválida: defina DB_HOST, DB_PORT, DB_USER, DB_PASSWORD e DB_NAME")
	}

	dbConn, err := db.OpenPostgres(cfg)
	if err != nil {
		log.Fatalf("falha ao conectar no Postgres: %v", err)
	}

	if err := model.AutoMigrate(dbConn); err != nil {
		log.Fatalf("falha no AutoMigrate: %v", err)
	}

	ctx := context.Background()

	pRepo := repository.NewPessoaRepository(dbConn)
	tRepo := repository.NewTicketRepository(dbConn)
	tuRepo := repository.NewTicketUsuarioRepository(dbConn)
	payRepo := repository.NewPagamentoRepository(dbConn)
	estRepo := repository.NewEstornoRepository(dbConn)

	unique := time.Now().UnixNano()
	cpf := fmt.Sprintf("%011d", unique%100000000000)

	pessoa := model.Pessoa{
		Nome:        "Smoke Test",
		TipoUsuario: model.TipoUsuarioUsuario,
		Senha:       "hash",
		CPF:         cpf,
		Idade:       30,
		Email:       fmt.Sprintf("smoke-%d@example.com", unique),
		Sexo:        model.SexoMasculino,
		EstadoUF:    model.EstadoUFSaoPaulo,
		EstadoCivil: model.EstadoCivilSolteiro,
	}
	if err := pRepo.Create(ctx, &pessoa); err != nil {
		log.Fatalf("Pessoa.Create: %v", err)
	}

	if _, err := pRepo.GetByCPF(ctx, pessoa.CPF); err != nil {
		log.Fatalf("Pessoa.GetByCPF: %v", err)
	}

	ticket := model.Ticket{
		Nome:                 "Ingresso Smoke",
		Descricao:            "Executado pelo main.go",
		Preco:                decimal.NewFromFloat(10.00),
		QuantidadeDisponivel: 1,
		DataEvento:           time.Now().UTC().Add(24 * time.Hour),
	}
	if err := tRepo.Create(ctx, &ticket); err != nil {
		log.Fatalf("Ticket.Create: %v", err)
	}

	tu := model.TicketUsuario{
		UsuarioID: pessoa.ID,
		TicketID:  ticket.ID,
		Status:    model.TicketsStatusPendente,
	}
	if err := tuRepo.Create(ctx, &tu); err != nil {
		log.Fatalf("TicketUsuario.Create: %v", err)
	}

	// Pagamento transacional: cria pagamento + marca TicketUsuario como Pago.
	pagamento := model.Pagamento{
		IDTransacao:      fmt.Sprintf("tx-%d", unique),
		Valor:            decimal.NewFromFloat(10.00),
		TicketsUsuarioID: tu.ID,
		Metodo:           model.MetodoPagamentoPix,
		DataPagamento:    time.Now().UTC(),
	}
	if err := payRepo.CreateAndMarkTicketPago(ctx, &pagamento); err != nil {
		log.Fatalf("Pagamento.CreateAndMarkTicketPago: %v", err)
	}

	gotTU, err := tuRepo.GetByID(ctx, tu.ID)
	if err != nil {
		log.Fatalf("TicketUsuario.GetByID: %v", err)
	}
	if gotTU.Status != model.TicketsStatusPago {
		log.Fatalf("status esperado %q, obtido %q", model.TicketsStatusPago, gotTU.Status)
	}

	if _, err := payRepo.GetByIDTransacao(ctx, pagamento.IDTransacao); err != nil {
		log.Fatalf("Pagamento.GetByIDTransacao: %v", err)
	}

	if pays, err := payRepo.ListByTicketsUsuarioID(ctx, tu.ID, 10, 0); err != nil {
		log.Fatalf("Pagamento.ListByTicketsUsuarioID: %v", err)
	} else {
		log.Printf("pagamentos do ticket_usuario=%d: %d", tu.ID, len(pays))
	}

	// Estorno transacional: cria estorno + marca TicketUsuario como Reembolsado.
	estorno := model.Estorno{
		PagamentoID:        pagamento.ID,
		IDTransacaoEstorno: fmt.Sprintf("st-%d", unique),
		Valor:              decimal.NewFromFloat(10.00),
		Motivo:             "smoke test",
		DataEstorno:        time.Now().UTC(),
	}
	if err := estRepo.CreateAndMarkTicketReembolsado(ctx, &estorno); err != nil {
		log.Fatalf("Estorno.CreateAndMarkTicketReembolsado: %v", err)
	}

	gotTU2, err := tuRepo.GetByID(ctx, tu.ID)
	if err != nil {
		log.Fatalf("TicketUsuario.GetByID (após estorno): %v", err)
	}
	if gotTU2.Status != model.TicketsStatusReembolsado {
		log.Fatalf("status esperado %q, obtido %q", model.TicketsStatusReembolsado, gotTU2.Status)
	}

	if _, err := estRepo.GetByIDTransacaoEstorno(ctx, estorno.IDTransacaoEstorno); err != nil {
		log.Fatalf("Estorno.GetByIDTransacaoEstorno: %v", err)
	}

	if ests, err := estRepo.ListByPagamentoID(ctx, pagamento.ID, 10, 0); err != nil {
		log.Fatalf("Estorno.ListByPagamentoID: %v", err)
	} else {
		log.Printf("estornos do pagamento=%d: %d", pagamento.ID, len(ests))
	}

	// Exercita updates simples.
	pessoa.Nome = "Smoke Test 2"
	if err := pRepo.Update(ctx, &pessoa); err != nil {
		log.Fatalf("Pessoa.Update: %v", err)
	}

	ticket.Descricao = "Atualizado pelo smoke test"
	if err := tRepo.Update(ctx, &ticket); err != nil {
		log.Fatalf("Ticket.Update: %v", err)
	}

	if err := tuRepo.UpdateStatus(ctx, tu.ID, model.TicketsStatusCancelado); err != nil {
		log.Fatalf("TicketUsuario.UpdateStatus: %v", err)
	}

	pagamento.Metodo = model.MetodoPagamentoBoleto
	if err := payRepo.Update(ctx, &pagamento); err != nil {
		log.Fatalf("Pagamento.Update: %v", err)
	}

	// Cleanup (ordem por constraints RESTRICT).
	if err := estRepo.Delete(ctx, estorno.ID); err != nil {
		log.Fatalf("Estorno.Delete: %v", err)
	}
	if err := payRepo.Delete(ctx, pagamento.ID); err != nil {
		log.Fatalf("Pagamento.Delete: %v", err)
	}
	if err := tuRepo.Delete(ctx, tu.ID); err != nil {
		log.Fatalf("TicketUsuario.Delete: %v", err)
	}
	if err := tRepo.Delete(ctx, ticket.ID); err != nil {
		log.Fatalf("Ticket.Delete: %v", err)
	}
	if err := pRepo.Delete(ctx, pessoa.ID); err != nil {
		log.Fatalf("Pessoa.Delete: %v", err)
	}

	log.Printf("smoke test finalizado com sucesso")
}
