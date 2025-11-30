package auth

import "time"

// User represents an authenticated user in the system.
type User struct {
	ID             string                 `json:"id"`
	Email          string                 `json:"email"`
	PasswordHash   string                 `json:"password_hash"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	FailedAttempts int                    `json:"failed_attempts"`
	LockedUntil    time.Time              `json:"locked_until"`
	Language       string                 `json:"language"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Session represents an authenticated session that can be revoked.
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	Metadata  string    `json:"metadata"`
}

// Role defines permissions granted to a user or a group of users.
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// AuditLog records critical authentication events for auditing purposes.
type AuditLog struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Action    string                 `json:"action"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// PasswordResetToken represents a temporary token used to reset passwords.
type PasswordResetToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// APIKey represents a long-lived credential tied to a user with limited scope.
type APIKey struct {
	Key         string    `json:"key"`
	UserID      string    `json:"user_id"`
	Description string    `json:"description"`
	Scopes      []string  `json:"scopes"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Revoked     bool      `json:"revoked"`
}
