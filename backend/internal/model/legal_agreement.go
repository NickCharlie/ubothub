package model

import (
	"time"
)

// LegalAgreement represents a versioned legal document (terms of service, privacy policy, etc.).
type LegalAgreement struct {
	ID        string    `gorm:"type:char(26);primaryKey" json:"id"`
	Type      string    `gorm:"type:varchar(32);not null;index:idx_agreement_type_version" json:"type"`
	Version   string    `gorm:"type:varchar(16);not null;index:idx_agreement_type_version" json:"version"`
	Locale    string    `gorm:"type:varchar(8);not null;index" json:"locale"`
	Title     string    `gorm:"type:varchar(256);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsActive  bool      `gorm:"default:false;not null" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the database table name.
func (LegalAgreement) TableName() string {
	return "legal_agreements"
}

// Agreement types.
const (
	AgreementTypeTerms   = "terms_of_service"
	AgreementTypePrivacy = "privacy_policy"
)

// UserAgreementAcceptance records a user's acceptance of a specific agreement version.
type UserAgreementAcceptance struct {
	ID          string    `gorm:"type:char(26);primaryKey" json:"id"`
	UserID      string    `gorm:"type:char(26);not null;index:idx_user_agreement,unique" json:"user_id"`
	AgreementID string    `gorm:"type:char(26);not null;index:idx_user_agreement,unique" json:"agreement_id"`
	AcceptedAt  time.Time `gorm:"autoCreateTime" json:"accepted_at"`
	IPAddress   string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent   string    `gorm:"type:varchar(512)" json:"user_agent"`

	Agreement LegalAgreement `gorm:"foreignKey:AgreementID" json:"-"`
}

// TableName specifies the database table name.
func (UserAgreementAcceptance) TableName() string {
	return "user_agreement_acceptances"
}
