package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/NickCharlie/ubothub/backend/pkg/email"
	"go.uber.org/zap"
)

const (
	verifyTokenTTL = 24 * time.Hour
	resetTokenTTL  = 1 * time.Hour
	verifyKeyFmt   = "email:verify:%s"
	resetKeyFmt    = "email:reset:%s"
)

// EmailService handles email verification and password reset token management.
type EmailService struct {
	sender      *email.Sender
	rdb         *redis.Client
	frontendURL string
	logger      *zap.Logger
}

// NewEmailService creates a new email service.
func NewEmailService(sender *email.Sender, rdb *redis.Client, frontendURL string, logger *zap.Logger) *EmailService {
	return &EmailService{
		sender:      sender,
		rdb:         rdb,
		frontendURL: frontendURL,
		logger:      logger,
	}
}

// SendVerificationEmail generates a verification token and sends the email.
func (s *EmailService) SendVerificationEmail(ctx context.Context, userID, username, emailAddr string) error {
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("generate verification token: %w", err)
	}

	key := fmt.Sprintf(verifyKeyFmt, token)
	if err := s.rdb.Set(ctx, key, userID, verifyTokenTTL).Err(); err != nil {
		return fmt.Errorf("store verification token: %w", err)
	}

	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.frontendURL, token)
	subject, body := email.VerificationEmail(username, verifyURL)

	if err := s.sender.Send(emailAddr, subject, body); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	s.logger.Info("verification email sent",
		zap.String("user_id", userID),
		zap.String("email", emailAddr),
	)
	return nil
}

// VerifyEmail validates a verification token and returns the associated user ID.
// The token is consumed after verification (one-time use).
func (s *EmailService) VerifyEmail(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf(verifyKeyFmt, token)
	userID, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrTokenInvalid
		}
		return "", fmt.Errorf("get verification token: %w", err)
	}

	// Consume the token.
	s.rdb.Del(ctx, key)

	return userID, nil
}

// SendPasswordResetEmail generates a reset token and sends the email.
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, userID, username, emailAddr string) error {
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}

	key := fmt.Sprintf(resetKeyFmt, token)
	if err := s.rdb.Set(ctx, key, userID, resetTokenTTL).Err(); err != nil {
		return fmt.Errorf("store reset token: %w", err)
	}

	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.frontendURL, token)
	subject, body := email.PasswordResetEmail(username, resetURL)

	if err := s.sender.Send(emailAddr, subject, body); err != nil {
		return fmt.Errorf("send reset email: %w", err)
	}

	s.logger.Info("password reset email sent",
		zap.String("user_id", userID),
		zap.String("email", emailAddr),
	)
	return nil
}

// ValidateResetToken validates a password reset token and returns the user ID.
// The token is consumed after validation (one-time use).
func (s *EmailService) ValidateResetToken(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf(resetKeyFmt, token)
	userID, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrTokenInvalid
		}
		return "", fmt.Errorf("get reset token: %w", err)
	}

	// Consume the token.
	s.rdb.Del(ctx, key)

	return userID, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
