package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Transaction type constants.
const (
	TxTypeTopUp      = "top_up"
	TxTypeUsage      = "usage"
	TxTypeEarning    = "earning"
	TxTypeWithdrawal = "withdrawal"
	TxTypeRefund     = "refund"
	TxTypeSubsidy    = "subsidy"
)

// Transaction status constants.
const (
	TxStatusPending   = "pending"
	TxStatusCompleted = "completed"
	TxStatusFailed    = "failed"
	TxStatusCancelled = "cancelled"
)

// Transaction represents a financial transaction record.
type Transaction struct {
	ID              string          `gorm:"type:char(26);primaryKey" json:"id"`
	UserID          string          `gorm:"type:char(26);index;not null;column:user_id" json:"user_id"`
	WalletID        string          `gorm:"type:char(26);index;not null;column:wallet_id" json:"wallet_id"`
	Type            string          `gorm:"type:varchar(16);not null;index" json:"type"`
	Amount          decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	BalanceBefore   decimal.Decimal `gorm:"type:decimal(12,2);not null;column:balance_before" json:"balance_before"`
	BalanceAfter    decimal.Decimal `gorm:"type:decimal(12,2);not null;column:balance_after" json:"balance_after"`
	Status          string          `gorm:"type:varchar(16);default:pending;not null" json:"status"`
	BotID           string          `gorm:"type:char(26);column:bot_id" json:"bot_id,omitempty"`
	ExternalOrderID string          `gorm:"type:varchar(128);column:external_order_id" json:"external_order_id,omitempty"`
	PaymentChannel  string          `gorm:"type:varchar(32);column:payment_channel" json:"payment_channel,omitempty"`
	Description     string          `gorm:"type:varchar(255)" json:"description"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`

	User   User   `gorm:"foreignKey:UserID" json:"-"`
	Wallet Wallet `gorm:"foreignKey:WalletID" json:"-"`
}

// TableName specifies the database table name for the Transaction model.
func (Transaction) TableName() string {
	return "transactions"
}
