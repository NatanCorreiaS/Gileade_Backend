package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/gateway"
	"gileade/gileade_backend/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
	"github.com/mercadopago/sdk-go/pkg/refund"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CheckoutRequest struct {
	UsuarioID       uint64
	TicketID        uint64
	SuccessURL      string
	FailureURL      string
	PendingURL      string
	NotificationURL string
}

type CheckoutResponse struct {
	PreferenceID    string
	InitPoint       string
	SandboxInit     string
	TicketUsuarioID uint64
}

var ErrNotificationURLObrigatoria = errors.New("notification_url obrigatoria")
var ErrExternalReferenceInvalida = errors.New("external_reference invalida")

type PagamentoService struct {
	pRepo   *repository.PessoaRepository
	tRepo   *repository.TicketRepository
	tuRepo  *repository.TicketUsuarioRepository
	payRepo *repository.PagamentoRepository
	estRepo *repository.EstornoRepository
	gw      *gateway.MercadoPagoGateway
}

func NewPagamentoService(db *gorm.DB, gw *gateway.MercadoPagoGateway) *PagamentoService {
	return &PagamentoService{
		pRepo:   repository.NewPessoaRepository(db),
		tRepo:   repository.NewTicketRepository(db),
		tuRepo:  repository.NewTicketUsuarioRepository(db),
		payRepo: repository.NewPagamentoRepository(db),
		estRepo: repository.NewEstornoRepository(db),
		gw:      gw,
	}
}

// CriarCheckout cria o checkout e persiste o ticket do usuario como pendente.
func (s *PagamentoService) CriarCheckout(ctx context.Context, req CheckoutRequest) (CheckoutResponse, error) {
	pessoa, err := s.pRepo.GetByID(ctx, req.UsuarioID)
	if err != nil {
		return CheckoutResponse{}, err
	}

	ticket, err := s.tRepo.GetByID(ctx, req.TicketID)
	if err != nil {
		return CheckoutResponse{}, err
	}

	tu := model.TicketUsuario{
		UsuarioID: pessoa.ID,
		TicketID:  ticket.ID,
		Status:    model.TicketsStatusPendente,
	}
	if err := s.tuRepo.Create(ctx, &tu); err != nil {
		return CheckoutResponse{}, err
	}

	price, _ := ticket.Preco.Float64()
	prefReq := preference.Request{
		Items: []preference.ItemRequest{
			{
				ID:          fmt.Sprintf("%d", ticket.ID),
				Title:       ticket.Nome,
				Description: ticket.Descricao,
				CurrencyID:  "BRL",
				UnitPrice:   price,
				Quantity:    1,
			},
		},
		// Configuração do Payer obrigatória para Cartões de Crédito e Testes de Cenário
		Payer: &preference.PayerRequest{
			Name:  pessoa.Nome, // Para testes, altere no banco para 'APRO', 'FUND', etc.
			Email: pessoa.Email,
			Identification: &preference.IdentificationRequest{
				Type:   "CPF",
				Number: "12345678909", // CPF de teste padrão da documentação
			},
		},
		ExternalReference: fmt.Sprintf("%d", tu.ID),
		BinaryMode:        true,
		Metadata: map[string]any{
			"ticket_usuario_id": tu.ID,
			"usuario_id":        pessoa.ID,
			"ticket_id":         ticket.ID,
		},
	}

	if req.SuccessURL != "" || req.FailureURL != "" || req.PendingURL != "" {
		prefReq.BackURLs = &preference.BackURLsRequest{
			Success: req.SuccessURL,
			Failure: req.FailureURL,
			Pending: req.PendingURL,
		}
	}
	
	if req.SuccessURL != "" {
		prefReq.AutoReturn = "approved"
	}

	notificationURL := req.NotificationURL
	if notificationURL == "" {
		notificationURL = os.Getenv("MERCADO_PAGO_NOTIFICATION_URL")
	}
	if notificationURL == "" {
		_ = s.tuRepo.Delete(ctx, tu.ID)
		return CheckoutResponse{}, ErrNotificationURLObrigatoria
	}
	prefReq.NotificationURL = notificationURL

	prefResp, err := s.gw.CreateCheckoutPro(ctx, prefReq)
	if err != nil {
		_ = s.tuRepo.Delete(ctx, tu.ID)
		return CheckoutResponse{}, err
	}
	
	if err := s.tuRepo.UpdatePreferenceID(ctx, tu.ID, prefResp.ID); err != nil {
		return CheckoutResponse{}, err
	}

	return CheckoutResponse{
		PreferenceID:    prefResp.ID,
		InitPoint:       prefResp.InitPoint,
		SandboxInit:     prefResp.SandboxInitPoint,
		TicketUsuarioID: tu.ID,
	}, nil
}

type WebhookResultado struct {
	Status          string
	TicketUsuarioID uint64
	UsuarioID       uint64
	TicketID        uint64
	CPF             string
}

// ProcessarPagamentoWebhook confirma pagamento aprovado e persiste o resultado.
func (s *PagamentoService) ProcessarPagamentoWebhook(ctx context.Context, paymentID int) (WebhookResultado, error) {
	payResp, err := s.gw.GetPayment(ctx, paymentID)
	if err != nil {
		return WebhookResultado{}, err
	}

	result := WebhookResultado{Status: payResp.Status}
	
	// Se o status for rejeitado ou outro que não seja aprovado, apenas retornamos o status
	if payResp.Status != "approved" {
		return result, nil
	}

	tuID, err := strconv.ParseUint(payResp.ExternalReference, 10, 64)
	if err != nil {
		return WebhookResultado{}, ErrExternalReferenceInvalida
	}

	tu, err := s.tuRepo.GetByID(ctx, tuID)
	if err != nil {
		return WebhookResultado{}, err
	}
	
	result.TicketUsuarioID = tu.ID
	result.UsuarioID = tu.UsuarioID
	result.TicketID = tu.TicketID
	result.CPF = tu.Usuario.CPF

	dataPagamento := time.Now().UTC()
	if !payResp.DateApproved.IsZero() {
		dataPagamento = payResp.DateApproved
	}

	pagamento := model.Pagamento{
		IDTransacao:      fmt.Sprintf("%d", payResp.ID),
		Valor:            decimal.NewFromFloat(payResp.TransactionAmount),
		TicketsUsuarioID: tuID,
		Metodo:           mapMetodoPagamento(payResp.PaymentTypeID),
		DataPagamento:    dataPagamento,
	}

	if err := s.payRepo.CreateAndMarkTicketPago(ctx, &pagamento); err != nil {
		if isUniqueViolation(err) {
			result.Status = "duplicado"
			return result, nil
		}
		if errors.Is(err, repository.ErrTicketIndisponivel) {
			result.Status = "ticket_indisponivel"
			return result, err
		}
		return WebhookResultado{}, err
	}

	result.Status = "ok"
	return result, nil
}

func (s *PagamentoService) SearchPayments(ctx context.Context, req payment.SearchRequest) (*payment.SearchResponse, error) {
	return s.gw.SearchPayments(ctx, req)
}

func (s *PagamentoService) CriarEstornoPorPagamentoID(ctx context.Context, pagamentoID uint64, motivo string, valor *decimal.Decimal) (model.Estorno, error) {
	pagamento, err := s.payRepo.GetByID(ctx, pagamentoID)
	if err != nil {
		return model.Estorno{}, err
	}

	paymentID, err := strconv.Atoi(pagamento.IDTransacao)
	if err != nil {
		return model.Estorno{}, fmt.Errorf("id_transacao invalido")
	}

	refundResp, err := s.criarRefund(ctx, paymentID, valor)
	if err != nil {
		return model.Estorno{}, err
	}

	dataEstorno := time.Now().UTC()
	if !refundResp.DateCreated.IsZero() {
		dataEstorno = refundResp.DateCreated
	}

	estorno := model.Estorno{
		PagamentoID:        pagamento.ID,
		IDTransacaoEstorno: fmt.Sprintf("%d", refundResp.ID),
		Valor:              decimal.NewFromFloat(refundResp.Amount),
		Motivo:             motivo,
		DataEstorno:        dataEstorno,
	}

	if err := s.estRepo.CreateAndMarkTicketReembolsado(ctx, &estorno); err != nil {
		if isUniqueViolation(err) {
			return estorno, nil
		}
		return model.Estorno{}, err
	}

	return estorno, nil
}

func (s *PagamentoService) CriarEstornoPorPaymentID(ctx context.Context, paymentID int, motivo string, valor *decimal.Decimal) (model.Estorno, error) {
	pagamento, err := s.payRepo.GetByIDTransacao(ctx, fmt.Sprintf("%d", paymentID))
	if err != nil {
		return model.Estorno{}, err
	}

	refundResp, err := s.criarRefund(ctx, paymentID, valor)
	if err != nil {
		return model.Estorno{}, err
	}

	dataEstorno := time.Now().UTC()
	if !refundResp.DateCreated.IsZero() {
		dataEstorno = refundResp.DateCreated
	}

	estorno := model.Estorno{
		PagamentoID:        pagamento.ID,
		IDTransacaoEstorno: fmt.Sprintf("%d", refundResp.ID),
		Valor:              decimal.NewFromFloat(refundResp.Amount),
		Motivo:             motivo,
		DataEstorno:        dataEstorno,
	}

	if err := s.estRepo.CreateAndMarkTicketReembolsado(ctx, &estorno); err != nil {
		if isUniqueViolation(err) {
			return estorno, nil
		}
		return model.Estorno{}, err
	}

	return estorno, nil
}

func (s *PagamentoService) criarRefund(ctx context.Context, paymentID int, valor *decimal.Decimal) (*refund.Response, error) {
	if valor != nil {
		floatVal, _ := valor.Float64()
		return s.gw.CreatePartialRefund(ctx, paymentID, floatVal)
	}
	return s.gw.CreateRefund(ctx, paymentID)
}

func mapMetodoPagamento(paymentTypeID string) model.MetodoPagamento {
	switch paymentTypeID {
	case "credit_card", "debit_card":
		return model.MetodoPagamentoCartaoCredito
	case "ticket":
		return model.MetodoPagamentoBoleto
	case "pix":
		return model.MetodoPagamentoPix
	default:
		return model.MetodoPagamentoPix
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}