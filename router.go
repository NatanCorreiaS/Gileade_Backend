package main

import (
	"gileade/gileade_backend/controller"
	"gileade/gileade_backend/gateway"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppDeps struct {
	DB *gorm.DB
	MP *gateway.MercadoPagoGateway
}

// NewRouter registra as rotas HTTP do servico.
func NewRouter(deps AppDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	_ = r.SetTrustedProxies(nil)

	api := r.Group("/api/v1")
	controller.NewAuthController(deps.DB).RegisterRoutes(api)
	controller.NewPessoaController(deps.DB).RegisterRoutes(api)
	controller.NewTicketController(deps.DB).RegisterRoutes(api)
	controller.NewTicketCompraController(deps.DB).RegisterRoutes(api)
	controller.NewPagamentoController(deps.DB, deps.MP).RegisterRoutes(api)
	controller.NewEstornoController(deps.DB, deps.MP).RegisterRoutes(api)

	return r
}
