# HTTP Server Package Plan

## Overview

This document outlines the comprehensive plan for creating a production-ready HTTP server package (`pkg/httpserver`) that follows the established patterns in this codebase. The package will be designed as a plug-and-play solution that allows upstream services to easily integrate their business logic.

## Design Goals

1. **Plug-and-Play**: Easy to set up with sensible defaults, minimal configuration required
2. **Composable**: Middleware-based architecture for extensibility
3. **Production-Ready**: Includes all essential features for production deployments
4. **Framework Agnostic**: No external dependencies, uses only Go standard library
5. **Testable**: Interface-driven design with testing utilities
6. **Consistent**: Follows existing codebase patterns (auth, httpclient packages)

---

## Package Structure

```
pkg/httpserver/
├── server.go           # Core Server type and lifecycle management
├── config.go           # Configuration with environment variable support
├── router.go           # Router with path parameters and method routing
├── context.go          # Request context utilities and value injection
├── middleware.go       # Middleware type definition and chaining
├── handler.go          # Handler types and adapter functions
├── request.go          # Request parsing and validation helpers
├── response.go         # Response writing utilities (JSON, error, etc.)
├── errors.go           # Structured error types and error responses
├── logger.go           # Logger interface (consistent with httpclient)
├── health.go           # Health check endpoint and readiness probes
├── graceful.go         # Graceful shutdown implementation
├── tls.go              # TLS configuration and certificate handling
├── recovery.go         # Panic recovery middleware
├── timeout.go          # Request timeout middleware
├── cors.go             # CORS middleware
├── requestid.go        # Request ID generation and propagation
├── compress.go         # Response compression (gzip, deflate)
├── ratelimit.go        # Rate limiting middleware (token bucket)
├── metrics.go          # Metrics collection interface
├── validation.go       # Request body and query validation
├── binding.go          # Request body binding (JSON, form, etc.)
├── static.go           # Static file serving
├── websocket.go        # WebSocket upgrade support (optional)
├── options.go          # Functional options pattern
├── doc.go              # Package documentation
├── server_test.go      # Server tests
├── router_test.go      # Router tests
├── middleware_test.go  # Middleware tests
├── *_test.go           # Other test files
├── testutil/           # Testing utilities
│   ├── mock_handler.go
│   ├── test_server.go
│   └── assertions.go
├── examples/           # Example implementations
│   ├── basic/          # Minimal setup example
│   ├── rest-api/       # Full REST API example
│   ├── middleware/     # Custom middleware example
│   └── graceful/       # Graceful shutdown example
└── README.md           # Package documentation
```

---

## Core Components

### 1. Server (`server.go`)

The main Server struct that wraps `http.Server` with additional functionality.

```go
// Server represents an HTTP server with production-ready features
type Server struct {
    server      *http.Server
    router      *Router
    middlewares []Middleware
    config      *Config
    logger      Logger

    // Lifecycle
    running     atomic.Bool
    shutdownCh  chan struct{}

    // Health
    healthChecker HealthChecker
}

// Key Methods:
// - NewServer(opts ...Option) *Server
// - Start() error
// - StartTLS(certFile, keyFile string) error
// - Shutdown(ctx context.Context) error
// - ListenAndServe() error
// - Router() *Router
// - Use(middlewares ...Middleware)
```

**Features:**
- Functional options pattern for configuration
- Automatic graceful shutdown on OS signals
- Health check endpoint registration
- Middleware chain management
- TLS support out of the box

---

### 2. Configuration (`config.go`)

Environment-driven configuration with validation.

```go
type Config struct {
    // Server
    Host            string        // Default: ""
    Port            int           // Default: 8080
    ReadTimeout     time.Duration // Default: 30s
    WriteTimeout    time.Duration // Default: 30s
    IdleTimeout     time.Duration // Default: 120s
    MaxHeaderBytes  int           // Default: 1MB
    ShutdownTimeout time.Duration // Default: 30s

    // TLS
    TLSEnabled      bool
    TLSCertFile     string
    TLSKeyFile      string
    TLSMinVersion   uint16        // Default: TLS 1.2

    // Request Limits
    MaxBodySize     int64         // Default: 10MB

    // Features
    EnableHealthCheck   bool      // Default: true
    HealthCheckPath     string    // Default: /health
    EnableReadinessCheck bool     // Default: true
    ReadinessCheckPath  string    // Default: /ready
    EnableMetrics       bool      // Default: false
    MetricsPath         string    // Default: /metrics

    // CORS
    CORSEnabled         bool
    CORSAllowedOrigins  []string
    CORSAllowedMethods  []string
    CORSAllowedHeaders  []string
    CORSExposedHeaders  []string
    CORSMaxAge          int
    CORSAllowCredentials bool

    // Rate Limiting
    RateLimitEnabled    bool
    RateLimitRequests   int       // Requests per window
    RateLimitWindow     time.Duration

    // Compression
    CompressionEnabled  bool      // Default: true
    CompressionLevel    int       // Default: gzip.DefaultCompression
    CompressionMinSize  int       // Default: 1024 bytes

    // Request ID
    RequestIDEnabled    bool      // Default: true
    RequestIDHeader     string    // Default: X-Request-ID

    // Logging
    LogRequests         bool      // Default: true
    LogResponseBody     bool      // Default: false
}

// Functions:
// - LoadConfig() (*Config, error)
// - LoadConfigFromEnv() (*Config, error)
// - (c *Config) Validate() error
// - DefaultConfig() *Config
```

**Environment Variables:**
```
HTTP_HOST, HTTP_PORT, HTTP_READ_TIMEOUT, HTTP_WRITE_TIMEOUT,
HTTP_IDLE_TIMEOUT, HTTP_MAX_HEADER_BYTES, HTTP_SHUTDOWN_TIMEOUT,
HTTP_TLS_ENABLED, HTTP_TLS_CERT_FILE, HTTP_TLS_KEY_FILE,
HTTP_MAX_BODY_SIZE, HTTP_CORS_ENABLED, HTTP_CORS_ALLOWED_ORIGINS,
HTTP_RATE_LIMIT_ENABLED, HTTP_RATE_LIMIT_REQUESTS, etc.
```

---

### 3. Router (`router.go`)

A lightweight router with path parameter support.

```go
type Router struct {
    routes      map[string]*routeNode // method -> trie
    middleware  []Middleware
    notFound    HandlerFunc
    methodNotAllowed HandlerFunc
}

type Route struct {
    Method      string
    Pattern     string
    Handler     HandlerFunc
    Middlewares []Middleware
    Name        string
}

// Key Methods:
// - NewRouter() *Router
// - Handle(method, pattern string, handler HandlerFunc, middlewares ...Middleware)
// - GET(pattern string, handler HandlerFunc, middlewares ...Middleware)
// - POST(pattern string, handler HandlerFunc, middlewares ...Middleware)
// - PUT(pattern string, handler HandlerFunc, middlewares ...Middleware)
// - PATCH(pattern string, handler HandlerFunc, middlewares ...Middleware)
// - DELETE(pattern string, handler HandlerFunc, middlewares ...Middleware)
// - OPTIONS(pattern string, handler HandlerFunc, middlewares ...Middleware)
// - HEAD(pattern string, handler HandlerFunc, middlewares ...Middleware)
// - Group(prefix string, middlewares ...Middleware) *RouteGroup
// - Use(middlewares ...Middleware)
// - ServeHTTP(w http.ResponseWriter, r *http.Request)
// - NotFound(handler HandlerFunc)
// - MethodNotAllowed(handler HandlerFunc)
```

**Path Parameter Patterns:**
```go
router.GET("/users/:id", handler)           // Named parameter
router.GET("/files/*filepath", handler)     // Wildcard/catch-all
router.GET("/users/:id/posts/:postId", handler) // Multiple parameters
```

---

### 4. Context Utilities (`context.go`)

Request context management and value injection.

```go
type contextKey string

const (
    RequestIDKey    contextKey = "request-id"
    RouteParamsKey  contextKey = "route-params"
    StartTimeKey    contextKey = "start-time"
    LoggerKey       contextKey = "logger"
)

// RouteParams holds path parameters
type RouteParams map[string]string

// Functions:
// - RequestID(ctx context.Context) string
// - SetRequestID(ctx context.Context, id string) context.Context
// - RouteParam(ctx context.Context, name string) string
// - RouteParamsFromContext(ctx context.Context) RouteParams
// - SetRouteParams(ctx context.Context, params RouteParams) context.Context
// - RequestStartTime(ctx context.Context) time.Time
// - WithLogger(ctx context.Context, logger Logger) context.Context
// - LoggerFromContext(ctx context.Context) Logger
```

---

### 5. Middleware (`middleware.go`)

Middleware type definition and chain utilities.

```go
// Middleware wraps an HTTP handler
type Middleware func(HandlerFunc) HandlerFunc

// MiddlewareFunc is the standard http.Handler middleware
type MiddlewareFunc func(http.Handler) http.Handler

// Chain combines multiple middlewares
func Chain(middlewares ...Middleware) Middleware

// Adapt converts standard http.Handler middleware to our Middleware type
func Adapt(m MiddlewareFunc) Middleware

// AdaptHandler converts http.Handler to HandlerFunc
func AdaptHandler(h http.Handler) HandlerFunc

// WrapHandler wraps HandlerFunc back to http.Handler
func WrapHandler(h HandlerFunc) http.Handler
```

---

### 6. Handler Types (`handler.go`)

Custom handler types with error return support.

```go
// HandlerFunc is our custom handler that can return errors
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handler interface for struct-based handlers
type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

// Adapters:
// - ToStdHandler(h HandlerFunc) http.HandlerFunc
// - FromStdHandler(h http.HandlerFunc) HandlerFunc
// - FromHandler(h Handler) HandlerFunc
```

**Benefits of error-returning handlers:**
```go
// Instead of this:
func handler(w http.ResponseWriter, r *http.Request) {
    data, err := service.GetData()
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(data)
}

// You can do this:
func handler(w http.ResponseWriter, r *http.Request) error {
    data, err := service.GetData()
    if err != nil {
        return err // Handled by error middleware
    }
    return JSON(w, http.StatusOK, data)
}
```

---

### 7. Request Helpers (`request.go`)

Request parsing and extraction utilities.

```go
// Body binding
func Bind(r *http.Request, v interface{}) error
func BindJSON(r *http.Request, v interface{}) error
func BindForm(r *http.Request, v interface{}) error
func BindQuery(r *http.Request, v interface{}) error

// Parameter extraction
func PathParam(r *http.Request, name string) string
func QueryParam(r *http.Request, name string) string
func QueryParamDefault(r *http.Request, name, defaultValue string) string
func QueryParams(r *http.Request, name string) []string
func QueryInt(r *http.Request, name string, defaultValue int) int
func QueryBool(r *http.Request, name string, defaultValue bool) bool

// Header helpers
func Header(r *http.Request, name string) string
func ContentType(r *http.Request) string
func Accept(r *http.Request) string
func BearerToken(r *http.Request) string

// Request info
func ClientIP(r *http.Request) string
func IsAJAX(r *http.Request) bool
func IsWebSocket(r *http.Request) bool
func Scheme(r *http.Request) string
func FullURL(r *http.Request) string
```

---

### 8. Response Helpers (`response.go`)

Response writing utilities.

```go
// ResponseWriter wraps http.ResponseWriter with additional features
type ResponseWriter struct {
    http.ResponseWriter
    status      int
    size        int64
    written     bool
}

// Constructor
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter

// Methods
func (w *ResponseWriter) Status() int
func (w *ResponseWriter) Size() int64
func (w *ResponseWriter) Written() bool
func (w *ResponseWriter) Flush()
func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error)

// Response helpers (standalone functions)
func JSON(w http.ResponseWriter, status int, data interface{}) error
func JSONPretty(w http.ResponseWriter, status int, data interface{}) error
func XML(w http.ResponseWriter, status int, data interface{}) error
func HTML(w http.ResponseWriter, status int, html string) error
func Text(w http.ResponseWriter, status int, text string) error
func Bytes(w http.ResponseWriter, status int, contentType string, data []byte) error
func Stream(w http.ResponseWriter, status int, contentType string, reader io.Reader) error
func File(w http.ResponseWriter, r *http.Request, filepath string) error
func Attachment(w http.ResponseWriter, r *http.Request, filepath, filename string) error
func Redirect(w http.ResponseWriter, r *http.Request, url string, status int) error
func NoContent(w http.ResponseWriter) error
func Created(w http.ResponseWriter, location string, data interface{}) error
```

---

### 9. Error Handling (`errors.go`)

Structured error types and error response formatting.

```go
// HTTPError represents an HTTP error with status code
type HTTPError struct {
    Code       int         `json:"code"`
    Message    string      `json:"message"`
    Details    interface{} `json:"details,omitempty"`
    Internal   error       `json:"-"`
    RequestID  string      `json:"request_id,omitempty"`
}

func (e *HTTPError) Error() string
func (e *HTTPError) Unwrap() error

// Error constructors
func NewHTTPError(code int, message string) *HTTPError
func NewHTTPErrorWithDetails(code int, message string, details interface{}) *HTTPError

// Common errors
var (
    ErrBadRequest          = NewHTTPError(400, "bad request")
    ErrUnauthorized        = NewHTTPError(401, "unauthorized")
    ErrForbidden           = NewHTTPError(403, "forbidden")
    ErrNotFound            = NewHTTPError(404, "not found")
    ErrMethodNotAllowed    = NewHTTPError(405, "method not allowed")
    ErrConflict            = NewHTTPError(409, "conflict")
    ErrUnprocessableEntity = NewHTTPError(422, "unprocessable entity")
    ErrTooManyRequests     = NewHTTPError(429, "too many requests")
    ErrInternalServer      = NewHTTPError(500, "internal server error")
    ErrServiceUnavailable  = NewHTTPError(503, "service unavailable")
)

// Error response helper
func Error(w http.ResponseWriter, err error) error
func ErrorWithStatus(w http.ResponseWriter, status int, message string) error

// ErrorHandler is the signature for custom error handlers
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// DefaultErrorHandler is the default error handler
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error)
```

---

### 10. Logger Interface (`logger.go`)

Pluggable logger interface (consistent with httpclient package).

```go
// Logger defines the logging interface
type Logger interface {
    Debug(msg string, keysAndValues ...interface{})
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
}

// NoopLogger is a no-operation logger
type NoopLogger struct{}

func (NoopLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (NoopLogger) Info(msg string, keysAndValues ...interface{})  {}
func (NoopLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (NoopLogger) Error(msg string, keysAndValues ...interface{}) {}

// StdLogger wraps the standard library logger
type StdLogger struct {
    logger *log.Logger
    level  LogLevel
}

type LogLevel int

const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelWarn
    LogLevelError
)

func NewStdLogger(out io.Writer, level LogLevel) *StdLogger
```

---

### 11. Health Checks (`health.go`)

Health and readiness check support.

```go
// HealthChecker defines the health check interface
type HealthChecker interface {
    // Check returns nil if healthy, error otherwise
    Check(ctx context.Context) error
    // Name returns the name of the health check
    Name() string
}

// HealthStatus represents the health status response
type HealthStatus struct {
    Status    string                 `json:"status"`    // "healthy", "unhealthy", "degraded"
    Timestamp time.Time              `json:"timestamp"`
    Checks    map[string]CheckResult `json:"checks,omitempty"`
    Version   string                 `json:"version,omitempty"`
}

type CheckResult struct {
    Status  string        `json:"status"`
    Latency time.Duration `json:"latency_ms"`
    Error   string        `json:"error,omitempty"`
}

// Built-in checkers
type CompositeHealthChecker struct {
    checkers []HealthChecker
}

func NewCompositeHealthChecker(checkers ...HealthChecker) *CompositeHealthChecker
func (c *CompositeHealthChecker) Add(checker HealthChecker)
func (c *CompositeHealthChecker) Check(ctx context.Context) *HealthStatus

// Health handler
func HealthHandler(checker *CompositeHealthChecker) HandlerFunc
func ReadinessHandler(checker *CompositeHealthChecker) HandlerFunc
func LivenessHandler() HandlerFunc
```

**Built-in health checkers:**
- `PingChecker` - Simple ping check
- `DatabaseChecker` - Database connectivity
- `HTTPChecker` - External HTTP service
- `DiskSpaceChecker` - Disk space availability
- `MemoryChecker` - Memory usage

---

### 12. Graceful Shutdown (`graceful.go`)

Graceful shutdown handling.

```go
// GracefulShutdown handles server shutdown
type GracefulShutdown struct {
    server   *http.Server
    timeout  time.Duration
    logger   Logger
    hooks    []ShutdownHook
}

// ShutdownHook is called during shutdown
type ShutdownHook func(ctx context.Context) error

// Methods
func NewGracefulShutdown(server *http.Server, timeout time.Duration, logger Logger) *GracefulShutdown
func (g *GracefulShutdown) OnShutdown(hook ShutdownHook)
func (g *GracefulShutdown) ListenForSignals(signals ...os.Signal)
func (g *GracefulShutdown) Shutdown(ctx context.Context) error
```

---

### 13. Built-in Middlewares

#### Recovery Middleware (`recovery.go`)
```go
func Recovery() Middleware
func RecoveryWithConfig(config RecoveryConfig) Middleware

type RecoveryConfig struct {
    StackSize         int  // Default: 4KB
    DisableStackAll   bool
    DisablePrintStack bool
    LogLevel          LogLevel
    Handler           func(w http.ResponseWriter, r *http.Request, err interface{})
}
```

#### Timeout Middleware (`timeout.go`)
```go
func Timeout(timeout time.Duration) Middleware
func TimeoutWithConfig(config TimeoutConfig) Middleware

type TimeoutConfig struct {
    Timeout        time.Duration
    ErrorMessage   string
    ErrorHandler   func(w http.ResponseWriter, r *http.Request)
}
```

#### CORS Middleware (`cors.go`)
```go
func CORS() Middleware
func CORSWithConfig(config CORSConfig) Middleware

type CORSConfig struct {
    AllowOrigins     []string // Default: ["*"]
    AllowMethods     []string // Default: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    AllowHeaders     []string
    ExposeHeaders    []string
    AllowCredentials bool
    MaxAge           int      // Preflight cache duration in seconds
}
```

#### Request ID Middleware (`requestid.go`)
```go
func RequestID() Middleware
func RequestIDWithConfig(config RequestIDConfig) Middleware

type RequestIDConfig struct {
    Header    string                    // Default: "X-Request-ID"
    Generator func() string             // Default: UUID v4
    SkipFunc  func(*http.Request) bool  // Skip certain requests
}
```

#### Compression Middleware (`compress.go`)
```go
func Compress() Middleware
func CompressWithConfig(config CompressConfig) Middleware

type CompressConfig struct {
    Level      int      // gzip.DefaultCompression
    MinSize    int      // Minimum size to compress
    Types      []string // Content types to compress
    Skipper    func(*http.Request) bool
}
```

#### Rate Limiting Middleware (`ratelimit.go`)
```go
func RateLimit(requests int, window time.Duration) Middleware
func RateLimitWithConfig(config RateLimitConfig) Middleware

type RateLimitConfig struct {
    Requests     int
    Window       time.Duration
    KeyFunc      func(*http.Request) string // Default: client IP
    ExceededHandler func(w http.ResponseWriter, r *http.Request)
    Store        RateLimitStore // In-memory by default
}

// RateLimitStore interface for custom storage
type RateLimitStore interface {
    Increment(key string, window time.Duration) (int, error)
    Reset(key string) error
}
```

#### Logging Middleware (`logging.go`)
```go
func Logging(logger Logger) Middleware
func LoggingWithConfig(config LoggingConfig) Middleware

type LoggingConfig struct {
    Logger          Logger
    SkipPaths       []string
    SkipFunc        func(*http.Request) bool
    FormatFunc      func(*http.Request, *ResponseWriter, time.Duration) string
    LogRequestBody  bool
    LogResponseBody bool
}
```

---

### 14. Request Validation (`validation.go`)

Request validation helpers.

```go
// Validator interface for custom validators
type Validator interface {
    Validate() error
}

// ValidationError represents validation failures
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   interface{} `json:"value,omitempty"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string

// Validation helpers
func Validate(v interface{}) error
func ValidateStruct(s interface{}) ValidationErrors

// Built-in validators
type Rules struct {
    Required    bool
    Min         int
    Max         int
    MinLength   int
    MaxLength   int
    Pattern     string
    Email       bool
    URL         bool
    UUID        bool
    OneOf       []string
    Custom      func(interface{}) error
}
```

---

### 15. Static File Serving (`static.go`)

Static file server utilities.

```go
// Static serves static files from a directory
func Static(urlPath, fsPath string) HandlerFunc

// StaticFS serves files from an http.FileSystem
func StaticFS(urlPath string, fs http.FileSystem) HandlerFunc

// StaticFile serves a single file
func StaticFile(urlPath, filePath string) HandlerFunc

// StaticConfig for advanced configuration
type StaticConfig struct {
    Root       string
    Index      string // Default: "index.html"
    Browse     bool   // Enable directory browsing
    MaxAge     int    // Cache-Control max-age
    Compress   bool   // Enable compression
    SkipFunc   func(*http.Request) bool
}

func StaticWithConfig(config StaticConfig) HandlerFunc
```

---

### 16. Metrics Interface (`metrics.go`)

Metrics collection interface for observability.

```go
// Metrics defines the metrics collection interface
type Metrics interface {
    IncrementRequestCount(method, path string, status int)
    ObserveRequestDuration(method, path string, duration time.Duration)
    IncrementActiveRequests()
    DecrementActiveRequests()
}

// NoopMetrics is a no-operation metrics implementation
type NoopMetrics struct{}

// MetricsMiddleware creates a middleware that collects metrics
func MetricsMiddleware(m Metrics) Middleware

// Built-in in-memory metrics for development
type InMemoryMetrics struct {
    // ...counters and histograms
}

func NewInMemoryMetrics() *InMemoryMetrics
func (m *InMemoryMetrics) Handler() HandlerFunc // Returns metrics as JSON
```

---

### 17. Testing Utilities (`testutil/`)

Testing helpers for users of the package.

```go
// testutil/test_server.go
type TestServer struct {
    Server   *Server
    URL      string
    Client   *http.Client
}

func NewTestServer(opts ...Option) *TestServer
func (s *TestServer) Close()
func (s *TestServer) Do(req *http.Request) (*http.Response, error)
func (s *TestServer) Get(path string) (*http.Response, error)
func (s *TestServer) Post(path string, body interface{}) (*http.Response, error)
func (s *TestServer) Put(path string, body interface{}) (*http.Response, error)
func (s *TestServer) Delete(path string) (*http.Response, error)

// testutil/assertions.go
func AssertStatus(t *testing.T, resp *http.Response, expected int)
func AssertJSON(t *testing.T, resp *http.Response, expected interface{})
func AssertHeader(t *testing.T, resp *http.Response, key, expected string)
func AssertBodyContains(t *testing.T, resp *http.Response, expected string)

// testutil/mock_handler.go
type MockHandler struct {
    Calls   []MockCall
    Handler HandlerFunc
}

type MockCall struct {
    Method  string
    Path    string
    Headers http.Header
    Body    []byte
}

func NewMockHandler(handler HandlerFunc) *MockHandler
func (m *MockHandler) AssertCalled(t *testing.T, method, path string)
```

---

## Functional Options Pattern (`options.go`)

```go
type Option func(*Server)

// Server options
func WithConfig(cfg *Config) Option
func WithLogger(logger Logger) Option
func WithRouter(router *Router) Option
func WithMiddleware(middlewares ...Middleware) Option
func WithErrorHandler(handler ErrorHandler) Option
func WithHealthChecker(checker *CompositeHealthChecker) Option
func WithMetrics(metrics Metrics) Option
func WithTLS(certFile, keyFile string) Option
func WithAddr(addr string) Option
func WithReadTimeout(timeout time.Duration) Option
func WithWriteTimeout(timeout time.Duration) Option
func WithIdleTimeout(timeout time.Duration) Option
func WithMaxHeaderBytes(size int) Option
func WithShutdownTimeout(timeout time.Duration) Option
func WithGracefulShutdown(enabled bool) Option
func WithRequestIDGenerator(gen func() string) Option
```

---

## Usage Examples

### Basic Usage
```go
package main

import (
    "net/http"
    "github.com/rompi/core-backend/pkg/httpserver"
)

func main() {
    // Create server with defaults
    server := httpserver.NewServer()

    // Register routes
    server.Router().GET("/", func(w http.ResponseWriter, r *http.Request) error {
        return httpserver.JSON(w, http.StatusOK, map[string]string{
            "message": "Hello, World!",
        })
    })

    server.Router().GET("/users/:id", getUser)
    server.Router().POST("/users", createUser)

    // Start server
    server.ListenAndServe()
}

func getUser(w http.ResponseWriter, r *http.Request) error {
    id := httpserver.PathParam(r, "id")
    // ... fetch user
    return httpserver.JSON(w, http.StatusOK, user)
}

func createUser(w http.ResponseWriter, r *http.Request) error {
    var req CreateUserRequest
    if err := httpserver.BindJSON(r, &req); err != nil {
        return err
    }
    // ... create user
    return httpserver.Created(w, "/users/"+user.ID, user)
}
```

### With Configuration
```go
server := httpserver.NewServer(
    httpserver.WithAddr(":8080"),
    httpserver.WithReadTimeout(30 * time.Second),
    httpserver.WithWriteTimeout(30 * time.Second),
    httpserver.WithLogger(myLogger),
    httpserver.WithMiddleware(
        httpserver.Recovery(),
        httpserver.RequestID(),
        httpserver.Logging(myLogger),
        httpserver.CORS(),
        httpserver.Compress(),
    ),
)
```

### Route Groups
```go
router := server.Router()

// API v1 group
v1 := router.Group("/api/v1", authMiddleware)
{
    v1.GET("/users", listUsers)
    v1.POST("/users", createUser)
    v1.GET("/users/:id", getUser)
    v1.PUT("/users/:id", updateUser)
    v1.DELETE("/users/:id", deleteUser)
}

// Public routes
router.GET("/health", httpserver.LivenessHandler())
router.GET("/ready", httpserver.ReadinessHandler(healthChecker))
```

### Integration with Auth Package
```go
import (
    "github.com/rompi/core-backend/pkg/auth"
    "github.com/rompi/core-backend/pkg/httpserver"
)

func main() {
    // Setup auth service
    authService, _ := auth.NewService(authConfig, repos)

    // Create server
    server := httpserver.NewServer(
        httpserver.WithMiddleware(
            httpserver.Recovery(),
            httpserver.RequestID(),
            httpserver.Logging(logger),
        ),
    )

    router := server.Router()

    // Public routes
    router.POST("/auth/login", loginHandler(authService))
    router.POST("/auth/register", registerHandler(authService))

    // Protected routes with auth middleware
    protected := router.Group("/api", httpserver.Adapt(authService.Middleware()))
    {
        protected.GET("/profile", profileHandler)
        protected.PUT("/profile", updateProfileHandler)
    }

    // Admin routes
    admin := router.Group("/admin",
        httpserver.Adapt(authService.Middleware()),
        httpserver.Adapt(authService.RequireRole("admin")),
    )
    {
        admin.GET("/users", adminListUsers)
    }

    server.ListenAndServe()
}
```

---

## Implementation Order

### Phase 1: Core Foundation
1. `errors.go` - Error types and handling
2. `logger.go` - Logger interface
3. `context.go` - Context utilities
4. `config.go` - Configuration
5. `handler.go` - Handler types
6. `response.go` - Response helpers
7. `request.go` - Request helpers
8. `middleware.go` - Middleware types and chaining
9. `router.go` - Router with path parameters
10. `server.go` - Main server implementation

### Phase 2: Essential Middleware
11. `recovery.go` - Panic recovery
12. `requestid.go` - Request ID
13. `timeout.go` - Request timeout
14. `cors.go` - CORS support

### Phase 3: Advanced Features
15. `health.go` - Health checks
16. `graceful.go` - Graceful shutdown
17. `compress.go` - Response compression
18. `ratelimit.go` - Rate limiting
19. `validation.go` - Request validation
20. `binding.go` - Request body binding

### Phase 4: Production Features
21. `tls.go` - TLS configuration
22. `metrics.go` - Metrics interface
23. `static.go` - Static file serving
24. `options.go` - Functional options

### Phase 5: Testing and Documentation
25. `testutil/` - Testing utilities
26. `examples/` - Example implementations
27. `README.md` - Package documentation
28. `doc.go` - Godoc documentation
29. All `*_test.go` files

---

## Non-Goals (Out of Scope)

1. **WebSocket support** - May be added in a future version
2. **HTTP/2 Server Push** - Deprecated in most browsers
3. **GraphQL support** - Should be a separate package
4. **Template rendering** - Use standard library or third-party
5. **Session management** - Handled by auth package
6. **Database integration** - Keep package focused on HTTP

---

## Success Criteria

1. Zero external dependencies (stdlib only)
2. 90%+ test coverage
3. Comprehensive documentation with examples
4. Seamless integration with existing auth and httpclient packages
5. Performance on par with net/http
6. Memory efficient with minimal allocations
7. Thread-safe for concurrent use
8. Easy to extend with custom middleware

---

## References

- Go net/http package documentation
- Existing patterns from pkg/auth and pkg/httpclient
- coding-guidelines.md for style conventions
