package auth

import (
	"sync"
	"time"
)

// RateLimiter enforces request counts per key within a sliding window.
type RateLimiter struct {
	window time.Duration
	max    int

	mu      sync.Mutex
	buckets map[string]*rateBucket
	now     func() time.Time
}

type rateBucket struct {
	count int
	start time.Time
}

// NewRateLimiter creates a rate limiter configured from cfg.
func NewRateLimiter(cfg *Config) *RateLimiter {
	return &RateLimiter{
		window:  cfg.RateLimitWindow,
		max:     cfg.RateLimitMaxRequests,
		buckets: make(map[string]*rateBucket),
		now:     time.Now,
	}
}

// Allow reports whether a key can proceed. Returns false when the limit is reached.
func (rl *RateLimiter) Allow(key string) bool {
	return rl.allowAt(key, rl.now())
}

func (rl *RateLimiter) allowAt(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, ok := rl.buckets[key]
	if !ok || now.Sub(bucket.start) >= rl.window {
		bucket = &rateBucket{start: now}
		rl.buckets[key] = bucket
	}

	if bucket.count >= rl.max {
		return false
	}

	bucket.count++
	return true
}

// setNow is a test hook that overrides the time source.
func (rl *RateLimiter) setNow(fn func() time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.now = fn
}
