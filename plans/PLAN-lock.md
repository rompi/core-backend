# Package Plan: pkg/lock

## Overview

A distributed locking package for coordinating access to shared resources across multiple instances. Supports multiple backends (Redis, PostgreSQL, etcd) with features like lock extension, fencing tokens, and deadlock prevention.

## Goals

1. **Multiple Backends** - Redis, PostgreSQL, etcd, in-memory
2. **Lock Types** - Exclusive locks, read-write locks, semaphores
3. **Automatic Extension** - Auto-extend locks for long operations
4. **Fencing Tokens** - Prevent stale lock holders from causing issues
5. **Deadlock Prevention** - TTL-based automatic expiration
6. **Observability** - Lock metrics and tracing
7. **Context Support** - Cancellation and timeout via context

## Architecture

```
pkg/lock/
├── lock.go               # Core Lock interface
├── config.go             # Configuration
├── options.go            # Functional options
├── errors.go             # Custom error types
├── provider/
│   ├── provider.go       # Provider interface
│   ├── memory.go         # In-memory (single instance)
│   ├── redis.go          # Redis (Redlock)
│   ├── postgres.go       # PostgreSQL advisory locks
│   └── etcd.go           # etcd distributed locks
├── types/
│   ├── mutex.go          # Exclusive lock
│   ├── rwmutex.go        # Read-write lock
│   └── semaphore.go      # Counting semaphore
├── examples/
│   ├── basic/
│   ├── redis/
│   ├── rwlock/
│   └── semaphore/
└── README.md
```

## Core Interfaces

```go
package lock

import (
    "context"
    "time"
)

// Locker provides distributed locking
type Locker interface {
    // NewMutex creates an exclusive lock
    NewMutex(name string, opts ...Option) Mutex

    // NewRWMutex creates a read-write lock
    NewRWMutex(name string, opts ...Option) RWMutex

    // NewSemaphore creates a counting semaphore
    NewSemaphore(name string, limit int, opts ...Option) Semaphore

    // Close releases resources
    Close() error
}

// Mutex is an exclusive lock
type Mutex interface {
    // Lock acquires the lock
    Lock(ctx context.Context) error

    // TryLock attempts to acquire without blocking
    TryLock(ctx context.Context) (bool, error)

    // Unlock releases the lock
    Unlock(ctx context.Context) error

    // Extend extends the lock TTL
    Extend(ctx context.Context, ttl time.Duration) error

    // Token returns the fencing token
    Token() string
}

// RWMutex is a read-write lock
type RWMutex interface {
    // Lock acquires write lock
    Lock(ctx context.Context) error

    // RLock acquires read lock
    RLock(ctx context.Context) error

    // Unlock releases write lock
    Unlock(ctx context.Context) error

    // RUnlock releases read lock
    RUnlock(ctx context.Context) error

    // TryLock attempts write lock without blocking
    TryLock(ctx context.Context) (bool, error)

    // TryRLock attempts read lock without blocking
    TryRLock(ctx context.Context) (bool, error)
}

// Semaphore limits concurrent access
type Semaphore interface {
    // Acquire acquires n permits
    Acquire(ctx context.Context, n int) error

    // TryAcquire attempts to acquire without blocking
    TryAcquire(ctx context.Context, n int) (bool, error)

    // Release releases n permits
    Release(ctx context.Context, n int) error

    // Available returns available permits
    Available(ctx context.Context) (int, error)
}

// LockInfo contains lock metadata
type LockInfo struct {
    Name      string
    Owner     string
    Token     string
    ExpiresAt time.Time
    Metadata  map[string]string
}
```

## Configuration

```go
// Config holds lock configuration
type Config struct {
    // Provider: "memory", "redis", "postgres", "etcd"
    Provider string `env:"LOCK_PROVIDER" default:"memory"`

    // Default TTL for locks
    DefaultTTL time.Duration `env:"LOCK_DEFAULT_TTL" default:"30s"`

    // Auto-extend interval
    ExtendInterval time.Duration `env:"LOCK_EXTEND_INTERVAL" default:"10s"`

    // Retry settings for lock acquisition
    Retry RetryConfig

    // Owner ID (auto-generated if empty)
    OwnerID string `env:"LOCK_OWNER_ID"`
}

type RetryConfig struct {
    // Maximum attempts
    MaxAttempts int `env:"LOCK_RETRY_MAX_ATTEMPTS" default:"3"`

    // Initial delay
    InitialDelay time.Duration `env:"LOCK_RETRY_INITIAL_DELAY" default:"50ms"`

    // Maximum delay
    MaxDelay time.Duration `env:"LOCK_RETRY_MAX_DELAY" default:"1s"`

    // Jitter factor (0-1)
    Jitter float64 `env:"LOCK_RETRY_JITTER" default:"0.1"`
}

// RedisConfig for Redis provider
type RedisConfig struct {
    // Redis URLs (multiple for Redlock)
    URLs []string `env:"LOCK_REDIS_URLS" default:"redis://localhost:6379"`

    // Key prefix
    KeyPrefix string `env:"LOCK_REDIS_PREFIX" default:"lock:"`

    // Quorum for Redlock
    Quorum int `env:"LOCK_REDIS_QUORUM" default:"0"` // 0 = auto (majority)
}

// PostgresConfig for PostgreSQL provider
type PostgresConfig struct {
    // Connection string
    URL string `env:"LOCK_POSTGRES_URL"`

    // Lock namespace (for advisory locks)
    Namespace int64 `env:"LOCK_POSTGRES_NAMESPACE" default:"1"`
}

// EtcdConfig for etcd provider
type EtcdConfig struct {
    // Endpoints
    Endpoints []string `env:"LOCK_ETCD_ENDPOINTS" default:"localhost:2379"`

    // Key prefix
    KeyPrefix string `env:"LOCK_ETCD_PREFIX" default:"/locks/"`
}
```

## Lock Options

```go
// Option configures lock behavior
type Option func(*lockOptions)

// WithTTL sets the lock TTL
func WithTTL(ttl time.Duration) Option

// WithAutoExtend enables automatic lock extension
func WithAutoExtend(enabled bool) Option

// WithExtendInterval sets auto-extend interval
func WithExtendInterval(interval time.Duration) Option

// WithMetadata adds lock metadata
func WithMetadata(meta map[string]string) Option

// WithRetry configures retry behavior
func WithRetry(config RetryConfig) Option

// WithOnLost sets callback when lock is lost
func WithOnLost(fn func()) Option
```

## Usage Examples

### Basic Mutex

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/user/core-backend/pkg/lock"
)

func main() {
    // Create locker
    locker, err := lock.New(lock.Config{
        Provider:   "redis",
        DefaultTTL: 30 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer locker.Close()

    // Create mutex
    mutex := locker.NewMutex("resource:123",
        lock.WithTTL(10*time.Second),
        lock.WithAutoExtend(true),
    )

    ctx := context.Background()

    // Acquire lock
    if err := mutex.Lock(ctx); err != nil {
        log.Fatal(err)
    }
    defer mutex.Unlock(ctx)

    // Do work with exclusive access
    processResource()
}
```

### Try Lock (Non-Blocking)

```go
func main() {
    locker, _ := lock.New(cfg)
    mutex := locker.NewMutex("resource:123")

    ctx := context.Background()

    acquired, err := mutex.TryLock(ctx)
    if err != nil {
        log.Fatal(err)
    }

    if !acquired {
        log.Println("Resource is busy, trying later...")
        return
    }
    defer mutex.Unlock(ctx)

    // Do work
}
```

### Read-Write Lock

```go
func main() {
    locker, _ := lock.New(cfg)
    rwmutex := locker.NewRWMutex("config:global")

    ctx := context.Background()

    // Multiple readers
    go func() {
        rwmutex.RLock(ctx)
        defer rwmutex.RUnlock(ctx)
        config := readConfig()
    }()

    go func() {
        rwmutex.RLock(ctx)
        defer rwmutex.RUnlock(ctx)
        config := readConfig()
    }()

    // Exclusive writer
    go func() {
        rwmutex.Lock(ctx)
        defer rwmutex.Unlock(ctx)
        writeConfig(newConfig)
    }()
}
```

### Semaphore (Connection Limiting)

```go
func main() {
    locker, _ := lock.New(cfg)

    // Limit to 10 concurrent connections
    sem := locker.NewSemaphore("connections:db", 10)

    ctx := context.Background()

    // Acquire one permit
    if err := sem.Acquire(ctx, 1); err != nil {
        log.Fatal(err)
    }
    defer sem.Release(ctx, 1)

    // Use connection
    conn := db.GetConnection()
    defer conn.Close()
}
```

### With Fencing Token

```go
func main() {
    locker, _ := lock.New(cfg)
    mutex := locker.NewMutex("resource:123")

    ctx := context.Background()
    mutex.Lock(ctx)
    defer mutex.Unlock(ctx)

    // Use fencing token to prevent stale operations
    token := mutex.Token()

    // Pass token to storage system
    storage.WriteWithFence(token, data)
}
```

### Auto-Extend for Long Operations

```go
func main() {
    locker, _ := lock.New(cfg)

    mutex := locker.NewMutex("long-task",
        lock.WithTTL(30*time.Second),
        lock.WithAutoExtend(true),
        lock.WithExtendInterval(10*time.Second),
        lock.WithOnLost(func() {
            log.Println("Lock lost! Aborting operation...")
            cancel()
        }),
    )

    ctx := context.Background()
    mutex.Lock(ctx)
    defer mutex.Unlock(ctx)

    // Long operation - lock is auto-extended
    processLargeDataset()
}
```

### Redis (Redlock Algorithm)

```go
import (
    "github.com/user/core-backend/pkg/lock"
    "github.com/user/core-backend/pkg/lock/provider"
)

func main() {
    // Create Redlock with multiple Redis instances
    redisProvider, err := provider.NewRedis(provider.RedisConfig{
        URLs: []string{
            "redis://redis1:6379",
            "redis://redis2:6379",
            "redis://redis3:6379",
        },
        Quorum: 2, // Majority required
    })
    if err != nil {
        log.Fatal(err)
    }

    locker, _ := lock.New(lock.Config{},
        lock.WithProvider(redisProvider),
    )

    // Distributed lock across Redis cluster
    mutex := locker.NewMutex("global-resource")
    mutex.Lock(ctx)
}
```

### PostgreSQL Advisory Locks

```go
import (
    "github.com/user/core-backend/pkg/lock/provider"
)

func main() {
    pgProvider, err := provider.NewPostgres(provider.PostgresConfig{
        URL:       "postgres://localhost/mydb",
        Namespace: 1,
    })

    locker, _ := lock.New(lock.Config{},
        lock.WithProvider(pgProvider),
    )

    // Uses PostgreSQL advisory locks
    mutex := locker.NewMutex("order:process")
    mutex.Lock(ctx)
}
```

### Context Timeout

```go
func main() {
    locker, _ := lock.New(cfg)
    mutex := locker.NewMutex("resource:123")

    // Timeout after 5 seconds
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := mutex.Lock(ctx); err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            log.Println("Failed to acquire lock within timeout")
            return
        }
        log.Fatal(err)
    }
    defer mutex.Unlock(context.Background()) // Use fresh context for unlock
}
```

### Lock with Metadata

```go
func main() {
    locker, _ := lock.New(cfg)

    mutex := locker.NewMutex("resource:123",
        lock.WithMetadata(map[string]string{
            "owner":   "worker-1",
            "purpose": "batch-processing",
        }),
    )

    // Metadata useful for debugging lock holders
}
```

## Error Handling

```go
var (
    // ErrLockNotAcquired is returned when lock cannot be acquired
    ErrLockNotAcquired = errors.New("lock: could not acquire lock")

    // ErrLockNotHeld is returned when unlocking a lock not held
    ErrLockNotHeld = errors.New("lock: lock not held")

    // ErrLockExpired is returned when lock has expired
    ErrLockExpired = errors.New("lock: lock expired")

    // ErrLockLost is returned when lock is lost during operation
    ErrLockLost = errors.New("lock: lock lost")

    // ErrProviderUnavailable is returned when backend is unavailable
    ErrProviderUnavailable = errors.New("lock: provider unavailable")
)

// IsLockError checks if error is a lock-related error
func IsLockError(err error) bool
```

## Dependencies

- **Required:** None (in-memory uses stdlib)
- **Optional:**
  - `github.com/redis/go-redis/v9` for Redis
  - `github.com/jackc/pgx/v5` for PostgreSQL
  - `go.etcd.io/etcd/client/v3` for etcd

## Implementation Phases

### Phase 1: Core Interface & In-Memory
1. Define Locker, Mutex interfaces
2. Implement in-memory provider
3. Basic mutex implementation
4. TTL and expiration

### Phase 2: Redis Provider
1. Simple Redis locking
2. Redlock algorithm
3. Lock extension

### Phase 3: Additional Lock Types
1. Read-write mutex
2. Semaphore
3. Fencing tokens

### Phase 4: PostgreSQL & etcd
1. PostgreSQL advisory locks
2. etcd distributed locks

### Phase 5: Advanced Features
1. Auto-extend
2. Lost lock detection
3. Metadata support

### Phase 6: Documentation & Examples
1. README
2. Basic example
3. Distributed example
