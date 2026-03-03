package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Bot pricing mode constants.
const (
	PricingModeFree       = "free"
	PricingModePerCall    = "per_call"
	PricingModeSubscribe  = "subscribe"
)

// BotPricing represents the monetization configuration for a public bot.
type BotPricing struct {
	ID              string          `gorm:"type:char(26);primaryKey" json:"id"`
	BotID           string          `gorm:"type:char(26);uniqueIndex;not null;column:bot_id" json:"bot_id"`
	Mode            string          `gorm:"type:varchar(16);default:free;not null" json:"mode"`
	PricePerCall    decimal.Decimal `gorm:"type:decimal(8,4);default:0;column:price_per_call" json:"price_per_call"`
	MonthlyPrice    decimal.Decimal `gorm:"type:decimal(8,2);default:0;column:monthly_price" json:"monthly_price"`
	PlatformRate    decimal.Decimal `gorm:"type:decimal(4,2);default:0.20;column:platform_rate" json:"platform_rate"`
	CreatorRate     decimal.Decimal `gorm:"type:decimal(4,2);default:0.80;column:creator_rate" json:"creator_rate"`
	FreeCallsPerDay int             `gorm:"default:0;column:free_calls_per_day" json:"free_calls_per_day"`
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`

	Bot Bot `gorm:"foreignKey:BotID" json:"-"`
}

// TableName specifies the database table name for the BotPricing model.
func (BotPricing) TableName() string {
	return "bot_pricing"
}

// CreatorEarning calculates the creator's share for a given amount.
func (p *BotPricing) CreatorEarning(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(p.CreatorRate).Round(2)
}

// PlatformFee calculates the platform's share for a given amount.
func (p *BotPricing) PlatformFee(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(p.PlatformRate).Round(2)
}
