package gateway

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

type MercadoPagoGateway struct {
	cfg        *config.Config
	preference preference.Client
	payment    payment.Client
}

func NewMercadoPagoGateway(cfg *config.Config, pref preference.Client, pay payment.Client) (*MercadoPagoGateway, error) {
	if cfg == nil {
		return nil, errors.New("config Mercado Pago inválida")
	}
	if pref == nil {
		pref = preference.NewClient(cfg)
	}
	if pay == nil {
		pay = payment.NewClient(cfg)
	}

	return &MercadoPagoGateway{
		cfg:        cfg,
		preference: pref,
		payment:    pay,
	}, nil
}

func NewMercadoPagoGatewayFromEnv() (*MercadoPagoGateway, error) {
	token := os.Getenv("MERCADO_PAGO_ACCESS_TOKEN_TEST")
	if token == "" {
		return nil, errors.New("MERCADO_PAGO_ACCESS_TOKEN_TEST não definido no .env")
	}

	cfg, err := config.New(token)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar configuração do Mercado Pago: %w", err)
	}

	return NewMercadoPagoGateway(cfg, nil, nil)
}

func (g *MercadoPagoGateway) CreateCheckoutPro(ctx context.Context, req preference.Request) (*preference.Response, error) {
	return g.preference.Create(ctx, req)
}

func (g *MercadoPagoGateway) GetPayment(ctx context.Context, id int) (*payment.Response, error) {
	return g.payment.Get(ctx, id)
}
