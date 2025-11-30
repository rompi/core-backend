package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rompi/core-backend/pkg/auth"
	"github.com/rompi/core-backend/pkg/auth/testutil"
)

func main() {
	cfg := basicConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	var stored *auth.User
	users := &testutil.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*auth.User, error) {
			if stored == nil || stored.Email != email {
				return nil, auth.ErrUserNotFound
			}
			return stored, nil
		},
		CreateFunc: func(ctx context.Context, user *auth.User) error {
			stored = user
			return nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{Users: users})
	if err != nil {
		log.Fatalf("setting up service: %v", err)
	}

	ctx := context.Background()
	req := auth.RegisterRequest{Email: "basic@example.com", Password: "BasicPass1!", Language: "en"}
	user, err := svc.Register(ctx, req)
	if err != nil {
		log.Fatalf("register: %v", err)
	}
	fmt.Printf("registered user: %s\n", user.Email)

	resp, err := svc.Login(ctx, auth.LoginRequest{Email: user.Email, Password: req.Password})
	if err != nil {
		log.Fatalf("login: %v", err)
	}
	fmt.Printf("issued token expires at: %s\n", resp.ExpiresAt.Format(time.RFC3339))
}

func basicConfig() *auth.Config {
	return &auth.Config{
		JWTSecret:              "example-secret",
		JWTExpirationDuration:  24 * time.Hour,
		JWTIssuer:              "rompi-auth",
		PasswordMinLength:      8,
		PasswordRequireUpper:   true,
		PasswordRequireLower:   true,
		PasswordRequireNumber:  true,
		PasswordRequireSpecial: true,
		BcryptCost:             4,
		MaxFailedAttempts:      5,
		LockoutDuration:        15 * time.Minute,
		RateLimitWindow:        time.Minute,
		RateLimitMaxRequests:   20,
		ResetTokenLength:       32,
		ResetTokenExpiration:   time.Hour,
		DefaultLanguage:        "en",
	}
}
