package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/rompi/core-backend/pkg/auth"
	"github.com/rompi/core-backend/pkg/auth/testutil"
)

func TestService_InitiatePasswordReset(t *testing.T) {
	cfg := newTestConfig()
	cfg.JWTSecret = "secret"
	cfg.RateLimitMaxRequests = 100

	user := &auth.User{ID: "user-1", Email: "user@example.com"}
	tokens := &testutil.MockPasswordResetTokenRepository{
		CreateFunc: func(ctx context.Context, token *auth.PasswordResetToken) error {
			if token.UserID != user.ID {
				t.Fatalf("unexpected user id %s", token.UserID)
			}
			if token.Token == "" {
				t.Fatal("expected token value")
			}
			return nil
		},
	}
	users := &testutil.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*auth.User, error) {
			return user, nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{
		Users:               users,
		PasswordResetTokens: tokens,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	reset, err := svc.InitiatePasswordReset(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("InitiatePasswordReset() error = %v", err)
	}
	if reset.UserID != user.ID {
		t.Fatalf("user id = %s", reset.UserID)
	}
	if reset.Token == "" {
		t.Fatal("expected token")
	}
}

func TestService_CompletePasswordReset(t *testing.T) {
	cfg := newTestConfig()
	cfg.JWTSecret = "secret"
	cfg.ResetTokenExpiration = time.Hour

	tokenValue := "reset-token"
	hashed, _ := auth.HashPassword("OldPass1!", cfg.BcryptCost)
	user := &auth.User{ID: "user-1", Email: "user@example.com", PasswordHash: hashed}

	tokens := &testutil.MockPasswordResetTokenRepository{
		GetByTokenFunc: func(ctx context.Context, token string) (*auth.PasswordResetToken, error) {
			if token != tokenValue {
				return nil, nil
			}
			return &auth.PasswordResetToken{
				Token:     token,
				UserID:    user.ID,
				IssuedAt:  time.Now().Add(-time.Hour),
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
		DeleteFunc: func(ctx context.Context, token string) error {
			if token != tokenValue {
				t.Fatalf("unexpected token %s", token)
			}
			return nil
		},
	}

	updated := false
	users := &testutil.MockUserRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*auth.User, error) {
			return user, nil
		},
		UpdateFunc: func(ctx context.Context, u *auth.User) error {
			updated = true
			if u.PasswordHash == hashed {
				t.Fatal("expected new hash")
			}
			return nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{
		Users:               users,
		PasswordResetTokens: tokens,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if err := svc.CompletePasswordReset(context.Background(), tokenValue, "New!Pass1"); err != nil {
		t.Fatalf("CompletePasswordReset() error = %v", err)
	}
	if !updated {
		t.Fatal("expected user update")
	}
}

func TestService_ChangePassword(t *testing.T) {
	cfg := newTestConfig()
	cfg.JWTSecret = "secret"

	hashed, _ := auth.HashPassword("OldPass1!", cfg.BcryptCost)
	user := &auth.User{ID: "user-1", PasswordHash: hashed}

	updated := false
	users := &testutil.MockUserRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*auth.User, error) {
			return user, nil
		},
		UpdateFunc: func(ctx context.Context, u *auth.User) error {
			updated = true
			if u.PasswordHash == hashed {
				t.Fatal("expected new hash")
			}
			return nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{Users: users})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if err := svc.ChangePassword(context.Background(), user.ID, "OldPass1!", "New!Pass1"); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if !updated {
		t.Fatal("expected user update")
	}
}

func TestService_ValidateAPIKeyAndPermissions(t *testing.T) {
	cfg := newTestConfig()
	cfg.JWTSecret = "secret"

	apiKey := &auth.APIKey{Key: "abc", UserID: "user-1", ExpiresAt: time.Now().Add(time.Hour)}
	keyRepo := &testutil.MockAPIKeyRepository{
		GetByKeyFunc: func(ctx context.Context, key string) (*auth.APIKey, error) {
			if key != apiKey.Key {
				return nil, nil
			}
			return apiKey, nil
		},
	}

	user := &auth.User{ID: "user-1"}
	users := &testutil.MockUserRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*auth.User, error) {
			return user, nil
		},
	}

	roles := &testutil.MockRoleRepository{
		GetByUserIDFunc: func(ctx context.Context, id string) ([]auth.Role, error) {
			return []auth.Role{{Permissions: []string{"read", "write"}}}, nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{
		Users:   users,
		APIKeys: keyRepo,
		Roles:   roles,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	authenticated, err := svc.ValidateAPIKey(context.Background(), apiKey.Key)
	if err != nil {
		t.Fatalf("ValidateAPIKey() error = %v", err)
	}
	if authenticated.ID != user.ID {
		t.Fatalf("user id = %s", authenticated.ID)
	}

	hasPerm, err := svc.CheckPermission(context.Background(), user.ID, "write")
	if err != nil {
		t.Fatalf("CheckPermission() error = %v", err)
	}
	if !hasPerm {
		t.Fatal("expected permission true")
	}
}
