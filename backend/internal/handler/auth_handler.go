package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/NickCharlie/ubothub/backend/internal/dto/request"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/pkg/errcode"
	"github.com/NickCharlie/ubothub/backend/pkg/response"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authSvc  *service.AuthService
	legalSvc *service.LegalService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authSvc *service.AuthService, legalSvc *service.LegalService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, legalSvc: legalSvc}
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Require explicit acceptance of terms and privacy policy.
	if !req.AcceptTerms || !req.AcceptPrivacy {
		response.Error(c, errcode.ErrAgreementRequired)
		return
	}

	result, err := h.authSvc.Register(c.Request.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			response.Error(c, errcode.ErrUserExists)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	// Record agreement acceptances asynchronously.
	ipAddr := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	ctx := c.Request.Context()

	termsAgreement, err := h.legalSvc.GetActiveAgreement(ctx, model.AgreementTypeTerms, "en")
	if err == nil && termsAgreement != nil {
		_ = h.legalSvc.RecordAcceptance(ctx, result.User.ID, termsAgreement.ID, ipAddr, userAgent)
	}
	privacyAgreement, err := h.legalSvc.GetActiveAgreement(ctx, model.AgreementTypePrivacy, "en")
	if err == nil && privacyAgreement != nil {
		_ = h.legalSvc.RecordAcceptance(ctx, result.User.ID, privacyAgreement.ID, ipAddr, userAgent)
	}

	response.OK(c, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    900,
		"user": gin.H{
			"id":           result.User.ID,
			"email":        result.User.Email,
			"username":     result.User.Username,
			"display_name": result.User.DisplayName,
			"role":         result.User.Role,
			"status":       result.User.Status,
			"created_at":   result.User.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	result, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.Error(c, errcode.ErrInvalidCredentials)
		case errors.Is(err, service.ErrAccountLocked):
			response.Error(c, errcode.ErrAccountLocked)
		default:
			response.Error(c, errcode.ErrInternalServer)
		}
		return
	}

	response.OK(c, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    900,
		"user": gin.H{
			"id":           result.User.ID,
			"email":        result.User.Email,
			"username":     result.User.Username,
			"display_name": result.User.DisplayName,
			"role":         result.User.Role,
			"status":       result.User.Status,
			"created_at":   result.User.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTokenInvalid):
			response.Error(c, errcode.ErrTokenInvalid)
		case errors.Is(err, service.ErrTokenRevoked):
			response.Error(c, errcode.ErrTokenBlacklisted)
		default:
			response.Error(c, errcode.ErrInternalServer)
		}
		return
	}

	response.OK(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
	})
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	if err := h.authSvc.Logout(c.Request.Context(), tokenStr); err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	response.OKWithMessage(c, "logged out successfully")
}
