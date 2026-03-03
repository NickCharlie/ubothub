package payment

import (
	"context"

	"github.com/shopspring/decimal"
)

// CreateOrderRequest holds parameters for creating a payment order.
type CreateOrderRequest struct {
	OrderID     string          `json:"order_id"`
	UserID      string          `json:"user_id"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	Channel     string          `json:"channel"`
	NotifyURL   string          `json:"notify_url"`
	ReturnURL   string          `json:"return_url"`
}

// CreateOrderResponse holds the result of a payment order creation.
type CreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	PayURL      string `json:"pay_url"`
	QRCodeURL   string `json:"qr_code_url,omitempty"`
	ExternalID  string `json:"external_id,omitempty"`
	Channel     string `json:"channel"`
}

// TransferRequest holds parameters for transferring funds to a user (withdrawal).
type TransferRequest struct {
	OrderID     string          `json:"order_id"`
	UserID      string          `json:"user_id"`
	Amount      decimal.Decimal `json:"amount"`
	Account     string          `json:"account"`
	AccountName string          `json:"account_name"`
	Channel     string          `json:"channel"`
	Description string          `json:"description"`
}

// TransferResponse holds the result of a transfer operation.
type TransferResponse struct {
	OrderID    string `json:"order_id"`
	ExternalID string `json:"external_id,omitempty"`
	Status     string `json:"status"`
}

// QueryOrderResponse holds the payment order query result.
type QueryOrderResponse struct {
	OrderID    string          `json:"order_id"`
	ExternalID string          `json:"external_id,omitempty"`
	Amount     decimal.Decimal `json:"amount"`
	Status     string          `json:"status"`
	Channel    string          `json:"channel"`
	PaidAt     string          `json:"paid_at,omitempty"`
}

// NotifyResult holds parsed payment notification data.
type NotifyResult struct {
	OrderID    string          `json:"order_id"`
	ExternalID string          `json:"external_id"`
	Amount     decimal.Decimal `json:"amount"`
	Status     string          `json:"status"`
	Channel    string          `json:"channel"`
}

// Provider defines the abstract interface for third-party payment platforms.
// Implementations can support YunGouOS, gopay, xunhupay, or any other provider.
type Provider interface {
	// Name returns the provider identifier (e.g., "yungouos", "xunhupay").
	Name() string

	// CreateOrder initiates a payment order and returns a pay URL or QR code.
	CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error)

	// QueryOrder retrieves the status of a payment order.
	QueryOrder(ctx context.Context, orderID string) (*QueryOrderResponse, error)

	// Transfer sends funds to a user's external account (for withdrawal).
	Transfer(ctx context.Context, req TransferRequest) (*TransferResponse, error)

	// ParseNotify parses an async payment notification from the provider.
	ParseNotify(ctx context.Context, body []byte) (*NotifyResult, error)

	// VerifyNotify verifies the signature of a payment notification.
	VerifyNotify(ctx context.Context, body []byte) (bool, error)
}
