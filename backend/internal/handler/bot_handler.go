package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/NickCharlie/ubothub/backend/internal/dto/request"
	"github.com/NickCharlie/ubothub/backend/internal/dto/response"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/pkg/errcode"
	resp "github.com/NickCharlie/ubothub/backend/pkg/response"
)

// BotHandler handles bot management HTTP endpoints.
type BotHandler struct {
	botSvc *service.BotService
}

// NewBotHandler creates a new bot handler.
func NewBotHandler(botSvc *service.BotService) *BotHandler {
	return &BotHandler{botSvc: botSvc}
}

// Create handles POST /api/v1/bots.
// @Summary Create bot
// @Description Create a new bot with the given configuration.
// @Tags Bot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.CreateBotRequest true "Bot creation payload"
// @Success 200 {object} response.CommonResponse{data=response.BotWithTokenResponse}
// @Failure 400 {object} response.CommonResponse
// @Failure 409 {object} response.CommonResponse
// @Router /bots [post]
func (h *BotHandler) Create(c *gin.Context) {
	var req request.CreateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	userID := c.GetString("user_id")
	bot, accessToken, err := h.botSvc.CreateBot(
		c.Request.Context(), userID,
		req.Name, req.Description, req.Framework, req.WebhookURL, req.Config, req.Visibility,
	)
	if err != nil {
		if errors.Is(err, service.ErrBotLimitExceeded) {
			resp.Error(c, errcode.ErrBotLimitExceeded)
			return
		}
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	resp.OK(c, response.BotWithTokenResponse{
		BotResponse: h.toBotResponse(bot, true),
		AccessToken: accessToken,
	})
}

// List handles GET /api/v1/bots.
// @Summary List user's bots
// @Description Returns paginated list of bots owned by the current user.
// @Tags Bot
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.CommonResponse
// @Router /bots [get]
func (h *BotHandler) List(c *gin.Context) {
	var req request.ListBotRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	userID := c.GetString("user_id")
	bots, total, err := h.botSvc.ListBots(c.Request.Context(), userID, req.Page, req.PageSize)
	if err != nil {
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	items := make([]response.BotResponse, 0, len(bots))
	for _, bot := range bots {
		items = append(items, h.toBotResponse(bot, true))
	}

	resp.OKPaged(c, items, total, req.Page, req.PageSize)
}

// Get handles GET /api/v1/bots/:id.
// @Summary Get bot details
// @Description Returns detailed information about a specific bot.
// @Tags Bot
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Success 200 {object} response.CommonResponse{data=response.BotResponse}
// @Failure 404 {object} response.CommonResponse
// @Router /bots/{id} [get]
func (h *BotHandler) Get(c *gin.Context) {
	botID := c.Param("id")
	userID := c.GetString("user_id")

	bot, err := h.botSvc.GetBot(c.Request.Context(), botID, userID)
	if err != nil {
		if errors.Is(err, service.ErrBotNotFound) {
			resp.Error(c, errcode.ErrBotNotFound)
			return
		}
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	resp.OK(c, h.toBotResponse(bot, true))
}

// Update handles PUT /api/v1/bots/:id.
// @Summary Update bot
// @Description Update bot name, description, webhook URL, config, or visibility.
// @Tags Bot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Param body body request.UpdateBotRequest true "Bot update payload"
// @Success 200 {object} response.CommonResponse{data=response.BotResponse}
// @Failure 404 {object} response.CommonResponse
// @Router /bots/{id} [put]
func (h *BotHandler) Update(c *gin.Context) {
	var req request.UpdateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	botID := c.Param("id")
	userID := c.GetString("user_id")

	bot, err := h.botSvc.UpdateBot(
		c.Request.Context(), botID, userID,
		req.Name, req.Description, req.WebhookURL, req.Config, req.Visibility,
	)
	if err != nil {
		if errors.Is(err, service.ErrBotNotFound) {
			resp.Error(c, errcode.ErrBotNotFound)
			return
		}
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	resp.OK(c, h.toBotResponse(bot, true))
}

// Delete handles DELETE /api/v1/bots/:id.
// @Summary Delete bot
// @Description Delete a bot owned by the current user.
// @Tags Bot
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /bots/{id} [delete]
func (h *BotHandler) Delete(c *gin.Context) {
	botID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.botSvc.DeleteBot(c.Request.Context(), botID, userID); err != nil {
		if errors.Is(err, service.ErrBotNotFound) {
			resp.Error(c, errcode.ErrBotNotFound)
			return
		}
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	resp.OKWithMessage(c, "bot deleted successfully")
}

// RegenerateToken handles POST /api/v1/bots/:id/regenerate-token.
// @Summary Regenerate bot access token
// @Description Regenerate the access token for a bot. The old token is invalidated.
// @Tags Bot
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /bots/{id}/regenerate-token [post]
func (h *BotHandler) RegenerateToken(c *gin.Context) {
	botID := c.Param("id")
	userID := c.GetString("user_id")

	newToken, err := h.botSvc.RegenerateToken(c.Request.Context(), botID, userID)
	if err != nil {
		if errors.Is(err, service.ErrBotNotFound) {
			resp.Error(c, errcode.ErrBotNotFound)
			return
		}
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	resp.OK(c, gin.H{"access_token": newToken})
}

// toBotResponse converts a model.Bot to a response DTO.
// isOwner controls whether sensitive config fields are masked or stripped.
func (h *BotHandler) toBotResponse(bot *model.Bot, isOwner bool) response.BotResponse {
	return response.BotResponse{
		ID:           bot.ID,
		Name:         bot.Name,
		Description:  bot.Description,
		Framework:    bot.Framework,
		Visibility:   bot.Visibility,
		Status:       bot.Status,
		WebhookURL:   bot.WebhookURL,
		Config:       h.botSvc.MaskBotConfig(bot.Config, isOwner),
		LastActiveAt: bot.LastActiveAt,
		CreatedAt:    bot.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    bot.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// SetupAvatar handles POST /api/v1/bots/:id/setup-avatar.
// @Summary One-click avatar setup
// @Description Creates an avatar, binds the 3D model asset, and links everything to the bot in one step.
// @Tags Bot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Param body body request.SetupAvatarRequest true "Avatar setup payload"
// @Success 200 {object} response.CommonResponse{data=response.AvatarResponse}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /bots/{id}/setup-avatar [post]
func (h *BotHandler) SetupAvatar(c *gin.Context) {
	var req request.SetupAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	botID := c.Param("id")
	userID := c.GetString("user_id")

	avatar, err := h.botSvc.SetupAvatar(c.Request.Context(), botID, userID, req.AssetID, req.AvatarName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBotNotFound):
			resp.Error(c, errcode.ErrBotNotFound)
		case errors.Is(err, service.ErrAssetNotFound):
			resp.Error(c, errcode.ErrAssetNotFound)
		default:
			resp.Error(c, errcode.ErrInternalServer)
		}
		return
	}

	resp.OK(c, toAvatarResponse(avatar))
}

// RemoveAvatar handles DELETE /api/v1/bots/:id/avatar.
// @Summary Remove avatar from bot
// @Description Removes the avatar configuration and asset bindings from the bot.
// @Tags Bot
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /bots/{id}/avatar [delete]
func (h *BotHandler) RemoveAvatar(c *gin.Context) {
	botID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.botSvc.RemoveAvatar(c.Request.Context(), botID, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrBotNotFound):
			resp.Error(c, errcode.ErrBotNotFound)
		case errors.Is(err, service.ErrAvatarNotFound):
			resp.Error(c, errcode.ErrAvatarNotFound)
		default:
			resp.Error(c, errcode.ErrInternalServer)
		}
		return
	}

	resp.OKWithMessage(c, "avatar removed successfully")
}

// GetAvatar handles GET /api/v1/bots/:id/avatar.
// @Summary Get bot's avatar
// @Description Returns the avatar configuration bound to the bot.
// @Tags Bot
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Success 200 {object} response.CommonResponse{data=response.AvatarResponse}
// @Failure 404 {object} response.CommonResponse
// @Router /bots/{id}/avatar [get]
func (h *BotHandler) GetAvatar(c *gin.Context) {
	botID := c.Param("id")
	userID := c.GetString("user_id")

	avatar, err := h.botSvc.GetBotAvatar(c.Request.Context(), botID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBotNotFound):
			resp.Error(c, errcode.ErrBotNotFound)
		case errors.Is(err, service.ErrAvatarNotFound):
			resp.Error(c, errcode.ErrAvatarNotFound)
		default:
			resp.Error(c, errcode.ErrInternalServer)
		}
		return
	}

	resp.OK(c, toAvatarResponse(avatar))
}

// ListPublic handles GET /api/v1/plaza/bots.
// @Summary List public bots
// @Description Returns paginated list of public bots for the bot plaza.
// @Tags Plaza
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.CommonResponse
// @Router /plaza/bots [get]
func (h *BotHandler) ListPublic(c *gin.Context) {
	var req request.ListBotRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	bots, total, err := h.botSvc.ListPublicBots(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	items := make([]response.BotResponse, 0, len(bots))
	for _, bot := range bots {
		items = append(items, h.toBotResponse(bot, false))
	}

	resp.OKPaged(c, items, total, req.Page, req.PageSize)
}

// GetPublic handles GET /api/v1/plaza/bots/:id.
// @Summary Get public bot details
// @Description Returns a public bot's details for the bot plaza.
// @Tags Plaza
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} response.CommonResponse{data=response.BotResponse}
// @Failure 404 {object} response.CommonResponse
// @Router /plaza/bots/{id} [get]
func (h *BotHandler) GetPublic(c *gin.Context) {
	botID := c.Param("id")

	bot, err := h.botSvc.GetPublicBot(c.Request.Context(), botID)
	if err != nil {
		if errors.Is(err, service.ErrBotNotFound) {
			resp.Error(c, errcode.ErrBotNotFound)
			return
		}
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	resp.OK(c, h.toBotResponse(bot, false))
}
