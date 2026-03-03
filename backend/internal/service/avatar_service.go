package service

import (
	"context"
	"fmt"

	"github.com/rs/xid"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"github.com/NickCharlie/ubothub/backend/pkg/sanitize"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AvatarService handles avatar configuration business logic.
type AvatarService struct {
	avatarRepo repository.AvatarRepository
	botRepo    repository.BotRepository
	logger     *zap.Logger
}

// NewAvatarService creates a new avatar service.
func NewAvatarService(avatarRepo repository.AvatarRepository, botRepo repository.BotRepository, logger *zap.Logger) *AvatarService {
	return &AvatarService{
		avatarRepo: avatarRepo,
		botRepo:    botRepo,
		logger:     logger,
	}
}

// CreateAvatar creates a new avatar configuration.
func (s *AvatarService) CreateAvatar(ctx context.Context, userID, name, description, renderType, sceneConfig, actionMapping string) (*model.AvatarConfig, error) {
	if sceneConfig == "" {
		sceneConfig = "{}"
	}
	if actionMapping == "" {
		actionMapping = "{}"
	}

	avatar := &model.AvatarConfig{
		ID:            xid.New().String(),
		UserID:        userID,
		Name:          sanitize.Text(name),
		Description:   sanitize.Text(description),
		RenderType:    renderType,
		SceneConfig:   sanitize.JSON(sceneConfig),
		ActionMapping: sanitize.JSON(actionMapping),
	}

	if err := s.avatarRepo.Create(ctx, avatar); err != nil {
		return nil, fmt.Errorf("create avatar: %w", err)
	}

	s.logger.Info("avatar created",
		zap.String("avatar_id", avatar.ID),
		zap.String("user_id", userID),
		zap.String("render_type", renderType),
	)

	return avatar, nil
}

// GetAvatar returns an avatar by ID with assets, verifying ownership.
func (s *AvatarService) GetAvatar(ctx context.Context, avatarID, userID string) (*model.AvatarConfig, error) {
	avatar, err := s.avatarRepo.FindByIDWithAssets(ctx, avatarID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAvatarNotFound
		}
		return nil, fmt.Errorf("find avatar: %w", err)
	}

	if avatar.UserID != userID {
		return nil, ErrAvatarNotFound
	}

	return avatar, nil
}

// ListAvatars returns paginated avatars for a user.
func (s *AvatarService) ListAvatars(ctx context.Context, userID string, page, pageSize int) ([]*model.AvatarConfig, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.avatarRepo.FindByUserID(ctx, userID, offset, pageSize)
}

// UpdateAvatar updates avatar configuration fields.
func (s *AvatarService) UpdateAvatar(ctx context.Context, avatarID, userID, name, description, sceneConfig, actionMapping string) (*model.AvatarConfig, error) {
	avatar, err := s.avatarRepo.FindByID(ctx, avatarID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAvatarNotFound
		}
		return nil, fmt.Errorf("find avatar: %w", err)
	}

	if avatar.UserID != userID {
		return nil, ErrAvatarNotFound
	}

	if name != "" {
		avatar.Name = sanitize.Text(name)
	}
	if description != "" {
		avatar.Description = sanitize.Text(description)
	}
	if sceneConfig != "" {
		avatar.SceneConfig = sanitize.JSON(sceneConfig)
	}
	if actionMapping != "" {
		avatar.ActionMapping = sanitize.JSON(actionMapping)
	}

	if err := s.avatarRepo.Update(ctx, avatar); err != nil {
		return nil, fmt.Errorf("update avatar: %w", err)
	}

	s.logger.Info("avatar updated", zap.String("avatar_id", avatarID))
	return avatar, nil
}

// DeleteAvatar deletes an avatar configuration.
func (s *AvatarService) DeleteAvatar(ctx context.Context, avatarID, userID string) error {
	avatar, err := s.avatarRepo.FindByID(ctx, avatarID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAvatarNotFound
		}
		return fmt.Errorf("find avatar: %w", err)
	}

	if avatar.UserID != userID {
		return ErrAvatarNotFound
	}

	if err := s.avatarRepo.Delete(ctx, avatarID); err != nil {
		return fmt.Errorf("delete avatar: %w", err)
	}

	s.logger.Info("avatar deleted", zap.String("avatar_id", avatarID), zap.String("user_id", userID))
	return nil
}

// BindBot binds a bot to an avatar.
func (s *AvatarService) BindBot(ctx context.Context, avatarID, userID, botID string) error {
	avatar, err := s.avatarRepo.FindByID(ctx, avatarID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAvatarNotFound
		}
		return fmt.Errorf("find avatar: %w", err)
	}

	if avatar.UserID != userID {
		return ErrAvatarNotFound
	}

	// Verify bot ownership.
	bot, err := s.botRepo.FindByID(ctx, botID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrBotNotFound
		}
		return fmt.Errorf("find bot: %w", err)
	}
	if bot.UserID != userID {
		return ErrBotNotFound
	}

	// Check if bot is already bound to another avatar.
	existing, err := s.avatarRepo.FindByBotID(ctx, botID)
	if err == nil && existing.ID != avatarID {
		return ErrAvatarBotConflict
	}

	avatar.BotID = botID
	if err := s.avatarRepo.Update(ctx, avatar); err != nil {
		return fmt.Errorf("bind bot: %w", err)
	}

	s.logger.Info("bot bound to avatar",
		zap.String("avatar_id", avatarID),
		zap.String("bot_id", botID),
	)
	return nil
}

// BindAsset binds an asset to an avatar.
func (s *AvatarService) BindAsset(ctx context.Context, avatarID, userID, assetID, role, config string, sortOrder int) error {
	avatar, err := s.avatarRepo.FindByID(ctx, avatarID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAvatarNotFound
		}
		return fmt.Errorf("find avatar: %w", err)
	}

	if avatar.UserID != userID {
		return ErrAvatarNotFound
	}

	if config == "" {
		config = "{}"
	}

	aa := &model.AvatarAsset{
		ID:        xid.New().String(),
		AvatarID:  avatarID,
		AssetID:   assetID,
		Role:      sanitize.Text(role),
		Config:    sanitize.JSON(config),
		SortOrder: sortOrder,
	}

	if err := s.avatarRepo.CreateAvatarAsset(ctx, aa); err != nil {
		return fmt.Errorf("bind asset: %w", err)
	}

	s.logger.Info("asset bound to avatar",
		zap.String("avatar_id", avatarID),
		zap.String("asset_id", assetID),
		zap.String("role", role),
	)
	return nil
}

// UnbindAsset removes an asset binding from an avatar.
func (s *AvatarService) UnbindAsset(ctx context.Context, avatarID, userID, assetID string) error {
	avatar, err := s.avatarRepo.FindByID(ctx, avatarID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAvatarNotFound
		}
		return fmt.Errorf("find avatar: %w", err)
	}

	if avatar.UserID != userID {
		return ErrAvatarNotFound
	}

	if err := s.avatarRepo.DeleteAvatarAsset(ctx, avatarID, assetID); err != nil {
		return fmt.Errorf("unbind asset: %w", err)
	}

	return nil
}

// UpdateActionMapping updates the action mapping for an avatar.
func (s *AvatarService) UpdateActionMapping(ctx context.Context, avatarID, userID, actionMapping string) error {
	avatar, err := s.avatarRepo.FindByID(ctx, avatarID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAvatarNotFound
		}
		return fmt.Errorf("find avatar: %w", err)
	}

	if avatar.UserID != userID {
		return ErrAvatarNotFound
	}

	avatar.ActionMapping = sanitize.JSON(actionMapping)
	if err := s.avatarRepo.Update(ctx, avatar); err != nil {
		return fmt.Errorf("update action mapping: %w", err)
	}

	return nil
}

// GetAvatarPublic returns an avatar with assets for public preview (no ownership check).
// Only returns avatars that are bound to a public bot.
func (s *AvatarService) GetAvatarPublic(ctx context.Context, avatarID string) (*model.AvatarConfig, error) {
	avatar, err := s.avatarRepo.FindByIDWithAssets(ctx, avatarID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAvatarNotFound
		}
		return nil, fmt.Errorf("find avatar: %w", err)
	}

	// Only allow viewing if the avatar's bot is public.
	if avatar.BotID != "" {
		bot, err := s.botRepo.FindByID(ctx, avatar.BotID)
		if err != nil || bot.Visibility != "public" {
			return nil, ErrAvatarNotFound
		}
	} else {
		return nil, ErrAvatarNotFound
	}

	return avatar, nil
}
