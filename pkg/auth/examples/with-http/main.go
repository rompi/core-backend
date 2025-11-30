package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rompi/core-backend/pkg/auth"
	"github.com/rompi/core-backend/pkg/auth/testutil"
)

func main() {
	cfg := &auth.Config{
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
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	demoUser := &auth.User{ID: "demo", Email: "demo@example.com"}

	users := &testutil.MockUserRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*auth.User, error) {
			if email == demoUser.Email {
				return demoUser, nil
			}
			return nil, auth.ErrUserNotFound
		},
		GetByIDFunc: func(ctx context.Context, id string) (*auth.User, error) {
			if id == demoUser.ID {
				return demoUser, nil
			}
			return nil, auth.ErrUserNotFound
		},
	}
	roles := &testutil.MockRoleRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) ([]auth.Role, error) {
			if userID == demoUser.ID {
				return []auth.Role{{Name: "admin"}}, nil
			}
			return nil, nil
		},
	}

	svc, err := auth.NewService(cfg, auth.Repositories{Users: users, Roles: roles})
	if err != nil {
		log.Fatalf("failed to create auth service: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/protected", svc.Middleware()(svc.RequireRole("admin")(http.HandlerFunc(profileHandler))))

	handler := svc.RateLimitMiddleware()(mux)
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	log.Println("auth example running on :8080")
	log.Fatal(srv.ListenAndServe())
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "Hello %s", user.Email)
}
