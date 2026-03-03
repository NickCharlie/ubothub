package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/pkg/errcode"
	"github.com/NickCharlie/ubothub/backend/pkg/response"
)

// AdminHandler handles admin management HTTP endpoints.
type AdminHandler struct {
	adminSvc *service.AdminService
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(adminSvc *service.AdminService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc}
}

// Dashboard handles GET /api/v1/admin/dashboard.
// Returns system-wide statistics for the admin panel.
func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.adminSvc.GetDashboard(c.Request.Context())
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OK(c, stats)
}

// ListUsers handles GET /api/v1/admin/users.
// Returns paginated user list with optional status and role filters.
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	role := c.Query("role")

	users, total, err := h.adminSvc.ListUsers(c.Request.Context(), page, pageSize, status, role)
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OKPaged(c, users, total, page, pageSize)
}

// BanUser handles PUT /api/v1/admin/users/:id/ban.
// Sets the target user's status to "banned".
func (h *AdminHandler) BanUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	if err := h.adminSvc.BanUser(c.Request.Context(), userID); err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OKWithMessage(c, "user banned successfully")
}

// UnbanUser handles PUT /api/v1/admin/users/:id/unban.
// Restores a banned user's status to "active".
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	if err := h.adminSvc.UnbanUser(c.Request.Context(), userID); err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OKWithMessage(c, "user unbanned successfully")
}

// ListBots handles GET /api/v1/admin/bots.
// Returns paginated bot list with optional status and framework filters.
func (h *AdminHandler) ListBots(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	framework := c.Query("framework")

	bots, total, err := h.adminSvc.ListBots(c.Request.Context(), page, pageSize, status, framework)
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OKPaged(c, bots, total, page, pageSize)
}

// ForceDeleteBot handles DELETE /api/v1/admin/bots/:id.
// Permanently removes a bot (bypasses soft delete).
func (h *AdminHandler) ForceDeleteBot(c *gin.Context) {
	botID := c.Param("id")
	if botID == "" {
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	if err := h.adminSvc.ForceDeleteBot(c.Request.Context(), botID); err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OKWithMessage(c, "bot deleted permanently")
}

// ListAssets handles GET /api/v1/admin/assets.
// Returns paginated asset list with optional category and status filters.
func (h *AdminHandler) ListAssets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	status := c.Query("status")

	assets, total, err := h.adminSvc.ListAssets(c.Request.Context(), page, pageSize, category, status)
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OKPaged(c, assets, total, page, pageSize)
}

// ListMessageLogs handles GET /api/v1/admin/logs.
// Returns paginated message logs with optional bot_id filter.
func (h *AdminHandler) ListMessageLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	botID := c.Query("bot_id")

	logs, total, err := h.adminSvc.ListMessageLogs(c.Request.Context(), page, pageSize, botID)
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OKPaged(c, logs, total, page, pageSize)
}
