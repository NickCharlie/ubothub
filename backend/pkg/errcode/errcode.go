package errcode

import "net/http"

// ErrCode represents a standardized error code with HTTP status and message.
type ErrCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// Error implements the error interface.
func (e *ErrCode) Error() string {
	return e.Message
}

// Common error codes (10000-10999).
var (
	Success          = &ErrCode{Code: 0, Message: "success", Status: http.StatusOK}
	ErrInternalServer = &ErrCode{Code: 10001, Message: "internal server error", Status: http.StatusInternalServerError}
	ErrBadRequest     = &ErrCode{Code: 10002, Message: "bad request", Status: http.StatusBadRequest}
	ErrUnauthorized   = &ErrCode{Code: 10003, Message: "unauthorized", Status: http.StatusUnauthorized}
	ErrForbidden      = &ErrCode{Code: 10004, Message: "forbidden", Status: http.StatusForbidden}
	ErrNotFound       = &ErrCode{Code: 10005, Message: "resource not found", Status: http.StatusNotFound}
	ErrConflict       = &ErrCode{Code: 10006, Message: "resource conflict", Status: http.StatusConflict}
	ErrTooManyRequests = &ErrCode{Code: 10007, Message: "too many requests", Status: http.StatusTooManyRequests}
	ErrValidation     = &ErrCode{Code: 10008, Message: "validation failed", Status: http.StatusUnprocessableEntity}
)

// Auth error codes (11000-11999).
var (
	ErrInvalidCredentials = &ErrCode{Code: 11001, Message: "invalid email or password", Status: http.StatusUnauthorized}
	ErrTokenExpired       = &ErrCode{Code: 11002, Message: "token expired", Status: http.StatusUnauthorized}
	ErrTokenInvalid       = &ErrCode{Code: 11003, Message: "invalid token", Status: http.StatusUnauthorized}
	ErrTokenBlacklisted   = &ErrCode{Code: 11004, Message: "token revoked", Status: http.StatusUnauthorized}
	ErrUserExists         = &ErrCode{Code: 11005, Message: "user already exists", Status: http.StatusConflict}
	ErrAccountLocked      = &ErrCode{Code: 11006, Message: "account temporarily locked", Status: http.StatusTooManyRequests}
	ErrWeakPassword       = &ErrCode{Code: 11007, Message: "password does not meet strength requirements", Status: http.StatusBadRequest}
	ErrAgreementRequired  = &ErrCode{Code: 11008, Message: "you must accept the terms of service and privacy policy", Status: http.StatusBadRequest}
)

// Bot error codes (12000-12999).
var (
	ErrBotNotFound      = &ErrCode{Code: 12001, Message: "bot not found", Status: http.StatusNotFound}
	ErrBotTokenInvalid  = &ErrCode{Code: 12002, Message: "invalid bot access token", Status: http.StatusUnauthorized}
	ErrBotLimitExceeded = &ErrCode{Code: 12003, Message: "bot count limit exceeded", Status: http.StatusForbidden}
)

// Asset error codes (13000-13999).
var (
	ErrAssetNotFound       = &ErrCode{Code: 13001, Message: "asset not found", Status: http.StatusNotFound}
	ErrAssetFormatInvalid  = &ErrCode{Code: 13002, Message: "unsupported asset format", Status: http.StatusBadRequest}
	ErrAssetSizeTooLarge   = &ErrCode{Code: 13003, Message: "file size exceeds limit", Status: http.StatusRequestEntityTooLarge}
	ErrAssetQuotaExceeded  = &ErrCode{Code: 13004, Message: "storage quota exceeded", Status: http.StatusForbidden}
	ErrAssetProcessFailed  = &ErrCode{Code: 13005, Message: "asset processing failed", Status: http.StatusInternalServerError}
)

// Avatar error codes (14000-14999).
var (
	ErrAvatarNotFound    = &ErrCode{Code: 14001, Message: "avatar not found", Status: http.StatusNotFound}
	ErrAvatarBotConflict = &ErrCode{Code: 14002, Message: "bot already bound to another avatar", Status: http.StatusConflict}
)

// Content moderation error codes (15000-15999).
var (
	ErrContentViolation = &ErrCode{Code: 15001, Message: "content violates platform policies", Status: http.StatusForbidden}
)
