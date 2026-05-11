package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

type fakePreferenceClient struct {
	createCalled bool
	createReq    preference.Request
	createResp   *preference.Response
	createErr    error
}

// Create registra chamada de criacao de preferencia.
func (f *fakePreferenceClient) Create(ctx context.Context, request preference.Request) (*preference.Response, error) {
	f.createCalled = true
	f.createReq = request
	return f.createResp, f.createErr
}

// Get nao e utilizado nos testes atuais.
func (f *fakePreferenceClient) Get(ctx context.Context, id string) (*preference.Response, error) {
	return nil, errors.New("not implemented")
}

// Update nao e utilizado nos testes atuais.
func (f *fakePreferenceClient) Update(ctx context.Context, id string, request preference.Request) (*preference.Response, error) {
	return nil, errors.New("not implemented")
}

// Search nao e utilizado nos testes atuais.
func (f *fakePreferenceClient) Search(ctx context.Context, request preference.SearchRequest) (*preference.PagingResponse, error) {
	return nil, errors.New("not implemented")
}

type fakePaymentClient struct {
	getCalled bool
	getID     int
	getResp   *payment.Response
	getErr    error
}

// Create nao e utilizado nos testes atuais.
func (f *fakePaymentClient) Create(ctx context.Context, request payment.Request) (*payment.Response, error) {
	return nil, errors.New("not implemented")
}

// Search nao e utilizado nos testes atuais.
func (f *fakePaymentClient) Search(ctx context.Context, request payment.SearchRequest) (*payment.SearchResponse, error) {
	return nil, errors.New("not implemented")
}

// Get retorna o pagamento configurado no fake.
func (f *fakePaymentClient) Get(ctx context.Context, id int) (*payment.Response, error) {
	f.getCalled = true
	f.getID = id
	return f.getResp, f.getErr
}

// Cancel nao e utilizado nos testes atuais.
func (f *fakePaymentClient) Cancel(ctx context.Context, id int) (*payment.Response, error) {
	return nil, errors.New("not implemented")
}

// Capture nao e utilizado nos testes atuais.
func (f *fakePaymentClient) Capture(ctx context.Context, id int) (*payment.Response, error) {
	return nil, errors.New("not implemented")
}

// CaptureAmount nao e utilizado nos testes atuais.
func (f *fakePaymentClient) CaptureAmount(ctx context.Context, id int, amount float64) (*payment.Response, error) {
	return nil, errors.New("not implemented")
}

// TestNewMercadoPagoGatewayFromEnv valida leitura de token do ambiente.
func TestNewMercadoPagoGatewayFromEnv(t *testing.T) {
	cases := []struct {
		name    string
		setEnv  func(t *testing.T)
		wantErr bool
	}{
		{
			name: "missing token",
			setEnv: func(t *testing.T) {
				t.Setenv("MERCADO_PAGO_ACCESS_TOKEN_TEST", "")
			},
			wantErr: true,
		},
		{
			name: "access token set",
			setEnv: func(t *testing.T) {
				t.Setenv("MERCADO_PAGO_ACCESS_TOKEN_TEST", "test-token")
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setEnv(t)
			gw, err := NewMercadoPagoGatewayFromEnv()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.wantErr && gw == nil {
				t.Fatalf("expected gateway")
			}
		})
	}
}

// TestGatewayCreateCheckoutPro valida criacao de preferencia via gateway.
func TestGatewayCreateCheckoutPro(t *testing.T) {
	cfg, err := config.New("token")
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}

	prefResp := &preference.Response{ID: "pref-1"}
	prefClient := &fakePreferenceClient{createResp: prefResp}
	payClient := &fakePaymentClient{}

	gw, err := NewMercadoPagoGateway(cfg, prefClient, payClient, nil)
	if err != nil {
		t.Fatalf("NewMercadoPagoGateway: %v", err)
	}

	resp, err := gw.CreateCheckoutPro(context.Background(), preference.Request{ExternalReference: "1"})
	if err != nil {
		t.Fatalf("CreateCheckoutPro: %v", err)
	}
	if resp.ID != "pref-1" {
		t.Fatalf("unexpected preference id: %s", resp.ID)
	}
	if !prefClient.createCalled {
		t.Fatalf("expected Create to be called")
	}
}

// TestGatewayGetPayment valida consulta de pagamento via gateway.
func TestGatewayGetPayment(t *testing.T) {
	cfg, err := config.New("token")
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}

	payResp := &payment.Response{ID: 123}
	prefClient := &fakePreferenceClient{}
	payClient := &fakePaymentClient{getResp: payResp}

	gw, err := NewMercadoPagoGateway(cfg, prefClient, payClient, nil)
	if err != nil {
		t.Fatalf("NewMercadoPagoGateway: %v", err)
	}

	resp, err := gw.GetPayment(context.Background(), 123)
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	if resp.ID != 123 {
		t.Fatalf("unexpected payment id: %d", resp.ID)
	}
	if !payClient.getCalled {
		t.Fatalf("expected Get to be called")
	}
	if payClient.getID != 123 {
		t.Fatalf("unexpected id passed: %d", payClient.getID)
	}
}

// TestNewMercadoPagoGatewayWithNilConfig valida erro para config nula.
func TestNewMercadoPagoGatewayWithNilConfig(t *testing.T) {
	if _, err := NewMercadoPagoGateway(nil, nil, nil, nil); err == nil {
		t.Fatalf("expected error")
	}
}
