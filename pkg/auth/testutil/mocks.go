package testutil

import (
	"context"

	"github.com/rompi/core-backend/pkg/auth"
)

// MockUserRepository provides noop implementations for the user repository.
type MockUserRepository struct {
	CreateFunc                  func(ctx context.Context, user *auth.User) error
	GetByIDFunc                 func(ctx context.Context, id string) (*auth.User, error)
	GetByEmailFunc              func(ctx context.Context, email string) (*auth.User, error)
	UpdateFunc                  func(ctx context.Context, user *auth.User) error
	DeleteFunc                  func(ctx context.Context, id string) error
	IncrementFailedAttemptsFunc func(ctx context.Context, userID string) error
	ResetFailedAttemptsFunc     func(ctx context.Context, userID string) error
	LockAccountFunc             func(ctx context.Context, userID string) error
	UnlockAccountFunc           func(ctx context.Context, userID string) error
}

// Create delegates to CreateFunc if provided.
func (m *MockUserRepository) Create(ctx context.Context, user *auth.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

// GetByID delegates to GetByIDFunc if provided.
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*auth.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetByEmail delegates to GetByEmailFunc if provided.
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, nil
}

// Update delegates to UpdateFunc if provided.
func (m *MockUserRepository) Update(ctx context.Context, user *auth.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

// Delete delegates to DeleteFunc if provided.
func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// IncrementFailedAttempts delegates to IncrementFailedAttemptsFunc if provided.
func (m *MockUserRepository) IncrementFailedAttempts(ctx context.Context, userID string) error {
	if m.IncrementFailedAttemptsFunc != nil {
		return m.IncrementFailedAttemptsFunc(ctx, userID)
	}
	return nil
}

// ResetFailedAttempts delegates to ResetFailedAttemptsFunc if provided.
func (m *MockUserRepository) ResetFailedAttempts(ctx context.Context, userID string) error {
	if m.ResetFailedAttemptsFunc != nil {
		return m.ResetFailedAttemptsFunc(ctx, userID)
	}
	return nil
}

// LockAccount delegates to LockAccountFunc if provided.
func (m *MockUserRepository) LockAccount(ctx context.Context, userID string) error {
	if m.LockAccountFunc != nil {
		return m.LockAccountFunc(ctx, userID)
	}
	return nil
}

// UnlockAccount delegates to UnlockAccountFunc if provided.
func (m *MockUserRepository) UnlockAccount(ctx context.Context, userID string) error {
	if m.UnlockAccountFunc != nil {
		return m.UnlockAccountFunc(ctx, userID)
	}
	return nil
}

// MockSessionRepository provides stub implementations for session persistence.
type MockSessionRepository struct {
	CreateFunc        func(ctx context.Context, session *auth.Session) error
	GetByTokenFunc    func(ctx context.Context, token string) (*auth.Session, error)
	GetByUserIDFunc   func(ctx context.Context, userID string) ([]*auth.Session, error)
	DeleteFunc        func(ctx context.Context, token string) error
	DeleteExpiredFunc func(ctx context.Context) error
}

// Create delegates to CreateFunc if provided.
func (m *MockSessionRepository) Create(ctx context.Context, session *auth.Session) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, session)
	}
	return nil
}

// GetByToken delegates to GetByTokenFunc if provided.
func (m *MockSessionRepository) GetByToken(ctx context.Context, token string) (*auth.Session, error) {
	if m.GetByTokenFunc != nil {
		return m.GetByTokenFunc(ctx, token)
	}
	return nil, nil
}

// GetByUserID delegates to GetByUserIDFunc if provided.
func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*auth.Session, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

// Delete delegates to DeleteFunc if provided.
func (m *MockSessionRepository) Delete(ctx context.Context, token string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, token)
	}
	return nil
}

// DeleteExpired delegates to DeleteExpiredFunc if provided.
func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	if m.DeleteExpiredFunc != nil {
		return m.DeleteExpiredFunc(ctx)
	}
	return nil
}

// MockRoleRepository provides stub implementations for role persistence.
type MockRoleRepository struct {
	CreateFunc         func(ctx context.Context, role *auth.Role) error
	GetByIDFunc        func(ctx context.Context, id string) (*auth.Role, error)
	GetByNameFunc      func(ctx context.Context, name string) (*auth.Role, error)
	GetByUserIDFunc    func(ctx context.Context, userID string) ([]auth.Role, error)
	AssignToUserFunc   func(ctx context.Context, userID, roleID string) error
	RemoveFromUserFunc func(ctx context.Context, userID, roleID string) error
}

// Create delegates to CreateFunc if provided.
func (m *MockRoleRepository) Create(ctx context.Context, role *auth.Role) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, role)
	}
	return nil
}

// GetByID delegates to GetByIDFunc if provided.
func (m *MockRoleRepository) GetByID(ctx context.Context, id string) (*auth.Role, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetByName delegates to GetByNameFunc if provided.
func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*auth.Role, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, name)
	}
	return nil, nil
}

// GetByUserID delegates to GetByUserIDFunc if provided.
func (m *MockRoleRepository) GetByUserID(ctx context.Context, userID string) ([]auth.Role, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

// AssignToUser delegates to AssignToUserFunc if provided.
func (m *MockRoleRepository) AssignToUser(ctx context.Context, userID, roleID string) error {
	if m.AssignToUserFunc != nil {
		return m.AssignToUserFunc(ctx, userID, roleID)
	}
	return nil
}

// RemoveFromUser delegates to RemoveFromUserFunc if provided.
func (m *MockRoleRepository) RemoveFromUser(ctx context.Context, userID, roleID string) error {
	if m.RemoveFromUserFunc != nil {
		return m.RemoveFromUserFunc(ctx, userID, roleID)
	}
	return nil
}

// MockAuditLogRepository provides stub implementations for audit logging.
type MockAuditLogRepository struct {
	CreateFunc      func(ctx context.Context, log *auth.AuditLog) error
	GetByUserIDFunc func(ctx context.Context, userID string, limit int) ([]*auth.AuditLog, error)
}

// Create delegates to CreateFunc if provided.
func (m *MockAuditLogRepository) Create(ctx context.Context, log *auth.AuditLog) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, log)
	}
	return nil
}

// GetByUserID delegates to GetByUserIDFunc if provided.
func (m *MockAuditLogRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*auth.AuditLog, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID, limit)
	}
	return nil, nil
}

// MockPasswordResetTokenRepository provides stub implementations for reset tokens.
type MockPasswordResetTokenRepository struct {
	CreateFunc        func(ctx context.Context, token *auth.PasswordResetToken) error
	GetByTokenFunc    func(ctx context.Context, token string) (*auth.PasswordResetToken, error)
	DeleteFunc        func(ctx context.Context, token string) error
	DeleteExpiredFunc func(ctx context.Context) error
}

// Create delegates to CreateFunc if provided.
func (m *MockPasswordResetTokenRepository) Create(ctx context.Context, token *auth.PasswordResetToken) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, token)
	}
	return nil
}

// GetByToken delegates to GetByTokenFunc if provided.
func (m *MockPasswordResetTokenRepository) GetByToken(ctx context.Context, token string) (*auth.PasswordResetToken, error) {
	if m.GetByTokenFunc != nil {
		return m.GetByTokenFunc(ctx, token)
	}
	return nil, nil
}

// Delete delegates to DeleteFunc if provided.
func (m *MockPasswordResetTokenRepository) Delete(ctx context.Context, token string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, token)
	}
	return nil
}

// DeleteExpired delegates to DeleteExpiredFunc if provided.
func (m *MockPasswordResetTokenRepository) DeleteExpired(ctx context.Context) error {
	if m.DeleteExpiredFunc != nil {
		return m.DeleteExpiredFunc(ctx)
	}
	return nil
}

// MockAPIKeyRepository provides stub implementations for API key persistence.
type MockAPIKeyRepository struct {
	GetByKeyFunc func(ctx context.Context, key string) (*auth.APIKey, error)
}

// GetByKey delegates to GetByKeyFunc if provided.
func (m *MockAPIKeyRepository) GetByKey(ctx context.Context, key string) (*auth.APIKey, error) {
	if m.GetByKeyFunc != nil {
		return m.GetByKeyFunc(ctx, key)
	}
	return nil, nil
}
