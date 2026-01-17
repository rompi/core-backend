# Package Plan: pkg/health

## Overview

A health check aggregation package for monitoring application and dependency health. Provides Kubernetes-compatible readiness/liveness probes, dependency checks, and health status reporting with configurable thresholds.

## Goals

1. **Health Aggregation** - Combine multiple health checks
2. **Kubernetes Compatible** - Liveness, readiness, startup probes
3. **Dependency Checks** - Database, cache, queue, external services
4. **Async Checks** - Background health checking with caching
5. **Degraded State** - Support for degraded vs unhealthy states
6. **HTTP Endpoints** - Ready-to-use health endpoints
7. **Metrics Integration** - Export health metrics

## Architecture

```
pkg/health/
├── health.go             # Core Health interface
├── config.go             # Configuration
├── options.go            # Functional options
├── status.go             # Health status types
├── checker/
│   ├── checker.go        # Checker interface
│   ├── database.go       # Database health check
│   ├── redis.go          # Redis health check
│   ├── http.go           # HTTP endpoint check
│   ├── grpc.go           # gRPC service check
│   ├── disk.go           # Disk space check
│   ├── memory.go         # Memory usage check
│   ├── custom.go         # Custom check wrapper
│   └── composite.go      # Composite checker
├── handler/
│   ├── http.go           # HTTP handlers
│   └── grpc.go           # gRPC health service
├── examples/
│   ├── basic/
│   ├── kubernetes/
│   └── with-metrics/
└── README.md
```

## Core Interfaces

```go
package health

import (
    "context"
    "time"
)

// Health manages health checks
type Health interface {
    // Register adds a health check
    Register(name string, checker Checker, opts ...CheckOption) error

    // Deregister removes a health check
    Deregister(name string) error

    // Check runs all health checks
    Check(ctx context.Context) *Report

    // CheckLiveness runs liveness checks only
    CheckLiveness(ctx context.Context) *Report

    // CheckReadiness runs readiness checks only
    CheckReadiness(ctx context.Context) *Report

    // Status returns current health status
    Status() Status

    // Start begins background health checking
    Start(ctx context.Context) error

    // Stop stops background checking
    Stop() error
}

// Checker performs a health check
type Checker interface {
    // Check performs the health check
    Check(ctx context.Context) *CheckResult

    // Name returns the checker name
    Name() string
}

// CheckResult contains check outcome
type CheckResult struct {
    // Status is the health status
    Status Status

    // Message describes the status
    Message string

    // Error contains error details (if unhealthy)
    Error error

    // Duration is how long the check took
    Duration time.Duration

    // Timestamp is when the check was performed
    Timestamp time.Time

    // Details contains additional info
    Details map[string]interface{}
}

// Report contains aggregated health status
type Report struct {
    // Status is the overall status
    Status Status

    // Checks contains individual check results
    Checks map[string]*CheckResult

    // Timestamp is when the report was generated
    Timestamp time.Time

    // Duration is total check duration
    Duration time.Duration
}

// Status represents health status
type Status string

const (
    StatusHealthy   Status = "healthy"
    StatusDegraded  Status = "degraded"
    StatusUnhealthy Status = "unhealthy"
    StatusUnknown   Status = "unknown"
)
```

## Configuration

```go
// Config holds health check configuration
type Config struct {
    // Check interval for background checks
    CheckInterval time.Duration `env:"HEALTH_CHECK_INTERVAL" default:"30s"`

    // Timeout for individual checks
    CheckTimeout time.Duration `env:"HEALTH_CHECK_TIMEOUT" default:"5s"`

    // Cache duration for check results
    CacheDuration time.Duration `env:"HEALTH_CACHE_DURATION" default:"5s"`

    // Parallel execution of checks
    Parallel bool `env:"HEALTH_PARALLEL" default:"true"`

    // Failure threshold before unhealthy
    FailureThreshold int `env:"HEALTH_FAILURE_THRESHOLD" default:"3"`
}

// CheckOption configures individual checks
type CheckOption func(*checkOptions)

// WithCheckInterval sets check-specific interval
func WithCheckInterval(interval time.Duration) CheckOption

// WithCheckTimeout sets check-specific timeout
func WithCheckTimeout(timeout time.Duration) CheckOption

// WithLiveness marks check for liveness probe
func WithLiveness() CheckOption

// WithReadiness marks check for readiness probe
func WithReadiness() CheckOption

// WithCritical marks check as critical (failure = unhealthy)
func WithCritical() CheckOption

// WithWeight sets check importance weight
func WithWeight(weight int) CheckOption
```

## Built-in Checkers

### Database Checker

```go
// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
    db      *sql.DB
    query   string
    timeout time.Duration
}

func NewDatabaseChecker(db *sql.DB, opts ...DatabaseOption) *DatabaseChecker

// Options
func WithQuery(query string) DatabaseOption // Default: "SELECT 1"
func WithTimeout(timeout time.Duration) DatabaseOption
```

### Redis Checker

```go
// RedisChecker checks Redis connectivity
type RedisChecker struct {
    client RedisClient
}

func NewRedisChecker(client RedisClient) *RedisChecker
```

### HTTP Checker

```go
// HTTPChecker checks HTTP endpoint availability
type HTTPChecker struct {
    url            string
    expectedStatus int
    client         *http.Client
}

func NewHTTPChecker(url string, opts ...HTTPOption) *HTTPChecker

// Options
func WithExpectedStatus(status int) HTTPOption
func WithHTTPClient(client *http.Client) HTTPOption
func WithHeaders(headers map[string]string) HTTPOption
```

### gRPC Checker

```go
// GRPCChecker checks gRPC service health
type GRPCChecker struct {
    conn    *grpc.ClientConn
    service string
}

func NewGRPCChecker(conn *grpc.ClientConn, opts ...GRPCOption) *GRPCChecker
```

### System Checkers

```go
// DiskChecker checks disk space
type DiskChecker struct {
    path      string
    threshold float64 // Percentage threshold (e.g., 90.0)
}

func NewDiskChecker(path string, threshold float64) *DiskChecker

// MemoryChecker checks memory usage
type MemoryChecker struct {
    threshold float64 // Percentage threshold
}

func NewMemoryChecker(threshold float64) *MemoryChecker

// CPUChecker checks CPU usage
type CPUChecker struct {
    threshold float64
    window    time.Duration
}

func NewCPUChecker(threshold float64, window time.Duration) *CPUChecker
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "github.com/user/core-backend/pkg/health"
    "github.com/user/core-backend/pkg/health/checker"
)

func main() {
    // Create health manager
    h := health.New(health.Config{
        CheckInterval: 30 * time.Second,
        CheckTimeout:  5 * time.Second,
    })

    // Register checks
    h.Register("database", checker.NewDatabaseChecker(db),
        health.WithCritical(),
        health.WithReadiness(),
    )

    h.Register("redis", checker.NewRedisChecker(redisClient),
        health.WithReadiness(),
    )

    h.Register("disk", checker.NewDiskChecker("/", 90.0))

    // Start background checking
    ctx := context.Background()
    h.Start(ctx)
    defer h.Stop()

    // Check health
    report := h.Check(ctx)
    fmt.Printf("Status: %s\n", report.Status)

    for name, result := range report.Checks {
        fmt.Printf("  %s: %s (%v)\n", name, result.Status, result.Duration)
    }
}
```

### Kubernetes Probes

```go
import (
    "github.com/user/core-backend/pkg/health"
    "github.com/user/core-backend/pkg/health/handler"
)

func main() {
    h := health.New(cfg)

    // Critical for liveness (if db fails, restart pod)
    h.Register("database", dbChecker,
        health.WithLiveness(),
        health.WithCritical(),
    )

    // Required for readiness (don't receive traffic until ready)
    h.Register("cache", cacheChecker,
        health.WithReadiness(),
    )

    h.Register("external-api", apiChecker,
        health.WithReadiness(),
    )

    h.Start(ctx)

    mux := http.NewServeMux()

    // Kubernetes endpoints
    mux.Handle("/healthz", handler.Liveness(h))     // Liveness probe
    mux.Handle("/readyz", handler.Readiness(h))     // Readiness probe
    mux.Handle("/health", handler.Full(h))          // Full health report

    http.ListenAndServe(":8080", mux)
}
```

### HTTP Handler Response

```go
// GET /health
// Response:
{
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "duration": "45ms",
    "checks": {
        "database": {
            "status": "healthy",
            "message": "Connected",
            "duration": "2ms"
        },
        "redis": {
            "status": "healthy",
            "message": "PONG",
            "duration": "1ms"
        },
        "disk": {
            "status": "degraded",
            "message": "85% used",
            "duration": "5ms",
            "details": {
                "used_percent": 85.2,
                "threshold": 90.0
            }
        }
    }
}

// GET /healthz (liveness)
// Returns 200 OK or 503 Service Unavailable

// GET /readyz (readiness)
// Returns 200 OK or 503 Service Unavailable
```

### Custom Checker

```go
// Create custom checker
type PaymentGatewayChecker struct {
    client PaymentClient
}

func (c *PaymentGatewayChecker) Name() string {
    return "payment-gateway"
}

func (c *PaymentGatewayChecker) Check(ctx context.Context) *health.CheckResult {
    start := time.Now()

    err := c.client.Ping(ctx)
    if err != nil {
        return &health.CheckResult{
            Status:   health.StatusUnhealthy,
            Message:  "Payment gateway unavailable",
            Error:    err,
            Duration: time.Since(start),
        }
    }

    return &health.CheckResult{
        Status:   health.StatusHealthy,
        Message:  "Connected",
        Duration: time.Since(start),
    }
}

// Register
h.Register("payment", &PaymentGatewayChecker{client: paymentClient})
```

### Function Checker

```go
// Quick inline checker
h.Register("config", health.CheckFunc(func(ctx context.Context) *health.CheckResult {
    if config.IsValid() {
        return health.Healthy("Configuration valid")
    }
    return health.Unhealthy("Invalid configuration", nil)
}))
```

### Degraded State

```go
type CacheChecker struct {
    primary   CacheClient
    fallback  CacheClient
}

func (c *CacheChecker) Check(ctx context.Context) *health.CheckResult {
    primaryErr := c.primary.Ping(ctx)
    fallbackErr := c.fallback.Ping(ctx)

    if primaryErr == nil {
        return health.Healthy("Primary cache connected")
    }

    if fallbackErr == nil {
        return &health.CheckResult{
            Status:  health.StatusDegraded,
            Message: "Using fallback cache",
            Details: map[string]interface{}{
                "primary_error": primaryErr.Error(),
            },
        }
    }

    return health.Unhealthy("All caches unavailable", primaryErr)
}
```

### With gRPC Health Service

```go
import (
    "github.com/user/core-backend/pkg/health/handler"
    "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
    h := health.New(cfg)
    h.Start(ctx)

    server := grpc.NewServer()

    // Register gRPC health service
    grpc_health_v1.RegisterHealthServer(server, handler.GRPCHealth(h))
}
```

### Metrics Integration

```go
import (
    "github.com/user/core-backend/pkg/health"
    "github.com/user/core-backend/pkg/observability"
)

func main() {
    h := health.New(cfg,
        health.WithMetrics(metricsCollector),
    )

    // Exports metrics:
    // health_check_status{check="database"} 1 (1=healthy, 0=unhealthy)
    // health_check_duration_seconds{check="database"} 0.002
    // health_check_last_success_timestamp{check="database"} 1705312200
}
```

## Error Handling

```go
var (
    // ErrCheckNotFound is returned for unknown check
    ErrCheckNotFound = errors.New("health: check not found")

    // ErrCheckTimeout is returned when check times out
    ErrCheckTimeout = errors.New("health: check timeout")

    // ErrAlreadyStarted is returned if already started
    ErrAlreadyStarted = errors.New("health: already started")
)
```

## Dependencies

- **Required:** None (uses stdlib)
- **Optional:**
  - Database drivers for database checker
  - Redis client for Redis checker

## Implementation Phases

### Phase 1: Core Interface
1. Define Health, Checker interfaces
2. Report and status types
3. Basic aggregation logic

### Phase 2: Built-in Checkers
1. Database checker
2. Redis checker
3. HTTP checker
4. System checkers (disk, memory)

### Phase 3: HTTP Handlers
1. Full health endpoint
2. Liveness endpoint
3. Readiness endpoint

### Phase 4: Background Checking
1. Periodic background checks
2. Result caching
3. Failure threshold

### Phase 5: Advanced Features
1. gRPC health service
2. Metrics integration
3. Degraded state support

### Phase 6: Documentation
1. README
2. Kubernetes example
3. Custom checker example
