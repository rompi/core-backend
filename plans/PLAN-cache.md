# Package Plan: pkg/cache

## Overview

A unified caching layer providing a consistent interface for both in-memory and distributed caching (Redis). Supports TTL, namespacing, and serialization with zero coupling to other packages.

## Goals

1. **Unified Interface** - Single API for all cache backends
2. **Multiple Backends** - In-memory (local) and Redis (distributed)
3. **Type Safety** - Generic methods for type-safe get/set operations
4. **TTL Support** - Automatic expiration with configurable TTL
5. **Namespacing** - Key prefixing for multi-tenant applications
6. **Serialization** - Pluggable serializers (JSON, msgpack, gob)
7. **Zero Dependencies** - Core functionality uses stdlib only; Redis is optional

## Architecture

```
pkg/cache/
├── cache.go           # Core Cache interface
├── config.go          # Configuration with env support
├── options.go         # Functional options
├── memory.go          # In-memory implementation
├── memory_test.go
├── redis.go           # Redis implementation
├── redis_test.go
├── serializer.go      # Serialization interface
├── serializer_json.go # JSON serializer
├── serializer_gob.go  # Gob serializer
├── errors.go          # Custom error types
├── stats.go           # Cache statistics
├── examples/
│   ├── basic/
│   ├── redis/
│   └── with-namespace/
└── README.md
```

## Core Interface

```go
package cache

import (
    "context"
    "time"
)

// Cache defines the unified caching interface
type Cache interface {
    // Get retrieves a value by key. Returns ErrNotFound if key doesn't exist.
    Get(ctx context.Context, key string) ([]byte, error)

    // Set stores a value with optional TTL. TTL of 0 means no expiration.
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

    // Delete removes a key from the cache
    Delete(ctx context.Context, key string) error

    // Exists checks if a key exists
    Exists(ctx context.Context, key string) (bool, error)

    // Clear removes all keys (or all keys in namespace)
    Clear(ctx context.Context) error

    // Close releases resources
    Close() error
}

// TypedCache provides type-safe operations using generics
type TypedCache[T any] interface {
    Get(ctx context.Context, key string) (T, error)
    Set(ctx context.Context, key string, value T, ttl time.Duration) error
    GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (T, error)) (T, error)
}

// Stats provides cache statistics
type Stats struct {
    Hits       int64
    Misses     int64
    Size       int64
    Evictions  int64
}

// StatsProvider is implemented by caches that support statistics
type StatsProvider interface {
    Stats() Stats
}
```

## Configuration

```go
// Config holds cache configuration
type Config struct {
    // Backend type: "memory" or "redis"
    Backend string `env:"CACHE_BACKEND" default:"memory"`

    // Namespace prefix for all keys
    Namespace string `env:"CACHE_NAMESPACE" default:""`

    // Default TTL for entries (0 = no expiration)
    DefaultTTL time.Duration `env:"CACHE_DEFAULT_TTL" default:"1h"`

    // Serializer type: "json" or "gob"
    Serializer string `env:"CACHE_SERIALIZER" default:"json"`

    // Memory-specific settings
    Memory MemoryConfig

    // Redis-specific settings
    Redis RedisConfig
}

type MemoryConfig struct {
    // Maximum number of items (0 = unlimited)
    MaxItems int `env:"CACHE_MEMORY_MAX_ITEMS" default:"10000"`

    // Cleanup interval for expired items
    CleanupInterval time.Duration `env:"CACHE_MEMORY_CLEANUP_INTERVAL" default:"5m"`
}

type RedisConfig struct {
    // Redis connection string
    URL string `env:"CACHE_REDIS_URL" default:"redis://localhost:6379"`

    // Connection pool size
    PoolSize int `env:"CACHE_REDIS_POOL_SIZE" default:"10"`

    // Connection timeout
    ConnTimeout time.Duration `env:"CACHE_REDIS_CONN_TIMEOUT" default:"5s"`

    // Read timeout
    ReadTimeout time.Duration `env:"CACHE_REDIS_READ_TIMEOUT" default:"3s"`

    // Write timeout
    WriteTimeout time.Duration `env:"CACHE_REDIS_WRITE_TIMEOUT" default:"3s"`
}
```

## In-Memory Implementation

```go
// Memory implements Cache using sync.Map with TTL support
type Memory struct {
    data       sync.Map
    namespace  string
    defaultTTL time.Duration
    maxItems   int
    stats      Stats
    cleanupMu  sync.Mutex
    stopCh     chan struct{}
}

// item represents a cached item with expiration
type item struct {
    value     []byte
    expiresAt time.Time
}

// New creates a new in-memory cache
func NewMemory(opts ...Option) *Memory

// Options
func WithNamespace(ns string) Option
func WithDefaultTTL(ttl time.Duration) Option
func WithMaxItems(max int) Option
func WithCleanupInterval(interval time.Duration) Option
```

### Memory Cache Features

1. **Thread-safe** using `sync.Map`
2. **Automatic cleanup** goroutine for expired items
3. **LRU eviction** when max items reached
4. **Statistics tracking** for hits/misses
5. **Zero allocations** on hot path where possible

## Redis Implementation

```go
// Redis implements Cache using Redis as backend
type Redis struct {
    client     RedisClient  // Interface for redis commands
    namespace  string
    defaultTTL time.Duration
    serializer Serializer
}

// RedisClient is the interface for Redis operations
// This allows using any Redis client library
type RedisClient interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Del(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, keys ...string) (int64, error)
    FlushDB(ctx context.Context) error
    Close() error
}

// NewRedis creates a new Redis cache
func NewRedis(client RedisClient, opts ...Option) *Redis
```

### Redis Features

1. **Connection pooling** via provided client
2. **Automatic serialization** of complex types
3. **Pipeline support** for batch operations
4. **Pub/Sub** for cache invalidation (optional)
5. **Cluster support** via client implementation

## Serialization

```go
// Serializer defines the interface for value serialization
type Serializer interface {
    Marshal(v interface{}) ([]byte, error)
    Unmarshal(data []byte, v interface{}) error
}

// JSONSerializer uses encoding/json
type JSONSerializer struct{}

// GobSerializer uses encoding/gob
type GobSerializer struct{}

// MsgpackSerializer uses msgpack (optional dependency)
type MsgpackSerializer struct{}
```

## Typed Cache Wrapper

```go
// Typed wraps a Cache to provide type-safe operations
func Typed[T any](c Cache, serializer Serializer) TypedCache[T]

// Example usage:
userCache := cache.Typed[User](redisCache, cache.JSONSerializer{})
user, err := userCache.Get(ctx, "user:123")
```

## Error Handling

```go
var (
    // ErrNotFound is returned when a key doesn't exist
    ErrNotFound = errors.New("cache: key not found")

    // ErrClosed is returned when operating on a closed cache
    ErrClosed = errors.New("cache: cache is closed")

    // ErrSerialize is returned on serialization failure
    ErrSerialize = errors.New("cache: serialization failed")

    // ErrConnection is returned on connection issues (Redis)
    ErrConnection = errors.New("cache: connection failed")
)

// IsNotFound checks if error is ErrNotFound
func IsNotFound(err error) bool
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "time"

    "github.com/user/core-backend/pkg/cache"
)

func main() {
    // Create in-memory cache
    c := cache.NewMemory(
        cache.WithDefaultTTL(5 * time.Minute),
        cache.WithMaxItems(1000),
    )
    defer c.Close()

    ctx := context.Background()

    // Set a value
    err := c.Set(ctx, "key", []byte("value"), 0) // 0 = use default TTL

    // Get a value
    data, err := c.Get(ctx, "key")
    if cache.IsNotFound(err) {
        // Handle cache miss
    }

    // Delete
    err = c.Delete(ctx, "key")
}
```

### Typed Cache

```go
type User struct {
    ID    string
    Name  string
    Email string
}

func main() {
    c := cache.NewMemory()
    userCache := cache.Typed[User](c, cache.JSONSerializer{})

    ctx := context.Background()

    // Type-safe set
    user := User{ID: "123", Name: "John", Email: "john@example.com"}
    err := userCache.Set(ctx, "user:123", user, time.Hour)

    // Type-safe get
    cachedUser, err := userCache.Get(ctx, "user:123")

    // Get or set pattern
    user, err = userCache.GetOrSet(ctx, "user:456", time.Hour, func() (User, error) {
        // Called on cache miss
        return fetchUserFromDB("456")
    })
}
```

### Redis Cache

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/user/core-backend/pkg/cache"
)

func main() {
    // Create Redis client (using go-redis)
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Wrap with cache interface
    c := cache.NewRedis(
        cache.NewGoRedisAdapter(rdb),
        cache.WithNamespace("myapp"),
        cache.WithDefaultTTL(time.Hour),
    )
    defer c.Close()

    // Use same interface as memory cache
    ctx := context.Background()
    err := c.Set(ctx, "session:abc", sessionData, 24*time.Hour)
}
```

### Namespaced Cache

```go
// Create namespaced caches for multi-tenant app
userCache := cache.NewMemory(cache.WithNamespace("tenant1:users"))
sessionCache := cache.NewMemory(cache.WithNamespace("tenant1:sessions"))

// Keys are automatically prefixed
userCache.Set(ctx, "123", data, 0) // Stored as "tenant1:users:123"
```

## Testing

```go
// Mock cache for testing
type MockCache struct {
    GetFunc    func(ctx context.Context, key string) ([]byte, error)
    SetFunc    func(ctx context.Context, key string, value []byte, ttl time.Duration) error
    DeleteFunc func(ctx context.Context, key string) error
    ExistsFunc func(ctx context.Context, key string) (bool, error)
    ClearFunc  func(ctx context.Context) error
    CloseFunc  func() error
}

// Test helpers
func NewTestCache() *MockCache
```

## Health Check

```go
// HealthChecker returns a health check function
func (c *Redis) HealthCheck() func(ctx context.Context) error {
    return func(ctx context.Context) error {
        return c.client.Ping(ctx).Err()
    }
}
```

## Metrics Integration

```go
// Hook interface for observability
type Hook interface {
    BeforeGet(ctx context.Context, key string)
    AfterGet(ctx context.Context, key string, hit bool, err error)
    BeforeSet(ctx context.Context, key string)
    AfterSet(ctx context.Context, key string, err error)
}

// WithHook adds observability hooks
func WithHook(hook Hook) Option
```

## Dependencies

- **Required:** None (stdlib only for memory cache)
- **Optional:**
  - `github.com/redis/go-redis/v9` for Redis (user provides client)

## Test Coverage Requirements

- Unit tests for all public functions
- Integration tests for Redis (with testcontainers)
- Benchmark tests for hot paths
- Race condition tests with `-race` flag
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & Memory Cache
1. Define Cache interface
2. Implement Memory cache with TTL
3. Add serializers (JSON, Gob)
4. Write comprehensive tests

### Phase 2: Redis Implementation
1. Define RedisClient interface
2. Implement Redis cache
3. Add go-redis adapter
4. Integration tests with testcontainers

### Phase 3: Advanced Features
1. Typed cache wrapper
2. Statistics tracking
3. Health check support
4. Observability hooks

### Phase 4: Documentation & Examples
1. README with full documentation
2. Basic usage example
3. Redis example
4. Namespacing example
