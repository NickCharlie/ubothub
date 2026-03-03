package model

import (
	"time"

	"gorm.io/gorm"
)

// AvatarConfig represents a virtual avatar configuration entity.
type AvatarConfig struct {
	ID            string         `gorm:"type:char(26);primaryKey" json:"id"`
	UserID        string         `gorm:"type:char(26);index;not null;column:user_id" json:"user_id"`
	BotID         string         `gorm:"type:char(26);index;column:bot_id" json:"bot_id"`
	Name          string         `gorm:"type:varchar(128);not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	RenderType    string         `gorm:"type:varchar(16);not null;column:render_type" json:"render_type"`
	SceneConfig   string         `gorm:"type:jsonb;default:'{}';column:scene_config" json:"scene_config"`
	ActionMapping string         `gorm:"type:jsonb;default:'{}';column:action_mapping" json:"action_mapping"`
	IsDefault     bool           `gorm:"default:false;column:is_default" json:"is_default"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	User         User          `gorm:"foreignKey:UserID" json:"-"`
	Bot          *Bot          `gorm:"foreignKey:BotID" json:"-"`
	AvatarAssets []AvatarAsset `gorm:"foreignKey:AvatarID" json:"avatar_assets,omitempty"`
}

// TableName specifies the database table name for the AvatarConfig model.
func (AvatarConfig) TableName() string {
	return "avatar_configs"
}

// AvatarAsset represents the many-to-many relationship between avatars and assets.
type AvatarAsset struct {
	ID        string `gorm:"type:char(26);primaryKey" json:"id"`
	AvatarID  string `gorm:"type:char(26);not null;uniqueIndex:idx_avatar_asset;column:avatar_id" json:"avatar_id"`
	AssetID   string `gorm:"type:char(26);not null;uniqueIndex:idx_avatar_asset;column:asset_id" json:"asset_id"`
	Role      string `gorm:"type:varchar(32);not null" json:"role"`
	Config    string `gorm:"type:jsonb;default:'{}'" json:"config"`
	SortOrder int    `gorm:"default:0;column:sort_order" json:"sort_order"`

	Avatar AvatarConfig `gorm:"foreignKey:AvatarID;constraint:OnDelete:CASCADE" json:"-"`
	Asset  Asset        `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// TableName specifies the database table name for the AvatarAsset model.
func (AvatarAsset) TableName() string {
	return "avatar_assets"
}
