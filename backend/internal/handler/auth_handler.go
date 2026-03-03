package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/NickCharlie/ubothub/backend/internal/captcha"
	"github.com/NickCharlie/ubothub/backend/internal/dto/request"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/pkg/errcode"
	"github.com/NickCharlie/ubothub/backend/pkg/response"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authSvc    *service.AuthService
	emailSvc   *service.EmailService
	legalSvc   *service.LegalService
	captchaSvc *captcha.Service
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authSvc *service.AuthService, emailSvc *service.EmailService, legalSvc *service.LegalService, captchaSvc *captcha.Service) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, emailSvc: emailSvc, legalSvc: legalSvc, captchaSvc: captchaSvc}
}

// Captcha handles GET /api/v1/auth/captcha.
// Generates a new captcha image and returns the captcha ID + base64 image.
func (h *AuthHandler) Captcha(c *gin.Context) {
	result, err := h.captchaSvc.Generate()
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OK(c, gin.H{
		"captcha_id":    result.CaptchaID,
		"captcha_image": result.Image,
	})
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Verify captcha first.
	if !h.captchaSvc.Verify(req.CaptchaID, req.CaptchaAnswer) {
		response.Error(c, errcode.ErrCaptchaInvalid)
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

	// Send verification email (non-blocking, log errors only).
	go func() {
		if h.emailSvc != nil {
			_ = h.emailSvc.SendVerificationEmail(c.Request.Context(), result.User.ID, result.User.Username, result.User.Email)
		}
	}()

	response.OK(c, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    900,
		"user": gin.H{
			"id":             result.User.ID,
			"email":          result.User.Email,
			"username":       result.User.Username,
			"display_name":   result.User.DisplayName,
			"email_verified": result.User.EmailVerified,
			"role":           result.User.Role,
			"status":         result.User.Status,
			"created_at":     result.User.CreatedAt.Format("2006-01-02T15:04:05Z"),
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

	// Verify captcha.
	if !h.captchaSvc.Verify(req.CaptchaID, req.CaptchaAnswer) {
		response.Error(c, errcode.ErrCaptchaInvalid)
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
			"id":             result.User.ID,
			"email":          result.User.Email,
			"username":       result.User.Username,
			"display_name":   result.User.DisplayName,
			"email_verified": result.User.EmailVerified,
			"role":           result.User.Role,
			"status":         result.User.Status,
			"created_at":     result.User.CreatedAt.Format("2006-01-02T15:04:05Z"),
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

// VerifyEmail handles GET /api/v1/auth/verify-email?token=xxx.
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	userID, err := h.emailSvc.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, service.ErrTokenInvalid) {
			response.Error(c, errcode.ErrResetTokenInvalid)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	if err := h.authSvc.MarkEmailVerified(c.Request.Context(), userID); err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	response.OKWithMessage(c, "email verified successfully")
}

// ForgotPassword handles POST /api/v1/auth/forgot-password.
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req request.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Always return success to prevent email enumeration.
	user, err := h.authSvc.FindUserByEmail(c.Request.Context(), req.Email)
	if err == nil && user != nil && h.emailSvc != nil {
		_ = h.emailSvc.SendPasswordResetEmail(c.Request.Context(), user.ID, user.Username, user.Email)
	}

	response.OKWithMessage(c, "if the email exists, a reset link has been sent")
}

// ResetPassword handles POST /api/v1/auth/reset-password.
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	userID, err := h.emailSvc.ValidateResetToken(c.Request.Context(), req.Token)
	if err != nil {
		if errors.Is(err, service.ErrTokenInvalid) {
			response.Error(c, errcode.ErrResetTokenInvalid)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	if err := h.authSvc.ResetPassword(c.Request.Context(), userID, req.NewPassword); err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	response.OKWithMessage(c, "password reset successfully")
}

// ResendVerification handles POST /api/v1/auth/resend-verification.
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	user, err := h.authSvc.FindUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	if user.EmailVerified {
		response.Error(c, errcode.ErrEmailAlreadyVerified)
		return
	}

	if h.emailSvc != nil {
		if err := h.emailSvc.SendVerificationEmail(c.Request.Context(), user.ID, user.Username, user.Email); err != nil {
			response.Error(c, errcode.ErrInternalServer)
			return
		}
	}

	response.OKWithMessage(c, "verification email sent")
}
