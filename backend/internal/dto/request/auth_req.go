package request

// RegisterRequest represents the user registration request payload.
type RegisterRequest struct {
	Email         string `json:"email" binding:"required,email,max=255"`
	Username      string `json:"username" binding:"required,min=3,max=64,alphanum"`
	Password      string `json:"password" binding:"required,min=8,max=128"`
	AcceptTerms   bool   `json:"accept_terms" binding:"required"`
	AcceptPrivacy bool   `json:"accept_privacy" binding:"required"`
	CaptchaID     string `json:"captcha_id" binding:"required"`
	CaptchaAnswer string `json:"captcha_answer" binding:"required"`
}

// LoginRequest represents the user login request payload.
type LoginRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required"`
	CaptchaID     string `json:"captcha_id" binding:"required"`
	CaptchaAnswer string `json:"captcha_answer" binding:"required"`
}

// RefreshTokenRequest represents the token refresh request payload.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UpdateProfileRequest represents the user profile update payload.
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"omitempty,max=128"`
	AvatarURL   string `json:"avatar_url" binding:"omitempty,url,max=2048"`
}

// ChangePasswordRequest represents the password change payload.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=128"`
}

// ForgotPasswordRequest represents the forgot password request payload.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the password reset request payload.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=128"`
}
