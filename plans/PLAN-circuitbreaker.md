# Package Plan: pkg/circuitbreaker

## Overview

A standalone circuit breaker package implementing the circuit breaker pattern for fault tolerance. Prevents cascading failures by temporarily blocking calls to failing services, with configurable thresholds, timeouts, and recovery strategies.

## Goals

1. **Three States** - Closed, Open, Half-Open
2. **Configurable Thresholds** - Failure count, rate, consecutive
3. **Recovery** - Automatic recovery with half-open state
4. **Metrics** - Success/failure counts, state changes
5. **Events** - Callbacks for state transitions
6. **Two-Step Pattern** - Allow/record for async operations
7. **Zero Dependencies** - Pure Go implementation

## Architecture

```
pkg/circuitbreaker/
├── circuitbreaker.go     # Core circuit breaker
├── config.go             # Configuration
├── options.go            # Functional options
├── state.go              # State definitions
├── counts.go             # Request counting
├── settings.go           # Settings variations
├── examples/
│   ├── basic/
│   ├── http-client/
│   └── grpc-client/
└── README.md
```

## Core Interfaces

```go
package circuitbreaker

import (
    "context"
    "time"
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker interface {
    // Execute runs the operation through the circuit breaker
    Execute(ctx context.Context, operation Operation) error

    // ExecuteWithResult runs and returns a result
    ExecuteWithResult[T any](ctx context.Context, operation OperationWithResult[T]) (T, error)

    // Allow checks if a request is allowed (two-step pattern)
    Allow() (done func(success bool), err error)

    // State returns the current state
    State() State

    // Counts returns current counts
    Counts() Counts

    // Reset resets the circuit breaker
    Reset()

    // Name returns the circuit breaker name
    Name() string
}

// Operation is the protected operation
type Operation func(ctx context.Context) error

// OperationWithResult returns a result
type OperationWithResult[T any] func(ctx context.Context) (T, error)

// State represents circuit breaker state
type State int

const (
    StateClosed   State = iota // Normal operation
    StateOpen                   // Failing, reject requests
    StateHalfOpen               // Testing recovery
)

// Counts holds request counts
type Counts struct {
    Requests             int64
    TotalSuccesses       int64
    TotalFailures        int64
    ConsecutiveSuccesses int64
    ConsecutiveFailures  int64
}
```

## Configuration

```go
// Config holds circuit breaker configuration
type Config struct {
    // Name for identification
    Name string

    // MaxRequests in half-open state
    MaxRequests int `default:"1"`

    // Interval to clear counts when closed (0 = never)
    Interval time.Duration `default:"0"`

    // Timeout in open state before half-open
    Timeout time.Duration `default:"60s"`

    // ReadyToTrip determines when to open
    ReadyToTrip func(counts Counts) bool

    // OnStateChange is called on state transitions
    OnStateChange func(name string, from, to State)

    // IsSuccessful determines if error is a failure
    IsSuccessful func(err error) bool
}

// Option configures the circuit breaker
type Option func(*Config)

// WithMaxRequests sets max requests in half-open
func WithMaxRequests(n int) Option

// WithTimeout sets open state timeout
func WithTimeout(timeout time.Duration) Option

// WithInterval sets count reset interval
func WithInterval(interval time.Duration) Option

// WithReadyToTrip sets the trip condition
func WithReadyToTrip(fn func(Counts) bool) Option

// WithOnStateChange sets state change callback
func WithOnStateChange(fn func(string, State, State)) Option

// WithFailureRatio trips on failure ratio
func WithFailureRatio(threshold float64, minRequests int) Option

// WithConsecutiveFailures trips on consecutive failures
func WithConsecutiveFailures(threshold int) Option
```

## Trip Conditions

```go
// ConsecutiveFailures trips after n consecutive failures
func ConsecutiveFailures(n int) func(Counts) bool {
    return func(counts Counts) bool {
        return counts.ConsecutiveFailures >= int64(n)
    }
}

// FailureRatio trips when failure ratio exceeds threshold
func FailureRatio(ratio float64, minRequests int) func(Counts) bool {
    return func(counts Counts) bool {
        if counts.Requests < int64(minRequests) {
            return false
        }
        return float64(counts.TotalFailures)/float64(counts.Requests) >= ratio
    }
}

// FailureCount trips after n total failures
func FailureCount(n int) func(Counts) bool {
    return func(counts Counts) bool {
        return counts.TotalFailures >= int64(n)
    }
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "github.com/user/core-backend/pkg/circuitbreaker"
)

func main() {
    cb := circuitbreaker.New("external-api",
        circuitbreaker.WithTimeout(30*time.Second),
        circuitbreaker.WithConsecutiveFailures(5),
    )

    ctx := context.Background()

    err := cb.Execute(ctx, func(ctx context.Context) error {
        return callExternalAPI()
    })

    if err != nil {
        if errors.Is(err, circuitbreaker.ErrOpen) {
            log.Println("Circuit breaker is open, using fallback")
            return useFallback()
        }
        log.Printf("Request failed: %v", err)
    }
}
```

### With Result

```go
func main() {
    cb := circuitbreaker.New("user-service",
        circuitbreaker.WithFailureRatio(0.5, 10),
        circuitbreaker.WithTimeout(60*time.Second),
    )

    ctx := context.Background()

    user, err := circuitbreaker.ExecuteWithResult(cb, ctx,
        func(ctx context.Context) (*User, error) {
            return userClient.GetUser(ctx, userID)
        },
    )

    if err != nil {
        if errors.Is(err, circuitbreaker.ErrOpen) {
            return getCachedUser(userID)
        }
        return nil, err
    }

    return user, nil
}
```

### Two-Step Pattern

```go
func main() {
    cb := circuitbreaker.New("async-service",
        circuitbreaker.WithConsecutiveFailures(5),
    )

    // Check if allowed before starting
    done, err := cb.Allow()
    if err != nil {
        // Circuit is open
        return useFallback()
    }

    // Perform async operation
    go func() {
        err := performAsyncOperation()
        // Report result
        done(err == nil)
    }()
}
```

### HTTP Client Integration

```go
type CircuitBreakerTransport struct {
    base http.RoundTripper
    cb   circuitbreaker.CircuitBreaker
}

func (t *CircuitBreakerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    var resp *http.Response

    err := t.cb.Execute(req.Context(), func(ctx context.Context) error {
        var err error
        resp, err = t.base.RoundTrip(req)
        if err != nil {
            return err
        }

        // Treat 5xx as failures
        if resp.StatusCode >= 500 {
            return fmt.Errorf("server error: %d", resp.StatusCode)
        }

        return nil
    })

    return resp, err
}

// Usage
client := &http.Client{
    Transport: &CircuitBreakerTransport{
        base: http.DefaultTransport,
        cb:   circuitbreaker.New("api", circuitbreaker.WithConsecutiveFailures(3)),
    },
}
```

### State Change Notifications

```go
func main() {
    cb := circuitbreaker.New("payment-gateway",
        circuitbreaker.WithConsecutiveFailures(5),
        circuitbreaker.WithOnStateChange(func(name string, from, to circuitbreaker.State) {
            log.Printf("Circuit breaker %s: %s -> %s", name, from, to)

            if to == circuitbreaker.StateOpen {
                // Alert on circuit open
                alerting.Send("Circuit breaker opened: " + name)
            }
        }),
    )
}
```

### Per-Service Circuit Breakers

```go
type ServiceClient struct {
    circuitBreakers map[string]circuitbreaker.CircuitBreaker
    mu              sync.RWMutex
}

func (c *ServiceClient) getCircuitBreaker(service string) circuitbreaker.CircuitBreaker {
    c.mu.RLock()
    cb, ok := c.circuitBreakers[service]
    c.mu.RUnlock()

    if ok {
        return cb
    }

    c.mu.Lock()
    defer c.mu.Unlock()

    // Double-check
    if cb, ok = c.circuitBreakers[service]; ok {
        return cb
    }

    cb = circuitbreaker.New(service,
        circuitbreaker.WithConsecutiveFailures(5),
        circuitbreaker.WithTimeout(30*time.Second),
    )
    c.circuitBreakers[service] = cb
    return cb
}

func (c *ServiceClient) Call(ctx context.Context, service, method string, req, resp interface{}) error {
    cb := c.getCircuitBreaker(service)

    return cb.Execute(ctx, func(ctx context.Context) error {
        return c.doCall(ctx, service, method, req, resp)
    })
}
```

### Custom Success Criteria

```go
func main() {
    cb := circuitbreaker.New("api",
        circuitbreaker.WithConsecutiveFailures(5),
        circuitbreaker.WithIsSuccessful(func(err error) bool {
            if err == nil {
                return true
            }

            // Rate limiting is not a failure
            var apiErr *APIError
            if errors.As(err, &apiErr) {
                return apiErr.StatusCode == 429
            }

            return false
        }),
    )
}
```

### With Metrics

```go
func main() {
    cb := circuitbreaker.New("service",
        circuitbreaker.WithConsecutiveFailures(5),
        circuitbreaker.WithOnStateChange(func(name string, from, to circuitbreaker.State) {
            stateGauge.WithLabelValues(name).Set(float64(to))
        }),
    )

    // Periodically export counts
    go func() {
        for {
            counts := cb.Counts()
            requestsCounter.WithLabelValues(cb.Name()).Add(float64(counts.Requests))
            failuresCounter.WithLabelValues(cb.Name()).Add(float64(counts.TotalFailures))
            time.Sleep(10 * time.Second)
        }
    }()
}
```

## Error Handling

```go
var (
    // ErrOpen is returned when circuit is open
    ErrOpen = errors.New("circuitbreaker: circuit breaker is open")

    // ErrTooManyRequests is returned in half-open when max reached
    ErrTooManyRequests = errors.New("circuitbreaker: too many requests")
)

// IsCircuitBreakerError checks if error is from circuit breaker
func IsCircuitBreakerError(err error) bool
```

## State Diagram

```
     ┌──────────────────────────────────────┐
     │                                      │
     ▼                                      │
┌─────────┐  failure threshold  ┌──────────┐│
│ CLOSED  │ ─────────────────▶  │   OPEN   ││
│         │                     │          ││
└─────────┘                     └──────────┘│
     ▲                               │      │
     │                               │      │
     │    success    ┌──────────┐    │      │
     └───────────────│HALF-OPEN │◀───┘      │
                     │          │  timeout  │
                     └──────────┘           │
                          │                 │
                          │ failure         │
                          └─────────────────┘
```

## Dependencies

- **Required:** None (pure Go)

## Implementation Phases

### Phase 1: Core Implementation
1. Three-state machine
2. Consecutive failures trigger
3. Timeout recovery

### Phase 2: Configuration
1. Configurable thresholds
2. Failure ratio trigger
3. Custom trip conditions

### Phase 3: Two-Step Pattern
1. Allow/record pattern
2. Async operation support

### Phase 4: Observability
1. State change callbacks
2. Metrics support
3. Counts access

### Phase 5: Documentation
1. README
2. HTTP client example
3. gRPC client example
