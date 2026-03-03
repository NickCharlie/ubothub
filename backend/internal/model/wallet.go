package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Wallet represents a user's platform wallet for billing and revenue.
type Wallet struct {
	ID        string          `gorm:"type:char(26);primaryKey" json:"id"`
	UserID    string          `gorm:"type:char(26);uniqueIndex;not null;column:user_id" json:"user_id"`
	Balance   decimal.Decimal `gorm:"type:decimal(12,2);default:0;not null" json:"balance"`
	Frozen    decimal.Decimal `gorm:"type:decimal(12,2);default:0;not null" json:"frozen"`
	Currency  string          `gorm:"type:varchar(8);default:CNY;not null" json:"currency"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the database table name for the Wallet model.
func (Wallet) TableName() string {
	return "wallets"
}

// AvailableBalance returns the balance minus frozen amount.
func (w *Wallet) AvailableBalance() decimal.Decimal {
	return w.Balance.Sub(w.Frozen)
}
