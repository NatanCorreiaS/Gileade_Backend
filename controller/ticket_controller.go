package controller

import (
	"net/http"
	"time"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/audit"
	"gileade/gileade_backend/repository"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TicketController struct {
	repo *repository.TicketRepository
}

// NewTicketController monta o controller de tickets.
func NewTicketController(db *gorm.DB) *TicketController {
	return &TicketController{repo: repository.NewTicketRepository(db)}
}

type TicketCreateRequest struct {
	Tipo                 model.TipoTicket `json:"tipo" binding:"required"`
	Nome                 string           `json:"nome" binding:"required"`
	Descricao            string           `json:"descricao"`
	Preco                string           `json:"preco" binding:"required"`
	QuantidadeDisponivel uint64           `json:"quantidade_disponivel" binding:"required"`
	DataEvento           string           `json:"data_evento" binding:"required"`
}

type TicketUpdateRequest struct {
	Tipo                 *model.TipoTicket `json:"tipo"`
	Nome                 *string           `json:"nome"`
	Descricao            *string           `json:"descricao"`
	Preco                *string           `json:"preco"`
	QuantidadeDisponivel *uint64           `json:"quantidade_disponivel"`
	DataEvento           *string           `json:"data_evento"`
}

type TicketResponse struct {
	ID                   uint64           `json:"id"`
	Tipo                 model.TipoTicket `json:"tipo"`
	Nome                 string           `json:"nome"`
	Descricao            string           `json:"descricao"`
	Preco                decimal.Decimal  `json:"preco"`
	QuantidadeDisponivel uint64           `json:"quantidade_disponivel"`
	DataEvento           string           `json:"data_evento"`
}

// RegisterRoutes registra os endpoints de tickets.
func (c *TicketController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/tickets", c.Create)
	rg.GET("/tickets", c.List)
	rg.GET("/tickets/:id", c.GetByID)
	rg.PUT("/tickets/:id", c.Update)
	rg.DELETE("/tickets/:id", c.Delete)
}

// Create cadastra um ticket de evento.
func (c *TicketController) Create(ctx *gin.Context) {
	var req TicketCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("ticket_criar", false, map[string]any{
			"nome": req.Nome,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	preco, err := decimal.NewFromString(req.Preco)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_criar", false, map[string]any{
			"nome": req.Nome,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "preco invalido"})
		return
	}

	dataEvento, err := parseDate(req.DataEvento)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_criar", false, map[string]any{
			"nome": req.Nome,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "data_evento invalida (use YYYY-MM-DD)"})
		return
	}

	ticket := model.Ticket{
		Tipo:                 req.Tipo,
		Nome:                 req.Nome,
		Descricao:            req.Descricao,
		Preco:                preco,
		QuantidadeDisponivel: req.QuantidadeDisponivel,
		DataEvento:           dataEvento,
	}

	if err := c.repo.Create(ctx, &ticket); err != nil {
		audit.GetLogger().LogEvent("ticket_criar", false, map[string]any{
			"nome": req.Nome,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao criar ticket"})
		return
	}

	audit.GetLogger().LogEvent("ticket_criar", true, map[string]any{
		"ticket_id": ticket.ID,
		"nome":      ticket.Nome,
	}, nil)

	ctx.JSON(http.StatusCreated, toTicketResponse(ticket))
}

// List lista tickets com paginacao simples.
func (c *TicketController) List(ctx *gin.Context) {
	limit, _ := strconvAtoiDefault(ctx.Query("limit"), 50)
	offset, _ := strconvAtoiDefault(ctx.Query("offset"), 0)

	tickets, err := c.repo.List(ctx, limit, offset)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_listar", false, map[string]any{
			"limit":  limit,
			"offset": offset,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao listar"})
		return
	}

	resp := make([]TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		resp = append(resp, toTicketResponse(t))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetByID busca um ticket pelo ID.
func (c *TicketController) GetByID(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	ticket, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_buscar", false, map[string]any{
			"ticket_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "ticket nao encontrado"})
		return
	}

	audit.GetLogger().LogEvent("ticket_buscar", true, map[string]any{
		"ticket_id": ticket.ID,
	}, nil)

	ctx.JSON(http.StatusOK, toTicketResponse(ticket))
}

// Update atualiza dados de um ticket.
func (c *TicketController) Update(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	var req TicketUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("ticket_atualizar", false, map[string]any{
			"ticket_id": id,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	ticket, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_atualizar", false, map[string]any{
			"ticket_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "ticket nao encontrado"})
		return
	}

	if req.Nome != nil {
		ticket.Nome = *req.Nome
	}
	if req.Tipo != nil {
		ticket.Tipo = *req.Tipo
	}
	if req.Descricao != nil {
		ticket.Descricao = *req.Descricao
	}
	if req.Preco != nil {
		preco, err := decimal.NewFromString(*req.Preco)
		if err != nil {
			audit.GetLogger().LogEvent("ticket_atualizar", false, map[string]any{
				"ticket_id": id,
			}, err)
			ctx.JSON(http.StatusBadRequest, gin.H{"erro": "preco invalido"})
			return
		}
		ticket.Preco = preco
	}
	if req.QuantidadeDisponivel != nil {
		ticket.QuantidadeDisponivel = *req.QuantidadeDisponivel
	}
	if req.DataEvento != nil {
		dataEvento, err := parseDate(*req.DataEvento)
		if err != nil {
			audit.GetLogger().LogEvent("ticket_atualizar", false, map[string]any{
				"ticket_id": id,
			}, err)
			ctx.JSON(http.StatusBadRequest, gin.H{"erro": "data_evento invalida (use YYYY-MM-DD)"})
			return
		}
		ticket.DataEvento = dataEvento
	}

	if err := c.repo.Update(ctx, &ticket); err != nil {
		audit.GetLogger().LogEvent("ticket_atualizar", false, map[string]any{
			"ticket_id": id,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao atualizar"})
		return
	}

	audit.GetLogger().LogEvent("ticket_atualizar", true, map[string]any{
		"ticket_id": ticket.ID,
	}, nil)

	ctx.JSON(http.StatusOK, toTicketResponse(ticket))
}

// Delete remove um ticket pelo ID.
func (c *TicketController) Delete(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	if err := c.repo.Delete(ctx, id); err != nil {
		audit.GetLogger().LogEvent("ticket_remover", false, map[string]any{
			"ticket_id": id,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao remover"})
		return
	}

	audit.GetLogger().LogEvent("ticket_remover", true, map[string]any{
		"ticket_id": id,
	}, nil)
	ctx.Status(http.StatusNoContent)
}

// toTicketResponse converte o modelo para resposta JSON.
func toTicketResponse(t model.Ticket) TicketResponse {
	return TicketResponse{
		ID:                   t.ID,
		Tipo:                 t.Tipo,
		Nome:                 t.Nome,
		Descricao:            t.Descricao,
		Preco:                t.Preco,
		QuantidadeDisponivel: t.QuantidadeDisponivel,
		DataEvento:           t.DataEvento.Format("2006-01-02"),
	}
}

// parseDate converte data no formato YYYY-MM-DD.
func parseDate(val string) (time.Time, error) {
	return time.Parse("2006-01-02", val)
}
