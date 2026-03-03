package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/NickCharlie/ubothub/backend/internal/dto/request"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/pkg/errcode"
	"github.com/NickCharlie/ubothub/backend/pkg/response"
)

// UserHandler handles user profile HTTP endpoints.
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GetMe handles GET /api/v1/users/me.
// @Summary Get current user profile
// @Description Returns the authenticated user's profile information.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.CommonResponse
// @Failure 401 {object} response.CommonResponse
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	user, err := h.userSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, errcode.ErrNotFound)
		return
	}

	response.OK(c, gin.H{
		"id":           user.ID,
		"email":        user.Email,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"avatar_url":   user.AvatarURL,
		"role":         user.Role,
		"status":       user.Status,
		"created_at":   user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"updated_at":   user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// UpdateMe handles PUT /api/v1/users/me.
// @Summary Update user profile
// @Description Update the current user's display name and avatar URL.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.UpdateProfileRequest true "Profile update payload"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 401 {object} response.CommonResponse
// @Router /users/me [put]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	user, err := h.userSvc.UpdateProfile(c.Request.Context(), userID, req.DisplayName, req.AvatarURL)
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	response.OK(c, gin.H{
		"id":           user.ID,
		"email":        user.Email,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"avatar_url":   user.AvatarURL,
		"role":         user.Role,
		"status":       user.Status,
	})
}

// ChangePassword handles PUT /api/v1/users/me/password.
// @Summary Change password
// @Description Change the current user's password.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.ChangePasswordRequest true "Password change payload"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 401 {object} response.CommonResponse
// @Router /users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	var req request.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	err := h.userSvc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Error(c, errcode.ErrInvalidCredentials)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	response.OKWithMessage(c, "password changed successfully")
}
