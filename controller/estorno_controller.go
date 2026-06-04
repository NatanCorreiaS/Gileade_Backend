package controller

import (
	"errors"
	"net/http"
	"time"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/audit"
	"gileade/gileade_backend/gateway"
	"gileade/gileade_backend/repository"
	"gileade/gileade_backend/service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type EstornoController struct {
	payService *service.PagamentoService
	payRepo    *repository.PagamentoRepository
	estRepo    *repository.EstornoRepository
}

// NewEstornoController monta o controller de estornos.
func NewEstornoController(db *gorm.DB, gw *gateway.MercadoPagoGateway) *EstornoController {
	return &EstornoController{
		payService: service.NewPagamentoService(db, gw),
		payRepo:    repository.NewPagamentoRepository(db),
		estRepo:    repository.NewEstornoRepository(db),
	}
}

type EstornoRequest struct {
	Valor  *string `json:"valor"`
	Motivo string  `json:"motivo"`
}

type EstornoResponse struct {
	ID                 uint64 `json:"id"`
	PagamentoID        uint64 `json:"pagamento_id"`
	IDTransacaoEstorno string `json:"id_transacao_estorno"`
	Valor              string `json:"valor"`
	Motivo             string `json:"motivo"`
}

// RegisterRoutes registra os endpoints de estorno.
func (c *EstornoController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/estornos", c.List)
	rg.GET("/estornos/:id", c.GetByID)
	rg.GET("/pagamentos/:id/estornos", c.ListByPagamentoID)
	rg.POST("/pagamentos/:id/estornos", c.Create)
}

// prazoEstornoDiasCompra e o prazo maximo em dias apos a compra para solicitar estorno.
const prazoEstornoDiasCompra = 7

// prazoEstornoDiasAntesEvento e o prazo minimo em dias antes do evento para solicitar estorno.
const prazoEstornoDiasAntesEvento = 5

// Create registra um estorno ligado a um pagamento.
// Admins podem estornar a qualquer momento.
// Usuarios comuns podem estornar compras proprias ate 7 dias apos a compra
// e ate 5 dias antes do evento.
func (c *EstornoController) Create(ctx *gin.Context) {
	pagamentoID, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	authID, isAuth := GetAuthUsuarioID(ctx)
	if !isAuth {
		ctx.JSON(http.StatusUnauthorized, gin.H{"erro": "autenticacao necessaria"})
		return
	}

	pagamento, err := c.payRepo.GetByID(ctx, pagamentoID)
	if err != nil {
		audit.GetLogger().LogEvent("estorno_criar", false, map[string]any{
			"pagamento_id": pagamentoID,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "pagamento nao encontrado"})
		return
	}

	eAdmin := IsAdmin(ctx)

	if !eAdmin {
		if pagamento.TicketCompra.UsuarioID != authID {
			audit.GetLogger().LogEvent("estorno_criar", false, map[string]any{
				"pagamento_id":  pagamentoID,
				"solicitante_id": authID,
			}, nil)
			ctx.JSON(http.StatusForbidden, gin.H{"erro": "voce so pode estornar suas proprias compras"})
			return
		}

		agora := time.Now().UTC()

		if agora.After(pagamento.DataPagamento.Add(prazoEstornoDiasCompra * 24 * time.Hour)) {
			audit.GetLogger().LogEvent("estorno_criar", false, map[string]any{
				"pagamento_id":    pagamentoID,
				"solicitante_id":   authID,
				"data_pagamento":  pagamento.DataPagamento,
			}, nil)
			ctx.JSON(http.StatusForbidden, gin.H{"erro": "prazo de 7 dias apos a compra excedido"})
			return
		}

		limiteEvento := pagamento.TicketCompra.Ticket.DataEvento.Add(-prazoEstornoDiasAntesEvento * 24 * time.Hour)
		if agora.After(limiteEvento) {
			audit.GetLogger().LogEvent("estorno_criar", false, map[string]any{
				"pagamento_id":    pagamentoID,
				"solicitante_id":   authID,
				"data_evento":     pagamento.TicketCompra.Ticket.DataEvento,
			}, nil)
			ctx.JSON(http.StatusForbidden, gin.H{"erro": "estorno permitido apenas ate 5 dias antes do evento"})
			return
		}
	}

	var req EstornoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("estorno_criar", false, map[string]any{
			"pagamento_id": pagamentoID,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	var valorDecimal *decimal.Decimal
	if req.Valor != nil {
		parsed, err := decimal.NewFromString(*req.Valor)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"erro": "valor invalido"})
			return
		}
		valorDecimal = &parsed
	}

	estorno, err := c.payService.CriarEstornoPorPagamentoID(ctx, pagamentoID, req.Motivo, valorDecimal)
	if err != nil {
		audit.GetLogger().LogEvent("estorno_criar", false, map[string]any{
			"pagamento_id": pagamentoID,
		}, err)
		if errors.Is(err, repository.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"erro": "pagamento nao encontrado"})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"erro": "falha ao estornar"})
		return
	}

	audit.GetLogger().LogEvent("estorno_criar", true, map[string]any{
		"pagamento_id":   pagamentoID,
		"estorno_id":     estorno.ID,
		"solicitante_id": authID,
	}, nil)

	ctx.JSON(http.StatusOK, toEstornoResponse(estorno))
}

// List lista estornos com paginacao.
func (c *EstornoController) List(ctx *gin.Context) {
	limit, _ := strconvAtoiDefault(ctx.Query("limit"), 50)
	offset, _ := strconvAtoiDefault(ctx.Query("offset"), 0)

	estornos, err := c.estRepo.List(ctx, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao listar estornos"})
		return
	}

	resp := make([]EstornoResponse, 0, len(estornos))
	for _, e := range estornos {
		resp = append(resp, toEstornoResponse(e))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetByID busca um estorno pelo ID.
func (c *EstornoController) GetByID(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	estorno, err := c.estRepo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("estorno_buscar", false, map[string]any{
			"estorno_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "estorno nao encontrado"})
		return
	}

	audit.GetLogger().LogEvent("estorno_buscar", true, map[string]any{
		"estorno_id":   estorno.ID,
		"pagamento_id": estorno.PagamentoID,
	}, nil)

	ctx.JSON(http.StatusOK, toEstornoResponse(estorno))
}

// ListByPagamentoID lista estornos de um pagamento.
func (c *EstornoController) ListByPagamentoID(ctx *gin.Context) {
	pagamentoID, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	limit, _ := strconvAtoiDefault(ctx.Query("limit"), 50)
	offset, _ := strconvAtoiDefault(ctx.Query("offset"), 0)

	estornos, err := c.estRepo.ListByPagamentoID(ctx, pagamentoID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao listar estornos"})
		return
	}

	resp := make([]EstornoResponse, 0, len(estornos))
	for _, e := range estornos {
		resp = append(resp, toEstornoResponse(e))
	}
	ctx.JSON(http.StatusOK, resp)
}

func toEstornoResponse(e model.Estorno) EstornoResponse {
	return EstornoResponse{
		ID:                 e.ID,
		PagamentoID:        e.PagamentoID,
		IDTransacaoEstorno: e.IDTransacaoEstorno,
		Valor:              e.Valor.StringFixed(2),
		Motivo:             e.Motivo,
	}
}
