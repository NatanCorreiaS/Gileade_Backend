package controller

import (
	"net/http"
	"strconv"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/audit"
	"gileade/gileade_backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PessoaController struct {
	repo *repository.PessoaRepository
}

func NewPessoaController(db *gorm.DB) *PessoaController {
	return &PessoaController{repo: repository.NewPessoaRepository(db)}
}

type PessoaCreateRequest struct {
	Nome         string             `json:"nome" binding:"required"`
	TipoUsuario  model.TipoUsuario  `json:"tipo_usuario" binding:"required"`
	Senha        string             `json:"senha" binding:"required"`
	CPF          string             `json:"cpf" binding:"required"`
	Idade        int16              `json:"idade"`
	Igreja       string             `json:"igreja"`
	PapelIgreja  model.PapelIgreja  `json:"papel_igreja"`
	EstadoCivil  model.EstadoCivil  `json:"estado_civil"`
	Email        string             `json:"email" binding:"required"`
	Sexo         model.Sexo         `json:"sexo" binding:"required"`
	Cidade       string             `json:"cidade"`
	EstadoUF     model.EstadoUF     `json:"estado_uf"`
	Escolaridade model.Escolaridade `json:"escolaridade"`
}

type PessoaUpdateRequest struct {
	Nome         *string             `json:"nome"`
	TipoUsuario  *model.TipoUsuario  `json:"tipo_usuario"`
	Senha        *string             `json:"senha"`
	CPF          *string             `json:"cpf"`
	Idade        *int16              `json:"idade"`
	Igreja       *string             `json:"igreja"`
	PapelIgreja  *model.PapelIgreja  `json:"papel_igreja"`
	EstadoCivil  *model.EstadoCivil  `json:"estado_civil"`
	Email        *string             `json:"email"`
	Sexo         *model.Sexo         `json:"sexo"`
	Cidade       *string             `json:"cidade"`
	EstadoUF     *model.EstadoUF     `json:"estado_uf"`
	Escolaridade *model.Escolaridade `json:"escolaridade"`
}

type PessoaResponse struct {
	ID           uint64             `json:"id"`
	Nome         string             `json:"nome"`
	TipoUsuario  model.TipoUsuario  `json:"tipo_usuario"`
	CPF          string             `json:"cpf"`
	Idade        int16              `json:"idade"`
	Igreja       string             `json:"igreja"`
	PapelIgreja  model.PapelIgreja  `json:"papel_igreja"`
	EstadoCivil  model.EstadoCivil  `json:"estado_civil"`
	Email        string             `json:"email"`
	Sexo         model.Sexo         `json:"sexo"`
	Cidade       string             `json:"cidade"`
	EstadoUF     model.EstadoUF     `json:"estado_uf"`
	Escolaridade model.Escolaridade `json:"escolaridade"`
}

func (c *PessoaController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/pessoas", c.Create)
	rg.GET("/pessoas", c.List)
	rg.GET("/pessoas/:id", c.GetByID)
	rg.PUT("/pessoas/:id", c.Update)
	rg.DELETE("/pessoas/:id", c.Delete)
}

func (c *PessoaController) Create(ctx *gin.Context) {
	var req PessoaCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("pessoa_criar", false, map[string]any{
			"nome":         req.Nome,
			"cpf":          req.CPF,
			"tipo_usuario": req.TipoUsuario,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	pessoa := model.Pessoa{
		Nome:         req.Nome,
		TipoUsuario:  req.TipoUsuario,
		Senha:        req.Senha,
		CPF:          req.CPF,
		Idade:        req.Idade,
		Igreja:       req.Igreja,
		PapelIgreja:  req.PapelIgreja,
		EstadoCivil:  req.EstadoCivil,
		Email:        req.Email,
		Sexo:         req.Sexo,
		Cidade:       req.Cidade,
		EstadoUF:     req.EstadoUF,
		Escolaridade: req.Escolaridade,
	}
	if err := c.repo.Create(ctx, &pessoa); err != nil {
		audit.GetLogger().LogEvent("pessoa_criar", false, map[string]any{
			"nome":         pessoa.Nome,
			"cpf":          pessoa.CPF,
			"tipo_usuario": pessoa.TipoUsuario,
		}, err)
		if isUniqueViolation(err) {
			ctx.JSON(http.StatusConflict, gin.H{"erro": "cpf ja cadastrado"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao criar usuario"})
		return
	}

	audit.GetLogger().LogEvent("pessoa_criar", true, map[string]any{
		"pessoa_id":    pessoa.ID,
		"cpf":          pessoa.CPF,
		"tipo_usuario": pessoa.TipoUsuario,
	}, nil)

	ctx.JSON(http.StatusCreated, toPessoaResponse(pessoa))
}

func (c *PessoaController) List(ctx *gin.Context) {
	limit, _ := strconvAtoiDefault(ctx.Query("limit"), 50)
	offset, _ := strconvAtoiDefault(ctx.Query("offset"), 0)

	pessoas, err := c.repo.List(ctx, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao listar"})
		return
	}

	resp := make([]PessoaResponse, 0, len(pessoas))
	for _, p := range pessoas {
		resp = append(resp, toPessoaResponse(p))
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *PessoaController) GetByID(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	pessoa, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("pessoa_buscar", false, map[string]any{
			"pessoa_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "usuario nao encontrado"})
		return
	}

	audit.GetLogger().LogEvent("pessoa_buscar", true, map[string]any{
		"pessoa_id":    pessoa.ID,
		"cpf":          pessoa.CPF,
		"tipo_usuario": pessoa.TipoUsuario,
	}, nil)

	ctx.JSON(http.StatusOK, toPessoaResponse(pessoa))
}

func (c *PessoaController) Update(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	var req PessoaUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("pessoa_atualizar", false, map[string]any{
			"pessoa_id": id,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "payload invalido"})
		return
	}

	pessoa, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("pessoa_atualizar", false, map[string]any{
			"pessoa_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "usuario nao encontrado"})
		return
	}

	if req.Nome != nil {
		pessoa.Nome = *req.Nome
	}
	if req.TipoUsuario != nil {
		pessoa.TipoUsuario = *req.TipoUsuario
	}
	if req.Senha != nil {
		pessoa.Senha = *req.Senha
	}
	if req.CPF != nil {
		pessoa.CPF = *req.CPF
	}
	if req.Idade != nil {
		pessoa.Idade = *req.Idade
	}
	if req.Igreja != nil {
		pessoa.Igreja = *req.Igreja
	}
	if req.PapelIgreja != nil {
		pessoa.PapelIgreja = *req.PapelIgreja
	}
	if req.EstadoCivil != nil {
		pessoa.EstadoCivil = *req.EstadoCivil
	}
	if req.Email != nil {
		pessoa.Email = *req.Email
	}
	if req.Sexo != nil {
		pessoa.Sexo = *req.Sexo
	}
	if req.Cidade != nil {
		pessoa.Cidade = *req.Cidade
	}
	if req.EstadoUF != nil {
		pessoa.EstadoUF = *req.EstadoUF
	}
	if req.Escolaridade != nil {
		pessoa.Escolaridade = *req.Escolaridade
	}

	if err := c.repo.Update(ctx, &pessoa); err != nil {
		audit.GetLogger().LogEvent("pessoa_atualizar", false, map[string]any{
			"pessoa_id":    pessoa.ID,
			"cpf":          pessoa.CPF,
			"tipo_usuario": pessoa.TipoUsuario,
		}, err)
		if isUniqueViolation(err) {
			ctx.JSON(http.StatusConflict, gin.H{"erro": "cpf ja cadastrado"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao atualizar"})
		return
	}

	audit.GetLogger().LogEvent("pessoa_atualizar", true, map[string]any{
		"pessoa_id":    pessoa.ID,
		"cpf":          pessoa.CPF,
		"tipo_usuario": pessoa.TipoUsuario,
	}, nil)

	ctx.JSON(http.StatusOK, toPessoaResponse(pessoa))
}

func (c *PessoaController) Delete(ctx *gin.Context) {
	id, ok := parseUintParam(ctx, "id")
	if !ok {
		return
	}

	pessoa, err := c.repo.GetByID(ctx, id)
	if err != nil {
		audit.GetLogger().LogEvent("pessoa_remover", false, map[string]any{
			"pessoa_id": id,
		}, err)
		ctx.JSON(http.StatusNotFound, gin.H{"erro": "usuario nao encontrado"})
		return
	}

	if err := c.repo.Delete(ctx, id); err != nil {
		audit.GetLogger().LogEvent("pessoa_remover", false, map[string]any{
			"pessoa_id":    pessoa.ID,
			"cpf":          pessoa.CPF,
			"tipo_usuario": pessoa.TipoUsuario,
		}, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao remover"})
		return
	}

	audit.GetLogger().LogEvent("pessoa_remover", true, map[string]any{
		"pessoa_id":    pessoa.ID,
		"cpf":          pessoa.CPF,
		"tipo_usuario": pessoa.TipoUsuario,
	}, nil)

	ctx.Status(http.StatusNoContent)
}

func toPessoaResponse(p model.Pessoa) PessoaResponse {
	return PessoaResponse{
		ID:           p.ID,
		Nome:         p.Nome,
		TipoUsuario:  p.TipoUsuario,
		CPF:          p.CPF,
		Idade:        p.Idade,
		Igreja:       p.Igreja,
		PapelIgreja:  p.PapelIgreja,
		EstadoCivil:  p.EstadoCivil,
		Email:        p.Email,
		Sexo:         p.Sexo,
		Cidade:       p.Cidade,
		EstadoUF:     p.EstadoUF,
		Escolaridade: p.Escolaridade,
	}
}

func strconvAtoiDefault(val string, def int) (int, bool) {
	if val == "" {
		return def, true
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return def, false
	}
	return parsed, true
}
