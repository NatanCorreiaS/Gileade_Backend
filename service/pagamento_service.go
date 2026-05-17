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
	Quantidade      uint64
	SuccessURL      string
	FailureURL      string
	PendingURL      string
	CPFBeneficiados []string
}

type CheckoutResponse struct {
	PreferenceID   string
	InitPoint      string
	SandboxInit    string
	TicketCompraID uint64
}

var ErrNotificationURLObrigatoria = errors.New("notification_url obrigatoria")
var ErrExternalReferenceInvalida = errors.New("external_reference invalida")
var ErrQuantidadeInvalida = errors.New("quantidade invalida")
var ErrCPFBeneficiadosInvalido = errors.New("cpf_beneficiados invalido")

type PagamentoService struct {
	pRepo   *repository.PessoaRepository
	tRepo   *repository.TicketRepository
	tcRepo  *repository.TicketCompraRepository
	payRepo *repository.PagamentoRepository
	estRepo *repository.EstornoRepository
	gw      *gateway.MercadoPagoGateway
}

func NewPagamentoService(db *gorm.DB, gw *gateway.MercadoPagoGateway) *PagamentoService {
	return &PagamentoService{
		pRepo:   repository.NewPessoaRepository(db),
		tRepo:   repository.NewTicketRepository(db),
		tcRepo:  repository.NewTicketCompraRepository(db),
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

	quantidade := req.Quantidade
	if quantidade == 0 {
		quantidade = 1
	}

	unidadesPorTicket, err := unidadesPorTicket(ticket.Tipo)
	if err != nil {
		return CheckoutResponse{}, err
	}
	quantidadeBeneficiarios := int(unidadesPorTicket * quantidade)
	beneficiados, err := normalizarBeneficiados(req.CPFBeneficiados, quantidadeBeneficiarios)
	if err != nil {
		return CheckoutResponse{}, err
	}

	tc := model.TicketCompra{
		UsuarioID:  pessoa.ID,
		TicketID:   ticket.ID,
		Status:     model.TicketsStatusPendente,
		Quantidade: quantidade,
	}

	var individuais []model.TicketIndividual
	var duos []model.TicketDuo
	var caravanas []model.TicketCaravana

	switch ticket.Tipo {
	case model.TipoTicketIndividual, "":
		individuais = buildTicketsIndividual(beneficiados)
	case model.TipoTicketDuo:
		duos, err = buildTicketsDuo(beneficiados)
		if err != nil {
			return CheckoutResponse{}, err
		}
	case model.TipoTicketCaravana:
		caravanas, err = buildTicketsCaravana(beneficiados)
		if err != nil {
			return CheckoutResponse{}, err
		}
	default:
		return CheckoutResponse{}, repository.ErrTipoTicketInvalido
	}

	if err := s.tcRepo.CreateWithDetalhes(ctx, &tc, individuais, duos, caravanas); err != nil {
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
				Quantity:    int(quantidade),
			},
		},
		// Configuração do Payer obrigatória para Cartões de Crédito e Testes de Cenário
		Payer: &preference.PayerRequest{
			Name:    pessoa.Nome, // Para testes, altere no banco para 'APRO', 'FUND', etc.
			Surname: pessoa.Nome, // ADICIONADO: O Mercado Pago costuma rejeitar cartões sem sobrenome
			Email:   pessoa.Email,
			Identification: &preference.IdentificationRequest{
				Type:   "CPF",
				Number: pessoa.CPF, // CORRIGIDO: Agora usa o CPF dinâmico. Para testar, cadastre a Pessoa com o CPF 12345678909
			},
		},
		ExternalReference: fmt.Sprintf("%d", tc.ID),
		BinaryMode:        true,
		Metadata: map[string]any{
			"ticket_compra_id": tc.ID,
			"usuario_id":       pessoa.ID,
			"ticket_id":        ticket.ID,
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

	notificationURL := os.Getenv("MERCADO_PAGO_NOTIFICATION_URL")
	if notificationURL == "" {
		_ = s.tcRepo.Delete(ctx, tc.ID)
		return CheckoutResponse{}, ErrNotificationURLObrigatoria
	}
	prefReq.NotificationURL = notificationURL

	prefResp, err := s.gw.CreateCheckoutPro(ctx, prefReq)
	if err != nil {
		_ = s.tcRepo.Delete(ctx, tc.ID)
		return CheckoutResponse{}, err
	}

	if err := s.tcRepo.UpdatePreferenceID(ctx, tc.ID, prefResp.ID); err != nil {
		return CheckoutResponse{}, err
	}

	return CheckoutResponse{
		PreferenceID:   prefResp.ID,
		InitPoint:      prefResp.InitPoint,
		SandboxInit:    prefResp.SandboxInitPoint,
		TicketCompraID: tc.ID,
	}, nil
}

type WebhookResultado struct {
	Status         string
	TicketCompraID uint64
	UsuarioID      uint64
	TicketID       uint64
	CPF            string
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

	tc, err := s.tcRepo.GetByID(ctx, tuID)
	if err != nil {
		return WebhookResultado{}, err
	}

	result.TicketCompraID = tc.ID
	result.UsuarioID = tc.UsuarioID
	result.TicketID = tc.TicketID
	result.CPF = tc.Usuario.CPF

	dataPagamento := time.Now().UTC()
	if !payResp.DateApproved.IsZero() {
		dataPagamento = payResp.DateApproved
	}

	pagamento := model.Pagamento{
		IDTransacao:    fmt.Sprintf("%d", payResp.ID),
		Valor:          decimal.NewFromFloat(payResp.TransactionAmount),
		TicketCompraID: tuID,
		Metodo:         mapMetodoPagamento(payResp.PaymentTypeID),
		DataPagamento:  dataPagamento,
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

func unidadesPorTicket(tipo model.TipoTicket) (uint64, error) {
	switch tipo {
	case model.TipoTicketIndividual, "":
		return 1, nil
	case model.TipoTicketDuo:
		return 2, nil
	case model.TipoTicketCaravana:
		return 10, nil
	default:
		return 0, repository.ErrTipoTicketInvalido
	}
}

func normalizarBeneficiados(beneficiados []string, total int) ([]string, error) {
	if total <= 0 {
		return nil, ErrQuantidadeInvalida
	}
	if len(beneficiados) == 0 {
		return make([]string, total), nil
	}
	if len(beneficiados) != total {
		return nil, ErrCPFBeneficiadosInvalido
	}
	return beneficiados, nil
}

func buildTicketsIndividual(beneficiados []string) []model.TicketIndividual {
	if len(beneficiados) == 0 {
		return nil
	}
	individuais := make([]model.TicketIndividual, 0, len(beneficiados))
	for _, cpf := range beneficiados {
		individuais = append(individuais, model.TicketIndividual{CPFBeneficiado: cpf})
	}
	return individuais
}

func buildTicketsDuo(beneficiados []string) ([]model.TicketDuo, error) {
	if len(beneficiados) == 0 {
		return nil, nil
	}
	if len(beneficiados)%2 != 0 {
		return nil, ErrCPFBeneficiadosInvalido
	}
	duos := make([]model.TicketDuo, 0, len(beneficiados)/2)
	for i := 0; i < len(beneficiados); i += 2 {
		duos = append(duos, model.TicketDuo{CPFBeneficiados: model.StringArray{beneficiados[i], beneficiados[i+1]}})
	}
	return duos, nil
}

func buildTicketsCaravana(beneficiados []string) ([]model.TicketCaravana, error) {
	if len(beneficiados) == 0 {
		return nil, nil
	}
	if len(beneficiados)%10 != 0 {
		return nil, ErrCPFBeneficiadosInvalido
	}
	caravanas := make([]model.TicketCaravana, 0, len(beneficiados)/10)
	for i := 0; i < len(beneficiados); i += 10 {
		grupo := make([]string, 10)
		copy(grupo, beneficiados[i:i+10])
		caravanas = append(caravanas, model.TicketCaravana{CPFBeneficiados: model.StringArray(grupo)})
	}
	return caravanas, nil
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
