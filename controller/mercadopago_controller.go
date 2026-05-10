package controller

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/audit"
	"gileade/gileade_backend/gateway"
	"gileade/gileade_backend/repository"

	"github.com/gin-gonic/gin"
	"github.com/mercadopago/sdk-go/pkg/preference"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MercadoPagoController struct {
	pRepo   *repository.PessoaRepository
	tRepo   *repository.TicketRepository
	tuRepo  *repository.TicketUsuarioRepository
	payRepo *repository.PagamentoRepository
	gw      *gateway.MercadoPagoGateway
}

func NewMercadoPagoController(
	db *gorm.DB,
	gw *gateway.MercadoPagoGateway,
) *MercadoPagoController {
	return &MercadoPagoController{
		pRepo:   repository.NewPessoaRepository(db),
		tRepo:   repository.NewTicketRepository(db),
		tuRepo:  repository.NewTicketUsuarioRepository(db),
		payRepo: repository.NewPagamentoRepository(db),
		gw:      gw,
	}
}

type CheckoutProRequest struct {
	UsuarioID       uint64 `json:"usuario_id" binding:"required"`
	TicketID        uint64 `json:"ticket_id" binding:"required"`
	SuccessURL      string `json:"success_url"`
	FailureURL      string `json:"failure_url"`
	PendingURL      string `json:"pending_url"`
	NotificationURL string `json:"notification_url"`
}

type CheckoutProResponse struct {
	PreferenceID    string `json:"preference_id"`
	InitPoint       string `json:"init_point"`
	SandboxInit     string `json:"sandbox_init_point"`
	TicketUsuarioID uint64 `json:"ticket_usuario_id"`
}

type WebhookPayload struct {
	Type string `json:"type"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (c *MercadoPagoController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/checkout/pro", c.CreateCheckoutPro)
	rg.POST("/mercadopago/webhook", c.HandleWebhook)
}

func (c *MercadoPagoController) CreateCheckoutPro(ctx *gin.Context) {
	var req CheckoutProRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("checkout_pro_criar", false, map[string]any{
			"usuario_id": req.UsuarioID,
			"ticket_id":  req.TicketID,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload inválido"})
		return
	}

	// Garante que usuário e ticket existem.
	pessoa, err := c.pRepo.GetByID(ctx, req.UsuarioID)
	if err != nil {
		audit.GetLogger().LogEvent("checkout_pro_criar", false, map[string]any{
			"usuario_id": req.UsuarioID,
			"ticket_id":  req.TicketID,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "usuário não encontrado"})
		return
	}
	ticket, err := c.tRepo.GetByID(ctx, req.TicketID)
	if err != nil {
		audit.GetLogger().LogEvent("checkout_pro_criar", false, map[string]any{
			"usuario_id": pessoa.ID,
			"ticket_id":  req.TicketID,
			"cpf":        pessoa.CPF,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "ticket não encontrado"})
		return
	}

	// Cria TicketUsuario pendente.
	tu := model.TicketUsuario{
		UsuarioID: pessoa.ID,
		TicketID:  ticket.ID,
		Status:    model.TicketsStatusPendente,
	}
	if err := c.tuRepo.Create(ctx, &tu); err != nil {
		audit.GetLogger().LogEvent("checkout_pro_criar", false, map[string]any{
			"usuario_id": pessoa.ID,
			"ticket_id":  ticket.ID,
			"cpf":        pessoa.CPF,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao criar ticket do usuário"})
		return
	}

	notificationURL := req.NotificationURL
	if notificationURL == "" {
		notificationURL = os.Getenv("MERCADO_PAGO_NOTIFICATION_URL")
	}
	if notificationURL == "" {
		_ = c.tuRepo.Delete(ctx, tu.ID)
		audit.GetLogger().LogEvent("checkout_pro_criar", false, map[string]any{
			"ticket_usuario_id": tu.ID,
			"usuario_id":        pessoa.ID,
			"ticket_id":         ticket.ID,
			"cpf":               pessoa.CPF,
		}, fmt.Errorf("notification_url obrigatório"))
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "notification_url obrigatório"})
		return
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
		BackURLs: &preference.BackURLsRequest{
			Success: req.SuccessURL,
			Failure: req.FailureURL,
			Pending: req.PendingURL,
		},
		AutoReturn:        "approved",
		ExternalReference: fmt.Sprintf("%d", tu.ID),
		NotificationURL:   notificationURL,
		BinaryMode:        true,
		Metadata: map[string]any{
			"ticket_usuario_id": tu.ID,
			"usuario_id":        pessoa.ID,
			"ticket_id":         ticket.ID,
		},
	}

	prefResp, err := c.gw.CreateCheckoutPro(ctx, prefReq)
	if err != nil {
		_ = c.tuRepo.Delete(ctx, tu.ID)
		audit.GetLogger().LogEvent("checkout_pro_criar", false, map[string]any{
			"ticket_usuario_id": tu.ID,
			"usuario_id":        pessoa.ID,
			"ticket_id":         ticket.ID,
			"cpf":               pessoa.CPF,
		}, err)
		ctx.JSON(http.StatusBadGateway, gin.H{"erro": "falha ao criar preferência no Mercado Pago"})
		return
	}

	audit.GetLogger().LogEvent("checkout_pro_criar", true, map[string]any{
		"ticket_usuario_id": tu.ID,
		"usuario_id":        pessoa.ID,
		"ticket_id":         ticket.ID,
		"cpf":               pessoa.CPF,
		"preference_id":     prefResp.ID,
	}, nil)

	ctx.JSON(http.StatusOK, CheckoutProResponse{
		PreferenceID:    prefResp.ID,
		InitPoint:       prefResp.InitPoint,
		SandboxInit:     prefResp.SandboxInitPoint,
		TicketUsuarioID: tu.ID,
	})
}

func (c *MercadoPagoController) HandleWebhook(ctx *gin.Context) {
	payload := WebhookPayload{}
	_ = ctx.ShouldBindJSON(&payload)

	paymentIDStr := strings.TrimSpace(ctx.Query("data.id"))
	if paymentIDStr == "" {
		paymentIDStr = strings.TrimSpace(ctx.Query("id"))
	}
	if paymentIDStr == "" {
		paymentIDStr = strings.TrimSpace(payload.Data.ID)
	}

	if paymentIDStr == "" {
		audit.GetLogger().LogEvent("pagamento_webhook", false, nil, fmt.Errorf("payment id ausente"))
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payment id ausente"})
		return
	}

	paymentID, err := strconv.Atoi(paymentIDStr)
	if err != nil {
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{
			"payment_id": paymentIDStr,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payment id inválido"})
		return
	}

	// Busca detalhes do pagamento para validar status e referência.
	payResp, err := c.gw.GetPayment(ctx, paymentID)
	if err != nil {
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{
			"payment_id": paymentID,
		}, err)
		ctx.JSON(http.StatusBadGateway, gin.H{"erro": "falha ao consultar pagamento"})
		return
	}

	if payResp.Status != "approved" {
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{
			"payment_id": paymentID,
			"status":     payResp.Status,
		}, fmt.Errorf("status=%s", payResp.Status))
		ctx.JSON(http.StatusOK, gin.H{"status": payResp.Status})
		return
	}

	tuID, err := strconv.ParseUint(payResp.ExternalReference, 10, 64)
	if err != nil {
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{
			"payment_id": paymentID,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "external_reference inválida"})
		return
	}

	tu, tuErr := c.tuRepo.GetByID(ctx, tuID)
	if tuErr != nil {
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{
			"payment_id":        paymentID,
			"ticket_usuario_id": tuID,
		}, tuErr)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "ticket do usuario não encontrado"})
		return
	}

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

	if err := c.payRepo.CreateAndMarkTicketPago(ctx, &pagamento); err != nil {
		if isUniqueViolation(err) {
			audit.GetLogger().LogEvent("pagamento_webhook", true, map[string]any{
				"payment_id":        paymentID,
				"ticket_usuario_id": tu.ID,
				"usuario_id":        tu.UsuarioID,
				"ticket_id":         tu.TicketID,
				"cpf":               tu.Usuario.CPF,
				"status":            "duplicado",
			}, nil)
			ctx.JSON(http.StatusOK, gin.H{"status": "duplicado"})
			return
		}
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{
			"payment_id":        paymentID,
			"ticket_usuario_id": tu.ID,
			"usuario_id":        tu.UsuarioID,
			"ticket_id":         tu.TicketID,
			"cpf":               tu.Usuario.CPF,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao registrar pagamento"})
		return
	}

	audit.GetLogger().LogEvent("pagamento_webhook", true, map[string]any{
		"payment_id":        paymentID,
		"ticket_usuario_id": tu.ID,
		"usuario_id":        tu.UsuarioID,
		"ticket_id":         tu.TicketID,
		"cpf":               tu.Usuario.CPF,
		"status":            "ok",
	}, nil)
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
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
