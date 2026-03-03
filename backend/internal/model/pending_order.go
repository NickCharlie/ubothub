package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Payment order status constants.
const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusFailed    = "failed"
	OrderStatusCancelled = "cancelled"
	OrderStatusExpired   = "expired"
)

// PendingOrder tracks a payment order from creation to completion.
// It maps order_id → user_id so the payment callback can credit the correct wallet.
type PendingOrder struct {
	ID             string          `gorm:"type:char(26);primaryKey" json:"id"`
	UserID         string          `gorm:"type:char(26);index;not null;column:user_id" json:"user_id"`
	WalletID       string          `gorm:"type:char(26);index;not null;column:wallet_id" json:"wallet_id"`
	Amount         decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	Status         string          `gorm:"type:varchar(16);index;not null;default:pending" json:"status"`
	PaymentChannel string          `gorm:"type:varchar(32);not null" json:"payment_channel"`
	ExternalID     string          `gorm:"type:varchar(128);column:external_id" json:"external_id,omitempty"`
	Description    string          `gorm:"type:varchar(255)" json:"description,omitempty"`
	PayURL         string          `gorm:"-" json:"pay_url,omitempty"`
	QRCodeURL      string          `gorm:"-" json:"qr_code_url,omitempty"`
	ExpiresAt      time.Time       `gorm:"not null" json:"expires_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	CreatedAt      time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`

	User   User   `gorm:"foreignKey:UserID" json:"-"`
	Wallet Wallet `gorm:"foreignKey:WalletID" json:"-"`
}
