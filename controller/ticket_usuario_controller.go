package controller

import (
	"net/http"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/audit"
	"gileade/gileade_backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketUsuarioController struct {
	repo  *repository.TicketUsuarioRepository
	pRepo *repository.PessoaRepository
}

// NewTicketUsuarioController monta o controller de tickets por usuario.
func NewTicketUsuarioController(db *gorm.DB) *TicketUsuarioController {
	return &TicketUsuarioController{
		repo:  repository.NewTicketUsuarioRepository(db),
		pRepo: repository.NewPessoaRepository(db),
	}
}

type TicketUsuarioCreateRequest struct {
	UsuarioID uint64              `json:"usuario_id" binding:"required"`
	TicketID  uint64              `json:"ticket_id" binding:"required"`
	Status    model.TicketsStatus `json:"status"`
}

type TicketUsuarioStatusRequest struct {
	Status model.TicketsStatus `json:"status" binding:"required"`
}

type TicketUsuarioResponse struct {
	ID        uint64              `json:"id"`
	UsuarioID uint64              `json:"usuario_id"`
	TicketID  uint64              `json:"ticket_id"`
	Status    model.TicketsStatus `json:"status"`
}

// RegisterRoutes registra os endpoints de tickets por usuario.
func (c *TicketUsuarioController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/tickets-usuario", c.Create)
	rg.GET("/tickets-usuario/:id", c.GetByID)
	rg.GET("/usuarios/:id/tickets", c.ListByUsuarioID)
	rg.PATCH("/tickets-usuario/:id/status", c.UpdateStatus)
	rg.DELETE("/tickets-usuario/:id", c.Delete)
}

// Create cria o vinculo entre usuario e ticket.
func (c *TicketUsuarioController) Create(ctx *gin.Context) {
	var req TicketUsuarioCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("ticket_usuario_criar", false, map[string]any{
			"usuario_id": req.UsuarioID,
			"ticket_id":  req.TicketID,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	pessoa, err := c.pRepo.GetByID(ctx, req.UsuarioID)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_usuario_criar", false, map[string]any{
			"usuario_id": req.UsuarioID,
			"ticket_id":  req.TicketID,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "usuario nao encontrado"})
		return
	}

	status := req.Status
	if status == "" {
		status = model.TicketsStatusPendente
	}

	tu := model.TicketUsuario{
		UsuarioID: req.UsuarioID,
		TicketID:  req.TicketID,
		Status:    status,
	}
	if err := c.repo.Create(ctx, &tu); err != nil {
		audit.GetLogger().LogEvent("ticket_usuario_criar", false, map[string]any{
			"usuario_id": pessoa.ID,
			"ticket_id":  req.TicketID,
			"cpf":        pessoa.CPF,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao criar"})
		return
	}

	audit.GetLogger().LogEvent("ticket_usuario_criar", true, map[string]any{
		"ticket_usuario_id": tu.ID,
		"usuario_id":        pessoa.ID,
		"ticket_id":         tu.TicketID,
		"cpf":               pessoa.CPF,
	}, nil)

	ctx.JSON(http.StatusCreated, toTicketUsuarioResponse(tu))
}

// GetByID busca o vinculo pelo ID.
func (c *TicketUsuarioController) GetByID(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	tu, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_usuario_buscar", false, map[string]any{
			"ticket_usuario_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "registro nao encontrado"})
		return
	}

	audit.GetLogger().LogEvent("ticket_usuario_buscar", true, map[string]any{
		"ticket_usuario_id": tu.ID,
		"usuario_id":        tu.UsuarioID,
		"ticket_id":         tu.TicketID,
		"cpf":               tu.Usuario.CPF,
	}, nil)

	ctx.JSON(http.StatusOK, toTicketUsuarioResponse(tu))
}

// ListByUsuarioID lista tickets associados ao usuario.
func (c *TicketUsuarioController) ListByUsuarioID(ctx *gin.Context) {
	usuarioID, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	limit, _ := strconvAtoiDefault(ctx.Query("limit"), 50)
	offset, _ := strconvAtoiDefault(ctx.Query("offset"), 0)

	tus, err := c.repo.ListByUsuarioID(ctx, usuarioID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao listar"})
		return
	}

	resp := make([]TicketUsuarioResponse, 0, len(tus))
	for _, tu := range tus {
		resp = append(resp, toTicketUsuarioResponse(tu))
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateStatus atualiza o status do vinculo.
func (c *TicketUsuarioController) UpdateStatus(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	var req TicketUsuarioStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("ticket_usuario_status", false, map[string]any{
			"ticket_usuario_id": id,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	tu, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_usuario_status", false, map[string]any{
			"ticket_usuario_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "registro nao encontrado"})
		return
	}

	if err := c.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		audit.GetLogger().LogEvent("ticket_usuario_status", false, map[string]any{
			"ticket_usuario_id": tu.ID,
			"usuario_id":        tu.UsuarioID,
			"ticket_id":         tu.TicketID,
			"status":            req.Status,
			"cpf":               tu.Usuario.CPF,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao atualizar"})
		return
	}

	audit.GetLogger().LogEvent("ticket_usuario_status", true, map[string]any{
		"ticket_usuario_id": tu.ID,
		"usuario_id":        tu.UsuarioID,
		"ticket_id":         tu.TicketID,
		"status":            req.Status,
		"cpf":               tu.Usuario.CPF,
	}, nil)

	ctx.Status(http.StatusNoContent)
}

// Delete remove o vinculo entre usuario e ticket.
func (c *TicketUsuarioController) Delete(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	tu, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("ticket_usuario_remover", false, map[string]any{
			"ticket_usuario_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "registro nao encontrado"})
		return
	}

	if err := c.repo.Delete(ctx, id); err != nil {
		audit.GetLogger().LogEvent("ticket_usuario_remover", false, map[string]any{
			"ticket_usuario_id": tu.ID,
			"usuario_id":        tu.UsuarioID,
			"ticket_id":         tu.TicketID,
			"cpf":               tu.Usuario.CPF,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao remover"})
		return
	}

	audit.GetLogger().LogEvent("ticket_usuario_remover", true, map[string]any{
		"ticket_usuario_id": tu.ID,
		"usuario_id":        tu.UsuarioID,
		"ticket_id":         tu.TicketID,
		"cpf":               tu.Usuario.CPF,
	}, nil)

	ctx.Status(http.StatusNoContent)
}

// toTicketUsuarioResponse converte o modelo para resposta JSON.
func toTicketUsuarioResponse(tu model.TicketUsuario) TicketUsuarioResponse {
	return TicketUsuarioResponse{
		ID:        tu.ID,
		UsuarioID: tu.UsuarioID,
		TicketID:  tu.TicketID,
		Status:    tu.Status,
	}
}
