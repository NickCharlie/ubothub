package repository

import (
	"context"

	"github.com/NickCharlie/ubothub/backend/internal/model"
	"gorm.io/gorm"
)

// LegalRepository defines the data access interface for legal agreement entities.
type LegalRepository interface {
	FindActiveByTypeAndLocale(ctx context.Context, agreementType, locale string) (*model.LegalAgreement, error)
	FindActiveByType(ctx context.Context, agreementType string) ([]model.LegalAgreement, error)
	Create(ctx context.Context, agreement *model.LegalAgreement) error
	CreateAcceptance(ctx context.Context, acceptance *model.UserAgreementAcceptance) error
	HasAccepted(ctx context.Context, userID, agreementID string) (bool, error)
	FindAcceptancesByUser(ctx context.Context, userID string) ([]model.UserAgreementAcceptance, error)
}

type legalRepository struct {
	db *gorm.DB
}

// NewLegalRepository creates a new GORM-backed legal repository.
func NewLegalRepository(db *gorm.DB) LegalRepository {
	return &legalRepository{db: db}
}

func (r *legalRepository) FindActiveByTypeAndLocale(ctx context.Context, agreementType, locale string) (*model.LegalAgreement, error) {
	var agreement model.LegalAgreement
	err := r.db.WithContext(ctx).
		Where("type = ? AND locale = ? AND is_active = ?", agreementType, locale, true).
		Order("created_at DESC").
		First(&agreement).Error
	if err != nil {
		return nil, err
	}
	return &agreement, nil
}

func (r *legalRepository) FindActiveByType(ctx context.Context, agreementType string) ([]model.LegalAgreement, error) {
	var agreements []model.LegalAgreement
	err := r.db.WithContext(ctx).
		Where("type = ? AND is_active = ?", agreementType, true).
		Order("locale ASC").
		Find(&agreements).Error
	return agreements, err
}

func (r *legalRepository) Create(ctx context.Context, agreement *model.LegalAgreement) error {
	return r.db.WithContext(ctx).Create(agreement).Error
}

func (r *legalRepository) CreateAcceptance(ctx context.Context, acceptance *model.UserAgreementAcceptance) error {
	return r.db.WithContext(ctx).Create(acceptance).Error
}

func (r *legalRepository) HasAccepted(ctx context.Context, userID, agreementID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserAgreementAcceptance{}).
		Where("user_id = ? AND agreement_id = ?", userID, agreementID).
		Count(&count).Error
	return count > 0, err
}

func (r *legalRepository) FindAcceptancesByUser(ctx context.Context, userID string) ([]model.UserAgreementAcceptance, error) {
	var acceptances []model.UserAgreementAcceptance
	err := r.db.WithContext(ctx).
		Preload("Agreement").
		Where("user_id = ?", userID).
		Order("accepted_at DESC").
		Find(&acceptances).Error
	return acceptances, err
}
