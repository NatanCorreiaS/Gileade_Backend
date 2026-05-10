package controller

import (
	"net/http"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketUsuarioController struct {
	repo *repository.TicketUsuarioRepository
}

func NewTicketUsuarioController(db *gorm.DB) *TicketUsuarioController {
	return &TicketUsuarioController{repo: repository.NewTicketUsuarioRepository(db)}
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

func (c *TicketUsuarioController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/tickets-usuario", c.Create)
	rg.GET("/tickets-usuario/:id", c.GetByID)
	rg.GET("/usuarios/:id/tickets", c.ListByUsuarioID)
	rg.PATCH("/tickets-usuario/:id/status", c.UpdateStatus)
	rg.DELETE("/tickets-usuario/:id", c.Delete)
}

func (c *TicketUsuarioController) Create(ctx *gin.Context) {
	var req TicketUsuarioCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao criar"})
		return
	}

	ctx.JSON(http.StatusCreated, toTicketUsuarioResponse(tu))
}

func (c *TicketUsuarioController) GetByID(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	tu, err := c.repo.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "registro nao encontrado"})
		return
	}

	ctx.JSON(http.StatusOK, toTicketUsuarioResponse(tu))
}

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

func (c *TicketUsuarioController) UpdateStatus(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	var req TicketUsuarioStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	if err := c.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao atualizar"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *TicketUsuarioController) Delete(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	if err := c.repo.Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao remover"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func toTicketUsuarioResponse(tu model.TicketUsuario) TicketUsuarioResponse {
	return TicketUsuarioResponse{
		ID:        tu.ID,
		UsuarioID: tu.UsuarioID,
		TicketID:  tu.TicketID,
		Status:    tu.Status,
	}
}
