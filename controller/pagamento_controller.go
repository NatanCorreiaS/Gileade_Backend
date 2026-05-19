package controller

import (
	"bytes"
	"errors"
	"fmt"
	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/audit"
	"gileade/gileade_backend/gateway"
	"gileade/gileade_backend/repository"
	"gileade/gileade_backend/service"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PagamentoController struct {
	payService *service.PagamentoService
	payRepo    *repository.PagamentoRepository
}

func NewPagamentoController(db *gorm.DB, gw *gateway.MercadoPagoGateway) *PagamentoController {
	return &PagamentoController{
		payService: service.NewPagamentoService(db, gw),
		payRepo:    repository.NewPagamentoRepository(db),
	}
}

type CheckoutRequest struct {
	UsuarioID    uint64               `json:"usuario_id" binding:"required"`
	TicketID     uint64               `json:"ticket_id" binding:"required"`
	Quantidade   uint64               `json:"quantidade"`
	SuccessURL   string               `json:"success_url"`
	FailureURL   string               `json:"failure_url"`
	PendingURL   string               `json:"pending_url"`
	Beneficiados []BeneficiadoRequest `json:"beneficiados"`
}

type BeneficiadoRequest struct {
	Nome         string             `json:"nome" binding:"required"`
	CPF          string             `json:"cpf" binding:"required"`
	Idade        int16              `json:"idade"`
	Celular      string             `json:"celular"`
	Igreja       string             `json:"igreja"`
	PapelIgreja  model.PapelIgreja  `json:"papel_igreja"`
	EstadoCivil  model.EstadoCivil  `json:"estado_civil"`
	Email        string             `json:"email" binding:"required"`
	Sexo         model.Sexo         `json:"sexo" binding:"required"`
	Cidade       string             `json:"cidade"`
	EstadoUF     model.EstadoUF     `json:"estado_uf"`
	Escolaridade model.Escolaridade `json:"escolaridade"`
}

type CheckoutResponse struct {
	PreferenceID   string `json:"preference_id"`
	InitPoint      string `json:"init_point"`
	SandboxInit    string `json:"sandbox_init_point"`
	TicketCompraID uint64 `json:"ticket_compra_id"`
}

type PagamentoConsultaResponse struct {
	ID             uint64                `json:"id"`
	IDTransacao    string                `json:"id_transacao"`
	Valor          string                `json:"valor"`
	Metodo         model.MetodoPagamento `json:"metodo"`
	DataPagamento  string                `json:"data_pagamento"`
	TicketCompraID uint64                `json:"ticket_compra_id"`
	UsuarioID      uint64                `json:"usuario_id"`
	TicketID       uint64                `json:"ticket_id"`
	Status         model.TicketsStatus   `json:"status"`
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
	rg.GET("/pagamentos", c.ListPayments)
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
		UsuarioID:    req.UsuarioID,
		TicketID:     req.TicketID,
		Quantidade:   req.Quantidade,
		SuccessURL:   req.SuccessURL,
		FailureURL:   req.FailureURL,
		PendingURL:   req.PendingURL,
		Beneficiados: toBeneficiadosInput(req.Beneficiados),
	})
	if err != nil {
		audit.GetLogger().LogEvent("pagamento_checkout_criar", false, map[string]any{
			"usuario_id": req.UsuarioID,
			"ticket_id":  req.TicketID,
		}, err)
		if errors.Is(err, repository.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"erro": "usuario ou ticket nao encontrado"})
			return
		}
		if errors.Is(err, repository.ErrTipoTicketInvalido) {
			ctx.JSON(http.StatusBadRequest, gin.H{"erro": "tipo de ticket invalido"})
			return
		}
		if errors.Is(err, service.ErrQuantidadeInvalida) || errors.Is(err, service.ErrBeneficiadosInvalidos) {
			ctx.JSON(http.StatusBadRequest, gin.H{"erro": "beneficiados invalidos"})
			return
		}
		if errors.Is(err, service.ErrNotificationURLObrigatoria) {
			ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "notification_url nao configurado"})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"erro": "falha ao criar preferência no Mercado Pago"})
		return
	}

	audit.GetLogger().LogEvent("pagamento_checkout_criar", true, map[string]any{
		"ticket_compra_id": resp.TicketCompraID,
		"usuario_id":       req.UsuarioID,
		"ticket_id":        req.TicketID,
		"preference_id":    resp.PreferenceID,
	}, nil)

	ctx.JSON(http.StatusOK, CheckoutResponse{
		PreferenceID:   resp.PreferenceID,
		InitPoint:      resp.InitPoint,
		SandboxInit:    resp.SandboxInit,
		TicketCompraID: resp.TicketCompraID,
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

func (c *PagamentoController) ListPayments(ctx *gin.Context) {
	usuarioIDs, err := parseUsuarioIDs(ctx.Query("usuario_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "usuario_id invalido"})
		return
	}
	if len(usuarioIDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "usuario_id obrigatorio"})
		return
	}

	var status *model.TicketsStatus
	if statusStr := strings.TrimSpace(ctx.Query("status")); statusStr != "" {
		st := model.TicketsStatus(statusStr)
		status = &st
	}

	var dataInicio *time.Time
	if val := strings.TrimSpace(ctx.Query("data_inicio")); val != "" {
		parsed, err := parseDate(val)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"erro": "data_inicio invalida"})
			return
		}
		dataInicio = &parsed
	}

	var dataFim *time.Time
	if val := strings.TrimSpace(ctx.Query("data_fim")); val != "" {
		parsed, err := parseDate(val)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"erro": "data_fim invalida"})
			return
		}
		dataFim = &parsed
	}

	limit, _ := strconvAtoiDefault(ctx.Query("limit"), 50)
	offset, _ := strconvAtoiDefault(ctx.Query("offset"), 0)

	pagamentos, err := c.payRepo.ListByUsuarios(ctx, usuarioIDs, status, dataInicio, dataFim, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao listar pagamentos"})
		return
	}

	resp := make([]PagamentoConsultaResponse, 0, len(pagamentos))
	for _, pagamento := range pagamentos {
		resp = append(resp, PagamentoConsultaResponse{
			ID:             pagamento.ID,
			IDTransacao:    pagamento.IDTransacao,
			Valor:          pagamento.Valor.StringFixed(2),
			Metodo:         pagamento.Metodo,
			DataPagamento:  pagamento.DataPagamento.Format(time.RFC3339),
			TicketCompraID: pagamento.TicketCompraID,
			UsuarioID:      pagamento.TicketCompra.UsuarioID,
			TicketID:       pagamento.TicketCompra.TicketID,
			Status:         pagamento.TicketCompra.Status,
		})
	}

	ctx.JSON(http.StatusOK, resp)
}

func parseUsuarioIDs(val string) ([]uint64, error) {
	if strings.TrimSpace(val) == "" {
		return nil, nil
	}
	parts := strings.Split(val, ",")
	ids := make([]uint64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func toBeneficiadosInput(reqs []BeneficiadoRequest) []service.BeneficiadoInput {
	if len(reqs) == 0 {
		return nil
	}
	inputs := make([]service.BeneficiadoInput, 0, len(reqs))
	for _, req := range reqs {
		inputs = append(inputs, service.BeneficiadoInput{
			Nome:         req.Nome,
			CPF:          req.CPF,
			Idade:        req.Idade,
			Celular:      req.Celular,
			Igreja:       req.Igreja,
			PapelIgreja:  req.PapelIgreja,
			EstadoCivil:  req.EstadoCivil,
			Email:        req.Email,
			Sexo:         req.Sexo,
			Cidade:       req.Cidade,
			EstadoUF:     req.EstadoUF,
			Escolaridade: req.Escolaridade,
		})
	}
	return inputs
}
