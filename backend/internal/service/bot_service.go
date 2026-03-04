package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/xid"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"github.com/NickCharlie/ubothub/backend/pkg/crypto"
	"github.com/NickCharlie/ubothub/backend/pkg/sanitize"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const maxBotsPerUser = 20

// sensitiveConfigKeys lists config fields that must be encrypted at rest.
var sensitiveConfigKeys = []string{"api_key"}

// BotService handles bot management business logic.
type BotService struct {
	botRepo    repository.BotRepository
	avatarRepo repository.AvatarRepository
	assetRepo  repository.AssetRepository
	encryptor  *crypto.Encryptor
	logger     *zap.Logger
}

// NewBotService creates a new bot service with optional config encryption.
func NewBotService(botRepo repository.BotRepository, avatarRepo repository.AvatarRepository, assetRepo repository.AssetRepository, encryptor *crypto.Encryptor, logger *zap.Logger) *BotService {
	return &BotService{
		botRepo:    botRepo,
		avatarRepo: avatarRepo,
		assetRepo:  assetRepo,
		encryptor:  encryptor,
		logger:     logger,
	}
}

// CreateBot creates a new bot with a generated access token.
func (s *BotService) CreateBot(ctx context.Context, userID, name, description, framework, webhookURL, configStr, visibility string) (*model.Bot, string, error) {
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

	if configStr == "" {
		configStr = "{}"
	}

	// Encrypt sensitive fields in config before storage.
	encryptedConfig, err := s.encryptConfig(sanitize.JSON(configStr))
	if err != nil {
		return nil, "", fmt.Errorf("encrypt config: %w", err)
	}

	if visibility == "" {
		visibility = model.BotVisibilityPrivate
	}

	bot := &model.Bot{
		ID:          xid.New().String(),
		UserID:      userID,
		Name:        sanitize.Text(name),
		Description: sanitize.Text(description),
		Framework:   framework,
		Visibility:  visibility,
		Status:      "offline",
		AccessToken: accessToken,
		WebhookURL:  webhookURL,
		Config:      encryptedConfig,
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

// GetBotByID returns a bot by ID without ownership verification.
// Used internally by the chat forwarder for message routing.
func (s *BotService) GetBotByID(ctx context.Context, botID string) (*model.Bot, error) {
	bot, err := s.botRepo.FindByID(ctx, botID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBotNotFound
		}
		return nil, fmt.Errorf("find bot: %w", err)
	}
	return bot, nil
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
func (s *BotService) UpdateBot(ctx context.Context, botID, userID, name, description, webhookURL, configStr, visibility string) (*model.Bot, error) {
	bot, err := s.GetBot(ctx, botID, userID)
	if err != nil {
		return nil, err
	}

	if name != "" {
		bot.Name = sanitize.Text(name)
	}
	if description != "" {
		bot.Description = sanitize.Text(description)
	}
	if webhookURL != "" {
		bot.WebhookURL = webhookURL
	}
	if configStr != "" {
		// Merge new config with existing: decrypt old → merge → encrypt.
		merged, err := s.mergeConfig(bot.Config, sanitize.JSON(configStr))
		if err != nil {
			return nil, fmt.Errorf("merge config: %w", err)
		}
		bot.Config = merged
	}
	if visibility != "" {
		bot.Visibility = visibility
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

// DecryptBotConfig returns the bot config with sensitive fields decrypted.
// Used internally by the gateway when sending outbound messages.
func (s *BotService) DecryptBotConfig(configStr string) (map[string]interface{}, error) {
	return s.decryptConfigMap(configStr)
}

// MaskBotConfig returns the bot config with sensitive fields masked for API responses.
// Owner sees masked values (e.g., "***abk_"); non-owner sees no sensitive fields.
func (s *BotService) MaskBotConfig(configStr string, isOwner bool) string {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &raw); err != nil {
		return "{}"
	}

	for _, key := range sensitiveConfigKeys {
		val, ok := raw[key]
		if !ok {
			continue
		}
		strVal, ok := val.(string)
		if !ok {
			continue
		}
		if !isOwner {
			delete(raw, key)
			continue
		}
		// For owner: decrypt then mask.
		if s.encryptor != nil {
			decrypted, err := s.encryptor.Decrypt(strVal)
			if err == nil {
				strVal = decrypted
			}
		}
		raw[key] = crypto.MaskSecret(strVal)
	}

	out, err := json.Marshal(raw)
	if err != nil {
		return "{}"
	}
	return string(out)
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

// encryptConfig encrypts sensitive fields in a JSON config string.
func (s *BotService) encryptConfig(configStr string) (string, error) {
	if s.encryptor == nil {
		return configStr, nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &raw); err != nil {
		return configStr, nil
	}

	for _, key := range sensitiveConfigKeys {
		val, ok := raw[key]
		if !ok {
			continue
		}
		strVal, ok := val.(string)
		if !ok || strVal == "" {
			continue
		}
		encrypted, err := s.encryptor.Encrypt(strVal)
		if err != nil {
			return "", fmt.Errorf("encrypt field %s: %w", key, err)
		}
		raw[key] = encrypted
	}

	out, err := json.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}
	return string(out), nil
}

// decryptConfigMap returns the config as a map with sensitive fields decrypted.
func (s *BotService) decryptConfigMap(configStr string) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &raw); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if s.encryptor == nil {
		return raw, nil
	}

	for _, key := range sensitiveConfigKeys {
		val, ok := raw[key]
		if !ok {
			continue
		}
		strVal, ok := val.(string)
		if !ok || strVal == "" {
			continue
		}
		decrypted, err := s.encryptor.Decrypt(strVal)
		if err != nil {
			// If decryption fails, the value may be plaintext (pre-encryption migration).
			continue
		}
		raw[key] = decrypted
	}

	return raw, nil
}

// mergeConfig decrypts existing config, overlays new values (encrypting sensitive
// ones), and returns the merged encrypted JSON string.
func (s *BotService) mergeConfig(existingConfig, newConfigStr string) (string, error) {
	// Decrypt existing config.
	existing, err := s.decryptConfigMap(existingConfig)
	if err != nil {
		existing = make(map[string]interface{})
	}

	// Parse new config.
	var incoming map[string]interface{}
	if err := json.Unmarshal([]byte(newConfigStr), &incoming); err != nil {
		return existingConfig, nil
	}

	// Overlay: skip empty string values in incoming (means "keep existing").
	// Also skip masked values (***xxx) that the frontend may send back.
	for k, v := range incoming {
		strVal, isStr := v.(string)
		if isStr && (strVal == "" || strings.HasPrefix(strVal, "***")) {
			continue
		}
		existing[k] = v
	}

	// Re-marshal and encrypt.
	merged, err := json.Marshal(existing)
	if err != nil {
		return existingConfig, nil
	}
	return s.encryptConfig(string(merged))
}

// SetupAvatar performs one-click avatar setup for a bot:
// 1. Creates a new avatar (or reuses the bot's existing avatar)
// 2. Binds the specified 3D model asset as primary_model
// 3. Links the avatar to the bot
// 4. Updates the bot config with avatar_id
func (s *BotService) SetupAvatar(ctx context.Context, botID, userID, assetID, avatarName string) (*model.AvatarConfig, error) {
	// 1. Verify bot ownership.
	bot, err := s.GetBot(ctx, botID, userID)
	if err != nil {
		return nil, err
	}

	// 2. Verify asset exists and belongs to user.
	asset, err := s.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("find asset: %w", err)
	}
	if asset.UserID != userID {
		return nil, ErrAssetNotFound
	}

	// Determine render type from asset category.
	renderType := "3d"
	if asset.Category == "model_live2d" {
		renderType = "live2d"
	}

	if avatarName == "" {
		avatarName = bot.Name + "'s Avatar"
	}

	// 3. Check if bot already has a bound avatar.
	var avatar *model.AvatarConfig
	existing, err := s.avatarRepo.FindByBotID(ctx, botID)
	if err == nil && existing != nil {
		// Reuse existing avatar — load it with assets to find old primary_model bindings.
		avatar, _ = s.avatarRepo.FindByIDWithAssets(ctx, existing.ID)
		if avatar == nil {
			avatar = existing
		}
		avatar.Name = sanitize.Text(avatarName)
		avatar.RenderType = renderType
		if err := s.avatarRepo.Update(ctx, avatar); err != nil {
			return nil, fmt.Errorf("update avatar: %w", err)
		}
		// Remove existing primary_model bindings.
		for _, aa := range avatar.AvatarAssets {
			if aa.Role == "primary_model" {
				_ = s.avatarRepo.DeleteAvatarAsset(ctx, avatar.ID, aa.AssetID)
			}
		}
	} else {
		// Create new avatar.
		avatar = &model.AvatarConfig{
			ID:            xid.New().String(),
			UserID:        userID,
			BotID:         &botID,
			Name:          sanitize.Text(avatarName),
			RenderType:    renderType,
			SceneConfig:   "{}",
			ActionMapping: "{}",
		}
		if err := s.avatarRepo.Create(ctx, avatar); err != nil {
			return nil, fmt.Errorf("create avatar: %w", err)
		}
	}

	// 4. Bind asset to avatar as primary_model.
	aa := &model.AvatarAsset{
		ID:        xid.New().String(),
		AvatarID:  avatar.ID,
		AssetID:   assetID,
		Role:      "primary_model",
		Config:    "{}",
		SortOrder: 0,
	}
	if err := s.avatarRepo.CreateAvatarAsset(ctx, aa); err != nil {
		return nil, fmt.Errorf("bind asset: %w", err)
	}

	// 5. Update bot config with avatar_id.
	configMap, _ := s.decryptConfigMap(bot.Config)
	if configMap == nil {
		configMap = make(map[string]interface{})
	}
	configMap["avatar_id"] = avatar.ID
	configJSON, err := json.Marshal(configMap)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	encryptedConfig, err := s.encryptConfig(string(configJSON))
	if err != nil {
		return nil, fmt.Errorf("encrypt config: %w", err)
	}
	bot.Config = encryptedConfig
	if err := s.botRepo.Update(ctx, bot); err != nil {
		return nil, fmt.Errorf("update bot config: %w", err)
	}

	// Reload avatar with assets for response.
	avatar, _ = s.avatarRepo.FindByIDWithAssets(ctx, avatar.ID)

	s.logger.Info("bot avatar setup complete",
		zap.String("bot_id", botID),
		zap.String("avatar_id", avatar.ID),
		zap.String("asset_id", assetID),
	)
	return avatar, nil
}

// RemoveAvatar removes the avatar link from a bot and cleans up the avatar config.
func (s *BotService) RemoveAvatar(ctx context.Context, botID, userID string) error {
	bot, err := s.GetBot(ctx, botID, userID)
	if err != nil {
		return err
	}

	// Find and delete the avatar bound to this bot.
	avatar, err := s.avatarRepo.FindByBotID(ctx, botID)
	if err != nil {
		return ErrAvatarNotFound
	}
	if avatar.UserID != userID {
		return ErrAvatarNotFound
	}

	if err := s.avatarRepo.Delete(ctx, avatar.ID); err != nil {
		return fmt.Errorf("delete avatar: %w", err)
	}

	// Remove avatar_id from bot config.
	configMap, _ := s.decryptConfigMap(bot.Config)
	if configMap != nil {
		delete(configMap, "avatar_id")
		configJSON, _ := json.Marshal(configMap)
		encryptedConfig, _ := s.encryptConfig(string(configJSON))
		bot.Config = encryptedConfig
		_ = s.botRepo.Update(ctx, bot)
	}

	s.logger.Info("bot avatar removed",
		zap.String("bot_id", botID),
		zap.String("avatar_id", avatar.ID),
	)
	return nil
}

// GetBotAvatar returns the avatar bound to a bot (if any).
func (s *BotService) GetBotAvatar(ctx context.Context, botID, userID string) (*model.AvatarConfig, error) {
	_, err := s.GetBot(ctx, botID, userID)
	if err != nil {
		return nil, err
	}

	avatar, err := s.avatarRepo.FindByBotID(ctx, botID)
	if err != nil {
		return nil, ErrAvatarNotFound
	}

	// Reload with assets for full response.
	full, err := s.avatarRepo.FindByIDWithAssets(ctx, avatar.ID)
	if err != nil {
		return avatar, nil
	}
	return full, nil
}

// generateAccessToken creates a cryptographically random 32-byte hex token.
func generateAccessToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
