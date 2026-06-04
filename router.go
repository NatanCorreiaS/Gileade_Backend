package main

import (
	"gileade/gileade_backend/controller"
	"gileade/gileade_backend/gateway"
	"gileade/gileade_backend/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppDeps struct {
	DB *gorm.DB
	MP *gateway.MercadoPagoGateway
}

// NewRouter registra as rotas HTTP do servico com middleware de autenticacao.
func NewRouter(deps AppDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	_ = r.SetTrustedProxies(nil)

	authService := service.NewAuthService(deps.DB)
	authMiddleware := controller.AuthMiddleware(authService)

	api := r.Group("/api/v1")

	controller.NewAuthController(deps.DB).RegisterRoutes(api)

	pessoas := api.Group("/pessoas")
	pessoas.POST("", controller.NewPessoaController(deps.DB).Create)
	pessoas.GET("", controller.NewPessoaController(deps.DB).List)
	pessoas.GET("/:id", controller.NewPessoaController(deps.DB).GetByID)
	pessoas.PUT("/:id", authMiddleware, controller.NewPessoaController(deps.DB).Update)
	pessoas.DELETE("/:id", authMiddleware, controller.NewPessoaController(deps.DB).Delete)

	tickets := api.Group("/tickets")
	tickets.POST("", authMiddleware, controller.NewTicketController(deps.DB).Create)
	tickets.GET("", controller.NewTicketController(deps.DB).List)
	tickets.GET("/:id", controller.NewTicketController(deps.DB).GetByID)
	tickets.PUT("/:id", authMiddleware, controller.NewTicketController(deps.DB).Update)
	tickets.DELETE("/:id", authMiddleware, controller.NewTicketController(deps.DB).Delete)

	ticketsCompra := api.Group("/tickets-compra")
	ticketsCompra.POST("", authMiddleware, controller.NewTicketCompraController(deps.DB).Create)
	ticketsCompra.GET("/:id", controller.NewTicketCompraController(deps.DB).GetByID)
	ticketsCompra.PATCH("/:id/status", authMiddleware, controller.NewTicketCompraController(deps.DB).UpdateStatus)
	ticketsCompra.DELETE("/:id", authMiddleware, controller.NewTicketCompraController(deps.DB).Delete)

	api.GET("/usuarios/:id/tickets-compra", controller.NewTicketCompraController(deps.DB).ListByUsuarioID)

	pagamentos := api.Group("/pagamentos")
	pagamentos.POST("/checkout", authMiddleware, controller.NewPagamentoController(deps.DB, deps.MP).CreateCheckout)
	pagamentos.GET("", controller.NewPagamentoController(deps.DB, deps.MP).ListPayments)
	pagamentos.POST("/webhook", controller.NewPagamentoController(deps.DB, deps.MP).HandleWebhook)
	pagamentos.POST("/:id/estornos", authMiddleware, controller.NewEstornoController(deps.DB, deps.MP).Create)

	return r
}
