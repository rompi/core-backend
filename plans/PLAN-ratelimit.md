# Package Plan: pkg/ratelimit

## Overview

A flexible rate limiting package supporting multiple algorithms (token bucket, sliding window, fixed window) and backends (in-memory, Redis). Provides HTTP/gRPC middleware and distributed rate limiting for multi-instance deployments.

## Goals

1. **Multiple Algorithms** - Token bucket, sliding window, fixed window, leaky bucket
2. **Distributed Support** - Redis backend for multi-instance rate limiting
3. **Key Strategies** - Rate limit by IP, user ID, API key, or custom keys
4. **Middleware** - HTTP and gRPC middleware out of the box
5. **Flexible Limits** - Different limits for different endpoints/users
6. **Headers** - Standard rate limit headers (X-RateLimit-*)
7. **Zero Dependencies** - In-memory implementation uses stdlib only

## Architecture

```
pkg/ratelimit/
├── ratelimit.go          # Core Limiter interface
├── config.go             # Configuration
├── options.go            # Functional options
├── errors.go             # Custom error types
├── result.go             # Rate limit result
├── algorithm/
│   ├── algorithm.go      # Algorithm interface
│   ├── tokenbucket.go    # Token bucket algorithm
│   ├── slidingwindow.go  # Sliding window algorithm
│   ├── fixedwindow.go    # Fixed window algorithm
│   └── leakybucket.go    # Leaky bucket algorithm
├── store/
│   ├── store.go          # Store interface
│   ├── memory.go         # In-memory store
│   └── redis.go          # Redis store
├── strategy/
│   ├── strategy.go       # Key strategy interface
│   ├── ip.go             # IP-based limiting
│   ├── user.go           # User ID-based limiting
│   ├── apikey.go         # API key-based limiting
│   └── composite.go      # Multiple strategies
├── middleware/
│   ├── http.go           # HTTP middleware
│   └── grpc.go           # gRPC interceptor
├── examples/
│   ├── basic/
│   ├── redis/
│   ├── per-user/
│   └── tiered/
└── README.md
```

## Core Interfaces

```go
package ratelimit

import (
    "context"
    "time"
)

// Limiter provides rate limiting functionality
type Limiter interface {
    // Allow checks if a request is allowed
    Allow(ctx context.Context, key string) (*Result, error)

    // AllowN checks if n requests are allowed
    AllowN(ctx context.Context, key string, n int) (*Result, error)

    // Reset resets the rate limit for a key
    Reset(ctx context.Context, key string) error

    // Close releases resources
    Close() error
}

// Result contains rate limit check results
type Result struct {
    // Allowed is true if request is allowed
    Allowed bool

    // Limit is the maximum requests allowed
    Limit int

    // Remaining is requests remaining in window
    Remaining int

    // ResetAt is when the limit resets
    ResetAt time.Time

    // RetryAfter is duration until retry (if not allowed)
    RetryAfter time.Duration
}

// Algorithm implements a rate limiting algorithm
type Algorithm interface {
    // Take attempts to take n tokens
    Take(ctx context.Context, store Store, key string, limit Limit, n int) (*Result, error)

    // Name returns the algorithm name
    Name() string
}

// Store persists rate limit state
type Store interface {
    // Get retrieves counter state
    Get(ctx context.Context, key string) (*State, error)

    // Set updates counter state
    Set(ctx context.Context, key string, state *State, ttl time.Duration) error

    // Increment atomically increments counter
    Increment(ctx context.Context, key string, delta int, ttl time.Duration) (int, error)

    // Delete removes a key
    Delete(ctx context.Context, key string) error

    // Close releases resources
    Close() error
}

// Limit defines rate limit parameters
type Limit struct {
    // Rate is requests per period
    Rate int

    // Period is the time window
    Period time.Duration

    // Burst is maximum burst size (for token bucket)
    Burst int
}

// KeyStrategy extracts rate limit key from request
type KeyStrategy interface {
    // Key returns the rate limit key
    Key(ctx context.Context, r interface{}) (string, error)
}
```

## Configuration

```go
// Config holds rate limiter configuration
type Config struct {
    // Algorithm: "token_bucket", "sliding_window", "fixed_window", "leaky_bucket"
    Algorithm string `env:"RATELIMIT_ALGORITHM" default:"token_bucket"`

    // Store: "memory" or "redis"
    Store string `env:"RATELIMIT_STORE" default:"memory"`

    // Default limit
    DefaultLimit Limit

    // Key prefix for store
    KeyPrefix string `env:"RATELIMIT_KEY_PREFIX" default:"ratelimit:"`
}

type RedisConfig struct {
    // Redis URL
    URL string `env:"RATELIMIT_REDIS_URL" default:"redis://localhost:6379"`

    // Key prefix
    KeyPrefix string `env:"RATELIMIT_REDIS_PREFIX" default:"ratelimit:"`
}
```

## Algorithms

### Token Bucket

```go
// TokenBucket allows burst traffic up to bucket size
// Tokens are added at a fixed rate
type TokenBucket struct {
    rate   int           // Tokens per period
    period time.Duration // Refill period
    burst  int           // Maximum bucket size
}

func NewTokenBucket() *TokenBucket

// Pros: Allows bursts, smooth rate limiting
// Cons: More complex state management
```

### Sliding Window

```go
// SlidingWindow counts requests in a rolling time window
// More accurate than fixed window
type SlidingWindow struct {
    windowSize time.Duration
}

func NewSlidingWindow() *SlidingWindow

// Pros: Smooth distribution, no edge spikes
// Cons: Higher memory usage
```

### Fixed Window

```go
// FixedWindow counts requests in fixed time intervals
// Simple but can have edge case issues
type FixedWindow struct {
    windowSize time.Duration
}

func NewFixedWindow() *FixedWindow

// Pros: Simple, low memory
// Cons: Can allow 2x burst at window edges
```

### Leaky Bucket

```go
// LeakyBucket processes requests at a fixed rate
// Excess requests are queued or rejected
type LeakyBucket struct {
    rate       int           // Requests per period
    period     time.Duration
    bucketSize int           // Queue size
}

func NewLeakyBucket() *LeakyBucket

// Pros: Smooth output rate
// Cons: Added latency for queued requests
```

## Key Strategies

```go
// IPStrategy extracts client IP
type IPStrategy struct {
    // TrustProxy trusts X-Forwarded-For header
    TrustProxy bool
    // TrustedProxies is list of trusted proxy IPs
    TrustedProxies []string
}

func NewIPStrategy(opts ...IPOption) *IPStrategy

// UserStrategy extracts user ID from context
type UserStrategy struct {
    // ContextKey is the key for user ID in context
    ContextKey interface{}
}

func NewUserStrategy(contextKey interface{}) *UserStrategy

// APIKeyStrategy extracts API key from header
type APIKeyStrategy struct {
    // Header is the API key header name
    Header string
}

func NewAPIKeyStrategy(header string) *APIKeyStrategy

// CompositeStrategy combines multiple strategies
type CompositeStrategy struct {
    strategies []KeyStrategy
    separator  string
}

func NewCompositeStrategy(strategies ...KeyStrategy) *CompositeStrategy

// Example: "user:123:ip:192.168.1.1"
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "time"
    "github.com/user/core-backend/pkg/ratelimit"
)

func main() {
    // Create limiter with default settings
    limiter, err := ratelimit.New(ratelimit.Config{
        Algorithm: "token_bucket",
        Store:     "memory",
        DefaultLimit: ratelimit.Limit{
            Rate:   100,              // 100 requests
            Period: time.Minute,      // per minute
            Burst:  10,               // allow burst of 10
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    ctx := context.Background()

    // Check rate limit
    result, err := limiter.Allow(ctx, "user:123")
    if err != nil {
        log.Fatal(err)
    }

    if !result.Allowed {
        fmt.Printf("Rate limited. Retry after: %v\n", result.RetryAfter)
    } else {
        fmt.Printf("Allowed. Remaining: %d/%d\n", result.Remaining, result.Limit)
    }
}
```

### HTTP Middleware

```go
import (
    "github.com/user/core-backend/pkg/ratelimit"
    "github.com/user/core-backend/pkg/ratelimit/middleware"
)

func main() {
    limiter, _ := ratelimit.New(cfg)

    // Create middleware with IP-based limiting
    rateLimitMiddleware := middleware.HTTP(limiter,
        middleware.WithKeyStrategy(ratelimit.NewIPStrategy()),
        middleware.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, result *ratelimit.Result) {
            w.Header().Set("Retry-After", fmt.Sprintf("%d", int(result.RetryAfter.Seconds())))
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
        }),
    )

    mux := http.NewServeMux()
    mux.HandleFunc("/api/", handleAPI)

    // Apply middleware
    http.ListenAndServe(":8080", rateLimitMiddleware(mux))
}
```

### Distributed Rate Limiting with Redis

```go
import (
    "github.com/user/core-backend/pkg/ratelimit"
    "github.com/user/core-backend/pkg/ratelimit/store"
)

func main() {
    // Create Redis store
    redisStore, err := store.NewRedis(store.RedisConfig{
        URL:       "redis://localhost:6379",
        KeyPrefix: "myapp:ratelimit:",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create limiter with Redis
    limiter, err := ratelimit.New(
        ratelimit.Config{
            Algorithm: "sliding_window",
        },
        ratelimit.WithStore(redisStore),
        ratelimit.WithLimit(ratelimit.Limit{
            Rate:   1000,
            Period: time.Hour,
        }),
    )

    // Works across multiple instances
}
```

### Per-User Rate Limiting

```go
func main() {
    limiter, _ := ratelimit.New(cfg)

    // Different limits for different user tiers
    limits := map[string]ratelimit.Limit{
        "free":       {Rate: 100, Period: time.Hour},
        "pro":        {Rate: 1000, Period: time.Hour},
        "enterprise": {Rate: 10000, Period: time.Hour},
    }

    http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
        user := getUserFromContext(r.Context())
        limit := limits[user.Tier]

        // Create tiered limiter
        result, err := limiter.AllowWithLimit(r.Context(),
            fmt.Sprintf("user:%s", user.ID),
            limit,
        )

        if !result.Allowed {
            w.WriteHeader(http.StatusTooManyRequests)
            return
        }

        // Handle request
    })
}
```

### Endpoint-Specific Limits

```go
func main() {
    limiter, _ := ratelimit.New(cfg)

    // Different limits per endpoint
    endpointLimits := map[string]ratelimit.Limit{
        "/api/search":   {Rate: 10, Period: time.Minute},   // Expensive
        "/api/users":    {Rate: 100, Period: time.Minute},  // Normal
        "/api/health":   {Rate: 1000, Period: time.Minute}, // Lightweight
    }

    middleware := middleware.HTTP(limiter,
        middleware.WithKeyFunc(func(r *http.Request) string {
            return fmt.Sprintf("%s:%s", getIP(r), r.URL.Path)
        }),
        middleware.WithLimitFunc(func(r *http.Request) ratelimit.Limit {
            if limit, ok := endpointLimits[r.URL.Path]; ok {
                return limit
            }
            return defaultLimit
        }),
    )
}
```

### gRPC Interceptor

```go
import (
    "github.com/user/core-backend/pkg/ratelimit/middleware"
)

func main() {
    limiter, _ := ratelimit.New(cfg)

    server := grpc.NewServer(
        grpc.UnaryInterceptor(middleware.UnaryServerInterceptor(limiter,
            middleware.WithKeyStrategy(ratelimit.NewUserStrategy(userContextKey)),
        )),
    )
}
```

### Rate Limit Headers

```go
// Middleware automatically adds headers:
// X-RateLimit-Limit: 100
// X-RateLimit-Remaining: 95
// X-RateLimit-Reset: 1640000000
// Retry-After: 60 (when rate limited)

middleware := middleware.HTTP(limiter,
    middleware.WithHeaders(true), // Enable headers
)
```

### Composite Keys

```go
// Rate limit by user + endpoint
strategy := ratelimit.NewCompositeStrategy(
    ratelimit.NewUserStrategy(userKey),
    ratelimit.NewEndpointStrategy(),
)

// Key: "user:123:endpoint:/api/search"
```

## Error Handling

```go
var (
    // ErrRateLimited is returned when rate limit exceeded
    ErrRateLimited = errors.New("ratelimit: rate limit exceeded")

    // ErrStoreUnavailable is returned when store is unavailable
    ErrStoreUnavailable = errors.New("ratelimit: store unavailable")

    // ErrInvalidKey is returned for invalid rate limit keys
    ErrInvalidKey = errors.New("ratelimit: invalid key")
)

// IsRateLimited checks if error is rate limit error
func IsRateLimited(err error) bool
```

## Dependencies

- **Required:** None (in-memory uses stdlib)
- **Optional:**
  - `github.com/redis/go-redis/v9` for Redis store

## Implementation Phases

### Phase 1: Core Interface & Token Bucket
1. Define Limiter, Store interfaces
2. Implement token bucket algorithm
3. In-memory store
4. Basic result type

### Phase 2: Additional Algorithms
1. Sliding window algorithm
2. Fixed window algorithm
3. Leaky bucket algorithm

### Phase 3: Redis Store
1. Implement Redis store
2. Lua scripts for atomicity
3. Distributed rate limiting

### Phase 4: Middleware
1. HTTP middleware
2. gRPC interceptor
3. Rate limit headers

### Phase 5: Key Strategies
1. IP strategy
2. User strategy
3. Composite strategy

### Phase 6: Documentation & Examples
1. README
2. Basic example
3. Redis example
4. Tiered limits example
