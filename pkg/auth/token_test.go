package auth

import (
	"fmt"
	"sync"
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

func TestTokenManager_Generate_UniqueTokens(t *testing.T) {
	cfg := defaultConfig()
	cfg.JWTSecret = "super-secret"
	cfg.JWTExpirationDuration = time.Minute

	manager := NewTokenManager(cfg)
	user := &User{ID: "user-1", Email: "test@rompi.com"}

	tokens := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		token, _, err := manager.Generate(user)
		if err != nil {
			t.Fatalf("Generate() iteration %d error = %v", i, err)
		}
		if _, exists := tokens[token]; exists {
			t.Fatalf("duplicate token on iteration %d", i)
		}
		tokens[token] = struct{}{}
	}
}

func TestTokenManager_Generate_UniqueTokensConcurrent(t *testing.T) {
	cfg := defaultConfig()
	cfg.JWTSecret = "super-secret"
	cfg.JWTExpirationDuration = time.Minute

	manager := NewTokenManager(cfg)
	user := &User{ID: "user-1", Email: "test@rompi.com"}

	const goroutines = 20
	var mu sync.Mutex
	tokens := make(map[string]struct{})
	errs := make(chan error, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			token, _, err := manager.Generate(user)
			if err != nil {
				errs <- err
				return
			}
			mu.Lock()
			if _, exists := tokens[token]; exists {
				mu.Unlock()
				errs <- fmt.Errorf("duplicate token: %s", token)
				return
			}
			tokens[token] = struct{}{}
			mu.Unlock()
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatal(err)
	}
}

func TestTokenManager_Generate_ContainsJTI(t *testing.T) {
	cfg := defaultConfig()
	cfg.JWTSecret = "super-secret"
	cfg.JWTExpirationDuration = time.Minute

	manager := NewTokenManager(cfg)
	user := &User{ID: "user-1", Email: "test@rompi.com"}

	token, _, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := manager.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.ID == "" {
		t.Fatal("expected jti claim to be set")
	}
}
