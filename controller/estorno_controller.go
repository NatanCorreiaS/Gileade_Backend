package controller

import (
	"errors"
	"net/http"

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
}

// NewEstornoController monta o controller de estornos.
func NewEstornoController(db *gorm.DB, gw *gateway.MercadoPagoGateway) *EstornoController {
	return &EstornoController{payService: service.NewPagamentoService(db, gw)}
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
}

// RegisterRoutes registra o endpoint de estorno.
func (c *EstornoController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/pagamentos/:id/estornos", c.Create)
}

// Create registra um estorno ligado a um pagamento.
func (c *EstornoController) Create(ctx *gin.Context) {
	pagamentoID, ok := parseUintParam(ctx, "id")
	if !ok {
		return
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
		"pagamento_id": pagamentoID,
		"estorno_id":   estorno.ID,
	}, nil)

	ctx.JSON(http.StatusOK, EstornoResponse{
		ID:                 estorno.ID,
		PagamentoID:        estorno.PagamentoID,
		IDTransacaoEstorno: estorno.IDTransacaoEstorno,
		Valor:              estorno.Valor.StringFixed(2),
	})
}
