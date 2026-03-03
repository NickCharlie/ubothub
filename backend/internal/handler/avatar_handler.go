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
// @Summary Create avatar
// @Description Create a new avatar configuration with render type and scene config.
// @Tags Avatar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.CreateAvatarRequest true "Avatar creation payload"
// @Success 200 {object} response.CommonResponse{data=response.AvatarResponse}
// @Failure 400 {object} response.CommonResponse
// @Router /avatars [post]
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
// @Summary List user's avatars
// @Description Returns paginated list of avatars owned by the current user.
// @Tags Avatar
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.CommonResponse
// @Router /avatars [get]
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
// @Summary Get avatar details
// @Description Returns detailed avatar configuration with bound assets.
// @Tags Avatar
// @Produce json
// @Security BearerAuth
// @Param id path string true "Avatar ID"
// @Success 200 {object} response.CommonResponse{data=response.AvatarResponse}
// @Failure 404 {object} response.CommonResponse
// @Router /avatars/{id} [get]
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
// @Summary Update avatar
// @Description Update avatar name, description, scene config, or action mapping.
// @Tags Avatar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Avatar ID"
// @Param body body request.UpdateAvatarRequest true "Avatar update payload"
// @Success 200 {object} response.CommonResponse{data=response.AvatarResponse}
// @Failure 404 {object} response.CommonResponse
// @Router /avatars/{id} [put]
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
// @Summary Delete avatar
// @Description Delete an avatar configuration and its asset bindings.
// @Tags Avatar
// @Produce json
// @Security BearerAuth
// @Param id path string true "Avatar ID"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /avatars/{id} [delete]
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
// @Summary Bind bot to avatar
// @Description Associate a bot with this avatar configuration.
// @Tags Avatar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Avatar ID"
// @Param body body request.BindBotRequest true "Bot binding payload"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Failure 409 {object} response.CommonResponse
// @Router /avatars/{id}/bind-bot [post]
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
// @Summary Bind asset to avatar
// @Description Associate a 3D/Live2D asset with an avatar role.
// @Tags Avatar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Avatar ID"
// @Param body body request.BindAssetRequest true "Asset binding payload"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /avatars/{id}/bind-asset [post]
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
// @Summary Unbind asset from avatar
// @Description Remove an asset from the avatar configuration.
// @Tags Avatar
// @Produce json
// @Security BearerAuth
// @Param id path string true "Avatar ID"
// @Param assetId path string true "Asset ID"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /avatars/{id}/assets/{assetId} [delete]
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
// @Summary Update action mapping
// @Description Update the action-to-animation mapping for an avatar.
// @Tags Avatar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Avatar ID"
// @Param body body request.UpdateActionMappingRequest true "Action mapping payload"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /avatars/{id}/action-mapping [put]
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

// GetPublic handles GET /api/v1/plaza/avatars/:id.
// @Summary Get public avatar preview
// @Description Returns avatar with its 3D model assets for public preview.
// @Tags Plaza
// @Produce json
// @Param id path string true "Avatar ID"
// @Success 200 {object} response.CommonResponse{data=response.AvatarResponse}
// @Failure 404 {object} response.CommonResponse
// @Router /plaza/avatars/{id} [get]
func (h *AvatarHandler) GetPublic(c *gin.Context) {
	avatarID := c.Param("id")

	avatar, err := h.avatarSvc.GetAvatarPublic(c.Request.Context(), avatarID)
	if err != nil {
		handleAvatarError(c, err)
		return
	}

	resp.OK(c, toAvatarResponse(avatar))
}
