package gateway

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/rompi/core-backend/pkg/server"
)

// RateLimitConfig defines HTTP rate limiting configuration.
type RateLimitConfig struct {
	// Rate is requests per second.
	Rate float64

	// Burst is maximum burst size.
	Burst int

	// KeyFunc extracts the rate limit key (default: client IP).
	KeyFunc func(r *http.Request) string

	// SkipFunc determines whether to skip rate limiting.
	SkipFunc func(r *http.Request) bool

	// ExceededHandler handles rate limit exceeded (default: 429).
	ExceededHandler http.HandlerFunc

	// Headers enables X-RateLimit-* response headers.
	Headers bool
}

// DefaultRateLimitConfig returns default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Rate:    20,
		Burst:   40,
		Headers: true,
	}
}

// rateLimiterStore stores rate limiters per key.
type httpRateLimiterStore struct {
	limiters map[string]*httpRateLimiterEntry
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	expiry   time.Duration
}

type httpRateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newHTTPRateLimiterStore(r float64, burst int, expiry time.Duration) *httpRateLimiterStore {
	store := &httpRateLimiterStore{
		limiters: make(map[string]*httpRateLimiterEntry),
		rate:     rate.Limit(r),
		burst:    burst,
		expiry:   expiry,
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

func (s *httpRateLimiterStore) getLimiter(key string) *rate.Limiter {
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
	s.limiters[key] = &httpRateLimiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

func (s *httpRateLimiterStore) cleanup() {
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

// RateLimitMiddleware creates rate limiting middleware.
func RateLimitMiddleware(config RateLimitConfig) Middleware {
	store := newHTTPRateLimiterStore(config.Rate, config.Burst, 3*time.Minute)

	keyFunc := config.KeyFunc
	if keyFunc == nil {
		keyFunc = defaultHTTPKeyFunc
	}

	exceededHandler := config.ExceededHandler
	if exceededHandler == nil {
		exceededHandler = func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if we should skip
			if config.SkipFunc != nil && config.SkipFunc(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := keyFunc(r)
			limiter := store.getLimiter(key)

			if !limiter.Allow() {
				if config.Headers {
					w.Header().Set("X-RateLimit-Limit", strconv.FormatFloat(config.Rate, 'f', -1, 64))
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("Retry-After", "1")
				}
				exceededHandler(w, r)
				return
			}

			if config.Headers {
				w.Header().Set("X-RateLimit-Limit", strconv.FormatFloat(config.Rate, 'f', -1, 64))
				// Note: This is an approximation
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(int(limiter.Tokens())))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitWithRate creates a simple rate limiter.
func RateLimitWithRate(rps float64, burst int) Middleware {
	return RateLimitMiddleware(RateLimitConfig{
		Rate:  rps,
		Burst: burst,
	})
}

// --- Presets ---

// RateLimitStrict - 5 req/s (login, password reset).
func RateLimitStrict() Middleware {
	return RateLimitWithRate(5, 10)
}

// RateLimitAuth - 10 req/s (token refresh).
func RateLimitAuth() Middleware {
	return RateLimitWithRate(10, 20)
}

// RateLimitNormal - 30 req/s (general API).
func RateLimitNormal() Middleware {
	return RateLimitWithRate(30, 60)
}

// RateLimitRelaxed - 100 req/s (read-heavy).
func RateLimitRelaxed() Middleware {
	return RateLimitWithRate(100, 200)
}

// --- Per-Path Rate Limiting ---

// PerPathRateLimits allows different limits per HTTP path.
type PerPathRateLimits map[string]RateLimitConfig

// RateLimitPerPath creates middleware with per-path limits.
// Key format: "METHOD /path" (e.g., "POST /api/v1/login")
func RateLimitPerPath(limits PerPathRateLimits, defaultConfig RateLimitConfig) Middleware {
	stores := make(map[string]*httpRateLimiterStore)
	for path, cfg := range limits {
		stores[path] = newHTTPRateLimiterStore(cfg.Rate, cfg.Burst, 3*time.Minute)
	}

	defaultStore := newHTTPRateLimiterStore(defaultConfig.Rate, defaultConfig.Burst, 3*time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pathKey := r.Method + " " + r.URL.Path

			var store *httpRateLimiterStore
			var config RateLimitConfig

			if s, ok := stores[pathKey]; ok {
				store = s
				config = limits[pathKey]
			} else {
				store = defaultStore
				config = defaultConfig
			}

			// Check if we should skip
			if config.SkipFunc != nil && config.SkipFunc(r) {
				next.ServeHTTP(w, r)
				return
			}

			keyFunc := config.KeyFunc
			if keyFunc == nil {
				keyFunc = defaultHTTPKeyFunc
			}

			key := keyFunc(r)
			limiter := store.getLimiter(key)

			if !limiter.Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// defaultHTTPKeyFunc returns the client IP as the rate limit key.
func defaultHTTPKeyFunc(r *http.Request) string {
	// Try X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return xff
	}

	// Try X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// UserKeyFunc returns a key function that uses the user ID if authenticated.
func UserHTTPKeyFunc() func(r *http.Request) string {
	return func(r *http.Request) string {
		if user := GetUser(r); user != nil {
			return "user:" + user.GetID()
		}
		return "ip:" + defaultHTTPKeyFunc(r)
	}
}
