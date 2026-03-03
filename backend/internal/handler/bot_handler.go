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
func (h *BotHandler) Create(c *gin.Context) {
	var req request.CreateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	userID := c.GetString("user_id")
	bot, accessToken, err := h.botSvc.CreateBot(
		c.Request.Context(), userID,
		req.Name, req.Description, req.Framework, req.WebhookURL, req.Config,
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
		BotResponse: toBotResponse(bot),
		AccessToken: accessToken,
	})
}

// List handles GET /api/v1/bots.
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
		items = append(items, toBotResponse(bot))
	}

	resp.OKPaged(c, items, total, req.Page, req.PageSize)
}

// Get handles GET /api/v1/bots/:id.
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

	resp.OK(c, toBotResponse(bot))
}

// Update handles PUT /api/v1/bots/:id.
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
		req.Name, req.Description, req.WebhookURL, req.Config,
	)
	if err != nil {
		if errors.Is(err, service.ErrBotNotFound) {
			resp.Error(c, errcode.ErrBotNotFound)
			return
		}
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	resp.OK(c, toBotResponse(bot))
}

// Delete handles DELETE /api/v1/bots/:id.
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
func toBotResponse(bot *model.Bot) response.BotResponse {
	return response.BotResponse{
		ID:           bot.ID,
		Name:         bot.Name,
		Description:  bot.Description,
		Framework:    bot.Framework,
		Status:       bot.Status,
		WebhookURL:   bot.WebhookURL,
		Config:       bot.Config,
		LastActiveAt: bot.LastActiveAt,
		CreatedAt:    bot.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    bot.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
