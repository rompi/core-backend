# Redis Integration Package

A comprehensive, configurable Redis integration package for Go applications.

## Features

- Full Redis support (Strings, Hashes, Lists, Sets, Sorted Sets, Streams, Pub/Sub)
- Environment-based configuration
- High-level Service API with serialization
- Distributed locking with retry support
- Rate limiting
- Connection pooling with Cluster and Sentinel support
- Mock implementations for testing

## Quick Start

```go
cfg := redis.DefaultConfig()
cfg.Host = "localhost"

svc, err := redis.NewService(cfg)
if err != nil {
    log.Fatal(err)
}
defer svc.Close()

ctx := context.Background()
svc.SetString(ctx, "key", "value", 5*time.Minute)
val, _ := svc.GetString(ctx, "key")
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis hostname | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Password | (empty) |
| `REDIS_DB` | Database number | `0` |
| `REDIS_POOL_SIZE` | Max connections | `10` |
| `REDIS_KEY_PREFIX` | Key prefix | (empty) |

## Distributed Locking

```go
lock, _ := svc.AcquireLock(ctx, "resource", 30*time.Second)
defer svc.ReleaseLock(ctx, lock)
// critical section
```

## Rate Limiting

```go
allowed, count, _ := svc.RateLimit(ctx, "api:user:123", 100, time.Minute)
if !allowed {
    // rate limit exceeded
}
```

## Testing

```go
import "core-backend/pkg/redis/testutil"

mockClient := testutil.NewMockClient()
svc, _ := redis.NewServiceWithClient(mockClient, redis.DefaultConfig())
```
