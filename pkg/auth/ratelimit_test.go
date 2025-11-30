package auth

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	cfg := defaultConfig()
	cfg.RateLimitWindow = time.Second
	cfg.RateLimitMaxRequests = 2

	rl := NewRateLimiter(cfg)
	base := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	rl.setNow(func() time.Time { return base })

	for i := 0; i < cfg.RateLimitMaxRequests; i++ {
		if !rl.Allow("key") {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	if rl.Allow("key") {
		t.Fatalf("expected request to be rejected after limit")
	}

	rl.setNow(func() time.Time { return base.Add(cfg.RateLimitWindow) })
	if !rl.Allow("key") {
		t.Fatalf("expected request after window reset")
	}
}
