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

// AvatarHandler handles avatar management HTTP endpoints.
type AvatarHandler struct {
	avatarSvc *service.AvatarService
}

// NewAvatarHandler creates a new avatar handler.
func NewAvatarHandler(avatarSvc *service.AvatarService) *AvatarHandler {
	return &AvatarHandler{avatarSvc: avatarSvc}
}

// Create handles POST /api/v1/avatars.
func (h *AvatarHandler) Create(c *gin.Context) {
	var req request.CreateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	userID := c.GetString("user_id")
	avatar, err := h.avatarSvc.CreateAvatar(
		c.Request.Context(), userID,
		req.Name, req.Description, req.RenderType, req.SceneConfig, req.ActionMapping,
	)
	if err != nil {
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	resp.OK(c, toAvatarResponse(avatar))
}

// List handles GET /api/v1/avatars.
func (h *AvatarHandler) List(c *gin.Context) {
	var req request.ListAvatarRequest
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
	avatars, total, err := h.avatarSvc.ListAvatars(c.Request.Context(), userID, req.Page, req.PageSize)
	if err != nil {
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	items := make([]response.AvatarResponse, 0, len(avatars))
	for _, a := range avatars {
		items = append(items, toAvatarResponse(a))
	}

	resp.OKPaged(c, items, total, req.Page, req.PageSize)
}

// Get handles GET /api/v1/avatars/:id.
func (h *AvatarHandler) Get(c *gin.Context) {
	avatarID := c.Param("id")
	userID := c.GetString("user_id")

	avatar, err := h.avatarSvc.GetAvatar(c.Request.Context(), avatarID, userID)
	if err != nil {
		handleAvatarError(c, err)
		return
	}

	resp.OK(c, toAvatarResponse(avatar))
}

// Update handles PUT /api/v1/avatars/:id.
func (h *AvatarHandler) Update(c *gin.Context) {
	var req request.UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	avatarID := c.Param("id")
	userID := c.GetString("user_id")

	avatar, err := h.avatarSvc.UpdateAvatar(
		c.Request.Context(), avatarID, userID,
		req.Name, req.Description, req.SceneConfig, req.ActionMapping,
	)
	if err != nil {
		handleAvatarError(c, err)
		return
	}

	resp.OK(c, toAvatarResponse(avatar))
}

// Delete handles DELETE /api/v1/avatars/:id.
func (h *AvatarHandler) Delete(c *gin.Context) {
	avatarID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.avatarSvc.DeleteAvatar(c.Request.Context(), avatarID, userID); err != nil {
		handleAvatarError(c, err)
		return
	}

	resp.OKWithMessage(c, "avatar deleted successfully")
}

// BindBot handles POST /api/v1/avatars/:id/bind-bot.
func (h *AvatarHandler) BindBot(c *gin.Context) {
	var req request.BindBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	avatarID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.avatarSvc.BindBot(c.Request.Context(), avatarID, userID, req.BotID); err != nil {
		handleAvatarError(c, err)
		return
	}

	resp.OKWithMessage(c, "bot bound successfully")
}

// BindAsset handles POST /api/v1/avatars/:id/bind-asset.
func (h *AvatarHandler) BindAsset(c *gin.Context) {
	var req request.BindAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	avatarID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.avatarSvc.BindAsset(
		c.Request.Context(), avatarID, userID,
		req.AssetID, req.Role, req.Config, req.SortOrder,
	); err != nil {
		handleAvatarError(c, err)
		return
	}

	resp.OKWithMessage(c, "asset bound successfully")
}

// UnbindAsset handles DELETE /api/v1/avatars/:id/assets/:assetId.
func (h *AvatarHandler) UnbindAsset(c *gin.Context) {
	avatarID := c.Param("id")
	assetID := c.Param("assetId")
	userID := c.GetString("user_id")

	if err := h.avatarSvc.UnbindAsset(c.Request.Context(), avatarID, userID, assetID); err != nil {
		handleAvatarError(c, err)
		return
	}

	resp.OKWithMessage(c, "asset unbound successfully")
}

// UpdateActionMapping handles PUT /api/v1/avatars/:id/action-mapping.
func (h *AvatarHandler) UpdateActionMapping(c *gin.Context) {
	var req request.UpdateActionMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	avatarID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.avatarSvc.UpdateActionMapping(c.Request.Context(), avatarID, userID, req.ActionMapping); err != nil {
		handleAvatarError(c, err)
		return
	}

	resp.OKWithMessage(c, "action mapping updated successfully")
}

func toAvatarResponse(avatar *model.AvatarConfig) response.AvatarResponse {
	r := response.AvatarResponse{
		ID:            avatar.ID,
		UserID:        avatar.UserID,
		BotID:         avatar.BotID,
		Name:          avatar.Name,
		Description:   avatar.Description,
		RenderType:    avatar.RenderType,
		SceneConfig:   avatar.SceneConfig,
		ActionMapping: avatar.ActionMapping,
		IsDefault:     avatar.IsDefault,
		CreatedAt:     avatar.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     avatar.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if avatar.AvatarAssets != nil {
		r.AvatarAssets = make([]response.AvatarAssetDetail, 0, len(avatar.AvatarAssets))
		for _, aa := range avatar.AvatarAssets {
			r.AvatarAssets = append(r.AvatarAssets, response.AvatarAssetDetail{
				AssetID:   aa.AssetID,
				AssetName: aa.Asset.Name,
				Role:      aa.Role,
				Config:    aa.Config,
				SortOrder: aa.SortOrder,
			})
		}
	}

	return r
}

func handleAvatarError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAvatarNotFound):
		resp.Error(c, errcode.ErrAvatarNotFound)
	case errors.Is(err, service.ErrAvatarBotConflict):
		resp.Error(c, errcode.ErrAvatarBotConflict)
	case errors.Is(err, service.ErrBotNotFound):
		resp.Error(c, errcode.ErrBotNotFound)
	default:
		resp.Error(c, errcode.ErrInternalServer)
	}
}
