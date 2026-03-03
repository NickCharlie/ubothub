package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a platform user entity.
type User struct {
	ID            string         `gorm:"type:char(26);primaryKey" json:"id"`
	Email         string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Username      string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	PasswordHash  string         `gorm:"type:varchar(255);not null;column:password_hash" json:"-"`
	DisplayName   string         `gorm:"type:varchar(128)" json:"display_name"`
	AvatarURL     string         `gorm:"type:text;column:avatar_url" json:"avatar_url"`
	EmailVerified bool           `gorm:"default:false;column:email_verified" json:"email_verified"`
	Role          string         `gorm:"type:varchar(16);default:user;not null" json:"role"`
	Status        string         `gorm:"type:varchar(16);default:active;not null" json:"status"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the database table name for the User model.
func (User) TableName() string {
	return "users"
}

// IsAdmin checks whether the user has admin role.
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}
