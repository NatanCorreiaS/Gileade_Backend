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

func NewRouter(deps AppDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	_ = r.SetTrustedProxies(nil)

	api := r.Group("/api/v1")
	controller.NewPessoaController(deps.DB).RegisterRoutes(api)
	controller.NewTicketController(deps.DB).RegisterRoutes(api)
	controller.NewTicketUsuarioController(deps.DB).RegisterRoutes(api)
	controller.NewMercadoPagoController(deps.DB, deps.MP).RegisterRoutes(api)

	return r
}
