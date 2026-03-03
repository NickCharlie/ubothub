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
// @Summary Admin dashboard
// @Description Returns system-wide statistics for the admin panel.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.CommonResponse
// @Router /admin/dashboard [get]
func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.adminSvc.GetDashboard(c.Request.Context())
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OK(c, stats)
}

// ListUsers handles GET /api/v1/admin/users.
// @Summary List all users
// @Description Returns paginated user list with optional status and role filters.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param role query string false "Filter by role"
// @Success 200 {object} response.CommonResponse
// @Router /admin/users [get]
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
// @Summary Ban user
// @Description Sets the target user's status to "banned".
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /admin/users/{id}/ban [put]
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
// @Summary Unban user
// @Description Restores a banned user's status to "active".
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /admin/users/{id}/unban [put]
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
// @Summary List all bots (admin)
// @Description Returns paginated bot list with optional status and framework filters.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param framework query string false "Filter by framework"
// @Success 200 {object} response.CommonResponse
// @Router /admin/bots [get]
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
// @Summary Force delete bot (admin)
// @Description Permanently removes a bot (bypasses soft delete).
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /admin/bots/{id} [delete]
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
// @Summary List all assets (admin)
// @Description Returns paginated asset list with optional category and status filters.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param category query string false "Filter by category"
// @Param status query string false "Filter by status"
// @Success 200 {object} response.CommonResponse
// @Router /admin/assets [get]
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
// @Summary List message logs (admin)
// @Description Returns paginated message logs with optional bot_id filter.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param bot_id query string false "Filter by bot ID"
// @Success 200 {object} response.CommonResponse
// @Router /admin/logs [get]
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
