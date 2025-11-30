package auth

import (
	"testing"
	"time"
)

func TestTokenManager_GenerateAndValidate(t *testing.T) {
	cfg := defaultConfig()
	cfg.JWTSecret = "super-secret"
	cfg.JWTExpirationDuration = time.Minute

	user := &User{ID: "user-1", Email: "test@rompi.com"}
	manager := NewTokenManager(cfg)

	token, expiresAt, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatalf("expiration should be in future: %v", expiresAt)
	}

	claims, err := manager.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("claims user id = %s, want %s", claims.UserID, user.ID)
	}
}
