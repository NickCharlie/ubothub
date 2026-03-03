package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ubothub/backend/internal/dto/request"
	"github.com/ubothub/backend/internal/dto/response"
	"github.com/ubothub/backend/internal/model"
	"github.com/ubothub/backend/internal/service"
	"github.com/ubothub/backend/pkg/errcode"
	resp "github.com/ubothub/backend/pkg/response"
)

// AssetHandler handles asset management HTTP endpoints.
type AssetHandler struct {
	assetSvc *service.AssetService
}

// NewAssetHandler creates a new asset handler.
func NewAssetHandler(assetSvc *service.AssetService) *AssetHandler {
	return &AssetHandler{assetSvc: assetSvc}
}

// PresignedUpload handles POST /api/v1/assets/upload/presigned.
func (h *AssetHandler) PresignedUpload(c *gin.Context) {
	var req request.PresignedUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	userID := c.GetString("user_id")
	url, fileKey, err := h.assetSvc.GeneratePresignedUpload(
		c.Request.Context(), userID, req.Filename, req.Category, req.FileSize,
	)
	if err != nil {
		handleAssetError(c, err)
		return
	}

	resp.OK(c, response.PresignedUploadResponse{
		UploadURL: url,
		FileKey:   fileKey,
		ExpiresIn: 3600,
	})
}

// CompleteUpload handles POST /api/v1/assets/upload/complete.
func (h *AssetHandler) CompleteUpload(c *gin.Context) {
	var req request.CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	userID := c.GetString("user_id")
	var tags []string
	if req.Tags != "" {
		tags = strings.Split(req.Tags, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	asset, err := h.assetSvc.CompleteUpload(
		c.Request.Context(), userID,
		req.FileKey, req.Name, req.Description, req.Category, req.Format,
		req.FileSize, req.IsPublic, tags,
	)
	if err != nil {
		handleAssetError(c, err)
		return
	}

	resp.OK(c, toAssetResponse(asset))
}

// List handles GET /api/v1/assets.
func (h *AssetHandler) List(c *gin.Context) {
	var req request.ListAssetRequest
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
	assets, total, err := h.assetSvc.ListAssets(
		c.Request.Context(), userID, req.Page, req.PageSize,
		req.Category, req.Format, req.Status,
	)
	if err != nil {
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	items := make([]response.AssetResponse, 0, len(assets))
	for _, a := range assets {
		items = append(items, toAssetResponse(a))
	}

	resp.OKPaged(c, items, total, req.Page, req.PageSize)
}

// ListPublic handles GET /api/v1/assets/public.
func (h *AssetHandler) ListPublic(c *gin.Context) {
	var req request.ListAssetRequest
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

	assets, total, err := h.assetSvc.ListPublicAssets(
		c.Request.Context(), req.Page, req.PageSize,
		req.Category, req.Format, req.Search,
	)
	if err != nil {
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	items := make([]response.AssetResponse, 0, len(assets))
	for _, a := range assets {
		items = append(items, toAssetResponse(a))
	}

	resp.OKPaged(c, items, total, req.Page, req.PageSize)
}

// Get handles GET /api/v1/assets/:id.
func (h *AssetHandler) Get(c *gin.Context) {
	assetID := c.Param("id")
	userID := c.GetString("user_id")

	asset, err := h.assetSvc.GetAsset(c.Request.Context(), assetID, userID)
	if err != nil {
		handleAssetError(c, err)
		return
	}

	resp.OK(c, toAssetResponse(asset))
}

// Update handles PUT /api/v1/assets/:id.
func (h *AssetHandler) Update(c *gin.Context) {
	var req request.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ValidationError(c, err.Error())
		return
	}

	assetID := c.Param("id")
	userID := c.GetString("user_id")

	var tags []string
	if req.Tags != "" {
		tags = strings.Split(req.Tags, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	asset, err := h.assetSvc.UpdateAsset(
		c.Request.Context(), assetID, userID,
		req.Name, req.Description, req.IsPublic, tags,
	)
	if err != nil {
		handleAssetError(c, err)
		return
	}

	resp.OK(c, toAssetResponse(asset))
}

// Delete handles DELETE /api/v1/assets/:id.
func (h *AssetHandler) Delete(c *gin.Context) {
	assetID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.assetSvc.DeleteAsset(c.Request.Context(), assetID, userID); err != nil {
		handleAssetError(c, err)
		return
	}

	resp.OKWithMessage(c, "asset deleted successfully")
}

// Download handles GET /api/v1/assets/:id/download.
func (h *AssetHandler) Download(c *gin.Context) {
	assetID := c.Param("id")
	userID := c.GetString("user_id")

	url, err := h.assetSvc.GetDownloadURL(c.Request.Context(), assetID, userID)
	if err != nil {
		handleAssetError(c, err)
		return
	}

	resp.OK(c, gin.H{"download_url": url, "expires_in": 3600})
}

// Thumbnail handles GET /api/v1/assets/:id/thumbnail.
func (h *AssetHandler) Thumbnail(c *gin.Context) {
	assetID := c.Param("id")
	userID := c.GetString("user_id")

	url, err := h.assetSvc.GetThumbnailURL(c.Request.Context(), assetID, userID)
	if err != nil {
		handleAssetError(c, err)
		return
	}

	resp.OK(c, gin.H{"thumbnail_url": url, "expires_in": 3600})
}

func toAssetResponse(asset *model.Asset) response.AssetResponse {
	tags := []string{}
	if asset.Tags != nil {
		tags = asset.Tags
	}
	return response.AssetResponse{
		ID:            asset.ID,
		UserID:        asset.UserID,
		Name:          asset.Name,
		Description:   asset.Description,
		Category:      asset.Category,
		Format:        asset.Format,
		FileSize:      asset.FileSize,
		Tags:          tags,
		IsPublic:      asset.IsPublic,
		DownloadCount: asset.DownloadCount,
		Status:        asset.Status,
		CreatedAt:     asset.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     asset.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func handleAssetError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAssetNotFound):
		resp.Error(c, errcode.ErrAssetNotFound)
	case errors.Is(err, service.ErrAssetFormatInvalid):
		resp.Error(c, errcode.ErrAssetFormatInvalid)
	case errors.Is(err, service.ErrAssetSizeTooLarge):
		resp.Error(c, errcode.ErrAssetSizeTooLarge)
	case errors.Is(err, service.ErrAssetQuotaExceeded):
		resp.Error(c, errcode.ErrAssetQuotaExceeded)
	default:
		resp.Error(c, errcode.ErrInternalServer)
	}
}
