package controller

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"gileade/gileade_backend/audit"
	"gileade/gileade_backend/gateway"
	"gileade/gileade_backend/repository"
	"gileade/gileade_backend/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PagamentoController struct {
	payService *service.PagamentoService
}

func NewPagamentoController(db *gorm.DB, gw *gateway.MercadoPagoGateway) *PagamentoController {
	return &PagamentoController{payService: service.NewPagamentoService(db, gw)}
}

type CheckoutRequest struct {
	UsuarioID       uint64 `json:"usuario_id" binding:"required"`
	TicketID        uint64 `json:"ticket_id" binding:"required"`
	SuccessURL      string `json:"success_url"`
	FailureURL      string `json:"failure_url"`
	PendingURL      string `json:"pending_url"`
	NotificationURL string `json:"notification_url"`
}

type CheckoutResponse struct {
	PreferenceID    string `json:"preference_id"`
	InitPoint       string `json:"init_point"`
	SandboxInit     string `json:"sandbox_init_point"`
	TicketUsuarioID uint64 `json:"ticket_usuario_id"`
}

type WebhookPayload struct {
	Action string `json:"action"`
	Type   string `json:"type"`
	Data   struct {
		ID any `json:"id"`
	} `json:"data"`
}

func (c *PagamentoController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/pagamentos/checkout", c.CreateCheckout)
	rg.POST("/pagamentos/webhook", c.HandleWebhook)
}

func (c *PagamentoController) CreateCheckout(ctx *gin.Context) {
	var req CheckoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("pagamento_checkout_criar", false, map[string]any{
			"usuario_id": req.UsuarioID,
			"ticket_id":  req.TicketID,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload inválido"})
		return
	}

	resp, err := c.payService.CriarCheckout(ctx, service.CheckoutRequest{
		UsuarioID:       req.UsuarioID,
		TicketID:        req.TicketID,
		SuccessURL:      req.SuccessURL,
		FailureURL:      req.FailureURL,
		PendingURL:      req.PendingURL,
		NotificationURL: req.NotificationURL,
	})
	if err != nil {
		audit.GetLogger().LogEvent("pagamento_checkout_criar", false, map[string]any{
			"usuario_id": req.UsuarioID,
			"ticket_id":  req.TicketID,
		}, err)
		if errors.Is(err, service.ErrNotificationURLObrigatoria) {
			ctx.JSON(http.StatusBadRequest, gin.H{"erro": "notification_url obrigatório"})
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"erro": "usuario ou ticket nao encontrado"})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"erro": "falha ao criar preferência no Mercado Pago"})
		return
	}

	audit.GetLogger().LogEvent("pagamento_checkout_criar", true, map[string]any{
		"ticket_usuario_id": resp.TicketUsuarioID,
		"usuario_id":        req.UsuarioID,
		"ticket_id":         req.TicketID,
		"preference_id":     resp.PreferenceID,
	}, nil)

	ctx.JSON(http.StatusOK, CheckoutResponse{
		PreferenceID:    resp.PreferenceID,
		InitPoint:       resp.InitPoint,
		SandboxInit:     resp.SandboxInit,
		TicketUsuarioID: resp.TicketUsuarioID,
	})
}

func (c *PagamentoController) HandleWebhook(ctx *gin.Context) {
	// 1. Ler o body bruto para registrar exatamente o que o Mercado Pago mandou
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		audit.GetLogger().LogEvent("pagamento_webhook_erro_leitura", false, nil, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "falha ao ler body"})
		return
	}

	// 2. Restaurar o body para o Gin conseguir fazer o ShouldBindJSON
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 3. Capturar headers para analisar as assinaturas x-signature
	headersObj := make(map[string]any)
	for k, v := range ctx.Request.Header {
		if len(v) == 1 {
			headersObj[k] = v[0]
		} else {
			headersObj[k] = v
		}
	}

	// Registra log detalhado
	audit.GetLogger().LogEvent("pagamento_webhook_recebido", true, map[string]any{
		"body":    string(bodyBytes),
		"headers": headersObj,
		"query":   ctx.Request.URL.RawQuery,
	}, nil)

	var payload WebhookPayload
	_ = ctx.ShouldBindJSON(&payload)

	topic := strings.TrimSpace(payload.Type)
	if topic == "" {
		topic = strings.TrimSpace(ctx.Query("topic"))
	}
	if topic == "" {
		topic = strings.TrimSpace(ctx.Query("type"))
	}
	action := strings.TrimSpace(payload.Action)

	if topic == "merchant_order" || strings.HasPrefix(action, "merchant_order") {
		ctx.JSON(http.StatusOK, gin.H{"status": "ignorado"})
		return
	}

	isPayment := topic == "payment" || strings.HasPrefix(action, "payment.")
	if !isPayment {
		ctx.JSON(http.StatusOK, gin.H{"status": "ignorado", "motivo": "não é evento de pagamento"})
		return
	}

	var paymentIDStr string
	if payload.Data.ID != nil {
		paymentIDStr = strings.TrimSpace(fmt.Sprintf("%v", payload.Data.ID))
	}
	if paymentIDStr == "" || paymentIDStr == "<nil>" {
		paymentIDStr = strings.TrimSpace(ctx.Query("data.id"))
	}
	if paymentIDStr == "" {
		paymentIDStr = strings.TrimSpace(ctx.Query("id"))
	}

	if paymentIDStr == "" {
		audit.GetLogger().LogEvent("pagamento_webhook", false, nil, fmt.Errorf("payment id ausente"))
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payment id ausente"})
		return
	}

	paymentID, err := strconv.Atoi(paymentIDStr)
	if err != nil {
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{"payment_id": paymentIDStr}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payment id inválido"})
		return
	}

	result, err := c.payService.ProcessarPagamentoWebhook(ctx, paymentID)
	if err != nil {
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{"payment_id": paymentID}, err)
		ctx.JSON(http.StatusOK, gin.H{"status": "erro_ignorado", "erro": err.Error()})
		return
	}

	sucesso := result.Status == "ok" || result.Status == "duplicado"
	audit.GetLogger().LogEvent("pagamento_webhook", sucesso, map[string]any{
		"payment_id": paymentID,
		"status":     result.Status,
	}, nil)

	ctx.JSON(http.StatusOK, gin.H{"status": result.Status})
}
