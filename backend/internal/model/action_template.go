package model

import (
	"time"
)

// ActionTemplate represents a predefined action mapping rule template.
type ActionTemplate struct {
	ID            string    `gorm:"type:char(26);primaryKey" json:"id"`
	Name          string    `gorm:"type:varchar(128);not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	TriggerType   string    `gorm:"type:varchar(32);not null;column:trigger_type" json:"trigger_type"`
	TriggerConfig string    `gorm:"type:jsonb;not null;column:trigger_config" json:"trigger_config"`
	ActionConfig  string    `gorm:"type:jsonb;not null;column:action_config" json:"action_config"`
	IsSystem      bool      `gorm:"default:false;column:is_system" json:"is_system"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the database table name for the ActionTemplate model.
func (ActionTemplate) TableName() string {
	return "action_templates"
}
