# HTTP Client Package Implementation Plan

## 1. Context & Goals

### Purpose
Build a production-grade, reusable HTTP client package (`pkg/httpclient`) that provides an enhanced wrapper around Go's `net/http` with enterprise-grade features for consuming REST APIs, service-to-service communication, and third-party integrations.

### Business/Technical Justification
- **Standardization**: Provide a consistent HTTP client across all rompi projects and external consumers
- **Reliability**: Built-in retry logic, circuit breaking, and timeout management reduce transient failures
- **Observability**: Structured logging and middleware enable debugging and monitoring
- **Developer Experience**: Simplified API reduces boilerplate and common pitfalls in HTTP communication
- **Reusability**: Standalone package usable across multiple GitHub repositories with zero coupling to core-backend

### Success Criteria
1. Package is independently importable via `go get github.com/rompi/core-backend/pkg/httpclient`
2. Core features (retry, circuit breaker, middleware, logging) working with comprehensive tests (>85% coverage)
3. Complete documentation with working examples in README and GoDoc
4. Zero breaking changes to existing core-backend code
5. Successfully consumed by at least one external rompi repository

---

## 2. In-scope / Out-of-scope

### In-Scope
- ✅ HTTP client with fluent API for common methods (GET, POST, PUT, PATCH, DELETE)
- ✅ Automatic retry with exponential backoff for transient failures (5xx, network errors)
- ✅ Circuit breaker pattern to prevent cascading failures
- ✅ Request/response middleware system (logging, auth, metrics, custom)
- ✅ Structured logging integration (compatible with slog or custom loggers)
- ✅ JSON request/response helpers with automatic encoding/decoding
- ✅ Timeout configuration at client and request levels
- ✅ Context propagation for cancellation and deadlines
- ✅ Support for custom headers, query parameters, request/response hooks
- ✅ Comprehensive unit and integration tests
- ✅ Benchmarks for performance validation

### Out-of-Scope
- ❌ HTTP server capabilities (this is client-only)
- ❌ GraphQL or gRPC support (HTTP REST focus)
- ❌ WebSocket connections
- ❌ OAuth flow implementations (auth headers via middleware is sufficient)
- ❌ Request queueing or batch processing
- ❌ Built-in metrics collection (expose hooks for consumers to add their own)
- ❌ Response caching (can be added via middleware by consumers)

---

## 3. API Surface & Contracts

### Core Types

```go
// Client is the main HTTP client
type Client struct {
    baseURL        string
    httpClient     *http.Client
    middleware     []Middleware
    retryPolicy    RetryPolicy
    circuitBreaker CircuitBreaker
    logger         Logger
}

// Config for creating a new client
type Config struct {
    BaseURL          string
    Timeout          time.Duration        // default: 30s
    MaxRetries       int                  // default: 3
    RetryWaitMin     time.Duration        // default: 1s
    RetryWaitMax     time.Duration        // default: 30s
    CircuitBreaker   *CircuitBreakerConfig
    Logger           Logger               // default: noop logger
    Transport        http.RoundTripper    // default: http.DefaultTransport
    FollowRedirects  bool                 // default: true
}

// Request represents an HTTP request
type Request struct {
    method  string
    url     string
    headers http.Header
    query   url.Values
    body    io.Reader
    ctx     context.Context
}

// Response wraps http.Response with helpers
type Response struct {
    *http.Response
    body []byte // cached body for multiple reads
}

// Middleware intercepts requests/responses
type Middleware func(next RoundTripper) RoundTripper

// RoundTripper matches http.RoundTripper signature
type RoundTripper interface {
    RoundTrip(*http.Request) (*http.Response, error)
}
```

### Exported Functions

```go
// New creates a client with config
func New(cfg Config) (*Client, error)

// NewDefault creates a client with sensible defaults
func NewDefault(baseURL string) *Client

// Request builders
func (c *Client) Get(ctx context.Context, path string) *RequestBuilder
func (c *Client) Post(ctx context.Context, path string) *RequestBuilder
func (c *Client) Put(ctx context.Context, path string) *RequestBuilder
func (c *Client) Patch(ctx context.Context, path string) *RequestBuilder
func (c *Client) Delete(ctx context.Context, path string) *RequestBuilder

// RequestBuilder fluent API
func (rb *RequestBuilder) Header(key, value string) *RequestBuilder
func (rb *RequestBuilder) Headers(headers map[string]string) *RequestBuilder
func (rb *RequestBuilder) Query(key, value string) *RequestBuilder
func (rb *RequestBuilder) QueryParams(params map[string]string) *RequestBuilder
func (rb *RequestBuilder) JSON(body interface{}) *RequestBuilder
func (rb *RequestBuilder) Body(body io.Reader, contentType string) *RequestBuilder
func (rb *RequestBuilder) Do() (*Response, error)

// Response helpers
func (r *Response) JSON(v interface{}) error
func (r *Response) String() (string, error)
func (r *Response) Bytes() ([]byte, error)
func (r *Response) IsSuccess() bool
func (r *Response) IsClientError() bool
func (r *Response) IsServerError() bool

// Built-in middleware
func LoggingMiddleware(logger Logger) Middleware
func AuthBearerMiddleware(token string) Middleware
func AuthAPIKeyMiddleware(headerName, apiKey string) Middleware
func UserAgentMiddleware(userAgent string) Middleware
func RetryMiddleware(policy RetryPolicy) Middleware
func CircuitBreakerMiddleware(cb CircuitBreaker) Middleware
```

### Error Handling

```go
// Errors are sentinel values with wrapped contexts
var (
    ErrTimeout            = errors.New("httpclient: request timeout")
    ErrCircuitOpen        = errors.New("httpclient: circuit breaker open")
    ErrMaxRetriesExceeded = errors.New("httpclient: max retries exceeded")
    ErrInvalidConfig      = errors.New("httpclient: invalid configuration")
)

// Error wraps HTTP errors with context
type Error struct {
    StatusCode int
    Message    string
    Body       []byte
    Request    *http.Request
    Response   *http.Response
    Err        error
}

func (e *Error) Error() string
func (e *Error) Unwrap() error
func (e *Error) Is(target error) bool
```

### Concurrency Expectations
- **Thread-safe**: `Client` is safe for concurrent use across goroutines
- **Context-aware**: All operations accept `context.Context` for cancellation and deadlines
- **No shared state**: `RequestBuilder` instances are independent and not reusable after `.Do()`

---

## 4. Internal Architecture

### Package Structure

```
pkg/httpclient/
├── README.md                    # Package documentation
├── client.go                    # Client, Config, New()
├── client_test.go               # Client tests
├── request.go                   # Request, RequestBuilder
├── request_test.go
├── response.go                  # Response wrapper and helpers
├── response_test.go
├── middleware.go                # Middleware type and built-ins
├── middleware_test.go
├── retry.go                     # Retry logic and policy
├── retry_test.go
├── circuitbreaker.go            # Circuit breaker implementation
├── circuitbreaker_test.go
├── logger.go                    # Logger interface and noop
├── errors.go                    # Error types
├── integration_test.go          # End-to-end tests with httptest
├── bench_test.go                # Benchmarks
├── examples/                    # Example usage
│   ├── basic/main.go
│   ├── middleware/main.go
│   └── advanced/main.go
└── internal/                    # Internal helpers (not exported)
    └── backoff/                 # Backoff algorithm
        ├── backoff.go
        └── backoff_test.go
```

### Dependency Graph

```
Client
├── uses → http.Client (stdlib)
├── uses → RetryPolicy
├── uses → CircuitBreaker
├── uses → []Middleware
└── uses → Logger

Middleware chain
├── LoggingMiddleware
├── AuthMiddleware
├── RetryMiddleware (wraps RetryPolicy)
├── CircuitBreakerMiddleware (wraps CircuitBreaker)
└── CustomMiddleware (user-defined)

RequestBuilder
├── builds → http.Request
└── executes via → Client
```

### External Dependencies

Based on "flexible dependencies" constraint, we'll use:
1. **github.com/cenkalti/backoff/v4** - battle-tested exponential backoff (optional, can implement ourselves)
2. **github.com/sony/gobreaker** - production-grade circuit breaker (optional, can implement ourselves)

**Decision**: Implement retry/backoff and circuit breaker ourselves for better control and zero deps, following patterns from these libraries.

---

## 5. Integration with Other Repos

### Module Path
```
github.com/rompi/core-backend/pkg/httpclient
```

### Import Statement
```go
import "github.com/rompi/core-backend/pkg/httpclient"
```

### Installation
```bash
go get github.com/rompi/core-backend/pkg/httpclient@v1.0.0
```

### Versioning Strategy
- **Semantic Versioning** following Go module conventions
- **v0.x.x**: Initial development, API may change
- **v1.0.0**: Stable release with compatibility guarantees
- **v1.x.x**: Backwards-compatible features and bug fixes
- **v2.0.0+**: Breaking changes require new major version

### Release Cadence
- **Patch releases**: Bug fixes as needed
- **Minor releases**: New features every 4-8 weeks during active development
- **Major releases**: Only when breaking changes are absolutely necessary

### Git Tags
```bash
git tag pkg/httpclient/v1.0.0
git push origin pkg/httpclient/v1.0.0
```

### Consuming from Other Repos - Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/rompi/core-backend/pkg/httpclient"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    // Create client
    client := httpclient.NewDefault("https://api.example.com")

    // Or with custom config
    client, err := httpclient.New(httpclient.Config{
        BaseURL:      "https://api.example.com",
        Timeout:      30 * time.Second,
        MaxRetries:   3,
        Logger:       myLogger,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Make request
    var user User
    resp, err := client.Get(context.Background(), "/users/123").Do()
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    if err := resp.JSON(&user); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("User: %+v\n", user)
}
```

---

## 6. Testing & Validation

### Unit Tests
- **client_test.go**: Config validation, client initialization, method routing
- **request_test.go**: RequestBuilder fluent API, header/query parameter handling
- **response_test.go**: JSON/String/Bytes decoding, status code helpers
- **middleware_test.go**: Middleware chain execution order, built-in middleware behavior
- **retry_test.go**: Retry logic, backoff timing, max retries enforcement
- **circuitbreaker_test.go**: State transitions (closed → open → half-open), threshold behavior
- **errors_test.go**: Error wrapping, unwrapping, Is/As behavior

### Integration Tests
- **integration_test.go**: Full request lifecycle using `httptest.Server`
  - Successful request/response
  - Retry on 5xx errors
  - Circuit breaker opens after consecutive failures
  - Timeout handling
  - Context cancellation
  - Middleware chain integration
  - JSON encoding/decoding end-to-end

### Benchmarks
- **bench_test.go**:
  - `BenchmarkClientGet`: Simple GET request
  - `BenchmarkClientPostJSON`: JSON encoding/decoding
  - `BenchmarkMiddlewareChain`: Overhead of 5-middleware chain
  - `BenchmarkRetryLogic`: Retry decision making
  - `BenchmarkCircuitBreaker`: Circuit breaker state checking

### Coverage Goals
- **Minimum**: 85% overall
- **Critical paths**: 95%+ (retry logic, circuit breaker, error handling)

### Testing Tools
- `go test -v -race -cover ./...` - run tests with race detector
- `go test -bench=. -benchmem` - run benchmarks
- `golangci-lint run` - static analysis (following core-backend patterns)

### Test Patterns (following coding-guidelines.md)
```go
func TestClient_Get_Success(t *testing.T) {
    // Arrange
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer server.Close()

    client := httpclient.NewDefault(server.URL)

    // Act
    resp, err := client.Get(context.Background(), "/test").Do()

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !resp.IsSuccess() {
        t.Errorf("expected success status, got %d", resp.StatusCode)
    }
}
```

---

## 7. Documentation & Examples

### README.md Structure

```markdown
# httpclient

Production-grade HTTP client for Go with built-in retry, circuit breaking, and middleware support.

## Features
- Fluent API for HTTP requests
- Automatic retry with exponential backoff
- Circuit breaker pattern
- Request/response middleware
- Structured logging
- JSON helpers
- Context-aware

## Installation
`go get github.com/rompi/core-backend/pkg/httpclient`

## Quick Start
[Basic example]

## Configuration
[Config options table]

## Middleware
[Middleware examples]

## Advanced Usage
- Custom retry policies
- Circuit breaker configuration
- Writing custom middleware
- Error handling

## API Reference
[Link to GoDoc]

## Testing
[How to run tests]

## Contributing
[Guidelines]

## License
[License info]
```

### GoDoc Comments
- Package-level overview with design philosophy
- Every exported type/function documented with examples
- Code examples in doc comments using `Example*` tests

### Examples Directory
1. **examples/basic/main.go**: Simple GET/POST requests
2. **examples/middleware/main.go**: Custom middleware, auth, logging
3. **examples/advanced/main.go**: Circuit breaker, retry policies, error handling
4. **examples/testing/main.go**: How to mock/test code using httpclient

### Documentation Checklist
- [ ] README.md with quick start and feature overview
- [ ] GoDoc comments on all exported types/functions
- [ ] At least 3 runnable examples in examples/
- [ ] CHANGELOG.md for version history
- [ ] Migration guide (when v2+ is released)

---

## 8. Risks & Mitigations

### Risk 1: Thread Safety Issues
**Impact**: Data races in concurrent usage
**Mitigation**:
- Use `sync.RWMutex` for circuit breaker state
- Make `Client` immutable after creation
- Run all tests with `-race` flag
- Document concurrency guarantees clearly

### Risk 2: Dependency on External Libraries
**Impact**: Supply chain vulnerabilities, version conflicts
**Mitigation**:
- Implement retry and circuit breaker internally (zero external deps beyond stdlib)
- If deps needed later, pin versions in go.mod with comments
- Regular `go mod tidy` and security audits

### Risk 3: Breaking Changes in Future Versions
**Impact**: Consumer code breaks on updates
**Mitigation**:
- Strict semver adherence
- Deprecation warnings before removal (one minor version minimum)
- Keep v1.x.x stable for 12+ months before v2
- Document breaking changes prominently in CHANGELOG

### Risk 4: Performance Overhead
**Impact**: Middleware/retry adds latency
**Mitigation**:
- Benchmark critical paths
- Make retry/circuit breaker opt-in via config
- Avoid reflection in hot paths
- Use efficient JSON encoding (encoding/json is fine, consider alternatives if benchmarks show issues)

### Risk 5: Complex Circuit Breaker State Management
**Impact**: Incorrect state transitions, memory leaks
**Mitigation**:
- Follow established patterns from sony/gobreaker
- Comprehensive state machine tests
- Time-based state resets with proper cleanup
- Document state transition diagram in code

### Risk 6: Logger Interface Compatibility
**Impact**: Conflicts with consumer's logging frameworks
**Mitigation**:
- Define minimal `Logger` interface (Debug, Info, Warn, Error methods)
- Provide noop logger as default
- Provide adapter examples for slog, zap, logrus in docs
- Make logging completely optional

---

## 9. Delivery Plan

### Phase 1: Foundation (Days 1-3)
1. **Setup package structure**
   - Create `pkg/httpclient/` directory
   - Initialize README.md, go files
   - Setup test infrastructure

2. **Core client implementation**
   - `client.go`: Client, Config, New() functions
   - `request.go`: Request, RequestBuilder with fluent API
   - `response.go`: Response wrapper with JSON/String helpers
   - `errors.go`: Error types and sentinels
   - Basic unit tests for each

3. **Initial integration test**
   - Simple GET/POST test with httptest.Server
   - Validate basic request/response flow

### Phase 2: Reliability Features (Days 4-6)
4. **Retry mechanism**
   - `retry.go`: RetryPolicy interface and default implementation
   - `internal/backoff/backoff.go`: Exponential backoff algorithm
   - Unit tests with mocked time (avoid time.Sleep in tests)
   - Integration test with failing server

5. **Circuit breaker**
   - `circuitbreaker.go`: Circuit breaker implementation
   - State machine (closed/open/half-open) with thresholds
   - Unit tests for all state transitions
   - Integration test with consecutive failures

6. **Logger interface**
   - `logger.go`: Logger interface and noop implementation
   - Integration with retry and circuit breaker logging

### Phase 3: Middleware System (Days 7-8)
7. **Middleware infrastructure**
   - `middleware.go`: Middleware type and chain execution
   - Built-in middleware: Logging, Auth (Bearer/APIKey), UserAgent
   - Unit tests for middleware ordering and execution

8. **Middleware integration**
   - Wire middleware into Client execution flow
   - Integration tests with multiple middleware
   - Performance benchmarks for middleware overhead

### Phase 4: Polish & Documentation (Days 9-10)
9. **Comprehensive testing**
   - Achieve 85%+ test coverage
   - Add edge case tests (nil checks, empty responses, context cancellation)
   - Run benchmarks and optimize hotspots
   - Add fuzz tests if applicable (JSON parsing, URL building)

10. **Documentation**
    - Complete README.md with all sections
    - Add GoDoc comments to all exports
    - Create 3+ examples in examples/
    - Write CHANGELOG.md for v0.1.0

### Phase 5: Validation & Release (Days 11-12)
11. **Internal dogfooding**
    - Use httpclient in a small core-backend internal service
    - Identify usability issues and refine API
    - Performance testing under realistic load

12. **Release preparation**
    - Run `golangci-lint` and fix issues
    - Final review against coding-guidelines.md
    - Tag v0.1.0
    - Announce in project documentation

### Phase 6: External Validation (Post-Release)
13. **Cross-repo testing**
    - Import into another rompi repository
    - Validate import path, versioning, documentation
    - Gather feedback and create backlog for v0.2.0

---

## Sequencing Summary

**Critical path**: Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5
**Can parallelize**: Documentation (Phase 4) can start during Phase 3
**Dependencies**:
- Circuit breaker (Phase 2) requires logger (Phase 2)
- Middleware (Phase 3) requires core client (Phase 1)
- Release (Phase 5) requires all tests passing (Phase 4)

**Estimated effort**: 10-12 engineering days for full implementation + testing + docs

---

## Approval & Next Steps

This plan is ready for review and approval. Once approved, implementation will begin with Phase 1.

**Plan version**: 1.0
**Created**: 2025-12-20
**Status**: Awaiting approval
