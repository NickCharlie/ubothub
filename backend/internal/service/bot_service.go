package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/rs/xid"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const maxBotsPerUser = 20

// BotService handles bot management business logic.
type BotService struct {
	botRepo repository.BotRepository
	logger  *zap.Logger
}

// NewBotService creates a new bot service.
func NewBotService(botRepo repository.BotRepository, logger *zap.Logger) *BotService {
	return &BotService{
		botRepo: botRepo,
		logger:  logger,
	}
}

// CreateBot creates a new bot with a generated access token.
func (s *BotService) CreateBot(ctx context.Context, userID, name, description, framework, webhookURL, config string) (*model.Bot, string, error) {
	count, err := s.botRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("count user bots: %w", err)
	}
	if count >= maxBotsPerUser {
		return nil, "", ErrBotLimitExceeded
	}

	accessToken, err := generateAccessToken()
	if err != nil {
		return nil, "", fmt.Errorf("generate access token: %w", err)
	}

	if config == "" {
		config = "{}"
	}

	bot := &model.Bot{
		ID:          xid.New().String(),
		UserID:      userID,
		Name:        name,
		Description: description,
		Framework:   framework,
		Status:      "offline",
		AccessToken: accessToken,
		WebhookURL:  webhookURL,
		Config:      config,
	}

	if err := s.botRepo.Create(ctx, bot); err != nil {
		return nil, "", fmt.Errorf("create bot: %w", err)
	}

	s.logger.Info("bot created",
		zap.String("bot_id", bot.ID),
		zap.String("user_id", userID),
		zap.String("framework", framework),
	)

	return bot, accessToken, nil
}

// GetBot returns a bot by ID, verifying ownership.
func (s *BotService) GetBot(ctx context.Context, botID, userID string) (*model.Bot, error) {
	bot, err := s.botRepo.FindByID(ctx, botID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBotNotFound
		}
		return nil, fmt.Errorf("find bot: %w", err)
	}

	if bot.UserID != userID {
		return nil, ErrBotNotFound
	}

	return bot, nil
}

// ListBots returns paginated bots for a user.
func (s *BotService) ListBots(ctx context.Context, userID string, page, pageSize int) ([]*model.Bot, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	bots, total, err := s.botRepo.FindByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list bots: %w", err)
	}

	return bots, total, nil
}

// UpdateBot updates bot fields for the given bot, verifying ownership.
func (s *BotService) UpdateBot(ctx context.Context, botID, userID, name, description, webhookURL, config string) (*model.Bot, error) {
	bot, err := s.GetBot(ctx, botID, userID)
	if err != nil {
		return nil, err
	}

	if name != "" {
		bot.Name = name
	}
	if description != "" {
		bot.Description = description
	}
	if webhookURL != "" {
		bot.WebhookURL = webhookURL
	}
	if config != "" {
		bot.Config = config
	}

	if err := s.botRepo.Update(ctx, bot); err != nil {
		return nil, fmt.Errorf("update bot: %w", err)
	}

	s.logger.Info("bot updated", zap.String("bot_id", botID), zap.String("user_id", userID))
	return bot, nil
}

// DeleteBot deletes a bot, verifying ownership.
func (s *BotService) DeleteBot(ctx context.Context, botID, userID string) error {
	bot, err := s.GetBot(ctx, botID, userID)
	if err != nil {
		return err
	}

	if err := s.botRepo.Delete(ctx, bot.ID); err != nil {
		return fmt.Errorf("delete bot: %w", err)
	}

	s.logger.Info("bot deleted", zap.String("bot_id", botID), zap.String("user_id", userID))
	return nil
}

// RegenerateToken creates a new access token for the bot.
func (s *BotService) RegenerateToken(ctx context.Context, botID, userID string) (string, error) {
	bot, err := s.GetBot(ctx, botID, userID)
	if err != nil {
		return "", err
	}

	newToken, err := generateAccessToken()
	if err != nil {
		return "", fmt.Errorf("generate access token: %w", err)
	}

	bot.AccessToken = newToken
	if err := s.botRepo.Update(ctx, bot); err != nil {
		return "", fmt.Errorf("update bot token: %w", err)
	}

	s.logger.Info("bot token regenerated", zap.String("bot_id", botID), zap.String("user_id", userID))
	return newToken, nil
}

// GetBotByAccessToken finds a bot by its access token (used by gateway).
func (s *BotService) GetBotByAccessToken(ctx context.Context, token string) (*model.Bot, error) {
	bot, err := s.botRepo.FindByAccessToken(ctx, token)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBotNotFound
		}
		return nil, fmt.Errorf("find bot by token: %w", err)
	}
	return bot, nil
}

// ListPublicBots returns paginated public bots for the plaza.
func (s *BotService) ListPublicBots(ctx context.Context, page, pageSize int) ([]*model.Bot, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.botRepo.FindPublic(ctx, offset, pageSize)
}

// GetPublicBot returns a public bot by ID (no ownership check).
func (s *BotService) GetPublicBot(ctx context.Context, botID string) (*model.Bot, error) {
	bot, err := s.botRepo.FindPublicByID(ctx, botID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBotNotFound
		}
		return nil, fmt.Errorf("find public bot: %w", err)
	}
	return bot, nil
}

// generateAccessToken creates a cryptographically random 32-byte hex token.
func generateAccessToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
