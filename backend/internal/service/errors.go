package service

import "errors"

// Sentinel errors for the service layer.
var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account temporarily locked")
	ErrTokenInvalid       = errors.New("invalid token")
	ErrTokenRevoked       = errors.New("token revoked")
)
