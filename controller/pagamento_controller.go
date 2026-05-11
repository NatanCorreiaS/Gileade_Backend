package controller

import (
	"errors"
	"fmt"
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

// NewPagamentoController monta o controller de pagamentos.
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
	Type string `json:"type"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

// RegisterRoutes registra os endpoints de pagamentos.
func (c *PagamentoController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/pagamentos/checkout", c.CreateCheckout)
	rg.POST("/pagamentos/webhook", c.HandleWebhook)
}

// CreateCheckout cria checkout e ticket pendente para o usuario.
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

// HandleWebhook processa notificacoes de pagamento aprovadas.
func (c *PagamentoController) HandleWebhook(ctx *gin.Context) {
	payload := WebhookPayload{}
	_ = ctx.ShouldBindJSON(&payload)

	topic := strings.TrimSpace(ctx.Query("topic"))
	if topic == "merchant_order" {
		ctx.JSON(http.StatusOK, gin.H{"status": "ignorado"})
		return
	}

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
		if topic != "" {
			ctx.JSON(http.StatusOK, gin.H{"status": "ignorado"})
			return
		}
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{
			"payment_id": paymentIDStr,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payment id inválido"})
		return
	}

	result, err := c.payService.ProcessarPagamentoWebhook(ctx, paymentID)
	if err != nil {
		audit.GetLogger().LogEvent("pagamento_webhook", false, map[string]any{
			"payment_id":        paymentID,
			"ticket_usuario_id": result.TicketUsuarioID,
			"usuario_id":        result.UsuarioID,
			"ticket_id":         result.TicketID,
			"cpf":               result.CPF,
			"status":            result.Status,
		}, err)
		if errors.Is(err, service.ErrExternalReferenceInvalida) {
			ctx.JSON(http.StatusBadRequest, gin.H{"erro": "external_reference inválida"})
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"erro": "ticket do usuario não encontrado"})
			return
		}
		if errors.Is(err, repository.ErrTicketIndisponivel) {
			ctx.JSON(http.StatusConflict, gin.H{"erro": "ticket indisponivel"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao registrar pagamento"})
		return
	}

	sucesso := result.Status == "ok" || result.Status == "duplicado"
	var logErr error
	if !sucesso {
		logErr = fmt.Errorf("status=%s", result.Status)
	}
	audit.GetLogger().LogEvent("pagamento_webhook", sucesso, map[string]any{
		"payment_id":        paymentID,
		"ticket_usuario_id": result.TicketUsuarioID,
		"usuario_id":        result.UsuarioID,
		"ticket_id":         result.TicketID,
		"cpf":               result.CPF,
		"status":            result.Status,
	}, logErr)
	ctx.JSON(http.StatusOK, gin.H{"status": result.Status})
}
