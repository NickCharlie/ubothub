package service

import (
	"context"
	"fmt"

	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"go.uber.org/zap"
)

// DashboardStats holds system-wide statistics for the admin dashboard.
type DashboardStats struct {
	UserCount        int64 `json:"user_count"`
	BotCount         int64 `json:"bot_count"`
	ActiveBotCount   int64 `json:"active_bot_count"`
	AssetCount       int64 `json:"asset_count"`
	StorageUsedBytes int64 `json:"storage_used_bytes"`
}

// AdminService handles admin business logic.
type AdminService struct {
	adminRepo repository.AdminRepository
	logger    *zap.Logger
}

// NewAdminService creates a new admin service.
func NewAdminService(adminRepo repository.AdminRepository, logger *zap.Logger) *AdminService {
	return &AdminService{adminRepo: adminRepo, logger: logger}
}

// GetDashboard returns system-wide statistics.
func (s *AdminService) GetDashboard(ctx context.Context) (*DashboardStats, error) {
	userCount, err := s.adminRepo.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	botCount, err := s.adminRepo.CountBots(ctx)
	if err != nil {
		return nil, fmt.Errorf("count bots: %w", err)
	}

	activeBots, err := s.adminRepo.CountActiveBots(ctx)
	if err != nil {
		return nil, fmt.Errorf("count active bots: %w", err)
	}

	assetCount, err := s.adminRepo.CountAssets(ctx)
	if err != nil {
		return nil, fmt.Errorf("count assets: %w", err)
	}

	storageUsed, err := s.adminRepo.SumAssetFileSize(ctx)
	if err != nil {
		return nil, fmt.Errorf("sum storage: %w", err)
	}

	return &DashboardStats{
		UserCount:        userCount,
		BotCount:         botCount,
		ActiveBotCount:   activeBots,
		AssetCount:       assetCount,
		StorageUsedBytes: storageUsed,
	}, nil
}

// ListUsers returns paginated users for admin management.
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int, status, role string) ([]*model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.adminRepo.ListUsers(ctx, offset, pageSize, status, role)
}

// BanUser sets a user's status to "banned".
func (s *AdminService) BanUser(ctx context.Context, userID string) error {
	if err := s.adminRepo.UpdateUserStatus(ctx, userID, "banned"); err != nil {
		return fmt.Errorf("ban user: %w", err)
	}
	s.logger.Info("user banned by admin", zap.String("user_id", userID))
	return nil
}

// UnbanUser restores a banned user's status to "active".
func (s *AdminService) UnbanUser(ctx context.Context, userID string) error {
	if err := s.adminRepo.UpdateUserStatus(ctx, userID, "active"); err != nil {
		return fmt.Errorf("unban user: %w", err)
	}
	s.logger.Info("user unbanned by admin", zap.String("user_id", userID))
	return nil
}

// ListBots returns paginated bots for admin management.
func (s *AdminService) ListBots(ctx context.Context, page, pageSize int, status, framework string) ([]*model.Bot, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.adminRepo.ListBots(ctx, offset, pageSize, status, framework)
}

// ForceDeleteBot permanently removes a bot (bypasses soft delete).
func (s *AdminService) ForceDeleteBot(ctx context.Context, botID string) error {
	if err := s.adminRepo.ForceDeleteBot(ctx, botID); err != nil {
		return fmt.Errorf("force delete bot: %w", err)
	}
	s.logger.Info("bot force deleted by admin", zap.String("bot_id", botID))
	return nil
}

// ListAssets returns paginated assets for admin management.
func (s *AdminService) ListAssets(ctx context.Context, page, pageSize int, category, status string) ([]*model.Asset, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.adminRepo.ListAssets(ctx, offset, pageSize, category, status)
}

// ListMessageLogs returns paginated message logs for admin auditing.
func (s *AdminService) ListMessageLogs(ctx context.Context, page, pageSize int, botID string) ([]*model.MessageLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.adminRepo.ListMessageLogs(ctx, offset, pageSize, botID)
}
