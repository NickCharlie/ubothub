package service

import (
	"context"
	"fmt"

	"github.com/ubothub/backend/internal/model"
	"github.com/ubothub/backend/internal/repository"
	"github.com/ubothub/backend/pkg/hash"
	"go.uber.org/zap"
)

// UserService handles user profile business logic.
type UserService struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

// NewUserService creates a new user service.
func NewUserService(userRepo repository.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetProfile returns the user profile for the given ID.
func (s *UserService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}

// UpdateProfile updates the user's display name and avatar URL.
func (s *UserService) UpdateProfile(ctx context.Context, userID, displayName, avatarURL string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if displayName != "" {
		user.DisplayName = displayName
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	s.logger.Info("user profile updated", zap.String("user_id", userID))
	return user, nil
}

// ChangePassword verifies the old password and updates to the new one.
func (s *UserService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	if !hash.CheckPassword(oldPassword, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	passwordHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user.PasswordHash = passwordHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	s.logger.Info("user password changed", zap.String("user_id", userID))
	return nil
}
