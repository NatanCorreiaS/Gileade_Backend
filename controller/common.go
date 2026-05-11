package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// parseUintParam valida parametro de rota e converte para uint64.
func parseUintParam(ctx *gin.Context, name string) (uint64, bool) {
	val := ctx.Param(name)
	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"erro": "parametro invalido"})
		return 0, false
	}
	return id, true
}

// isUniqueViolation detecta violacao de unique constraint do Postgres.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
