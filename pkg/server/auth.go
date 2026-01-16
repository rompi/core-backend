package server

import (
	"context"
	"net/http"
)

// User represents an authenticated user.
// Implement this interface to integrate with your auth system.
type User interface {
	// GetID returns the user's unique identifier.
	GetID() string

	// GetRole returns the user's role (e.g., "admin", "user").
	GetRole() string

	// GetPermissions returns the user's permissions.
	GetPermissions() []string

	// HasRole checks if user has the specified role.
	HasRole(role string) bool

	// HasPermission checks if user has the specified permission.
	HasPermission(permission string) bool
}

// Authenticator validates tokens and returns user information.
// Implement this interface to integrate with your auth system (e.g., pkg/auth).
type Authenticator interface {
	// ValidateToken validates a token and returns the authenticated user.
	// Returns nil user and nil error if token is missing (for optional auth).
	// Returns nil user and error if token is invalid.
	ValidateToken(ctx context.Context, token string) (User, error)
}

// RoleChecker checks if a user has required roles.
// Optional: implement for custom role checking logic.
type RoleChecker interface {
	// CheckRoles returns nil if user has any of the required roles.
	CheckRoles(user User, roles ...string) error
}

// PermissionChecker checks if a user has required permissions.
// Optional: implement for custom permission checking logic.
type PermissionChecker interface {
	// CheckPermissions returns nil if user has all required permissions.
	CheckPermissions(user User, permissions ...string) error
}

// DefaultUser is a simple User implementation.
type DefaultUser struct {
	ID          string
	Role        string
	Permissions []string
}

// GetID implements User.
func (u *DefaultUser) GetID() string {
	return u.ID
}

// GetRole implements User.
func (u *DefaultUser) GetRole() string {
	return u.Role
}

// GetPermissions implements User.
func (u *DefaultUser) GetPermissions() []string {
	return u.Permissions
}

// HasRole implements User.
func (u *DefaultUser) HasRole(role string) bool {
	return u.Role == role
}

// HasPermission implements User.
func (u *DefaultUser) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// Ensure DefaultUser implements User.
var _ User = (*DefaultUser)(nil)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const userContextKey contextKey = "server-user"

// UserFromContext extracts the authenticated user from context.
// Returns nil if no user is present.
func UserFromContext(ctx context.Context) User {
	if user, ok := ctx.Value(userContextKey).(User); ok {
		return user
	}
	return nil
}

// ContextWithUser adds an authenticated user to the context.
func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromRequest extracts the authenticated user from HTTP request context.
// Returns nil if no user is present.
func UserFromRequest(r *http.Request) User {
	return UserFromContext(r.Context())
}

// DefaultRoleChecker provides a simple role checking implementation.
type DefaultRoleChecker struct{}

// CheckRoles implements RoleChecker.
// Returns nil if user has any of the required roles.
func (c *DefaultRoleChecker) CheckRoles(user User, roles ...string) error {
	if user == nil {
		return ErrUnauthenticated
	}
	for _, role := range roles {
		if user.HasRole(role) {
			return nil
		}
	}
	return ErrPermissionDenied
}

// Ensure DefaultRoleChecker implements RoleChecker.
var _ RoleChecker = (*DefaultRoleChecker)(nil)

// DefaultPermissionChecker provides a simple permission checking implementation.
type DefaultPermissionChecker struct{}

// CheckPermissions implements PermissionChecker.
// Returns nil if user has all required permissions.
func (c *DefaultPermissionChecker) CheckPermissions(user User, permissions ...string) error {
	if user == nil {
		return ErrUnauthenticated
	}
	for _, permission := range permissions {
		if !user.HasPermission(permission) {
			return ErrPermissionDenied
		}
	}
	return nil
}

// Ensure DefaultPermissionChecker implements PermissionChecker.
var _ PermissionChecker = (*DefaultPermissionChecker)(nil)

// NoopAuthenticator is an authenticator that always returns nil user.
// Useful for testing or when auth is disabled.
type NoopAuthenticator struct{}

// ValidateToken implements Authenticator.
func (NoopAuthenticator) ValidateToken(ctx context.Context, token string) (User, error) {
	return nil, nil
}

// Ensure NoopAuthenticator implements Authenticator.
var _ Authenticator = NoopAuthenticator{}
