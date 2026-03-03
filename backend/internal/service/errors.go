package service

import "errors"

// Sentinel errors for the service layer.
var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account temporarily locked")
	ErrTokenInvalid       = errors.New("invalid token")
	ErrTokenRevoked       = errors.New("token revoked")

	ErrBotNotFound      = errors.New("bot not found")
	ErrBotLimitExceeded = errors.New("bot count limit exceeded")

	ErrAssetNotFound      = errors.New("asset not found")
	ErrAssetFormatInvalid = errors.New("unsupported asset format")
	ErrAssetSizeTooLarge  = errors.New("file size exceeds limit")
	ErrAssetQuotaExceeded = errors.New("storage quota exceeded")

	ErrAvatarNotFound    = errors.New("avatar not found")
	ErrAvatarBotConflict = errors.New("bot already bound to another avatar")

	ErrInsufficientBalance = errors.New("insufficient wallet balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrWithdrawalTooSmall  = errors.New("withdrawal amount below minimum")
)
