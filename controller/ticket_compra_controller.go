package controller

import (
	"net/http"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/audit"
	"gileade/gileade_backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketCompraController struct {
	repo  *repository.TicketCompraRepository
	pRepo *repository.PessoaRepository
}

// NewTicketCompraController monta o controller de tickets por compra.
func NewTicketCompraController(db *gorm.DB) *TicketCompraController {
	return &TicketCompraController{
		repo:  repository.NewTicketCompraRepository(db),
		pRepo: repository.NewPessoaRepository(db),
	}
}

type TicketCompraCreateRequest struct {
	UsuarioID  uint64              `json:"usuario_id" binding:"required"`
	TicketID   uint64              `json:"ticket_id" binding:"required"`
	Quantidade uint64              `json:"quantidade"`
	Status     model.TicketsStatus `json:"status"`
}

type TicketCompraStatusRequest struct {
	Status model.TicketsStatus `json:"status" binding:"required"`
}

type TicketCompraResponse struct {
	ID           uint64              `json:"id"`
	UsuarioID    uint64              `json:"usuario_id"`
	TicketID     uint64              `json:"ticket_id"`
	Quantidade   uint64              `json:"quantidade"`
	Status       model.TicketsStatus `json:"status"`
	PreferenceID string              `json:"preference_id"`
}

// RegisterRoutes registra os endpoints de tickets por compra.
func (c *TicketCompraController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/tickets-compra", c.Create)
	rg.GET("/tickets-compra/:id", c.GetByID)
	rg.GET("/usuarios/:id/tickets-compra", c.ListByUsuarioID)
	rg.PATCH("/tickets-compra/:id/status", c.UpdateStatus)
	rg.DELETE("/tickets-compra/:id", c.Delete)
}

// Create cria o vinculo entre usuario e ticket.
func (c *TicketCompraController) Create(ctx *gin.Context) {
	var req TicketCompraCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("ticket_compra_criar", false, map[string]any{
			"usuario_id": req.UsuarioID,
			"ticket_id":  req.TicketID,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	pessoa, err := c.pRepo.GetByID(ctx, req.UsuarioID)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_compra_criar", false, map[string]any{
			"usuario_id": req.UsuarioID,
			"ticket_id":  req.TicketID,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "usuario nao encontrado"})
		return
	}

	quantidade := req.Quantidade
	if quantidade == 0 {
		quantidade = 1
	}

	status := req.Status
	if status == "" {
		status = model.TicketsStatusPendente
	}

	tc := model.TicketCompra{
		UsuarioID:  req.UsuarioID,
		TicketID:   req.TicketID,
		Quantidade: quantidade,
		Status:     status,
	}
	if err := c.repo.Create(ctx, &tc); err != nil {
		audit.GetLogger().LogEvent("ticket_compra_criar", false, map[string]any{
			"usuario_id": pessoa.ID,
			"ticket_id":  req.TicketID,
			"cpf":        pessoa.CPF,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao criar"})
		return
	}

	audit.GetLogger().LogEvent("ticket_compra_criar", true, map[string]any{
		"ticket_compra_id": tc.ID,
		"usuario_id":       pessoa.ID,
		"ticket_id":        tc.TicketID,
		"cpf":              pessoa.CPF,
	}, nil)

	ctx.JSON(http.StatusCreated, toTicketCompraResponse(tc))
}

// GetByID busca o vinculo pelo ID.
func (c *TicketCompraController) GetByID(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	tc, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_compra_buscar", false, map[string]any{
			"ticket_compra_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "registro nao encontrado"})
		return
	}

	audit.GetLogger().LogEvent("ticket_compra_buscar", true, map[string]any{
		"ticket_compra_id": tc.ID,
		"usuario_id":       tc.UsuarioID,
		"ticket_id":        tc.TicketID,
		"cpf":              tc.Usuario.CPF,
	}, nil)

	ctx.JSON(http.StatusOK, toTicketCompraResponse(tc))
}

// ListByUsuarioID lista tickets associados ao usuario.
func (c *TicketCompraController) ListByUsuarioID(ctx *gin.Context) {
	usuarioID, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	limit, _ := strconvAtoiDefault(ctx.Query("limit"), 50)
	offset, _ := strconvAtoiDefault(ctx.Query("offset"), 0)

	tcs, err := c.repo.ListByUsuarioID(ctx, usuarioID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao listar"})
		return
	}

	resp := make([]TicketCompraResponse, 0, len(tcs))
	for _, tc := range tcs {
		resp = append(resp, toTicketCompraResponse(tc))
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateStatus atualiza o status do vinculo.
func (c *TicketCompraController) UpdateStatus(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	var req TicketCompraStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("ticket_compra_status", false, map[string]any{
			"ticket_compra_id": id,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	tc, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_compra_status", false, map[string]any{
			"ticket_compra_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "registro nao encontrado"})
		return
	}

	if err := c.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		audit.GetLogger().LogEvent("ticket_compra_status", false, map[string]any{
			"ticket_compra_id": tc.ID,
			"usuario_id":       tc.UsuarioID,
			"ticket_id":        tc.TicketID,
			"status":           req.Status,
			"cpf":              tc.Usuario.CPF,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao atualizar"})
		return
	}

	audit.GetLogger().LogEvent("ticket_compra_status", true, map[string]any{
		"ticket_compra_id": tc.ID,
		"usuario_id":       tc.UsuarioID,
		"ticket_id":        tc.TicketID,
		"status":           req.Status,
		"cpf":              tc.Usuario.CPF,
	}, nil)

	ctx.Status(http.StatusNoContent)
}

// Delete remove o vinculo entre usuario e ticket.
func (c *TicketCompraController) Delete(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	tc, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_compra_remover", false, map[string]any{
			"ticket_compra_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "registro nao encontrado"})
		return
	}

	if err := c.repo.Delete(ctx, id); err != nil {
		audit.GetLogger().LogEvent("ticket_compra_remover", false, map[string]any{
			"ticket_compra_id": tc.ID,
			"usuario_id":       tc.UsuarioID,
			"ticket_id":        tc.TicketID,
			"cpf":              tc.Usuario.CPF,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao remover"})
		return
	}

	audit.GetLogger().LogEvent("ticket_compra_remover", true, map[string]any{
		"ticket_compra_id": tc.ID,
		"usuario_id":       tc.UsuarioID,
		"ticket_id":        tc.TicketID,
		"cpf":              tc.Usuario.CPF,
	}, nil)

	ctx.Status(http.StatusNoContent)
}

// toTicketCompraResponse converte o modelo para resposta JSON.
func toTicketCompraResponse(tc model.TicketCompra) TicketCompraResponse {
	return TicketCompraResponse{
		ID:           tc.ID,
		UsuarioID:    tc.UsuarioID,
		TicketID:     tc.TicketID,
		Quantidade:   tc.Quantidade,
		Status:       tc.Status,
		PreferenceID: tc.PreferenceID,
	}
}
