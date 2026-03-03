package model

import (
	"time"

	"github.com/lib/pq"
)

// Asset represents an uploaded model or motion file entity.
type Asset struct {
	ID            string         `gorm:"type:char(26);primaryKey" json:"id"`
	UserID        string         `gorm:"type:char(26);index;not null;column:user_id" json:"user_id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	Category      string         `gorm:"type:varchar(32);not null;index" json:"category"`
	Format        string         `gorm:"type:varchar(16);not null;index" json:"format"`
	FileKey       string         `gorm:"type:varchar(512);not null;column:file_key" json:"file_key"`
	FileSize      int64          `gorm:"not null;column:file_size" json:"file_size"`
	ThumbnailKey  string         `gorm:"type:varchar(512);column:thumbnail_key" json:"thumbnail_key"`
	Metadata      string         `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	Tags          pq.StringArray `gorm:"type:text[];default:'{}'" json:"tags"`
	IsPublic      bool           `gorm:"default:false;column:is_public;index" json:"is_public"`
	DownloadCount int            `gorm:"default:0;column:download_count" json:"download_count"`
	Status        string         `gorm:"type:varchar(16);default:processing;not null" json:"status"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the database table name for the Asset model.
func (Asset) TableName() string {
	return "assets"
}
