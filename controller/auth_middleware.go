package controller

import (
	"errors"
	"net/http"
	"strings"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/service"

	"github.com/gin-gonic/gin"
)

const (
	ctxKeyClaims      = "auth_claims"
	ctxKeyUsuarioID   = "usuario_id"
	ctxKeyTipoUsuario = "tipo_usuario"
)

// AuthMiddleware extrai e valida o token JWT do header Authorization.
// Injeta as claims e dados do usuario no contexto.
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := extractBearerToken(ctx)
		if tokenStr == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "token de autorizacao ausente"})
			return
		}

		claims, err := authService.ValidateToken(tokenStr)
		if err != nil {
			if errors.Is(err, service.ErrTokenExpirado) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "token expirado"})
				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "token invalido"})
			return
		}

		if authService.IsTokenBlacklisted(claims) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "token revogado"})
			return
		}

		ctx.Set(ctxKeyClaims, claims)
		ctx.Set(ctxKeyUsuarioID, claims.UsuarioID)
		ctx.Set(ctxKeyTipoUsuario, claims.TipoUsuario)

		ctx.Next()
	}
}

// AdminMiddleware valida que o usuario autenticado e um admin.
// Deve ser usado apos o AuthMiddleware.
func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tipoUsuario, _ := ctx.Get(ctxKeyTipoUsuario)
		if tipoUsuario != string(model.TipoUsuarioAdmin) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"erro": "acesso restrito a administradores"})
			return
		}
		ctx.Next()
	}
}

// GetAuthUsuarioID retorna o ID do usuario autenticado do contexto.
func GetAuthUsuarioID(ctx *gin.Context) (uint64, bool) {
	val, exists := ctx.Get(ctxKeyUsuarioID)
	if !exists {
		return 0, false
	}
	id, ok := val.(uint64)
	return id, ok
}

// GetAuthTipoUsuario retorna o tipo de usuario autenticado do contexto.
func GetAuthTipoUsuario(ctx *gin.Context) string {
	val, exists := ctx.Get(ctxKeyTipoUsuario)
	if !exists {
		return ""
	}
	tipo, _ := val.(string)
	return tipo
}

// IsAdmin verifica se o usuario autenticado e admin.
func IsAdmin(ctx *gin.Context) bool {
	return GetAuthTipoUsuario(ctx) == string(model.TipoUsuarioAdmin)
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
