package model

import (
	"time"
)

// MessageLog represents a bot message log entry for auditing and analysis.
type MessageLog struct {
	ID              string    `gorm:"type:char(26);primaryKey" json:"id"`
	BotID           string    `gorm:"type:char(26);not null;column:bot_id" json:"bot_id"`
	Direction       string    `gorm:"type:varchar(16);not null" json:"direction"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	Metadata        string    `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	ActionTriggered string    `gorm:"type:varchar(64);column:action_triggered" json:"action_triggered"`
	CreatedAt       time.Time `gorm:"autoCreateTime;index:idx_bot_created,priority:2" json:"created_at"`

	Bot Bot `gorm:"foreignKey:BotID" json:"-"`
}

// TableName specifies the database table name for the MessageLog model.
func (MessageLog) TableName() string {
	return "message_logs"
}
