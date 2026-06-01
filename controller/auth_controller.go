package controller

import (
	"errors"
	"net/http"
	"strings"

	"gileade/gileade_backend/audit"
	"gileade/gileade_backend/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	authService *service.AuthService
}

// NewAuthController monta o controller de autenticacao.
func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{authService: service.NewAuthService(db)}
}

type LoginRequest struct {
	CPF   string `json:"cpf" binding:"required"`
	Senha string `json:"senha" binding:"required"`
}

type LoginResponse struct {
	Token   string         `json:"token"`
	Usuario PessoaResponse `json:"usuario"`
}

// RegisterRoutes registra os endpoints de autenticacao.
func (c *AuthController) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.POST("/login", c.Login)
	auth.POST("/logout", c.Logout)
}

// Login autentica um usuario e retorna o token JWT.
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		audit.GetLogger().LogEvent("auth_login", false, map[string]any{
			"cpf": req.CPF,
		}, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "cpf e senha sao obrigatorios"})
		return
	}

	pessoa, token, err := c.authService.Login(ctx, req.CPF, req.Senha)
	if err != nil {
		audit.GetLogger().LogEvent("auth_login", false, map[string]any{
			"cpf": req.CPF,
		}, err)

		if errors.Is(err, service.ErrCredenciaisInvalidas) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"erro": "cpf ou senha invalidos"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao autenticar"})
		return
	}

	audit.GetLogger().LogEvent("auth_login", true, map[string]any{
		"pessoa_id":    pessoa.ID,
		"cpf":          pessoa.CPF,
		"tipo_usuario": pessoa.TipoUsuario,
	}, nil)

	ctx.JSON(http.StatusOK, LoginResponse{
		Token:   token,
		Usuario: toPessoaResponse(pessoa),
	})
}

// Logout invalida o token JWT do usuario.
func (c *AuthController) Logout(ctx *gin.Context) {
	tokenStr := extractBearerToken(ctx)
	if tokenStr == "" {
		audit.GetLogger().LogEvent("auth_logout", false, nil, errors.New("token ausente"))
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "token de autorizacao ausente"})
		return
	}

	if err := c.authService.Logout(tokenStr); err != nil {
		audit.GetLogger().LogEvent("auth_logout", false, map[string]any{
			"token": tokenStr,
		}, err)

		if errors.Is(err, service.ErrTokenInvalido) || errors.Is(err, service.ErrTokenExpirado) {
			ctx.JSON(http.StatusOK, gin.H{"mensagem": "logout realizado"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao realizar logout"})
		return
	}

	audit.GetLogger().LogEvent("auth_logout", true, nil, nil)
	ctx.JSON(http.StatusOK, gin.H{"mensagem": "logout realizado"})
}

// extractBearerToken extrai o token do header Authorization.
func extractBearerToken(ctx *gin.Context) string {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
