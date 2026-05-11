package jobs

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/gateway"
	"gileade/gileade_backend/repository"
	"gileade/gileade_backend/service"

	"github.com/mercadopago/sdk-go/pkg/payment"
	"gorm.io/gorm"
)

const (
	defaultPollIntervalSeconds = 60
	defaultPollBatchSize       = 50
)

// StartPagamentoPolling inicia o job de confirmacao periodica de pagamentos.
func StartPagamentoPolling(ctx context.Context, db *gorm.DB, gw *gateway.MercadoPagoGateway) {
	interval := readEnvInt("PAGAMENTO_POLL_INTERVAL_SECONDS", defaultPollIntervalSeconds)
	batchSize := readEnvInt("PAGAMENTO_POLL_BATCH_SIZE", defaultPollBatchSize)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	service := service.NewPagamentoService(db, gw)
	tuRepo := repository.NewTicketUsuarioRepository(db)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processarPendentes(ctx, service, tuRepo, batchSize)
		}
	}
}

// processarPendentes busca tickets pendentes e confirma pagamentos aprovados.
func processarPendentes(ctx context.Context, svc *service.PagamentoService, tuRepo *repository.TicketUsuarioRepository, batchSize int) {
	pendentes, err := tuRepo.ListByStatus(ctx, model.TicketsStatusPendente, batchSize, 0)
	if err != nil {
		log.Printf("poll pagamentos: falha ao listar pendentes: %v", err)
		return
	}

	for _, tu := range pendentes {
		resp, err := consultarPagamento(ctx, svc, tu.ID)
		if err != nil {
			log.Printf("poll pagamentos: tu=%d erro=%v", tu.ID, err)
			continue
		}
		_ = resp
	}
}

// consultarPagamento busca pagamentos aprovados para um ticket_usuario.
func consultarPagamento(ctx context.Context, svc *service.PagamentoService, ticketUsuarioID uint64) (*payment.Response, error) {
	searchReq := payment.SearchRequest{
		Limit:  1,
		Offset: 0,
		Filters: map[string]string{
			"external_reference": strconv.FormatUint(ticketUsuarioID, 10),
			"status":             "approved",
		},
	}

	searchResp, err := svc.SearchPayments(ctx, searchReq)
	if err != nil {
		return nil, err
	}
	if len(searchResp.Results) == 0 {
		return nil, nil
	}

	result := searchResp.Results[0]
	_, err = svc.ConfirmarPagamento(ctx, ticketUsuarioID, result.ID)
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// readEnvInt le um inteiro do ambiente com fallback seguro.
func readEnvInt(name string, def int) int {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		return def
	}
	return parsed
}
