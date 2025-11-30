package auth

import (
	"errors"
	"fmt"
)

// #nosec G101 -- these are error codes, not credentials.
const (
	CodeInvalidCredentials = "invalid_credentials"
	CodeUserAlreadyExists  = "user_already_exists"
	CodeUserNotFound       = "user_not_found"
	CodeAccountLocked      = "account_locked"
	CodeInvalidToken       = "invalid_token"
	CodeWeakPassword       = "weak_password"
	CodeRateLimitExceeded  = "rate_limit_exceeded"
	CodePermissionDenied   = "permission_denied"
	CodeSessionExpired     = "session_expired"
	CodeInvalidResetToken  = "invalid_reset_token"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrAccountLocked      = errors.New("account is locked due to too many failed attempts")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrWeakPassword       = errors.New("password does not meet complexity requirements")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrSessionExpired     = errors.New("session has expired")
	ErrInvalidResetToken  = errors.New("invalid or expired reset token")
	ErrNotImplemented     = errors.New("feature not implemented")
)

// AuthError contains structured details for API error responses.
type AuthError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"status_code"`
	Details    map[string]interface{} `json:"details"`
}

// Error returns a user-friendly string for the auth error.
func (e AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAuthError builds an AuthError with localized messaging.
func NewAuthError(code string, status int, language string, details map[string]interface{}, args ...interface{}) AuthError {
	return AuthError{
		Code:       code,
		Message:    DefaultTranslator.Message(language, code, args...),
		StatusCode: status,
		Details:    details,
	}
}

// WithLanguage returns a copy of the error translated into the provided language.
func (e AuthError) WithLanguage(language string) AuthError {
	if e.Code == "" || language == "" {
		return e
	}
	e.Message = DefaultTranslator.Message(language, e.Code)
	return e
}
