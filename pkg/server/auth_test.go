package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultUser(t *testing.T) {
	user := &DefaultUser{
		ID:          "user-123",
		Role:        "admin",
		Permissions: []string{"read", "write", "delete"},
	}

	t.Run("GetID", func(t *testing.T) {
		if got := user.GetID(); got != "user-123" {
			t.Errorf("GetID() = %q, want %q", got, "user-123")
		}
	})

	t.Run("GetRole", func(t *testing.T) {
		if got := user.GetRole(); got != "admin" {
			t.Errorf("GetRole() = %q, want %q", got, "admin")
		}
	})

	t.Run("GetPermissions", func(t *testing.T) {
		perms := user.GetPermissions()
		if len(perms) != 3 {
			t.Errorf("GetPermissions() len = %d, want 3", len(perms))
		}
	})

	t.Run("HasRole", func(t *testing.T) {
		if !user.HasRole("admin") {
			t.Error("HasRole(admin) = false, want true")
		}
		if user.HasRole("user") {
			t.Error("HasRole(user) = true, want false")
		}
	})

	t.Run("HasPermission", func(t *testing.T) {
		if !user.HasPermission("read") {
			t.Error("HasPermission(read) = false, want true")
		}
		if !user.HasPermission("write") {
			t.Error("HasPermission(write) = false, want true")
		}
		if user.HasPermission("execute") {
			t.Error("HasPermission(execute) = true, want false")
		}
	})
}

func TestDefaultUser_EmptyPermissions(t *testing.T) {
	user := &DefaultUser{
		ID:          "user-456",
		Role:        "guest",
		Permissions: nil,
	}

	if user.HasPermission("read") {
		t.Error("HasPermission should return false for nil permissions")
	}

	perms := user.GetPermissions()
	if perms != nil {
		t.Error("GetPermissions should return nil")
	}
}

func TestUserFromContext(t *testing.T) {
	t.Run("no user in context", func(t *testing.T) {
		ctx := context.Background()
		if user := UserFromContext(ctx); user != nil {
			t.Error("UserFromContext should return nil for empty context")
		}
	})

	t.Run("user in context", func(t *testing.T) {
		user := &DefaultUser{ID: "user-1", Role: "admin"}
		ctx := ContextWithUser(context.Background(), user)

		got := UserFromContext(ctx)
		if got == nil {
			t.Fatal("UserFromContext returned nil")
		}
		if got.GetID() != "user-1" {
			t.Errorf("GetID() = %q, want %q", got.GetID(), "user-1")
		}
	})
}

func TestContextWithUser(t *testing.T) {
	user := &DefaultUser{ID: "user-1"}
	ctx := ContextWithUser(context.Background(), user)

	// Verify the user can be retrieved
	got := UserFromContext(ctx)
	if got == nil || got.GetID() != "user-1" {
		t.Error("ContextWithUser did not properly store user")
	}
}

func TestUserFromRequest(t *testing.T) {
	t.Run("no user in request context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if user := UserFromRequest(req); user != nil {
			t.Error("UserFromRequest should return nil when no user in context")
		}
	})

	t.Run("user in request context", func(t *testing.T) {
		user := &DefaultUser{ID: "user-1", Role: "user"}
		ctx := ContextWithUser(context.Background(), user)
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

		got := UserFromRequest(req)
		if got == nil {
			t.Fatal("UserFromRequest returned nil")
		}
		if got.GetID() != "user-1" {
			t.Errorf("GetID() = %q, want %q", got.GetID(), "user-1")
		}
	})
}

func TestDefaultRoleChecker(t *testing.T) {
	checker := &DefaultRoleChecker{}

	t.Run("nil user", func(t *testing.T) {
		err := checker.CheckRoles(nil, "admin")
		if err != ErrUnauthenticated {
			t.Errorf("CheckRoles(nil) error = %v, want %v", err, ErrUnauthenticated)
		}
	})

	t.Run("user has role", func(t *testing.T) {
		user := &DefaultUser{ID: "1", Role: "admin"}
		err := checker.CheckRoles(user, "admin")
		if err != nil {
			t.Errorf("CheckRoles() error = %v, want nil", err)
		}
	})

	t.Run("user has one of multiple roles", func(t *testing.T) {
		user := &DefaultUser{ID: "1", Role: "editor"}
		err := checker.CheckRoles(user, "admin", "editor", "viewer")
		if err != nil {
			t.Errorf("CheckRoles() error = %v, want nil", err)
		}
	})

	t.Run("user does not have role", func(t *testing.T) {
		user := &DefaultUser{ID: "1", Role: "user"}
		err := checker.CheckRoles(user, "admin")
		if err != ErrPermissionDenied {
			t.Errorf("CheckRoles() error = %v, want %v", err, ErrPermissionDenied)
		}
	})

	t.Run("user does not have any of multiple roles", func(t *testing.T) {
		user := &DefaultUser{ID: "1", Role: "user"}
		err := checker.CheckRoles(user, "admin", "editor")
		if err != ErrPermissionDenied {
			t.Errorf("CheckRoles() error = %v, want %v", err, ErrPermissionDenied)
		}
	})
}

func TestDefaultPermissionChecker(t *testing.T) {
	checker := &DefaultPermissionChecker{}

	t.Run("nil user", func(t *testing.T) {
		err := checker.CheckPermissions(nil, "read")
		if err != ErrUnauthenticated {
			t.Errorf("CheckPermissions(nil) error = %v, want %v", err, ErrUnauthenticated)
		}
	})

	t.Run("user has permission", func(t *testing.T) {
		user := &DefaultUser{ID: "1", Permissions: []string{"read", "write"}}
		err := checker.CheckPermissions(user, "read")
		if err != nil {
			t.Errorf("CheckPermissions() error = %v, want nil", err)
		}
	})

	t.Run("user has all permissions", func(t *testing.T) {
		user := &DefaultUser{ID: "1", Permissions: []string{"read", "write", "delete"}}
		err := checker.CheckPermissions(user, "read", "write")
		if err != nil {
			t.Errorf("CheckPermissions() error = %v, want nil", err)
		}
	})

	t.Run("user missing permission", func(t *testing.T) {
		user := &DefaultUser{ID: "1", Permissions: []string{"read"}}
		err := checker.CheckPermissions(user, "write")
		if err != ErrPermissionDenied {
			t.Errorf("CheckPermissions() error = %v, want %v", err, ErrPermissionDenied)
		}
	})

	t.Run("user missing one of multiple permissions", func(t *testing.T) {
		user := &DefaultUser{ID: "1", Permissions: []string{"read"}}
		err := checker.CheckPermissions(user, "read", "write")
		if err != ErrPermissionDenied {
			t.Errorf("CheckPermissions() error = %v, want %v", err, ErrPermissionDenied)
		}
	})

	t.Run("user with no permissions", func(t *testing.T) {
		user := &DefaultUser{ID: "1", Permissions: nil}
		err := checker.CheckPermissions(user, "read")
		if err != ErrPermissionDenied {
			t.Errorf("CheckPermissions() error = %v, want %v", err, ErrPermissionDenied)
		}
	})
}

func TestNoopAuthenticator(t *testing.T) {
	auth := NoopAuthenticator{}

	user, err := auth.ValidateToken(context.Background(), "any-token")
	if err != nil {
		t.Errorf("ValidateToken() error = %v, want nil", err)
	}
	if user != nil {
		t.Error("ValidateToken() should return nil user")
	}

	user, err = auth.ValidateToken(context.Background(), "")
	if err != nil {
		t.Errorf("ValidateToken() error = %v, want nil", err)
	}
	if user != nil {
		t.Error("ValidateToken() should return nil user for empty token")
	}
}

// Verify interface implementations
func TestInterfaceCompliance(t *testing.T) {
	// These compile-time checks ensure interface compliance
	var _ User = (*DefaultUser)(nil)
	var _ RoleChecker = (*DefaultRoleChecker)(nil)
	var _ PermissionChecker = (*DefaultPermissionChecker)(nil)
	var _ Authenticator = NoopAuthenticator{}
}
