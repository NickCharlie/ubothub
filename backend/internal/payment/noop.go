package payment

import (
	"context"
	"fmt"
	"net/http"

	"github.com/shopspring/decimal"
)

// noopProvider is a no-op payment provider used when payment is disabled.
type noopProvider struct{}

// NewNoopProvider creates a no-op payment provider for development and testing.
func NewNoopProvider() Provider {
	return &noopProvider{}
}

func (p *noopProvider) Name() string { return "noop" }

func (p *noopProvider) CreateOrder(_ context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	return &CreateOrderResponse{
		OrderID:    req.OrderID,
		PayURL:     fmt.Sprintf("https://example.com/pay/%s", req.OrderID),
		ExternalID: "noop-" + req.OrderID,
		Channel:    req.Channel,
	}, nil
}

func (p *noopProvider) QueryOrder(_ context.Context, orderID string) (*QueryOrderResponse, error) {
	return &QueryOrderResponse{
		OrderID: orderID,
		Amount:  decimal.NewFromFloat(0),
		Status:  "completed",
		Channel: "noop",
	}, nil
}

func (p *noopProvider) Transfer(_ context.Context, req TransferRequest) (*TransferResponse, error) {
	return &TransferResponse{
		OrderID:    req.OrderID,
		ExternalID: "noop-" + req.OrderID,
		Status:     "completed",
	}, nil
}

func (p *noopProvider) ParseNotify(_ context.Context, _ *http.Request) (*NotifyResult, error) {
	return nil, fmt.Errorf("noop provider does not support notifications")
}

func (p *noopProvider) VerifyNotify(_ context.Context, _ *http.Request) (bool, error) {
	return false, fmt.Errorf("noop provider does not support notifications")
}
