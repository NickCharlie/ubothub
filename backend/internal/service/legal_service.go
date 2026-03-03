package service

import (
	"context"
	"fmt"

	"github.com/rs/xid"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LegalService handles legal agreement business logic.
type LegalService struct {
	legalRepo repository.LegalRepository
	logger    *zap.Logger
}

// NewLegalService creates a new legal service.
func NewLegalService(legalRepo repository.LegalRepository, logger *zap.Logger) *LegalService {
	return &LegalService{
		legalRepo: legalRepo,
		logger:    logger,
	}
}

// GetActiveAgreement returns the current active agreement for a given type and locale.
// Falls back to "en" locale if the requested locale is not found.
func (s *LegalService) GetActiveAgreement(ctx context.Context, agreementType, locale string) (*model.LegalAgreement, error) {
	agreement, err := s.legalRepo.FindActiveByTypeAndLocale(ctx, agreementType, locale)
	if err != nil {
		if err == gorm.ErrRecordNotFound && locale != "en" {
			return s.legalRepo.FindActiveByTypeAndLocale(ctx, agreementType, "en")
		}
		return nil, fmt.Errorf("find agreement: %w", err)
	}
	return agreement, nil
}

// GetAllActiveAgreements returns all active agreements of a given type (all locales).
func (s *LegalService) GetAllActiveAgreements(ctx context.Context, agreementType string) ([]model.LegalAgreement, error) {
	return s.legalRepo.FindActiveByType(ctx, agreementType)
}

// RecordAcceptance records that a user has accepted a specific agreement.
func (s *LegalService) RecordAcceptance(ctx context.Context, userID, agreementID, ipAddress, userAgent string) error {
	accepted, err := s.legalRepo.HasAccepted(ctx, userID, agreementID)
	if err != nil {
		return fmt.Errorf("check acceptance: %w", err)
	}
	if accepted {
		return nil
	}

	acceptance := &model.UserAgreementAcceptance{
		ID:          xid.New().String(),
		UserID:      userID,
		AgreementID: agreementID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}

	if err := s.legalRepo.CreateAcceptance(ctx, acceptance); err != nil {
		return fmt.Errorf("record acceptance: %w", err)
	}

	s.logger.Info("agreement accepted",
		zap.String("user_id", userID),
		zap.String("agreement_id", agreementID),
	)
	return nil
}

// GetUserAcceptances returns all agreement acceptances for a user.
func (s *LegalService) GetUserAcceptances(ctx context.Context, userID string) ([]model.UserAgreementAcceptance, error) {
	return s.legalRepo.FindAcceptancesByUser(ctx, userID)
}

// SeedDefaultAgreements creates default agreement documents if none exist.
func (s *LegalService) SeedDefaultAgreements(ctx context.Context) error {
	agreements := []model.LegalAgreement{
		// Terms of Service - Chinese
		{
			ID:       xid.New().String(),
			Type:     model.AgreementTypeTerms,
			Version:  "1.0.0",
			Locale:   "zh",
			Title:    "UBotHub 用户服务协议",
			Content:  termsOfServiceZH,
			IsActive: true,
		},
		// Terms of Service - English
		{
			ID:       xid.New().String(),
			Type:     model.AgreementTypeTerms,
			Version:  "1.0.0",
			Locale:   "en",
			Title:    "UBotHub Terms of Service",
			Content:  termsOfServiceEN,
			IsActive: true,
		},
		// Privacy Policy - Chinese
		{
			ID:       xid.New().String(),
			Type:     model.AgreementTypePrivacy,
			Version:  "1.0.0",
			Locale:   "zh",
			Title:    "UBotHub 隐私政策",
			Content:  privacyPolicyZH,
			IsActive: true,
		},
		// Privacy Policy - English
		{
			ID:       xid.New().String(),
			Type:     model.AgreementTypePrivacy,
			Version:  "1.0.0",
			Locale:   "en",
			Title:    "UBotHub Privacy Policy",
			Content:  privacyPolicyEN,
			IsActive: true,
		},
	}

	for i := range agreements {
		existing, err := s.legalRepo.FindActiveByTypeAndLocale(ctx, agreements[i].Type, agreements[i].Locale)
		if err == nil && existing != nil {
			continue
		}
		if err := s.legalRepo.Create(ctx, &agreements[i]); err != nil {
			return fmt.Errorf("seed agreement %s/%s: %w", agreements[i].Type, agreements[i].Locale, err)
		}
		s.logger.Info("seeded agreement",
			zap.String("type", agreements[i].Type),
			zap.String("locale", agreements[i].Locale),
			zap.String("version", agreements[i].Version),
		)
	}
	return nil
}
