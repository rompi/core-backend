package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rompi/core-backend/pkg/auth"
)

func main() {
	cfg := &auth.Config{
		JWTSecret:              "custom-secret",
		JWTExpirationDuration:  24 * time.Hour,
		JWTIssuer:              "rompi-auth",
		PasswordMinLength:      10,
		PasswordRequireUpper:   true,
		PasswordRequireLower:   true,
		PasswordRequireNumber:  true,
		PasswordRequireSpecial: false, // require special via custom rule
		BcryptCost:             4,
		MaxFailedAttempts:      5,
		LockoutDuration:        15 * time.Minute,
		RateLimitWindow:        time.Minute,
		RateLimitMaxRequests:   20,
		ResetTokenLength:       32,
		ResetTokenExpiration:   time.Hour,
		DefaultLanguage:        "en",
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	password := "RompiStrong1"
	if err := auth.ValidatePassword(password, cfg); err != nil {
		log.Fatalf("default validation: %v", err)
	}

	if !strings.Contains(password, "Rompi") {
		log.Fatalf("custom validation: password must include 'Rompi'")
	}

	svc, err := auth.NewService(cfg, auth.Repositories{Users: &noopUserRepo{}})
	if err != nil {
		log.Fatalf("new service: %v", err)
	}

	if _, err := svc.Register(context.Background(), auth.RegisterRequest{Email: "custom@example.com", Password: password}); err != nil {
		log.Fatalf("register: %v", err)
	}
	fmt.Println("Registration passed both default and custom validators.")
}

// noopUserRepo satisfies the UserRepository interface for example purposes.
type noopUserRepo struct{}

func (noopUserRepo) Create(ctx context.Context, user *auth.User) error { return nil }
func (noopUserRepo) GetByID(ctx context.Context, id string) (*auth.User, error) {
	return nil, auth.ErrUserNotFound
}
func (noopUserRepo) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	return nil, auth.ErrUserNotFound
}
func (noopUserRepo) Update(ctx context.Context, user *auth.User) error                { return nil }
func (noopUserRepo) Delete(ctx context.Context, id string) error                      { return nil }
func (noopUserRepo) IncrementFailedAttempts(ctx context.Context, userID string) error { return nil }
func (noopUserRepo) ResetFailedAttempts(ctx context.Context, userID string) error     { return nil }
func (noopUserRepo) LockAccount(ctx context.Context, userID string) error             { return nil }
func (noopUserRepo) UnlockAccount(ctx context.Context, userID string) error           { return nil }
