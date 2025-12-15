# HTTP Client Package Plan

## Overview

This document outlines the implementation plan for a new HTTP client package (`pkg/httpclient`) that provides a configurable circuit breaker pattern with proper error handling. The package will serve as a library/wrapper that makes it easy for core logic to integrate with external HTTP services.

## Goals

1. **Resilient HTTP Client**: Wrap `net/http` with circuit breaker, retry, and timeout capabilities
2. **Configurable**: Environment-based configuration with sensible defaults
3. **Easy Integration**: Simple API that follows existing codebase patterns
4. **Observable**: Built-in metrics and logging hooks for monitoring
5. **Thread-Safe**: Safe for concurrent use across goroutines

## Package Structure

```
pkg/httpclient/
├── client.go           # Main Client interface and implementation
├── config.go           # Configuration with env loading and validation
├── circuit_breaker.go  # Circuit breaker state machine
├── errors.go           # Error types and codes
├── retry.go            # Retry logic with exponential backoff
├── middleware.go       # Request/response middleware chain
├── options.go          # Functional options for requests
├── transport.go        # Custom http.RoundTripper implementation
├── metrics.go          # Metrics collection interface
├── examples/
│   ├── basic/main.go           # Basic usage example
│   ├── with-retry/main.go      # Retry configuration example
│   └── with-metrics/main.go    # Metrics integration example
├── testutil/
│   └── mocks.go        # Mock server and helpers for testing
└── README.md           # Package documentation
```

## Core Components

### 1. Client Interface (`client.go`)

```go
// Client defines the HTTP client interface with circuit breaker support
type Client interface {
    // HTTP methods with context support
    Get(ctx context.Context, url string, opts ...RequestOption) (*Response, error)
    Post(ctx context.Context, url string, body io.Reader, opts ...RequestOption) (*Response, error)
    Put(ctx context.Context, url string, body io.Reader, opts ...RequestOption) (*Response, error)
    Patch(ctx context.Context, url string, body io.Reader, opts ...RequestOption) (*Response, error)
    Delete(ctx context.Context, url string, opts ...RequestOption) (*Response, error)

    // Generic request method
    Do(ctx context.Context, req *http.Request, opts ...RequestOption) (*Response, error)

    // Circuit breaker status
    CircuitState() CircuitState

    // Metrics access
    Metrics() Metrics

    // Close and cleanup resources
    Close() error
}

// Response wraps http.Response with additional metadata
type Response struct {
    *http.Response
    Body        []byte        // Pre-read body for convenience
    Duration    time.Duration // Request duration
    Attempt     int           // Which retry attempt succeeded
    CircuitOpen bool          // Whether circuit was open during request
}
```

### 2. Configuration (`config.go`)

Following the pattern from `pkg/auth/config.go`:

```go
type Config struct {
    // HTTP Client Settings
    BaseURL         string        // Optional base URL for all requests
    Timeout         time.Duration // Request timeout (default: 30s)
    MaxIdleConns    int           // Max idle connections (default: 100)
    IdleConnTimeout time.Duration // Idle connection timeout (default: 90s)

    // Circuit Breaker Settings
    CircuitBreakerEnabled   bool          // Enable circuit breaker (default: true)
    CircuitBreakerThreshold int           // Failures before opening (default: 5)
    CircuitBreakerTimeout   time.Duration // Time in open state (default: 30s)
    CircuitBreakerInterval  time.Duration // Interval to clear counts (default: 60s)

    // Retry Settings
    RetryEnabled     bool          // Enable retries (default: true)
    RetryMaxAttempts int           // Maximum retry attempts (default: 3)
    RetryWaitMin     time.Duration // Minimum wait between retries (default: 1s)
    RetryWaitMax     time.Duration // Maximum wait between retries (default: 30s)
    RetryableStatus  []int         // HTTP status codes to retry (default: 502, 503, 504)

    // Logging/Observability
    EnableMetrics bool           // Enable metrics collection (default: true)
    Logger        Logger         // Optional logger interface
}

// Environment Variables:
// - HTTPCLIENT_BASE_URL
// - HTTPCLIENT_TIMEOUT
// - HTTPCLIENT_MAX_IDLE_CONNS
// - HTTPCLIENT_IDLE_CONN_TIMEOUT
// - HTTPCLIENT_CB_ENABLED
// - HTTPCLIENT_CB_THRESHOLD
// - HTTPCLIENT_CB_TIMEOUT
// - HTTPCLIENT_CB_INTERVAL
// - HTTPCLIENT_RETRY_ENABLED
// - HTTPCLIENT_RETRY_MAX_ATTEMPTS
// - HTTPCLIENT_RETRY_WAIT_MIN
// - HTTPCLIENT_RETRY_WAIT_MAX
// - HTTPCLIENT_RETRY_STATUS_CODES
// - HTTPCLIENT_ENABLE_METRICS
```

### 3. Circuit Breaker (`circuit_breaker.go`)

Implements the circuit breaker state machine:

```go
// CircuitState represents the current state of the circuit breaker
type CircuitState int

const (
    StateClosed   CircuitState = iota // Normal operation, requests flow through
    StateOpen                          // Circuit is open, requests fail fast
    StateHalfOpen                      // Testing if service recovered
)

// CircuitBreaker manages the circuit breaker state machine
type CircuitBreaker struct {
    mu              sync.RWMutex
    state           CircuitState
    failures        int64          // Current failure count
    successes       int64          // Current success count (for half-open)
    lastFailureTime time.Time      // When last failure occurred
    lastStateChange time.Time      // When state last changed

    // Configuration
    threshold       int            // Failures to open circuit
    timeout         time.Duration  // Time to stay open before half-open
    interval        time.Duration  // Interval to reset failure count
    halfOpenMax     int            // Successes needed to close from half-open

    // Callbacks
    onStateChange   func(from, to CircuitState)
}

// Core methods
func (cb *CircuitBreaker) Allow() (bool, error)           // Check if request allowed
func (cb *CircuitBreaker) RecordSuccess()                  // Record successful request
func (cb *CircuitBreaker) RecordFailure()                  // Record failed request
func (cb *CircuitBreaker) State() CircuitState             // Get current state
func (cb *CircuitBreaker) Reset()                          // Reset to closed state
```

**State Transitions:**

```
                    failure threshold reached
    ┌─────────────────────────────────────────────┐
    │                                             ▼
┌───────┐                                    ┌────────┐
│CLOSED │◄──────── success threshold ────────│HALF-   │
└───────┘             reached                │ OPEN   │
    ▲                                        └────────┘
    │                                             │
    │           timeout expired                   │
    │         ┌───────────────────────────────────┘
    │         ▼
    │    ┌────────┐
    └────│  OPEN  │──── any failure ────► stays OPEN
         └────────┘
```

### 4. Error Types (`errors.go`)

Following the pattern from `pkg/auth/errors.go`:

```go
// HTTPClientError represents an HTTP client error with context
type HTTPClientError struct {
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    StatusCode int                    `json:"status_code,omitempty"`
    Cause      error                  `json:"-"`
    Details    map[string]interface{} `json:"details,omitempty"`
}

// Error codes
const (
    CodeCircuitOpen      = "circuit_open"
    CodeRequestTimeout   = "request_timeout"
    CodeConnectionFailed = "connection_failed"
    CodeMaxRetriesExhausted = "max_retries_exhausted"
    CodeInvalidRequest   = "invalid_request"
    CodeResponseError    = "response_error"
    CodeContextCanceled  = "context_canceled"
)

// Sentinel errors
var (
    ErrCircuitOpen        = errors.New("circuit breaker is open")
    ErrRequestTimeout     = errors.New("request timed out")
    ErrConnectionFailed   = errors.New("connection failed")
    ErrMaxRetriesExhausted = errors.New("maximum retries exhausted")
    ErrInvalidRequest     = errors.New("invalid request")
    ErrContextCanceled    = errors.New("context canceled")
)

// Helper functions
func IsCircuitOpen(err error) bool
func IsRetryable(err error) bool
func IsTimeout(err error) bool
```

### 5. Retry Logic (`retry.go`)

```go
// RetryPolicy defines the retry behavior
type RetryPolicy struct {
    MaxAttempts     int
    WaitMin         time.Duration
    WaitMax         time.Duration
    RetryableErrors []error
    RetryableStatus []int
    Backoff         BackoffStrategy
}

// BackoffStrategy calculates wait time between retries
type BackoffStrategy interface {
    Duration(attempt int) time.Duration
}

// Built-in strategies
type ExponentialBackoff struct {
    Base       time.Duration
    Max        time.Duration
    Jitter     bool          // Add randomness to prevent thundering herd
    Multiplier float64       // Default: 2.0
}

// Retry executes function with retry logic
func Retry(ctx context.Context, policy RetryPolicy, fn func() error) error
```

### 6. Request Options (`options.go`)

Functional options pattern for per-request configuration:

```go
type RequestOption func(*requestOptions)

type requestOptions struct {
    headers         http.Header
    timeout         time.Duration
    skipRetry       bool
    skipCircuitBreaker bool
    retryPolicy     *RetryPolicy
}

// Option functions
func WithHeader(key, value string) RequestOption
func WithHeaders(headers http.Header) RequestOption
func WithTimeout(timeout time.Duration) RequestOption
func WithRetryPolicy(policy RetryPolicy) RequestOption
func WithoutRetry() RequestOption
func WithoutCircuitBreaker() RequestOption
func WithBasicAuth(username, password string) RequestOption
func WithBearerToken(token string) RequestOption
func WithContentType(contentType string) RequestOption
```

### 7. Middleware (`middleware.go`)

```go
// RoundTripperFunc type for middleware
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// Middleware wraps a RoundTripper
type Middleware func(http.RoundTripper) http.RoundTripper

// Built-in middleware
func LoggingMiddleware(logger Logger) Middleware
func MetricsMiddleware(metrics MetricsCollector) Middleware
func HeaderMiddleware(headers http.Header) Middleware
func TracingMiddleware(tracer Tracer) Middleware
```

### 8. Metrics (`metrics.go`)

```go
// Metrics holds collected metrics
type Metrics struct {
    TotalRequests      int64
    SuccessfulRequests int64
    FailedRequests     int64
    RetryCount         int64
    CircuitOpenCount   int64
    AverageLatency     time.Duration
    P99Latency         time.Duration
}

// MetricsCollector interface for custom metrics backends
type MetricsCollector interface {
    RecordRequest(method, url string, statusCode int, duration time.Duration)
    RecordRetry(method, url string, attempt int)
    RecordCircuitState(state CircuitState)
    RecordError(method, url string, err error)
}

// NoopMetricsCollector for when metrics are disabled
type NoopMetricsCollector struct{}
```

## Integration Examples

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/rompi/core-backend/pkg/httpclient"
)

func main() {
    // Create client with default configuration
    client, err := httpclient.New(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Make a GET request
    ctx := context.Background()
    resp, err := client.Get(ctx, "https://api.example.com/users")
    if err != nil {
        if httpclient.IsCircuitOpen(err) {
            log.Println("Service unavailable, circuit breaker is open")
        }
        log.Fatal(err)
    }

    log.Printf("Status: %d, Body: %s", resp.StatusCode, resp.Body)
}
```

### With Custom Configuration

```go
cfg := &httpclient.Config{
    BaseURL:                 "https://api.example.com",
    Timeout:                 10 * time.Second,
    CircuitBreakerThreshold: 3,
    CircuitBreakerTimeout:   60 * time.Second,
    RetryMaxAttempts:        5,
}

client, err := httpclient.New(cfg)
```

### With Request Options

```go
resp, err := client.Post(ctx, "/users",
    strings.NewReader(`{"name": "John"}`),
    httpclient.WithContentType("application/json"),
    httpclient.WithBearerToken("my-token"),
    httpclient.WithTimeout(5 * time.Second),
)
```

### With Custom Retry Policy

```go
customRetry := httpclient.RetryPolicy{
    MaxAttempts:     5,
    WaitMin:         500 * time.Millisecond,
    WaitMax:         10 * time.Second,
    RetryableStatus: []int{500, 502, 503, 504},
    Backoff: &httpclient.ExponentialBackoff{
        Base:   500 * time.Millisecond,
        Max:    10 * time.Second,
        Jitter: true,
    },
}

resp, err := client.Get(ctx, "/api/resource",
    httpclient.WithRetryPolicy(customRetry),
)
```

### Integration with Auth Package

```go
// Create HTTP client for external auth provider
authClient, _ := httpclient.New(&httpclient.Config{
    BaseURL: "https://auth.provider.com",
    Timeout: 5 * time.Second,
    CircuitBreakerThreshold: 5,
})

// Use in custom auth service
type ExternalAuthService struct {
    client httpclient.Client
}

func (s *ExternalAuthService) ValidateExternalToken(ctx context.Context, token string) (*User, error) {
    resp, err := s.client.Get(ctx, "/validate",
        httpclient.WithBearerToken(token),
    )
    if err != nil {
        return nil, fmt.Errorf("external auth validation: %w", err)
    }
    // Parse response...
}
```

## Implementation Phases

### Phase 1: Core Implementation
1. [ ] Create package structure (`pkg/httpclient/`)
2. [ ] Implement `Config` with environment loading and validation
3. [ ] Implement `CircuitBreaker` state machine
4. [ ] Implement core `Client` interface and implementation
5. [ ] Implement `Response` wrapper

### Phase 2: Retry and Error Handling
1. [ ] Implement error types and codes
2. [ ] Implement `RetryPolicy` with exponential backoff
3. [ ] Add jitter to prevent thundering herd
4. [ ] Implement retry logic integration with circuit breaker

### Phase 3: Request Options and Middleware
1. [ ] Implement functional options pattern
2. [ ] Implement middleware chain
3. [ ] Add logging middleware
4. [ ] Add metrics middleware

### Phase 4: Testing
1. [ ] Unit tests for circuit breaker state machine
2. [ ] Unit tests for retry logic
3. [ ] Integration tests with httptest
4. [ ] Mock utilities for consumers
5. [ ] Target: ≥80% coverage

### Phase 5: Documentation and Examples
1. [ ] Package README.md
2. [ ] Code examples (basic, retry, metrics)
3. [ ] Integration guide

## Testing Strategy

### Unit Tests

```go
func TestCircuitBreaker_OpenAfterThreshold(t *testing.T) {
    cb := NewCircuitBreaker(CircuitBreakerConfig{
        Threshold: 3,
        Timeout:   5 * time.Second,
    })

    // Record failures up to threshold
    for i := 0; i < 3; i++ {
        cb.RecordFailure()
    }

    // Circuit should be open
    allowed, err := cb.Allow()
    assert.False(t, allowed)
    assert.ErrorIs(t, err, ErrCircuitOpen)
    assert.Equal(t, StateOpen, cb.State())
}
```

### Integration Tests

```go
func TestClient_RetryOnServerError(t *testing.T) {
    attempts := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attempts++
        if attempts < 3 {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "ok"}`))
    }))
    defer server.Close()

    client, _ := httpclient.New(&httpclient.Config{
        RetryMaxAttempts: 5,
        RetryWaitMin:     10 * time.Millisecond,
    })

    resp, err := client.Get(context.Background(), server.URL)

    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    assert.Equal(t, 3, attempts)
    assert.Equal(t, 3, resp.Attempt)
}
```

## Dependencies

```go
// go.mod additions
require (
    // No external dependencies for core functionality
    // Uses only standard library
)
```

The package will rely solely on the Go standard library:
- `net/http` - HTTP client
- `sync` - Thread safety
- `time` - Timeouts and durations
- `context` - Cancellation and deadlines
- `math/rand` - Jitter calculation

## Configuration Reference

| Environment Variable | Type | Default | Description |
|---------------------|------|---------|-------------|
| `HTTPCLIENT_BASE_URL` | string | "" | Base URL for all requests |
| `HTTPCLIENT_TIMEOUT` | duration | 30s | Request timeout |
| `HTTPCLIENT_MAX_IDLE_CONNS` | int | 100 | Max idle connections |
| `HTTPCLIENT_IDLE_CONN_TIMEOUT` | duration | 90s | Idle connection timeout |
| `HTTPCLIENT_CB_ENABLED` | bool | true | Enable circuit breaker |
| `HTTPCLIENT_CB_THRESHOLD` | int | 5 | Failures before opening |
| `HTTPCLIENT_CB_TIMEOUT` | duration | 30s | Time in open state |
| `HTTPCLIENT_CB_INTERVAL` | duration | 60s | Interval to clear counts |
| `HTTPCLIENT_RETRY_ENABLED` | bool | true | Enable retries |
| `HTTPCLIENT_RETRY_MAX_ATTEMPTS` | int | 3 | Maximum retry attempts |
| `HTTPCLIENT_RETRY_WAIT_MIN` | duration | 1s | Minimum wait between retries |
| `HTTPCLIENT_RETRY_WAIT_MAX` | duration | 30s | Maximum wait between retries |
| `HTTPCLIENT_RETRY_STATUS_CODES` | []int | 502,503,504 | Status codes to retry |
| `HTTPCLIENT_ENABLE_METRICS` | bool | true | Enable metrics collection |

## Error Handling Guide

| Error Code | Description | Recommended Action |
|------------|-------------|-------------------|
| `circuit_open` | Circuit breaker is open | Wait and retry, or failover |
| `request_timeout` | Request timed out | Retry with longer timeout |
| `connection_failed` | Could not connect | Check network, retry |
| `max_retries_exhausted` | All retries failed | Alert, manual intervention |
| `invalid_request` | Invalid request configuration | Fix request parameters |
| `response_error` | Non-2xx response | Handle based on status code |
| `context_canceled` | Context was canceled | Normal cancellation, no retry |

## Thread Safety

All public methods are thread-safe:
- `CircuitBreaker` uses `sync.RWMutex` for state protection
- `Client` is safe for concurrent use
- Metrics collection uses atomic operations
- Configuration is immutable after creation

## Performance Considerations

1. **Connection Pooling**: Uses `http.Transport` connection pooling by default
2. **Context Deadlines**: All operations respect context deadlines
3. **Minimal Allocations**: Request options use sync.Pool for reuse
4. **Efficient Locking**: Read-heavy operations use `RWMutex`
5. **No External Dependencies**: Zero external dependencies for minimal overhead

## Future Enhancements

1. **Bulkhead Pattern**: Limit concurrent requests per host
2. **Fallback Support**: Define fallback responses for failures
3. **Request Hedging**: Send duplicate requests and use first response
4. **Adaptive Timeouts**: Dynamically adjust timeouts based on latency
5. **OpenTelemetry Integration**: Native tracing support
6. **Rate Limiting**: Client-side rate limiting per host
