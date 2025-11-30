package auth

import (
	"testing"
	"time"
)

func BenchmarkHashPassword(b *testing.B) {
	cfg := defaultConfig()
	password := "BenchPass1!"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := HashPassword(password, cfg.BcryptCost); err != nil {
			b.Fatalf("hash: %v", err)
		}
	}
}

func BenchmarkTokenManagerGenerate(b *testing.B) {
	cfg := defaultConfig()
	cfg.JWTSecret = "bench"
	cfg.JWTExpirationDuration = time.Minute
	manager := NewTokenManager(cfg)
	user := &User{ID: "bench-user", Email: "bench@example.com"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, _, err := manager.Generate(user); err != nil {
			b.Fatalf("generate: %v", err)
		}
	}
}
