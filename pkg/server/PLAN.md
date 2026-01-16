# Server Package Plan

## Overview

This document outlines the comprehensive plan for creating a production-ready server package (`pkg/server`) that provides **unified gRPC and HTTP/REST API** using **gRPC-Gateway**. The package uses Protocol Buffers as the single source of truth for API definitions, automatically generating both gRPC services and RESTful HTTP endpoints.

## Why gRPC-Gateway?

- **Single Source of Truth**: Define APIs once in `.proto` files, serve both gRPC and HTTP
- **Type Safety**: Protocol Buffers provide strong typing and validation
- **Auto-Generated Documentation**: OpenAPI/Swagger specs generated from protos
- **Consistent Behavior**: Same business logic serves both gRPC and REST clients
- **Performance**: Native gRPC for internal services, HTTP/JSON for external clients

## Design Goals

1. **Proto-First**: All APIs defined in Protocol Buffer files
2. **Unified Server**: Single package manages both gRPC and HTTP servers
3. **Production-Ready**: Built-in interceptors, middleware, health checks, graceful shutdown
4. **Plug-and-Play**: Easy to set up with sensible defaults
5. **Zero Internal Dependencies**: Uses interfaces for auth, logging - no coupling to internal packages
6. **Consistent**: Follows existing codebase patterns

---

## Architecture

```
                                    ┌─────────────────────────────────────────┐
                                    │            Your Application             │
                                    └─────────────────────────────────────────┘
                                                       │
                         ┌─────────────────────────────┴─────────────────────────────┐
                         │                        pkg/server                          │
                         │  ┌─────────────────────────────────────────────────────┐  │
                         │  │                   Server (unified)                   │  │
                         │  └─────────────────────────────────────────────────────┘  │
                         │         │                                     │           │
                         │         ▼                                     ▼           │
                         │  ┌─────────────┐                      ┌─────────────┐     │
                         │  │gRPC Server  │◀────────────────────▶│ HTTP Server │     │
                         │  │  :9090      │   gRPC-Gateway       │   :8080     │     │
                         │  │             │   (reverse proxy)    │  (net/http) │     │
                         │  └─────────────┘                      └─────────────┘     │
                         │         │                                     │           │
                         │         ▼                                     ▼           │
                         │  ┌─────────────┐                      ┌─────────────┐     │
                         │  │Interceptors │                      │ Middleware  │     │
                         │  │- Auth       │                      │- Auth       │     │
                         │  │- Logging    │                      │- Logging    │     │
                         │  │- Recovery   │                      │- Recovery   │     │
                         │  │- RateLimit  │                      │- RateLimit  │     │
                         │  │- Validation │                      │- CORS       │     │
                         │  └─────────────┘                      └─────────────┘     │
                         └───────────────────────────────────────────────────────────┘
                                    │                                     │
                         ┌──────────┴──────────┐               ┌─────────┴─────────┐
                         │    gRPC Clients     │               │   HTTP Clients    │
                         │ (internal services) │               │ (web, mobile)     │
                         └─────────────────────┘               └───────────────────┘
```

---

## Dependencies

```go
// go.mod additions
require (
    google.golang.org/grpc v1.60.0
    google.golang.org/protobuf v1.32.0
    github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0
    github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.1
    golang.org/x/time v0.5.0  // for rate limiting
)
```

**Required protoc plugins:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

---

## Package Structure

```
pkg/server/
├── server.go              # Unified server managing gRPC + HTTP
├── config.go              # Configuration with environment variable support
├── options.go             # Functional options pattern
├── errors.go              # Unified error handling (gRPC status ↔ HTTP)
├── logger.go              # Logger interface
├── auth.go                # Auth interfaces (Authenticator, User) - no internal deps
├── grpc/
│   ├── server.go          # gRPC server wrapper
│   ├── interceptor.go     # Interceptor chain builder
│   ├── auth.go            # Auth interceptor (integrates with pkg/auth)
│   ├── logging.go         # Logging interceptor
│   ├── recovery.go        # Panic recovery interceptor
│   ├── ratelimit.go       # Rate limiting interceptor (per-method)
│   ├── validation.go      # Request validation interceptor
│   ├── requestid.go       # Request ID interceptor
│   └── metadata.go        # Metadata utilities
├── gateway/
│   ├── gateway.go         # gRPC-Gateway setup and registration
│   ├── mux.go             # HTTP mux with custom handlers support
│   ├── middleware.go      # HTTP middleware chain builder
│   ├── auth.go            # Auth middleware for HTTP
│   ├── logging.go         # Request logging middleware
│   ├── recovery.go        # Panic recovery middleware
│   ├── ratelimit.go       # Rate limiting middleware (per-endpoint)
│   ├── cors.go            # CORS middleware
│   ├── requestid.go       # Request ID middleware
│   └── compress.go        # Response compression middleware
├── health/
│   ├── health.go          # Health check service (gRPC + HTTP)
│   ├── checker.go         # Health checker interface and implementations
│   └── probes.go          # Kubernetes liveness/readiness probes
├── doc.go                 # Package documentation
├── server_test.go         # Server tests
├── grpc/
│   └── *_test.go          # gRPC interceptor tests
├── gateway/
│   └── *_test.go          # HTTP middleware tests
├── testutil/
│   ├── test_server.go     # Test server helper
│   ├── grpc_client.go     # Test gRPC client helper
│   └── assertions.go      # Test assertions
├── examples/
│   ├── proto/             # Example proto definitions
│   │   ├── user/v1/
│   │   │   └── user.proto
│   │   └── buf.yaml
│   ├── basic/             # Basic server example
│   ├── with-auth/         # Auth integration example
│   ├── rate-limiting/     # Per-endpoint rate limiting example
│   └── custom-handlers/   # Custom HTTP handlers example
└── README.md              # Package documentation
```

---

## Core Components

### 1. Server (`server.go`)

The unified server that manages both gRPC and HTTP servers.

```go
package server

import (
    "context"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "google.golang.org/grpc"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Server manages both gRPC and HTTP servers
type Server struct {
    config        *Config
    logger        Logger
    grpcServer    *grpc.Server
    httpServer    *http.Server
    gatewayMux    *runtime.ServeMux
    httpMux       *http.ServeMux
    healthChecker *health.Checker
    shutdownHooks []ShutdownHook

    // Interceptors and middleware
    unaryInterceptors  []grpc.UnaryServerInterceptor
    streamInterceptors []grpc.StreamServerInterceptor
    httpMiddleware     []Middleware
}

// ShutdownHook is called during graceful shutdown
type ShutdownHook func(ctx context.Context) error

// NewServer creates a new unified server
func NewServer(opts ...Option) (*Server, error)

// --- gRPC Registration ---

// GRPCServer returns the underlying gRPC server for service registration
func (s *Server) GRPCServer() *grpc.Server

// RegisterService is a helper to register a gRPC service
// Example: server.RegisterService(&pb.UserService_ServiceDesc, userServiceImpl)
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{})

// --- Gateway Registration ---

// RegisterGateway registers a gRPC-Gateway handler
// The handler is generated by protoc-gen-grpc-gateway
// Example: server.RegisterGateway(pb.RegisterUserServiceHandlerFromEndpoint)
func (s *Server) RegisterGateway(
    registerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error,
) error

// RegisterGatewayHandler registers a gateway handler directly (for in-process)
// Example: server.RegisterGatewayHandler(pb.RegisterUserServiceHandler, userServiceImpl)
func (s *Server) RegisterGatewayHandler(
    registerFunc func(ctx context.Context, mux *runtime.ServeMux, server interface{}) error,
    impl interface{},
) error

// --- Custom HTTP Handlers ---

// HandleFunc registers a custom HTTP handler (for non-proto endpoints)
// Example: server.HandleFunc("POST /webhooks/stripe", stripeHandler)
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc)

// Handle registers a custom HTTP handler
func (s *Server) Handle(pattern string, handler http.Handler)

// --- Lifecycle ---

// Start starts both gRPC and HTTP servers (non-blocking)
func (s *Server) Start() error

// ListenAndServe starts the servers and blocks until shutdown
// Handles OS signals (SIGINT, SIGTERM) for graceful shutdown
func (s *Server) ListenAndServe() error

// Shutdown gracefully shuts down both servers
func (s *Server) Shutdown(ctx context.Context) error

// OnShutdown registers a hook to be called during shutdown
func (s *Server) OnShutdown(hook ShutdownHook)

// --- Health ---

// AddHealthCheck adds a named health checker
func (s *Server) AddHealthCheck(name string, checker health.CheckerFunc)

// --- Accessors ---

// GRPCAddr returns the gRPC server address
func (s *Server) GRPCAddr() string

// HTTPAddr returns the HTTP server address
func (s *Server) HTTPAddr() string
```

---

### 2. Configuration (`config.go`)

Environment-driven configuration with validation.

```go
type Config struct {
    // gRPC Server
    GRPCHost string `env:"GRPC_HOST" envDefault:""`
    GRPCPort int    `env:"GRPC_PORT" envDefault:"9090"`

    // HTTP Server (Gateway)
    HTTPHost         string        `env:"HTTP_HOST" envDefault:""`
    HTTPPort         int           `env:"HTTP_PORT" envDefault:"8080"`
    HTTPReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"30s"`
    HTTPWriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
    HTTPIdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"120s"`

    // Shutdown
    ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`

    // TLS (applies to both gRPC and HTTP)
    TLSEnabled  bool   `env:"TLS_ENABLED" envDefault:"false"`
    TLSCertFile string `env:"TLS_CERT_FILE"`
    TLSKeyFile  string `env:"TLS_KEY_FILE"`

    // Health Checks
    HealthEnabled     bool   `env:"HEALTH_ENABLED" envDefault:"true"`
    HealthHTTPPath    string `env:"HEALTH_HTTP_PATH" envDefault:"/health"`
    LivenessHTTPPath  string `env:"LIVENESS_HTTP_PATH" envDefault:"/health/live"`
    ReadinessHTTPPath string `env:"READINESS_HTTP_PATH" envDefault:"/health/ready"`

    // CORS (HTTP only)
    CORSEnabled          bool     `env:"CORS_ENABLED" envDefault:"false"`
    CORSAllowOrigins     []string `env:"CORS_ALLOW_ORIGINS" envDefault:"*"`
    CORSAllowMethods     []string `env:"CORS_ALLOW_METHODS" envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
    CORSAllowHeaders     []string `env:"CORS_ALLOW_HEADERS" envDefault:"Origin,Content-Type,Accept,Authorization,X-Request-ID"`
    CORSExposeHeaders    []string `env:"CORS_EXPOSE_HEADERS"`
    CORSAllowCredentials bool     `env:"CORS_ALLOW_CREDENTIALS" envDefault:"false"`
    CORSMaxAge           int      `env:"CORS_MAX_AGE" envDefault:"86400"`

    // Global Rate Limiting (applies to all endpoints as default)
    // For per-endpoint rate limiting, use RateLimit interceptor/middleware on specific methods
    RateLimitEnabled bool          `env:"RATE_LIMIT_ENABLED" envDefault:"false"`
    RateLimitRate    float64       `env:"RATE_LIMIT_RATE" envDefault:"20"`
    RateLimitBurst   int           `env:"RATE_LIMIT_BURST" envDefault:"40"`
    RateLimitExpiry  time.Duration `env:"RATE_LIMIT_EXPIRY" envDefault:"3m"`

    // Compression (HTTP only)
    CompressionEnabled bool `env:"COMPRESSION_ENABLED" envDefault:"true"`

    // Request ID
    RequestIDEnabled bool   `env:"REQUEST_ID_ENABLED" envDefault:"true"`
    RequestIDHeader  string `env:"REQUEST_ID_HEADER" envDefault:"X-Request-ID"`

    // Logging
    LogRequests  bool     `env:"LOG_REQUESTS" envDefault:"true"`
    LogSkipPaths []string `env:"LOG_SKIP_PATHS"`

    // Debug
    Debug bool `env:"DEBUG" envDefault:"false"`
}

// Functions
func LoadConfig() (*Config, error)
func LoadConfigFromEnv() (*Config, error)
func DefaultConfig() *Config
func (c *Config) Validate() error
func (c *Config) GRPCAddr() string  // Returns "host:port"
func (c *Config) HTTPAddr() string  // Returns "host:port"
```

---

### 3. Functional Options (`options.go`)

```go
type Option func(*Server) error

// --- Configuration ---
func WithConfig(cfg *Config) Option
func WithLogger(logger Logger) Option

// --- gRPC Options ---
func WithGRPCAddr(addr string) Option
func WithUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) Option
func WithStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) Option
func WithGRPCServerOption(opts ...grpc.ServerOption) Option

// --- HTTP/Gateway Options ---
func WithHTTPAddr(addr string) Option
func WithHTTPMiddleware(middleware ...Middleware) Option
func WithGatewayOption(opts ...runtime.ServeMuxOption) Option

// --- Features ---
func WithTLS(certFile, keyFile string) Option
func WithCORS(origins ...string) Option
func WithRateLimit(rate float64, burst int) Option
func WithHealthChecker(checker *health.Checker) Option
func WithGracefulShutdown(enabled bool) Option
func WithShutdownTimeout(timeout time.Duration) Option

// --- Auth Integration (uses interfaces, no internal deps) ---
func WithAuthenticator(auth Authenticator) Option
```

---

### 4. Error Handling (`errors.go`)

Unified error handling that maps between gRPC status codes and HTTP status codes.

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// Error represents an application error that can be returned from both gRPC and HTTP
type Error struct {
    Code     codes.Code  `json:"-"`
    HTTPCode int         `json:"code"`
    Message  string      `json:"message"`
    Details  interface{} `json:"details,omitempty"`
    Internal error       `json:"-"`
}

func (e *Error) Error() string
func (e *Error) GRPCStatus() *status.Status
func (e *Error) Unwrap() error

// Error constructors
func NewError(code codes.Code, message string) *Error
func NewErrorWithDetails(code codes.Code, message string, details interface{}) *Error
func WrapError(code codes.Code, message string, err error) *Error

// FromGRPCError converts a gRPC error to our Error type
func FromGRPCError(err error) *Error

// FromHTTPStatus creates an Error from HTTP status code
func FromHTTPStatus(httpCode int, message string) *Error

// Common errors
var (
    ErrInvalidArgument    = NewError(codes.InvalidArgument, "invalid argument")
    ErrNotFound           = NewError(codes.NotFound, "not found")
    ErrAlreadyExists      = NewError(codes.AlreadyExists, "already exists")
    ErrPermissionDenied   = NewError(codes.PermissionDenied, "permission denied")
    ErrUnauthenticated    = NewError(codes.Unauthenticated, "unauthenticated")
    ErrResourceExhausted  = NewError(codes.ResourceExhausted, "rate limit exceeded")
    ErrInternal           = NewError(codes.Internal, "internal error")
    ErrUnavailable        = NewError(codes.Unavailable, "service unavailable")
)

// gRPC code to HTTP status mapping
var grpcToHTTP = map[codes.Code]int{
    codes.OK:                 200,
    codes.Canceled:           499,
    codes.Unknown:            500,
    codes.InvalidArgument:    400,
    codes.DeadlineExceeded:   504,
    codes.NotFound:           404,
    codes.AlreadyExists:      409,
    codes.PermissionDenied:   403,
    codes.ResourceExhausted:  429,
    codes.FailedPrecondition: 400,
    codes.Aborted:            409,
    codes.OutOfRange:         400,
    codes.Unimplemented:      501,
    codes.Internal:           500,
    codes.Unavailable:        503,
    codes.DataLoss:           500,
    codes.Unauthenticated:    401,
}

func GRPCCodeToHTTP(code codes.Code) int
func HTTPToGRPCCode(httpCode int) codes.Code
```

---

### 5. Auth Interfaces (`auth.go`)

Auth interfaces that allow integration with any authentication system without coupling to internal packages.

```go
package server

import (
    "context"
    "net/http"
)

// User represents an authenticated user
// Implement this interface to integrate with your auth system
type User interface {
    // GetID returns the user's unique identifier
    GetID() string

    // GetRole returns the user's role (e.g., "admin", "user")
    GetRole() string

    // GetPermissions returns the user's permissions
    GetPermissions() []string

    // HasRole checks if user has the specified role
    HasRole(role string) bool

    // HasPermission checks if user has the specified permission
    HasPermission(permission string) bool
}

// Authenticator validates tokens and returns user information
// Implement this interface to integrate with your auth system (e.g., pkg/auth)
type Authenticator interface {
    // ValidateToken validates a token and returns the authenticated user
    // Returns nil user and nil error if token is missing (for optional auth)
    // Returns nil user and error if token is invalid
    ValidateToken(ctx context.Context, token string) (User, error)
}

// RoleChecker checks if a user has required roles
// Optional: implement for custom role checking logic
type RoleChecker interface {
    // CheckRoles returns nil if user has any of the required roles
    CheckRoles(user User, roles ...string) error
}

// PermissionChecker checks if a user has required permissions
// Optional: implement for custom permission checking logic
type PermissionChecker interface {
    // CheckPermissions returns nil if user has all required permissions
    CheckPermissions(user User, permissions ...string) error
}

// --- Default Implementations ---

// DefaultUser is a simple User implementation
type DefaultUser struct {
    ID          string
    Role        string
    Permissions []string
}

func (u *DefaultUser) GetID() string            { return u.ID }
func (u *DefaultUser) GetRole() string          { return u.Role }
func (u *DefaultUser) GetPermissions() []string { return u.Permissions }

func (u *DefaultUser) HasRole(role string) bool {
    return u.Role == role
}

func (u *DefaultUser) HasPermission(permission string) bool {
    for _, p := range u.Permissions {
        if p == permission {
            return true
        }
    }
    return false
}

// --- Context Helpers ---

type contextKey string

const userContextKey contextKey = "server-user"

// UserFromContext extracts the authenticated user from context
func UserFromContext(ctx context.Context) User

// ContextWithUser adds an authenticated user to the context
func ContextWithUser(ctx context.Context, user User) context.Context

// UserFromRequest extracts the authenticated user from HTTP request context
func UserFromRequest(r *http.Request) User
```

**Example: Integrating with pkg/auth**

```go
package main

import (
    "context"

    "github.com/rompi/core-backend/pkg/auth"
    "github.com/rompi/core-backend/pkg/server"
)

// AuthAdapter adapts pkg/auth to server.Authenticator interface
type AuthAdapter struct {
    authService auth.Service
}

func NewAuthAdapter(authService auth.Service) *AuthAdapter {
    return &AuthAdapter{authService: authService}
}

func (a *AuthAdapter) ValidateToken(ctx context.Context, token string) (server.User, error) {
    if token == "" {
        return nil, nil // No token provided (for optional auth)
    }

    // Use pkg/auth to validate the token
    authUser, err := a.authService.ValidateSession(ctx, token)
    if err != nil {
        return nil, err
    }

    // Wrap auth.User to implement server.User interface
    return &UserWrapper{user: authUser}, nil
}

// UserWrapper wraps auth.User to implement server.User interface
type UserWrapper struct {
    user *auth.User
}

func (w *UserWrapper) GetID() string            { return w.user.ID }
func (w *UserWrapper) GetRole() string          { return w.user.Role }
func (w *UserWrapper) GetPermissions() []string { return w.user.Permissions }
func (w *UserWrapper) HasRole(role string) bool { return w.user.Role == role }
func (w *UserWrapper) HasPermission(p string) bool {
    for _, perm := range w.user.Permissions {
        if perm == p {
            return true
        }
    }
    return false
}

// Usage:
func main() {
    authService, _ := auth.NewService(config, repos)
    authAdapter := NewAuthAdapter(authService)

    srv, _ := server.NewServer(
        server.WithAuthenticator(authAdapter),
        // ...
    )
}
```

---

### 6. gRPC Interceptors (`grpc/`)

#### Auth Interceptor (`grpc/auth.go`)
```go
// Note: Uses server.Authenticator interface - no internal package dependencies

// AuthInterceptor creates an authentication interceptor
// Requires valid token, returns Unauthenticated error if invalid
func AuthInterceptor(auth server.Authenticator) grpc.UnaryServerInterceptor

// AuthStreamInterceptor creates a streaming auth interceptor
func AuthStreamInterceptor(auth server.Authenticator) grpc.StreamServerInterceptor

// OptionalAuthInterceptor sets user if token present but doesn't require it
func OptionalAuthInterceptor(auth server.Authenticator) grpc.UnaryServerInterceptor

// RequireRoleInterceptor creates an interceptor that requires specific roles
// Must be used after AuthInterceptor
func RequireRoleInterceptor(roles ...string) grpc.UnaryServerInterceptor

// RequirePermissionInterceptor creates an interceptor that requires specific permissions
// Must be used after AuthInterceptor
func RequirePermissionInterceptor(permissions ...string) grpc.UnaryServerInterceptor

// GetUser extracts authenticated user from context (returns server.User interface)
func GetUser(ctx context.Context) server.User
```

#### Logging Interceptor (`grpc/logging.go`)
```go
// LoggingInterceptor creates a logging interceptor
func LoggingInterceptor(logger Logger) grpc.UnaryServerInterceptor

// LoggingStreamInterceptor creates a streaming logging interceptor
func LoggingStreamInterceptor(logger Logger) grpc.StreamServerInterceptor

type LoggingConfig struct {
    Logger       Logger
    SkipMethods  []string  // Methods to skip logging (e.g., health checks)
    LogPayloads  bool      // Log request/response payloads (careful with sensitive data)
}

func LoggingInterceptorWithConfig(config LoggingConfig) grpc.UnaryServerInterceptor
```

#### Recovery Interceptor (`grpc/recovery.go`)
```go
// RecoveryInterceptor creates a panic recovery interceptor
func RecoveryInterceptor(logger Logger) grpc.UnaryServerInterceptor

// RecoveryStreamInterceptor creates a streaming panic recovery interceptor
func RecoveryStreamInterceptor(logger Logger) grpc.StreamServerInterceptor

type RecoveryConfig struct {
    Logger         Logger
    RecoveryFunc   func(p interface{}) error  // Custom recovery handler
    EnableStack    bool                       // Include stack trace in logs
}

func RecoveryInterceptorWithConfig(config RecoveryConfig) grpc.UnaryServerInterceptor
```

#### Rate Limit Interceptor (`grpc/ratelimit.go`)
```go
// RateLimitConfig defines per-method rate limiting configuration
type RateLimitConfig struct {
    // Rate is requests per second
    Rate float64

    // Burst is maximum burst size
    Burst int

    // KeyFunc extracts the rate limit key from context (default: peer IP)
    KeyFunc func(ctx context.Context, method string) string

    // SkipFunc determines whether to skip rate limiting
    SkipFunc func(ctx context.Context, method string) bool

    // Store for distributed rate limiting (default: in-memory)
    Store RateLimitStore
}

// RateLimitInterceptor creates a rate limiting interceptor
func RateLimitInterceptor(config RateLimitConfig) grpc.UnaryServerInterceptor

// --- Preset Rate Limiters ---

// RateLimitStrict - 5 req/s, burst 10 (login, password reset)
func RateLimitStrict() grpc.UnaryServerInterceptor

// RateLimitAuth - 10 req/s, burst 20 (token refresh)
func RateLimitAuth() grpc.UnaryServerInterceptor

// RateLimitNormal - 30 req/s, burst 60 (general API)
func RateLimitNormal() grpc.UnaryServerInterceptor

// RateLimitRelaxed - 100 req/s, burst 200 (read-heavy)
func RateLimitRelaxed() grpc.UnaryServerInterceptor

// --- Per-Method Rate Limiting ---

// PerMethodRateLimits allows different limits per gRPC method
type PerMethodRateLimits map[string]RateLimitConfig

// RateLimitPerMethod creates an interceptor with per-method limits
func RateLimitPerMethod(limits PerMethodRateLimits) grpc.UnaryServerInterceptor
```

#### Request ID Interceptor (`grpc/requestid.go`)
```go
// RequestIDInterceptor adds a request ID to context
func RequestIDInterceptor() grpc.UnaryServerInterceptor

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string

// RequestIDKey is the metadata key for request ID
const RequestIDKey = "x-request-id"
```

---

### 6. HTTP Middleware (`gateway/`)

#### Middleware Type (`gateway/middleware.go`)
```go
// Middleware is the standard HTTP middleware type
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middlewares
func Chain(middlewares ...Middleware) Middleware
```

#### Auth Middleware (`gateway/auth.go`)
```go
// Note: Uses server.Authenticator interface - no internal package dependencies

// AuthMiddleware creates HTTP auth middleware
// Requires valid token, returns 401 Unauthorized if invalid
func AuthMiddleware(auth server.Authenticator) Middleware

// OptionalAuthMiddleware sets user if authenticated, but doesn't require it
func OptionalAuthMiddleware(auth server.Authenticator) Middleware

// RequireRoleMiddleware requires specific roles
// Must be used after AuthMiddleware
func RequireRoleMiddleware(roles ...string) Middleware

// RequirePermissionMiddleware requires specific permissions
// Must be used after AuthMiddleware
func RequirePermissionMiddleware(permissions ...string) Middleware

// GetUser extracts authenticated user from request context (returns server.User interface)
func GetUser(r *http.Request) server.User
```

#### Rate Limit Middleware (`gateway/ratelimit.go`)
```go
// RateLimitConfig defines HTTP rate limiting configuration
type RateLimitConfig struct {
    Rate            float64
    Burst           int
    KeyFunc         func(r *http.Request) string  // Default: client IP
    ExceededHandler http.HandlerFunc              // Custom 429 handler
    SkipFunc        func(r *http.Request) bool
    Store           RateLimitStore
    Headers         bool  // Add X-RateLimit-* headers
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(config RateLimitConfig) Middleware

// RateLimitWithRate creates a simple rate limiter
func RateLimitWithRate(rps float64, burst int) Middleware

// --- Presets ---
func RateLimitStrict() Middleware   // 5 req/s
func RateLimitAuth() Middleware     // 10 req/s
func RateLimitNormal() Middleware   // 30 req/s
func RateLimitRelaxed() Middleware  // 100 req/s

// --- Per-Path Rate Limiting ---
type PerPathRateLimits map[string]RateLimitConfig

// RateLimitPerPath creates middleware with per-path limits
func RateLimitPerPath(limits PerPathRateLimits) Middleware
```

#### CORS Middleware (`gateway/cors.go`)
```go
type CORSConfig struct {
    AllowOrigins     []string
    AllowMethods     []string
    AllowHeaders     []string
    ExposeHeaders    []string
    AllowCredentials bool
    MaxAge           int
}

func CORSMiddleware(config CORSConfig) Middleware
func CORSWithOrigins(origins ...string) Middleware
```

#### Other Middleware
```go
// gateway/logging.go
func LoggingMiddleware(logger Logger) Middleware
func LoggingMiddlewareWithConfig(config LoggingConfig) Middleware

// gateway/recovery.go
func RecoveryMiddleware(logger Logger) Middleware

// gateway/requestid.go
func RequestIDMiddleware() Middleware
func GetRequestID(r *http.Request) string

// gateway/compress.go
func CompressionMiddleware() Middleware
```

---

### 7. Health Checks (`health/`)

```go
// CheckerFunc is a function that performs a health check
type CheckerFunc func(ctx context.Context) error

// Checker manages health checks
type Checker struct {
    checks map[string]CheckerFunc
    mu     sync.RWMutex
}

func NewChecker() *Checker
func (c *Checker) Add(name string, check CheckerFunc)
func (c *Checker) Remove(name string)
func (c *Checker) Check(ctx context.Context) *Status

// Status is the health check response
type Status struct {
    Status    string                 `json:"status"`  // "healthy", "unhealthy", "degraded"
    Timestamp time.Time              `json:"timestamp"`
    Duration  string                 `json:"duration"`
    Checks    map[string]CheckResult `json:"checks,omitempty"`
}

type CheckResult struct {
    Status   string `json:"status"`
    Duration string `json:"duration"`
    Error    string `json:"error,omitempty"`
}

// Built-in checkers
func PingChecker() CheckerFunc
func DatabaseChecker(db interface{ PingContext(context.Context) error }) CheckerFunc
func GRPCChecker(conn *grpc.ClientConn) CheckerFunc
func HTTPChecker(url string, timeout time.Duration) CheckerFunc
func RedisChecker(client interface{ Ping(context.Context) error }) CheckerFunc

// gRPC Health Service (implements grpc.health.v1.Health)
func RegisterHealthServer(server *grpc.Server, checker *Checker)

// HTTP Health Handlers
func HealthHandler(checker *Checker) http.HandlerFunc
func LivenessHandler() http.HandlerFunc
func ReadinessHandler(checker *Checker) http.HandlerFunc
```

---

## Proto File Example

### User Service (`examples/proto/user/v1/user.proto`)

```protobuf
syntax = "proto3";

package user.v1;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/timestamp.proto";
import "protoc-gen-openapiv2/options/annotations.proto";

option go_package = "github.com/rompi/core-backend/pkg/server/examples/gen/user/v1;userv1";

// OpenAPI metadata
option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_swagger) = {
  info: {
    title: "User Service API";
    version: "1.0";
    description: "API for managing users";
  };
  schemes: HTTPS;
  consumes: "application/json";
  produces: "application/json";
  security_definitions: {
    security: {
      key: "Bearer";
      value: {
        type: TYPE_API_KEY;
        in: IN_HEADER;
        name: "Authorization";
        description: "JWT token with Bearer prefix";
      };
    };
  };
};

// User represents a user in the system
message User {
  string id = 1;
  string email = 2;
  string name = 3;
  string role = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}

// --- Request/Response Messages ---

message CreateUserRequest {
  string email = 1;
  string password = 2;
  string name = 3;
  string role = 4;
}

message CreateUserResponse {
  User user = 1;
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  User user = 1;
}

message UpdateUserRequest {
  string id = 1;
  optional string email = 2;
  optional string name = 3;
  optional string role = 4;
}

message UpdateUserResponse {
  User user = 1;
}

message DeleteUserRequest {
  string id = 1;
}

message ListUsersRequest {
  int32 page = 1;
  int32 page_size = 2;
  string filter = 3;  // Optional filter query
}

message ListUsersResponse {
  repeated User users = 1;
  int32 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}

// --- Service Definition ---

service UserService {
  // CreateUser creates a new user
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/api/v1/users"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Create a new user";
      tags: "Users";
    };
  }

  // GetUser retrieves a user by ID
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{id}"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Get user by ID";
      tags: "Users";
      security: {
        security_requirement: {
          key: "Bearer";
          value: {};
        };
      };
    };
  }

  // UpdateUser updates an existing user
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option (google.api.http) = {
      patch: "/api/v1/users/{id}"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Update user";
      tags: "Users";
      security: {
        security_requirement: {
          key: "Bearer";
          value: {};
        };
      };
    };
  }

  // DeleteUser deletes a user
  rpc DeleteUser(DeleteUserRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/api/v1/users/{id}"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Delete user";
      tags: "Users";
      security: {
        security_requirement: {
          key: "Bearer";
          value: {};
        };
      };
    };
  }

  // ListUsers lists users with pagination
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {
    option (google.api.http) = {
      get: "/api/v1/users"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "List users";
      tags: "Users";
      security: {
        security_requirement: {
          key: "Bearer";
          value: {};
        };
      };
    };
  }
}
```

---

## Usage Examples

### Basic Server Setup

```go
package main

import (
    "context"
    "log"

    "github.com/rompi/core-backend/pkg/server"
    "github.com/rompi/core-backend/pkg/server/examples/gen/user/v1"
    "google.golang.org/protobuf/types/known/emptypb"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
    // Create server with defaults
    srv, err := server.NewServer(
        server.WithGRPCAddr(":9090"),
        server.WithHTTPAddr(":8080"),
        server.WithCORS("*"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create service implementation
    userService := &UserServiceImpl{
        // inject dependencies
    }

    // Register gRPC service
    userv1.RegisterUserServiceServer(srv.GRPCServer(), userService)

    // Register gRPC-Gateway (HTTP routes auto-generated from proto)
    srv.RegisterGateway(userv1.RegisterUserServiceHandlerFromEndpoint)

    // Start server (blocks until shutdown)
    if err := srv.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}

// UserServiceImpl implements the gRPC UserService
type UserServiceImpl struct {
    userv1.UnimplementedUserServiceServer
    // Add your dependencies here (repositories, services, etc.)
}

// CreateUser implements userv1.UserServiceServer
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
    // Validate request
    if req.Email == "" {
        return nil, server.NewError(codes.InvalidArgument, "email is required")
    }
    if req.Password == "" {
        return nil, server.NewError(codes.InvalidArgument, "password is required")
    }

    // Create user (your business logic)
    user := &userv1.User{
        Id:        "user-123",
        Email:     req.Email,
        Name:      req.Name,
        Role:      req.Role,
        CreatedAt: timestamppb.Now(),
        UpdatedAt: timestamppb.Now(),
    }

    return &userv1.CreateUserResponse{User: user}, nil
}

// GetUser implements userv1.UserServiceServer
func (s *UserServiceImpl) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    if req.Id == "" {
        return nil, server.NewError(codes.InvalidArgument, "id is required")
    }

    // Fetch user (your business logic)
    user := &userv1.User{
        Id:        req.Id,
        Email:     "user@example.com",
        Name:      "John Doe",
        Role:      "user",
        CreatedAt: timestamppb.Now(),
        UpdatedAt: timestamppb.Now(),
    }

    return &userv1.GetUserResponse{User: user}, nil
}

// UpdateUser implements userv1.UserServiceServer
func (s *UserServiceImpl) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
    if req.Id == "" {
        return nil, server.NewError(codes.InvalidArgument, "id is required")
    }

    // Update user (your business logic)
    user := &userv1.User{
        Id:        req.Id,
        Email:     "updated@example.com",
        Name:      "Updated Name",
        Role:      "user",
        UpdatedAt: timestamppb.Now(),
    }

    return &userv1.UpdateUserResponse{User: user}, nil
}

// DeleteUser implements userv1.UserServiceServer
func (s *UserServiceImpl) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*emptypb.Empty, error) {
    if req.Id == "" {
        return nil, server.NewError(codes.InvalidArgument, "id is required")
    }

    // Delete user (your business logic)

    return &emptypb.Empty{}, nil
}

// ListUsers implements userv1.UserServiceServer
func (s *UserServiceImpl) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
    // List users (your business logic)
    users := []*userv1.User{
        {Id: "1", Email: "user1@example.com", Name: "User 1"},
        {Id: "2", Email: "user2@example.com", Name: "User 2"},
    }

    return &userv1.ListUsersResponse{
        Users:    users,
        Total:    100,
        Page:     req.Page,
        PageSize: req.PageSize,
    }, nil
}
```

### With Auth Integration

```go
package main

import (
    "context"
    "log"

    "github.com/rompi/core-backend/pkg/auth"      // Your auth package
    "github.com/rompi/core-backend/pkg/server"
    "github.com/rompi/core-backend/pkg/server/grpc"
    "github.com/rompi/core-backend/pkg/server/gateway"
    "github.com/rompi/core-backend/pkg/server/examples/gen/user/v1"
)

// --- Step 1: Create an adapter to implement server.Authenticator interface ---

type AuthAdapter struct {
    authService auth.Service
}

func (a *AuthAdapter) ValidateToken(ctx context.Context, token string) (server.User, error) {
    if token == "" {
        return nil, nil
    }
    user, err := a.authService.ValidateSession(ctx, token)
    if err != nil {
        return nil, err
    }
    return &UserWrapper{user: user}, nil
}

// UserWrapper adapts auth.User to server.User interface
type UserWrapper struct {
    user *auth.User
}

func (w *UserWrapper) GetID() string              { return w.user.ID }
func (w *UserWrapper) GetRole() string            { return w.user.Role }
func (w *UserWrapper) GetPermissions() []string   { return w.user.Permissions }
func (w *UserWrapper) HasRole(role string) bool   { return w.user.Role == role }
func (w *UserWrapper) HasPermission(p string) bool {
    for _, perm := range w.user.Permissions {
        if perm == p {
            return true
        }
    }
    return false
}

// --- Step 2: Use the adapter with the server ---

func main() {
    // Setup your auth service
    authService, err := auth.NewService(authConfig, repos)
    if err != nil {
        log.Fatal(err)
    }

    // Create adapter that implements server.Authenticator
    authAdapter := &AuthAdapter{authService: authService}

    // Create server with auth
    srv, err := server.NewServer(
        server.WithGRPCAddr(":9090"),
        server.WithHTTPAddr(":8080"),
        server.WithLogger(logger),

        // gRPC interceptors (using interface-based auth)
        server.WithUnaryInterceptor(
            grpc.RecoveryInterceptor(logger),
            grpc.RequestIDInterceptor(),
            grpc.LoggingInterceptor(logger),
            grpc.AuthInterceptor(authAdapter),         // Uses Authenticator interface
            grpc.RequireRoleInterceptor("user"),       // Optional: require specific role
        ),

        // HTTP middleware (using interface-based auth)
        server.WithHTTPMiddleware(
            gateway.RecoveryMiddleware(logger),
            gateway.RequestIDMiddleware(),
            gateway.LoggingMiddleware(logger),
            gateway.CORSWithOrigins("*"),
            gateway.AuthMiddleware(authAdapter),       // Uses Authenticator interface
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Register services
    userService := &UserServiceImpl{}
    userv1.RegisterUserServiceServer(srv.GRPCServer(), userService)
    srv.RegisterGateway(userv1.RegisterUserServiceHandlerFromEndpoint)

    srv.ListenAndServe()
}

// --- Step 3: Access authenticated user in your handlers ---

func (s *UserServiceImpl) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    // Get authenticated user from context (returns server.User interface)
    user := grpc.GetUser(ctx)
    if user == nil {
        return nil, server.ErrUnauthenticated
    }

    // Use interface methods
    if !user.HasRole("admin") && user.GetID() != req.Id {
        return nil, server.ErrPermissionDenied
    }

    // ... rest of implementation
}
```

### Per-Endpoint Rate Limiting

```go
package main

import (
    "log"

    "github.com/rompi/core-backend/pkg/server"
    "github.com/rompi/core-backend/pkg/server/grpc"
    "github.com/rompi/core-backend/pkg/server/gateway"
)

func main() {
    // Define per-method rate limits for gRPC
    grpcLimits := grpc.PerMethodRateLimits{
        // Strict limits for auth methods
        "/user.v1.UserService/CreateUser": {Rate: 5, Burst: 10},

        // Normal limits for regular methods
        "/user.v1.UserService/GetUser":    {Rate: 30, Burst: 60},
        "/user.v1.UserService/UpdateUser": {Rate: 30, Burst: 60},
        "/user.v1.UserService/DeleteUser": {Rate: 10, Burst: 20},

        // Relaxed limits for list operations
        "/user.v1.UserService/ListUsers":  {Rate: 100, Burst: 200},
    }

    // Define per-path rate limits for HTTP
    httpLimits := gateway.PerPathRateLimits{
        // Strict limits for auth endpoints
        "POST /api/v1/users":              {Rate: 5, Burst: 10},

        // Normal limits for regular endpoints
        "GET /api/v1/users/{id}":          {Rate: 30, Burst: 60},
        "PATCH /api/v1/users/{id}":        {Rate: 30, Burst: 60},
        "DELETE /api/v1/users/{id}":       {Rate: 10, Burst: 20},

        // Relaxed limits for list operations
        "GET /api/v1/users":               {Rate: 100, Burst: 200},

        // Custom endpoints
        "POST /webhooks/stripe":           {Rate: 50, Burst: 100},
    }

    srv, err := server.NewServer(
        server.WithGRPCAddr(":9090"),
        server.WithHTTPAddr(":8080"),

        // Per-method gRPC rate limiting
        server.WithUnaryInterceptor(
            grpc.RecoveryInterceptor(logger),
            grpc.RateLimitPerMethod(grpcLimits),
        ),

        // Per-path HTTP rate limiting
        server.WithHTTPMiddleware(
            gateway.RecoveryMiddleware(logger),
            gateway.RateLimitPerPath(httpLimits),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    // ... register services

    srv.ListenAndServe()
}
```

### Custom HTTP Handlers (Non-Proto Endpoints)

```go
package main

import (
    "encoding/json"
    "io"
    "net/http"
    "log"

    "github.com/rompi/core-backend/pkg/server"
)

func main() {
    srv, err := server.NewServer(
        server.WithHTTPAddr(":8080"),
        server.WithGRPCAddr(":9090"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Register gRPC services and gateway...
    // userv1.RegisterUserServiceServer(srv.GRPCServer(), userService)
    // srv.RegisterGateway(userv1.RegisterUserServiceHandlerFromEndpoint)

    // --- Custom HTTP handlers for non-proto endpoints ---

    // Webhook endpoint (doesn't fit proto model)
    srv.HandleFunc("POST /webhooks/stripe", func(w http.ResponseWriter, r *http.Request) {
        // Verify Stripe signature
        payload, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "failed to read body", http.StatusBadRequest)
            return
        }

        // Process webhook
        log.Printf("Received Stripe webhook: %s", string(payload))

        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"received": true}`))
    })

    // File upload endpoint
    srv.HandleFunc("POST /api/v1/upload", func(w http.ResponseWriter, r *http.Request) {
        // Parse multipart form
        err := r.ParseMultipartForm(32 << 20) // 32MB max
        if err != nil {
            http.Error(w, "failed to parse form", http.StatusBadRequest)
            return
        }

        file, header, err := r.FormFile("file")
        if err != nil {
            http.Error(w, "failed to get file", http.StatusBadRequest)
            return
        }
        defer file.Close()

        // Process file upload
        log.Printf("Uploaded file: %s (%d bytes)", header.Filename, header.Size)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "filename": header.Filename,
            "size":     header.Size,
            "url":      "/files/" + header.Filename,
        })
    })

    // Health check (custom, if you need different behavior than built-in)
    srv.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("pong"))
    })

    srv.ListenAndServe()
}
```

### gRPC Client Example

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"

    "github.com/rompi/core-backend/pkg/server/examples/gen/user/v1"
)

func main() {
    // Connect to gRPC server
    conn, err := grpc.Dial(
        "localhost:9090",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Create client
    client := userv1.NewUserServiceClient(conn)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Add auth token to metadata
    ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer <token>")

    // Create user
    createResp, err := client.CreateUser(ctx, &userv1.CreateUserRequest{
        Email:    "newuser@example.com",
        Password: "securepassword123",
        Name:     "New User",
        Role:     "user",
    })
    if err != nil {
        log.Fatalf("CreateUser failed: %v", err)
    }
    log.Printf("Created user: %+v", createResp.User)

    // Get user
    getResp, err := client.GetUser(ctx, &userv1.GetUserRequest{
        Id: createResp.User.Id,
    })
    if err != nil {
        log.Fatalf("GetUser failed: %v", err)
    }
    log.Printf("Got user: %+v", getResp.User)

    // List users
    listResp, err := client.ListUsers(ctx, &userv1.ListUsersRequest{
        Page:     1,
        PageSize: 10,
    })
    if err != nil {
        log.Fatalf("ListUsers failed: %v", err)
    }
    log.Printf("Listed %d users (total: %d)", len(listResp.Users), listResp.Total)
}
```

### HTTP Client Example (via Gateway)

```bash
# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe",
    "role": "user"
  }'

# Get user (with auth)
curl http://localhost:8080/api/v1/users/user-123 \
  -H "Authorization: Bearer <token>"

# Update user
curl -X PATCH http://localhost:8080/api/v1/users/user-123 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Name"}'

# Delete user
curl -X DELETE http://localhost:8080/api/v1/users/user-123 \
  -H "Authorization: Bearer <token>"

# List users with pagination
curl "http://localhost:8080/api/v1/users?page=1&page_size=10" \
  -H "Authorization: Bearer <token>"
```

---

## Implementation Order

### Phase 1: Core Foundation
1. `errors.go` - Unified error handling
2. `logger.go` - Logger interface
3. `config.go` - Configuration
4. `options.go` - Functional options
5. `server.go` - Main unified server

### Phase 2: gRPC
6. `grpc/server.go` - gRPC server wrapper
7. `grpc/interceptor.go` - Interceptor chain builder
8. `grpc/recovery.go` - Panic recovery
9. `grpc/requestid.go` - Request ID
10. `grpc/logging.go` - Logging interceptor
11. `grpc/ratelimit.go` - Rate limiting interceptor

### Phase 3: Gateway/HTTP
12. `gateway/gateway.go` - gRPC-Gateway setup
13. `gateway/mux.go` - HTTP mux with custom handlers
14. `gateway/middleware.go` - Middleware chain
15. `gateway/recovery.go` - Panic recovery
16. `gateway/requestid.go` - Request ID
17. `gateway/logging.go` - Request logging
18. `gateway/cors.go` - CORS
19. `gateway/ratelimit.go` - Rate limiting middleware
20. `gateway/compress.go` - Response compression

### Phase 4: Features
21. `health/` - Health checks (gRPC + HTTP)
22. `grpc/auth.go` - Auth interceptor
23. `gateway/auth.go` - Auth middleware

### Phase 5: Testing and Documentation
24. `testutil/` - Testing utilities
25. `examples/` - Example implementations
26. `README.md` - Package documentation
27. All `*_test.go` files

---

## Success Criteria

1. Single source of truth (proto files) for API definitions
2. Both gRPC and HTTP clients can access the same services
3. 90%+ test coverage
4. **Zero internal package dependencies** - uses interfaces for auth, logging
5. Easy integration with any auth system via Authenticator interface
6. Per-endpoint rate limiting for both gRPC and HTTP
7. Production-ready defaults (recovery, logging, health checks)
8. Easy to extend with custom interceptors/middleware
9. Generated OpenAPI documentation

---

## References

- [gRPC-Gateway Documentation](https://grpc-ecosystem.github.io/grpc-gateway/)
- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
- [Protocol Buffers](https://protobuf.dev/)
- [go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)
- coding-guidelines.md for style conventions
