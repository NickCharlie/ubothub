package model

import (
	"time"
)

// Bot represents a connected chat bot entity.
type Bot struct {
	ID           string     `gorm:"type:char(26);primaryKey" json:"id"`
	UserID       string     `gorm:"type:char(26);index;not null;column:user_id" json:"user_id"`
	Name         string     `gorm:"type:varchar(128);not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	Framework    string     `gorm:"type:varchar(32);not null" json:"framework"`
	Visibility   string     `gorm:"type:varchar(16);default:private;not null" json:"visibility"`
	Status       string     `gorm:"type:varchar(16);default:offline;not null" json:"status"`
	AccessToken  string     `gorm:"type:varchar(64);uniqueIndex;not null;column:access_token" json:"-"`
	WebhookURL   string     `gorm:"type:text;column:webhook_url" json:"webhook_url"`
	Config       string     `gorm:"type:jsonb;default:'{}'" json:"config"`
	LastActiveAt *time.Time `gorm:"column:last_active_at" json:"last_active_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the database table name for the Bot model.
func (Bot) TableName() string {
	return "bots"
}

// Bot visibility constants.
const (
	BotVisibilityPublic  = "public"
	BotVisibilityPrivate = "private"
)
