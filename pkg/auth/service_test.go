package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rompi/core-backend/pkg/auth"
	"github.com/rompi/core-backend/pkg/auth/testutil"
)

func newTestConfig() *auth.Config {
	return &auth.Config{
		JWTSecret:              "secret",
		JWTExpirationDuration:  time.Hour,
		JWTIssuer:              "rompi-auth",
		PasswordMinLength:      8,
		PasswordRequireUpper:   true,
		PasswordRequireLower:   true,
		PasswordRequireNumber:  true,
		PasswordRequireSpecial: true,
		BcryptCost:             4,
		MaxFailedAttempts:      5,
		LockoutDuration:        time.Minute,
		RateLimitWindow:        time.Second,
		RateLimitMaxRequests:   5,
		ResetTokenLength:       32,
		ResetTokenExpiration:   time.Minute,
		DefaultLanguage:        "en",
	}
}

func TestService_RegisterSuccess(t *testing.T) {
	cfg := newTestConfig()

	created := false
	users := &testutil.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*auth.User, error) {
			return nil, auth.ErrUserNotFound
		},
		CreateFunc: func(ctx context.Context, user *auth.User) error {
			created = true
			if user.ID == "" {
				user.ID = uuid.NewString()
			}
			return nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{Users: users})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	req := auth.RegisterRequest{Email: "Test@Example.com", Password: "Str0ng!Pass", Language: "id"}
	user, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("email = %s", user.Email)
	}
	if !created {
		t.Fatal("expected user created")
	}
}

func TestService_RegisterConflict(t *testing.T) {
	cfg := newTestConfig()

	users := &testutil.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*auth.User, error) {
			return &auth.User{Email: email}, nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{Users: users})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if _, err := svc.Register(context.Background(), auth.RegisterRequest{Email: "exists@example.com", Password: "Str0ng!Pass"}); !errors.Is(err, auth.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestService_LoginSuccess(t *testing.T) {
	cfg := newTestConfig()
	hash, err := auth.HashPassword("Str0ng!Pass", cfg.BcryptCost)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	sessionCreated := false
	users := &testutil.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*auth.User, error) {
			return &auth.User{ID: "user-1", Email: email, PasswordHash: hash}, nil
		},
		ResetFailedAttemptsFunc: func(ctx context.Context, userID string) error {
			return nil
		},
	}
	sessions := &testutil.MockSessionRepository{
		CreateFunc: func(ctx context.Context, session *auth.Session) error {
			sessionCreated = true
			if session.Token == "" {
				return errors.New("token missing")
			}
			return nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{Users: users, Sessions: sessions})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	resp, err := svc.Login(context.Background(), auth.LoginRequest{Email: "user@example.com", Password: "Str0ng!Pass"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token")
	}
	if !sessionCreated {
		t.Fatal("expected session created")
	}
}

func TestService_LoginFailure(t *testing.T) {
	cfg := newTestConfig()

	users := &testutil.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*auth.User, error) {
			return &auth.User{ID: "user-1", Email: email, PasswordHash: "invalid"}, nil
		},
		IncrementFailedAttemptsFunc: func(ctx context.Context, userID string) error {
			return nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{Users: users})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if _, err := svc.Login(context.Background(), auth.LoginRequest{Email: "user@example.com", Password: "wrong"}); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
