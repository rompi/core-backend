package auth

import "context"

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	IncrementFailedAttempts(ctx context.Context, userID string) error
	ResetFailedAttempts(ctx context.Context, userID string) error
	LockAccount(ctx context.Context, userID string) error
	UnlockAccount(ctx context.Context, userID string) error
}

// SessionRepository defines persistence operations for user sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByToken(ctx context.Context, token string) (*Session, error)
	GetByUserID(ctx context.Context, userID string) ([]*Session, error)
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

// RoleRepository defines persistence operations for roles.
type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id string) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	GetByUserID(ctx context.Context, userID string) ([]Role, error)
	AssignToUser(ctx context.Context, userID, roleID string) error
	RemoveFromUser(ctx context.Context, userID, roleID string) error
}

// AuditLogRepository defines persistence operations for audit logs.
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByUserID(ctx context.Context, userID string, limit int) ([]*AuditLog, error)
}

// PasswordResetTokenRepository defines persistence for password reset tokens.
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	GetByToken(ctx context.Context, token string) (*PasswordResetToken, error)
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

// APIKeyRepository defines persistence for API keys.
type APIKeyRepository interface {
	GetByKey(ctx context.Context, key string) (*APIKey, error)
}
