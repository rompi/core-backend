package auth

import (
	"context"
	"net/http"
	"time"
)

// Service defines the core authentication operations provided by the package.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (*User, error)
	RefreshToken(ctx context.Context, token string) (*LoginResponse, error)
	InitiatePasswordReset(ctx context.Context, email string) (*PasswordResetToken, error)
	CompletePasswordReset(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error
	ValidateAPIKey(ctx context.Context, apiKey string) (*User, error)
	GetUserRoles(ctx context.Context, userID string) ([]Role, error)
	CheckPermission(ctx context.Context, userID string, permission string) (bool, error)
	// Middleware validates JWT bearer tokens and injects the user into the request context.
	Middleware() func(http.Handler) http.Handler
	// RequireRole only allows requests for users holding at least one of the requested roles.
	RequireRole(roles ...string) func(http.Handler) http.Handler
	// RequirePermission only allows requests for users owning all requested permissions.
	RequirePermission(permissions ...string) func(http.Handler) http.Handler
	// RateLimitMiddleware applies per-origin rate limiting to HTTP requests.
	RateLimitMiddleware() func(http.Handler) http.Handler
}

// RegisterRequest captures required data for creating a new user account.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Language string `json:"language"`
}

// LoginRequest carries credentials for authenticating a user.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse returns tokens and metadata after a successful login.
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      *User     `json:"user"`
}
