# HTTP Server Package Plan

## Overview

This document outlines the comprehensive plan for creating a production-ready HTTP server package (`pkg/httpserver`) built on top of the **Echo framework**. The package will provide a plug-and-play solution that wraps Echo with sensible defaults, custom configurations, and seamless integration with existing codebase packages (auth, httpclient).

## Why Echo Framework?

Echo is a high-performance, extensible, minimalist Go web framework that provides:
- **High Performance**: One of the fastest Go web frameworks
- **Extensible Middleware**: Rich middleware ecosystem
- **Optimized Router**: Radix tree based routing with zero dynamic memory allocation
- **Data Binding**: Automatic binding of request payload (JSON, XML, form)
- **Data Validation**: Built-in validation with go-playground/validator
- **Error Handling**: Centralized HTTP error handling
- **Template Rendering**: Support for any template engine
- **Scalability**: Battle-tested in production environments

## Design Goals

1. **Plug-and-Play**: Easy to set up with sensible defaults, minimal configuration required
2. **Echo-Powered**: Leverage Echo's performance and feature set
3. **Production-Ready**: Pre-configured with essential production features
4. **Wrapper Pattern**: Thin abstraction over Echo for consistency and ease of use
5. **Testable**: Interface-driven design with testing utilities
6. **Consistent**: Follows existing codebase patterns (auth, httpclient packages)

---

## Dependencies

```go
// go.mod additions
require (
    github.com/labstack/echo/v4 v4.11.4
    github.com/labstack/echo/v4/middleware
)
```

---

## Package Structure

```
pkg/httpserver/
├── server.go           # Core Server wrapper around Echo
├── config.go           # Configuration with environment variable support
├── context.go          # Extended context utilities
├── middleware.go       # Custom middleware and middleware adapters
├── handler.go          # Handler type adapters and utilities
├── errors.go           # Structured error types and error handlers
├── logger.go           # Logger adapter for Echo (bridges to our Logger interface)
├── health.go           # Health check handlers
├── binder.go           # Custom binder extensions
├── validator.go        # Custom validator setup
├── response.go         # Response helper utilities
├── auth_adapter.go     # Adapter for pkg/auth middleware integration
├── options.go          # Functional options pattern
├── doc.go              # Package documentation
├── server_test.go      # Server tests
├── middleware_test.go  # Middleware tests
├── *_test.go           # Other test files
├── testutil/           # Testing utilities
│   ├── test_server.go  # Test server helper
│   └── assertions.go   # Test assertions
├── examples/           # Example implementations
│   ├── basic/          # Minimal setup example
│   ├── rest-api/       # Full REST API example
│   ├── with-auth/      # Integration with auth package
│   └── custom-middleware/ # Custom middleware example
└── README.md           # Package documentation
```

---

## Core Components

### 1. Server (`server.go`)

The main Server struct that wraps Echo with additional lifecycle management.

```go
package httpserver

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

// Server wraps Echo with production-ready defaults
type Server struct {
    echo            *echo.Echo
    config          *Config
    logger          Logger
    healthChecker   *HealthChecker
    shutdownHooks   []ShutdownHook
}

// ShutdownHook is called during graceful shutdown
type ShutdownHook func(ctx context.Context) error

// NewServer creates a new HTTP server with the given options
func NewServer(opts ...Option) *Server

// Echo returns the underlying Echo instance for advanced configuration
func (s *Server) Echo() *echo.Echo

// --- Route Registration (delegated to Echo) ---

// GET registers a GET route
func (s *Server) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route

// POST registers a POST route
func (s *Server) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route

// PUT registers a PUT route
func (s *Server) PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route

// PATCH registers a PATCH route
func (s *Server) PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route

// DELETE registers a DELETE route
func (s *Server) DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route

// Group creates a route group with optional middleware
func (s *Server) Group(prefix string, m ...echo.MiddlewareFunc) *echo.Group

// Use adds middleware to the server
func (s *Server) Use(middleware ...echo.MiddlewareFunc)

// Static serves static files from a directory
func (s *Server) Static(prefix, root string)

// File serves a single file
func (s *Server) File(path, file string)

// --- Lifecycle ---

// Start starts the HTTP server
func (s *Server) Start() error

// StartTLS starts the HTTPS server
func (s *Server) StartTLS(certFile, keyFile string) error

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error

// ListenAndServe starts the server and blocks until shutdown
// Handles OS signals for graceful shutdown automatically
func (s *Server) ListenAndServe() error

// OnShutdown registers a hook to be called during shutdown
func (s *Server) OnShutdown(hook ShutdownHook)

// --- Health ---

// RegisterHealthCheck registers the health check endpoint
func (s *Server) RegisterHealthCheck()

// AddHealthChecker adds a health checker
func (s *Server) AddHealthChecker(checker HealthCheckerFunc)
```

**Default Middleware Stack:**
When `NewServer()` is called, the following middlewares are automatically configured:
1. Recovery (panic recovery)
2. Request ID
3. Logger (if logger provided)
4. Secure headers
5. CORS (if enabled in config)
6. Rate limiting (if enabled in config)
7. Gzip compression (if enabled in config)
8. Body limit
9. Timeout

---

### 2. Configuration (`config.go`)

Environment-driven configuration with validation.

```go
type Config struct {
    // Server
    Host            string        `env:"HTTP_HOST" envDefault:""`
    Port            int           `env:"HTTP_PORT" envDefault:"8080"`
    ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"30s"`
    WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
    ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"30s"`

    // TLS
    TLSEnabled      bool   `env:"HTTP_TLS_ENABLED" envDefault:"false"`
    TLSCertFile     string `env:"HTTP_TLS_CERT_FILE"`
    TLSKeyFile      string `env:"HTTP_TLS_KEY_FILE"`

    // Request Limits
    BodyLimit       string `env:"HTTP_BODY_LIMIT" envDefault:"10M"`

    // Features
    Debug           bool   `env:"HTTP_DEBUG" envDefault:"false"`

    // Health Check
    HealthCheckEnabled bool   `env:"HTTP_HEALTH_ENABLED" envDefault:"true"`
    HealthCheckPath    string `env:"HTTP_HEALTH_PATH" envDefault:"/health"`
    LivenessPath       string `env:"HTTP_LIVENESS_PATH" envDefault:"/health/live"`
    ReadinessPath      string `env:"HTTP_READINESS_PATH" envDefault:"/health/ready"`

    // CORS
    CORSEnabled          bool     `env:"HTTP_CORS_ENABLED" envDefault:"false"`
    CORSAllowOrigins     []string `env:"HTTP_CORS_ALLOW_ORIGINS" envDefault:"*"`
    CORSAllowMethods     []string `env:"HTTP_CORS_ALLOW_METHODS" envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
    CORSAllowHeaders     []string `env:"HTTP_CORS_ALLOW_HEADERS" envDefault:"Origin,Content-Type,Accept,Authorization,X-Request-ID"`
    CORSExposeHeaders    []string `env:"HTTP_CORS_EXPOSE_HEADERS"`
    CORSAllowCredentials bool     `env:"HTTP_CORS_ALLOW_CREDENTIALS" envDefault:"false"`
    CORSMaxAge           int      `env:"HTTP_CORS_MAX_AGE" envDefault:"86400"`

    // Rate Limiting
    RateLimitEnabled  bool          `env:"HTTP_RATE_LIMIT_ENABLED" envDefault:"false"`
    RateLimitRate     float64       `env:"HTTP_RATE_LIMIT_RATE" envDefault:"10"`
    RateLimitBurst    int           `env:"HTTP_RATE_LIMIT_BURST" envDefault:"30"`
    RateLimitExpiry   time.Duration `env:"HTTP_RATE_LIMIT_EXPIRY" envDefault:"3m"`

    // Compression
    GzipEnabled bool `env:"HTTP_GZIP_ENABLED" envDefault:"true"`
    GzipLevel   int  `env:"HTTP_GZIP_LEVEL" envDefault:"5"`

    // Request ID
    RequestIDEnabled bool   `env:"HTTP_REQUEST_ID_ENABLED" envDefault:"true"`
    RequestIDHeader  string `env:"HTTP_REQUEST_ID_HEADER" envDefault:"X-Request-ID"`

    // Timeout
    TimeoutEnabled bool          `env:"HTTP_TIMEOUT_ENABLED" envDefault:"true"`
    Timeout        time.Duration `env:"HTTP_TIMEOUT" envDefault:"30s"`

    // Secure Headers
    SecureEnabled bool `env:"HTTP_SECURE_ENABLED" envDefault:"true"`

    // Logging
    LogRequests      bool `env:"HTTP_LOG_REQUESTS" envDefault:"true"`
    LogSkipPaths     []string `env:"HTTP_LOG_SKIP_PATHS"`
}

// Functions:
// - LoadConfig() (*Config, error)
// - LoadConfigFromEnv() (*Config, error)
// - (c *Config) Validate() error
// - DefaultConfig() *Config
// - (c *Config) Address() string // Returns "host:port"
```

---

### 3. Context Utilities (`context.go`)

Extended context utilities that work with Echo's context.

```go
// Context keys for custom values
type contextKey string

const (
    LoggerContextKey contextKey = "httpserver-logger"
    UserContextKey   contextKey = "httpserver-user"
)

// GetRequestID returns the request ID from Echo context
func GetRequestID(c echo.Context) string

// GetLogger returns the logger from context
func GetLogger(c echo.Context) Logger

// SetLogger sets the logger in context
func SetLogger(c echo.Context, logger Logger)

// GetClientIP returns the real client IP (handles proxies)
func GetClientIP(c echo.Context) string

// GetUser returns the authenticated user from context (integration with auth package)
func GetUser(c echo.Context) interface{}

// SetUser sets the authenticated user in context
func SetUser(c echo.Context, user interface{})

// MustBind binds and validates request, returns HTTPError on failure
func MustBind(c echo.Context, v interface{}) error

// BindAndValidate binds request body and validates
func BindAndValidate(c echo.Context, v interface{}) error
```

---

### 4. Middleware (`middleware.go`)

Custom middleware and adapters.

```go
// MiddlewareFunc is an alias for echo.MiddlewareFunc
type MiddlewareFunc = echo.MiddlewareFunc

// HandlerFunc is an alias for echo.HandlerFunc
type HandlerFunc = echo.HandlerFunc

// --- Middleware Adapters ---

// AdaptStdMiddleware converts standard http middleware to Echo middleware
func AdaptStdMiddleware(m func(http.Handler) http.Handler) echo.MiddlewareFunc

// --- Custom Middlewares ---

// LoggerMiddleware creates a logging middleware using our Logger interface
func LoggerMiddleware(logger Logger) echo.MiddlewareFunc

// LoggerMiddlewareWithConfig creates a configurable logging middleware
func LoggerMiddlewareWithConfig(config LoggerMiddlewareConfig) echo.MiddlewareFunc

type LoggerMiddlewareConfig struct {
    Logger      Logger
    SkipPaths   []string
    SkipFunc    func(c echo.Context) bool
}

// MetricsMiddleware creates a metrics collection middleware
func MetricsMiddleware(metrics Metrics) echo.MiddlewareFunc

// RequestContextMiddleware adds request-scoped values to context
func RequestContextMiddleware(logger Logger) echo.MiddlewareFunc
```

**Using Echo's Built-in Middlewares:**
```go
import "github.com/labstack/echo/v4/middleware"

// These are automatically configured but can be customized:
// - middleware.Recover()
// - middleware.RequestID()
// - middleware.Logger()
// - middleware.CORSWithConfig(...)
// - middleware.RateLimiter(...)
// - middleware.GzipWithConfig(...)
// - middleware.BodyLimit(...)
// - middleware.TimeoutWithConfig(...)
// - middleware.Secure()
```

---

### 5. Handler Utilities (`handler.go`)

Handler adapters and utilities.

```go
// Handler is an alias for echo.HandlerFunc
type Handler = echo.HandlerFunc

// ErrorHandler is the custom error handler type
type ErrorHandler func(err error, c echo.Context)

// WrapHandler wraps a standard http.Handler to Echo handler
func WrapHandler(h http.Handler) echo.HandlerFunc

// WrapHandlerFunc wraps a standard http.HandlerFunc to Echo handler
func WrapHandlerFunc(h http.HandlerFunc) echo.HandlerFunc

// ToStdHandler converts Echo handler to standard http.Handler
func ToStdHandler(h echo.HandlerFunc, e *echo.Echo) http.Handler
```

---

### 6. Error Handling (`errors.go`)

Structured error types and centralized error handling.

```go
// HTTPError represents an HTTP error response
type HTTPError struct {
    Code      int         `json:"code"`
    Message   string      `json:"message"`
    Details   interface{} `json:"details,omitempty"`
    RequestID string      `json:"request_id,omitempty"`
    Internal  error       `json:"-"`
}

func (e *HTTPError) Error() string
func (e *HTTPError) Unwrap() error

// Error constructors
func NewHTTPError(code int, message string) *HTTPError
func NewHTTPErrorWithDetails(code int, message string, details interface{}) *HTTPError
func WrapError(code int, message string, err error) *HTTPError

// Common errors (pre-defined for convenience)
var (
    ErrBadRequest          = NewHTTPError(http.StatusBadRequest, "bad request")
    ErrUnauthorized        = NewHTTPError(http.StatusUnauthorized, "unauthorized")
    ErrForbidden           = NewHTTPError(http.StatusForbidden, "forbidden")
    ErrNotFound            = NewHTTPError(http.StatusNotFound, "not found")
    ErrMethodNotAllowed    = NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
    ErrConflict            = NewHTTPError(http.StatusConflict, "conflict")
    ErrUnprocessableEntity = NewHTTPError(http.StatusUnprocessableEntity, "unprocessable entity")
    ErrTooManyRequests     = NewHTTPError(http.StatusTooManyRequests, "too many requests")
    ErrInternalServer      = NewHTTPError(http.StatusInternalServerError, "internal server error")
    ErrServiceUnavailable  = NewHTTPError(http.StatusServiceUnavailable, "service unavailable")
)

// DefaultErrorHandler is the default centralized error handler
func DefaultErrorHandler(err error, c echo.Context)

// CustomErrorHandler creates an error handler with custom options
func CustomErrorHandler(logger Logger, debug bool) echo.HTTPErrorHandler
```

**Error Response Format:**
```json
{
    "code": 400,
    "message": "validation failed",
    "details": {
        "field": "email",
        "error": "invalid email format"
    },
    "request_id": "abc-123-xyz"
}
```

---

### 7. Logger Adapter (`logger.go`)

Logger interface and Echo adapter.

```go
// Logger defines the logging interface (consistent with httpclient package)
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

// EchoLoggerAdapter adapts our Logger interface to Echo's logger
type EchoLoggerAdapter struct {
    logger Logger
}

func NewEchoLoggerAdapter(logger Logger) *EchoLoggerAdapter

// Implements echo.Logger interface
func (l *EchoLoggerAdapter) Output() io.Writer
func (l *EchoLoggerAdapter) SetOutput(w io.Writer)
func (l *EchoLoggerAdapter) Prefix() string
func (l *EchoLoggerAdapter) SetPrefix(p string)
func (l *EchoLoggerAdapter) Level() log.Lvl
func (l *EchoLoggerAdapter) SetLevel(v log.Lvl)
func (l *EchoLoggerAdapter) SetHeader(h string)
func (l *EchoLoggerAdapter) Print(i ...interface{})
func (l *EchoLoggerAdapter) Printf(format string, args ...interface{})
func (l *EchoLoggerAdapter) Printj(j log.JSON)
func (l *EchoLoggerAdapter) Debug(i ...interface{})
func (l *EchoLoggerAdapter) Debugf(format string, args ...interface{})
func (l *EchoLoggerAdapter) Debugj(j log.JSON)
func (l *EchoLoggerAdapter) Info(i ...interface{})
func (l *EchoLoggerAdapter) Infof(format string, args ...interface{})
func (l *EchoLoggerAdapter) Infoj(j log.JSON)
func (l *EchoLoggerAdapter) Warn(i ...interface{})
func (l *EchoLoggerAdapter) Warnf(format string, args ...interface{})
func (l *EchoLoggerAdapter) Warnj(j log.JSON)
func (l *EchoLoggerAdapter) Error(i ...interface{})
func (l *EchoLoggerAdapter) Errorf(format string, args ...interface{})
func (l *EchoLoggerAdapter) Errorj(j log.JSON)
func (l *EchoLoggerAdapter) Fatal(i ...interface{})
func (l *EchoLoggerAdapter) Fatalf(format string, args ...interface{})
func (l *EchoLoggerAdapter) Fatalj(j log.JSON)
func (l *EchoLoggerAdapter) Panic(i ...interface{})
func (l *EchoLoggerAdapter) Panicf(format string, args ...interface{})
func (l *EchoLoggerAdapter) Panicj(j log.JSON)
```

---

### 8. Health Checks (`health.go`)

Health and readiness check support.

```go
// HealthCheckerFunc is a function that performs a health check
type HealthCheckerFunc func(ctx context.Context) error

// HealthChecker manages multiple health checks
type HealthChecker struct {
    checkers map[string]HealthCheckerFunc
    mu       sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker

// Add adds a named health checker
func (h *HealthChecker) Add(name string, checker HealthCheckerFunc)

// Remove removes a health checker
func (h *HealthChecker) Remove(name string)

// Check runs all health checks
func (h *HealthChecker) Check(ctx context.Context) *HealthStatus

// HealthStatus represents the health status response
type HealthStatus struct {
    Status    string                 `json:"status"`    // "healthy", "unhealthy", "degraded"
    Timestamp time.Time              `json:"timestamp"`
    Duration  string                 `json:"duration"`
    Checks    map[string]CheckResult `json:"checks,omitempty"`
    Version   string                 `json:"version,omitempty"`
}

type CheckResult struct {
    Status   string `json:"status"`
    Duration string `json:"duration"`
    Error    string `json:"error,omitempty"`
}

// Health handlers
func HealthHandler(checker *HealthChecker) echo.HandlerFunc
func LivenessHandler() echo.HandlerFunc
func ReadinessHandler(checker *HealthChecker) echo.HandlerFunc

// Built-in health checkers
func PingChecker() HealthCheckerFunc
func DatabaseChecker(db interface{ PingContext(context.Context) error }) HealthCheckerFunc
func HTTPChecker(url string, timeout time.Duration) HealthCheckerFunc
```

---

### 9. Response Helpers (`response.go`)

Convenient response utilities.

```go
// JSON sends a JSON response (wrapper around c.JSON)
func JSON(c echo.Context, code int, data interface{}) error

// JSONPretty sends a pretty-printed JSON response
func JSONPretty(c echo.Context, code int, data interface{}) error

// Success sends a success response with data
func Success(c echo.Context, data interface{}) error

// Created sends a 201 Created response
func Created(c echo.Context, data interface{}) error

// NoContent sends a 204 No Content response
func NoContent(c echo.Context) error

// Error sends an error response
func Error(c echo.Context, err error) error

// ErrorWithCode sends an error response with a specific code
func ErrorWithCode(c echo.Context, code int, message string) error

// Paginated sends a paginated response
func Paginated(c echo.Context, data interface{}, total int64, page, pageSize int) error

// PaginatedResponse is the standard pagination response format
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
    Total      int64 `json:"total"`
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalPages int   `json:"total_pages"`
}

// Stream sends a streaming response
func Stream(c echo.Context, contentType string, reader io.Reader) error

// File sends a file response
func File(c echo.Context, filepath string) error

// Attachment sends a file as attachment
func Attachment(c echo.Context, filepath, filename string) error
```

---

### 10. Validator (`validator.go`)

Request validation setup.

```go
import "github.com/go-playground/validator/v10"

// CustomValidator wraps go-playground/validator
type CustomValidator struct {
    validator *validator.Validate
}

// NewValidator creates a new custom validator
func NewValidator() *CustomValidator

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error

// RegisterValidation registers a custom validation function
func (cv *CustomValidator) RegisterValidation(tag string, fn validator.Func) error

// ValidationError represents a validation error
type ValidationError struct {
    Field   string `json:"field"`
    Tag     string `json:"tag"`
    Value   string `json:"value,omitempty"`
    Message string `json:"message"`
}

// ValidationErrors is a slice of validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string

// TranslateValidationErrors converts validator errors to our format
func TranslateValidationErrors(err error) ValidationErrors
```

---

### 11. Auth Adapter (`auth_adapter.go`)

Integration with the existing auth package.

```go
import "github.com/rompi/core-backend/pkg/auth"

// AuthMiddleware adapts auth.Service middleware to Echo middleware
func AuthMiddleware(authService auth.Service) echo.MiddlewareFunc

// RequireRoleMiddleware creates middleware that requires specific roles
func RequireRoleMiddleware(authService auth.Service, roles ...string) echo.MiddlewareFunc

// RequirePermissionMiddleware creates middleware that requires specific permissions
func RequirePermissionMiddleware(authService auth.Service, permissions ...string) echo.MiddlewareFunc

// GetAuthUser returns the authenticated user from context
func GetAuthUser(c echo.Context) *auth.User

// OptionalAuthMiddleware sets user if authenticated but doesn't require it
func OptionalAuthMiddleware(authService auth.Service) echo.MiddlewareFunc
```

---

### 12. Functional Options (`options.go`)

```go
type Option func(*Server)

// WithConfig sets the server configuration
func WithConfig(cfg *Config) Option

// WithLogger sets the logger
func WithLogger(logger Logger) Option

// WithEcho sets a pre-configured Echo instance
func WithEcho(e *echo.Echo) Option

// WithMiddleware adds middleware to the server
func WithMiddleware(middleware ...echo.MiddlewareFunc) Option

// WithErrorHandler sets a custom error handler
func WithErrorHandler(handler echo.HTTPErrorHandler) Option

// WithValidator sets a custom validator
func WithValidator(validator echo.Validator) Option

// WithHealthChecker sets the health checker
func WithHealthChecker(checker *HealthChecker) Option

// WithBinder sets a custom binder
func WithBinder(binder echo.Binder) Option

// WithRenderer sets a custom renderer (for HTML templates)
func WithRenderer(renderer echo.Renderer) Option

// WithDebug enables debug mode
func WithDebug(debug bool) Option

// WithAddr sets the server address
func WithAddr(addr string) Option

// WithTLS configures TLS
func WithTLS(certFile, keyFile string) Option

// WithGracefulShutdown enables graceful shutdown with signals
func WithGracefulShutdown(enabled bool) Option

// WithShutdownTimeout sets the shutdown timeout
func WithShutdownTimeout(timeout time.Duration) Option

// WithCORS enables CORS with the given origins
func WithCORS(origins ...string) Option

// WithRateLimit enables rate limiting
func WithRateLimit(rate float64, burst int) Option

// WithMetrics sets the metrics collector
func WithMetrics(metrics Metrics) Option
```

---

### 13. Metrics Interface (`metrics.go`)

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

func (NoopMetrics) IncrementRequestCount(method, path string, status int)              {}
func (NoopMetrics) ObserveRequestDuration(method, path string, duration time.Duration) {}
func (NoopMetrics) IncrementActiveRequests()                                           {}
func (NoopMetrics) DecrementActiveRequests()                                           {}
```

---

### 14. Testing Utilities (`testutil/`)

```go
// testutil/test_server.go
package testutil

import (
    "github.com/rompi/core-backend/pkg/httpserver"
    "github.com/labstack/echo/v4"
    "net/http/httptest"
)

// TestServer provides testing utilities for httpserver
type TestServer struct {
    Server *httpserver.Server
    Echo   *echo.Echo
}

// NewTestServer creates a new test server
func NewTestServer(opts ...httpserver.Option) *TestServer

// Request creates a new test request
func (s *TestServer) Request(method, path string) *TestRequest

// TestRequest is a test request builder
type TestRequest struct {
    server  *TestServer
    method  string
    path    string
    body    interface{}
    headers map[string]string
    query   map[string]string
}

func (r *TestRequest) WithBody(body interface{}) *TestRequest
func (r *TestRequest) WithHeader(key, value string) *TestRequest
func (r *TestRequest) WithQuery(key, value string) *TestRequest
func (r *TestRequest) WithAuth(token string) *TestRequest
func (r *TestRequest) Do() *TestResponse

// TestResponse wraps the response for assertions
type TestResponse struct {
    *httptest.ResponseRecorder
}

func (r *TestResponse) Status() int
func (r *TestResponse) JSON(v interface{}) error
func (r *TestResponse) String() string

// testutil/assertions.go
func AssertStatus(t *testing.T, resp *TestResponse, expected int)
func AssertJSON(t *testing.T, resp *TestResponse, expected interface{})
func AssertHeader(t *testing.T, resp *TestResponse, key, expected string)
func AssertBodyContains(t *testing.T, resp *TestResponse, expected string)
func AssertNoError(t *testing.T, err error)
```

---

## Usage Examples

### Basic Usage
```go
package main

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/rompi/core-backend/pkg/httpserver"
)

func main() {
    // Create server with defaults
    server := httpserver.NewServer()

    // Register routes
    server.GET("/", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "message": "Hello, World!",
        })
    })

    server.GET("/users/:id", getUser)
    server.POST("/users", createUser)

    // Start server (blocks until shutdown)
    server.ListenAndServe()
}

func getUser(c echo.Context) error {
    id := c.Param("id")
    // ... fetch user
    return c.JSON(http.StatusOK, user)
}

func createUser(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return err
    }
    if err := c.Validate(&req); err != nil {
        return err
    }
    // ... create user
    return c.JSON(http.StatusCreated, user)
}
```

### With Configuration
```go
config := httpserver.DefaultConfig()
config.Port = 3000
config.CORSEnabled = true
config.RateLimitEnabled = true
config.RateLimitRate = 100

server := httpserver.NewServer(
    httpserver.WithConfig(config),
    httpserver.WithLogger(myLogger),
    httpserver.WithDebug(true),
)
```

### Route Groups
```go
server := httpserver.NewServer()

// API v1 group
v1 := server.Group("/api/v1")
{
    v1.GET("/users", listUsers)
    v1.POST("/users", createUser)
    v1.GET("/users/:id", getUser)
    v1.PUT("/users/:id", updateUser)
    v1.DELETE("/users/:id", deleteUser)
}

// Admin routes with additional middleware
admin := server.Group("/admin", adminMiddleware)
{
    admin.GET("/stats", getStats)
}

server.ListenAndServe()
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
        httpserver.WithLogger(logger),
    )

    // Public routes
    server.POST("/auth/login", loginHandler(authService))
    server.POST("/auth/register", registerHandler(authService))

    // Protected routes with auth middleware
    api := server.Group("/api", httpserver.AuthMiddleware(authService))
    {
        api.GET("/profile", profileHandler)
        api.PUT("/profile", updateProfileHandler)
    }

    // Admin routes with role requirement
    admin := server.Group("/admin",
        httpserver.AuthMiddleware(authService),
        httpserver.RequireRoleMiddleware(authService, "admin"),
    )
    {
        admin.GET("/users", adminListUsers)
        admin.DELETE("/users/:id", adminDeleteUser)
    }

    server.ListenAndServe()
}

func profileHandler(c echo.Context) error {
    user := httpserver.GetAuthUser(c)
    return c.JSON(http.StatusOK, user)
}
```

### Custom Health Checks
```go
server := httpserver.NewServer()

// Add custom health checkers
server.AddHealthChecker("database", httpserver.DatabaseChecker(db))
server.AddHealthChecker("redis", func(ctx context.Context) error {
    return redisClient.Ping(ctx).Err()
})
server.AddHealthChecker("external-api", httpserver.HTTPChecker("https://api.example.com/health", 5*time.Second))

// Health endpoints are automatically registered:
// GET /health       - Combined health status
// GET /health/live  - Liveness probe (always returns 200 if server is running)
// GET /health/ready - Readiness probe (checks all health checkers)

server.ListenAndServe()
```

### Custom Error Handling
```go
server := httpserver.NewServer(
    httpserver.WithErrorHandler(func(err error, c echo.Context) {
        // Custom error handling logic
        code := http.StatusInternalServerError
        message := "Internal Server Error"

        if he, ok := err.(*httpserver.HTTPError); ok {
            code = he.Code
            message = he.Message
        } else if he, ok := err.(*echo.HTTPError); ok {
            code = he.Code
            message = fmt.Sprintf("%v", he.Message)
        }

        // Log error
        logger.Error("request error",
            "error", err,
            "code", code,
            "request_id", httpserver.GetRequestID(c),
        )

        // Send response
        c.JSON(code, map[string]interface{}{
            "error":      message,
            "request_id": httpserver.GetRequestID(c),
        })
    }),
)
```

### Using Response Helpers
```go
func listUsers(c echo.Context) error {
    page := c.QueryParam("page")
    pageSize := c.QueryParam("page_size")

    users, total, err := userService.List(page, pageSize)
    if err != nil {
        return httpserver.Error(c, err)
    }

    return httpserver.Paginated(c, users, total, page, pageSize)
}

func createUser(c echo.Context) error {
    var req CreateUserRequest
    if err := httpserver.BindAndValidate(c, &req); err != nil {
        return err
    }

    user, err := userService.Create(req)
    if err != nil {
        return httpserver.WrapError(http.StatusConflict, "user already exists", err)
    }

    return httpserver.Created(c, user)
}
```

---

## Implementation Order

### Phase 1: Core Foundation
1. `errors.go` - Error types and handling
2. `logger.go` - Logger interface and Echo adapter
3. `config.go` - Configuration
4. `options.go` - Functional options
5. `server.go` - Main server wrapper

### Phase 2: Request/Response
6. `context.go` - Context utilities
7. `response.go` - Response helpers
8. `validator.go` - Custom validator
9. `binder.go` - Custom binder extensions

### Phase 3: Middleware & Features
10. `middleware.go` - Custom middleware
11. `handler.go` - Handler utilities
12. `health.go` - Health checks
13. `metrics.go` - Metrics interface

### Phase 4: Integration
14. `auth_adapter.go` - Auth package integration

### Phase 5: Testing and Documentation
15. `testutil/` - Testing utilities
16. `examples/` - Example implementations
17. `README.md` - Package documentation
18. `doc.go` - Godoc documentation
19. All `*_test.go` files

---

## Comparison: Echo vs Custom Implementation

| Feature | Custom (Previous Plan) | Echo-Based (This Plan) |
|---------|----------------------|------------------------|
| Router Performance | Good | Excellent (radix tree) |
| Path Parameters | Custom impl needed | Built-in |
| Middleware System | Custom impl needed | Built-in, extensive |
| Request Binding | Custom impl needed | Built-in (JSON, XML, Form) |
| Validation | Custom impl needed | Built-in (go-playground/validator) |
| Error Handling | Custom impl needed | Built-in, centralized |
| WebSocket | Out of scope | Supported |
| HTTP/2 | Manual setup | Built-in |
| Dependencies | stdlib only | Echo + validator |
| Development Time | ~4-6 weeks | ~1-2 weeks |
| Maintenance | High | Low (community maintained) |
| Documentation | Custom | Extensive (echo.labstack.com) |

---

## Non-Goals (Out of Scope)

1. **GraphQL support** - Should be a separate package
2. **Template rendering** - Use Echo's built-in or third-party
3. **Session management** - Handled by auth package
4. **Database integration** - Keep package focused on HTTP
5. **Custom router** - Use Echo's optimized router

---

## Success Criteria

1. Seamless integration with Echo ecosystem
2. Easy plug-and-play for upstream services
3. 90%+ test coverage
4. Comprehensive documentation with examples
5. Seamless integration with existing auth and httpclient packages
6. Sensible production-ready defaults
7. Support for all common HTTP patterns
8. Easy to extend with custom middleware

---

## References

- [Echo Framework Documentation](https://echo.labstack.com/)
- [Echo GitHub Repository](https://github.com/labstack/echo)
- [go-playground/validator](https://github.com/go-playground/validator)
- Existing patterns from pkg/auth and pkg/httpclient
- coding-guidelines.md for style conventions
