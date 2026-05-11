package gateway

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
	"github.com/mercadopago/sdk-go/pkg/refund"
)

type MercadoPagoGateway struct {
	cfg        *config.Config
	preference preference.Client
	payment    payment.Client
	refund     refund.Client
}

// NewMercadoPagoGateway monta o gateway com clientes configurados.
func NewMercadoPagoGateway(cfg *config.Config, pref preference.Client, pay payment.Client, ref refund.Client) (*MercadoPagoGateway, error) {
	if cfg == nil {
		return nil, errors.New("config Mercado Pago inválida")
	}
	if pref == nil {
		pref = preference.NewClient(cfg)
	}
	if pay == nil {
		pay = payment.NewClient(cfg)
	}
	if ref == nil {
		ref = refund.NewClient(cfg)
	}

	return &MercadoPagoGateway{
		cfg:        cfg,
		preference: pref,
		payment:    pay,
		refund:     ref,
	}, nil
}

// NewMercadoPagoGatewayFromEnv carrega a configuracao do gateway via .env.
func NewMercadoPagoGatewayFromEnv() (*MercadoPagoGateway, error) {
	token := os.Getenv("MERCADO_PAGO_ACCESS_TOKEN_TEST")
	if token == "" {
		return nil, errors.New("MERCADO_PAGO_ACCESS_TOKEN_TEST não definido no .env")
	}

	cfg, err := config.New(token)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar configuração do Mercado Pago: %w", err)
	}

	return NewMercadoPagoGateway(cfg, nil, nil, nil)
}

// CreateCheckoutPro cria a preferencia de checkout no Mercado Pago.
func (g *MercadoPagoGateway) CreateCheckoutPro(ctx context.Context, req preference.Request) (*preference.Response, error) {
	return g.preference.Create(ctx, req)
}

// GetPayment consulta um pagamento por ID no Mercado Pago.
func (g *MercadoPagoGateway) GetPayment(ctx context.Context, id int) (*payment.Response, error) {
	return g.payment.Get(ctx, id)
}

// SearchPayments busca pagamentos no Mercado Pago.
func (g *MercadoPagoGateway) SearchPayments(ctx context.Context, req payment.SearchRequest) (*payment.SearchResponse, error) {
	return g.payment.Search(ctx, req)
}

// CreateRefund solicita estorno total de um pagamento.
func (g *MercadoPagoGateway) CreateRefund(ctx context.Context, paymentID int) (*refund.Response, error) {
	return g.refund.Create(ctx, paymentID)
}

// CreatePartialRefund solicita estorno parcial de um pagamento.
func (g *MercadoPagoGateway) CreatePartialRefund(ctx context.Context, paymentID int, amount float64) (*refund.Response, error) {
	return g.refund.CreatePartialRefund(ctx, paymentID, amount)
}
