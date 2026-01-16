package grpc

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// RateLimitConfig defines per-method rate limiting configuration.
type RateLimitConfig struct {
	// Rate is requests per second.
	Rate float64

	// Burst is maximum burst size.
	Burst int

	// KeyFunc extracts the rate limit key from context (default: peer IP).
	KeyFunc func(ctx context.Context, method string) string

	// SkipFunc determines whether to skip rate limiting.
	SkipFunc func(ctx context.Context, method string) bool
}

// DefaultRateLimitConfig returns a default rate limit config.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Rate:  20,
		Burst: 40,
	}
}

// rateLimiterStore stores rate limiters per key.
type rateLimiterStore struct {
	limiters map[string]*rateLimiterEntry
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	expiry   time.Duration
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newRateLimiterStore(r float64, burst int, expiry time.Duration) *rateLimiterStore {
	store := &rateLimiterStore{
		limiters: make(map[string]*rateLimiterEntry),
		rate:     rate.Limit(r),
		burst:    burst,
		expiry:   expiry,
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

func (s *rateLimiterStore) getLimiter(key string) *rate.Limiter {
	s.mu.RLock()
	entry, exists := s.limiters[key]
	s.mu.RUnlock()

	if exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := s.limiters[key]; exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(s.rate, s.burst)
	s.limiters[key] = &rateLimiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

func (s *rateLimiterStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for key, entry := range s.limiters {
			if time.Since(entry.lastSeen) > s.expiry {
				delete(s.limiters, key)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimitInterceptor creates a rate limiting interceptor.
func RateLimitInterceptor(config RateLimitConfig) grpc.UnaryServerInterceptor {
	store := newRateLimiterStore(config.Rate, config.Burst, 3*time.Minute)

	keyFunc := config.KeyFunc
	if keyFunc == nil {
		keyFunc = defaultKeyFunc
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if we should skip
		if config.SkipFunc != nil && config.SkipFunc(ctx, info.FullMethod) {
			return handler(ctx, req)
		}

		key := keyFunc(ctx, info.FullMethod)
		limiter := store.getLimiter(key)

		if !limiter.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// RateLimitStreamInterceptor creates a streaming rate limiting interceptor.
func RateLimitStreamInterceptor(config RateLimitConfig) grpc.StreamServerInterceptor {
	store := newRateLimiterStore(config.Rate, config.Burst, 3*time.Minute)

	keyFunc := config.KeyFunc
	if keyFunc == nil {
		keyFunc = defaultKeyFunc
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// Check if we should skip
		if config.SkipFunc != nil && config.SkipFunc(ctx, info.FullMethod) {
			return handler(srv, ss)
		}

		key := keyFunc(ctx, info.FullMethod)
		limiter := store.getLimiter(key)

		if !limiter.Allow() {
			return status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(srv, ss)
	}
}

// --- Preset Rate Limiters ---

// RateLimitStrict - 5 req/s, burst 10 (login, password reset).
func RateLimitStrict() grpc.UnaryServerInterceptor {
	return RateLimitInterceptor(RateLimitConfig{Rate: 5, Burst: 10})
}

// RateLimitAuth - 10 req/s, burst 20 (token refresh).
func RateLimitAuth() grpc.UnaryServerInterceptor {
	return RateLimitInterceptor(RateLimitConfig{Rate: 10, Burst: 20})
}

// RateLimitNormal - 30 req/s, burst 60 (general API).
func RateLimitNormal() grpc.UnaryServerInterceptor {
	return RateLimitInterceptor(RateLimitConfig{Rate: 30, Burst: 60})
}

// RateLimitRelaxed - 100 req/s, burst 200 (read-heavy).
func RateLimitRelaxed() grpc.UnaryServerInterceptor {
	return RateLimitInterceptor(RateLimitConfig{Rate: 100, Burst: 200})
}

// --- Per-Method Rate Limiting ---

// PerMethodRateLimits allows different limits per gRPC method.
type PerMethodRateLimits map[string]RateLimitConfig

// RateLimitPerMethod creates an interceptor with per-method limits.
func RateLimitPerMethod(limits PerMethodRateLimits, defaultConfig RateLimitConfig) grpc.UnaryServerInterceptor {
	stores := make(map[string]*rateLimiterStore)
	for method, cfg := range limits {
		stores[method] = newRateLimiterStore(cfg.Rate, cfg.Burst, 3*time.Minute)
	}

	defaultStore := newRateLimiterStore(defaultConfig.Rate, defaultConfig.Burst, 3*time.Minute)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var store *rateLimiterStore
		var config RateLimitConfig

		if s, ok := stores[info.FullMethod]; ok {
			store = s
			config = limits[info.FullMethod]
		} else {
			store = defaultStore
			config = defaultConfig
		}

		// Check if we should skip
		if config.SkipFunc != nil && config.SkipFunc(ctx, info.FullMethod) {
			return handler(ctx, req)
		}

		keyFunc := config.KeyFunc
		if keyFunc == nil {
			keyFunc = defaultKeyFunc
		}

		key := keyFunc(ctx, info.FullMethod)
		limiter := store.getLimiter(key)

		if !limiter.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// defaultKeyFunc returns the peer IP as the rate limit key.
func defaultKeyFunc(ctx context.Context, method string) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "unknown"
}

// MethodKeyFunc returns a key function that includes the method name.
func MethodKeyFunc() func(ctx context.Context, method string) string {
	return func(ctx context.Context, method string) string {
		ip := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			ip = p.Addr.String()
		}
		return ip + ":" + method
	}
}

// UserKeyFunc returns a key function that uses the user ID if authenticated.
func UserKeyFunc() func(ctx context.Context, method string) string {
	return func(ctx context.Context, method string) string {
		if user := GetUser(ctx); user != nil {
			return "user:" + user.GetID()
		}
		if p, ok := peer.FromContext(ctx); ok {
			return "ip:" + p.Addr.String()
		}
		return "unknown"
	}
}
