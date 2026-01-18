# Package Plan: pkg/b2boauth

## Overview

A B2B (Business-to-Business) OAuth 2.0 authorization package designed for machine-to-machine authentication. This package provides a complete OAuth 2.0 implementation with Client Credentials flow, supporting easy integration via client ID and client secret pairs. The design is storage-agnostic, allowing organizations to plug in any database or storage backend through repository interfaces.

## Goals

1. **OAuth 2.0 Compliance** - Full implementation of RFC 6749 Client Credentials Grant
2. **Easy Integration** - Simple client ID/secret registration and token issuance
3. **Storage Agnostic** - Repository interfaces for any storage backend (PostgreSQL, MongoDB, Redis, etc.)
4. **Separate from Auth** - Independent package with no dependencies on `pkg/auth`
5. **Scoped Access** - Fine-grained permission control via OAuth scopes
6. **Token Management** - JWT-based access tokens with configurable expiration
7. **Security First** - Secure secret hashing, rate limiting, audit logging
8. **Zero External Dependencies** - Core uses stdlib only

## Architecture

```
pkg/b2boauth/
├── b2boauth.go           # Core Service interface
├── service.go            # Service implementation
├── repository.go         # Repository interfaces (storage agnostic)
├── models.go             # Domain models (Client, Token, Scope, etc.)
├── config.go             # Configuration with env var support
├── errors.go             # Error definitions with codes
├── token.go              # JWT token manager
├── hash.go               # Client secret hashing utilities
├── validator.go          # Request validation
├── middleware/
│   ├── http.go           # HTTP middleware for token validation
│   └── grpc.go           # gRPC interceptor for token validation
├── testutil/
│   └── mocks.go          # Mock repositories for testing
├── examples/
│   ├── basic/            # Basic client credentials flow
│   ├── http-server/      # HTTP server with middleware
│   └── scopes/           # Scope-based authorization
└── README.md
```

## Core Interfaces

```go
package b2boauth

import (
    "context"
    "time"
)

// Service provides B2B OAuth 2.0 functionality
type Service interface {
    // Client Management
    CreateClient(ctx context.Context, req CreateClientRequest) (*Client, error)
    GetClient(ctx context.Context, clientID string) (*Client, error)
    UpdateClient(ctx context.Context, clientID string, req UpdateClientRequest) (*Client, error)
    DeleteClient(ctx context.Context, clientID string) error
    RotateClientSecret(ctx context.Context, clientID string) (*ClientCredentials, error)
    ListClients(ctx context.Context, opts ListOptions) (*ClientList, error)

    // Token Operations (OAuth 2.0 Client Credentials Grant)
    TokenExchange(ctx context.Context, req TokenRequest) (*TokenResponse, error)
    ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
    RevokeToken(ctx context.Context, token string) error
    IntrospectToken(ctx context.Context, token string) (*IntrospectionResponse, error)

    // Scope Management
    CreateScope(ctx context.Context, scope *Scope) error
    GetScope(ctx context.Context, name string) (*Scope, error)
    ListScopes(ctx context.Context) ([]*Scope, error)
    DeleteScope(ctx context.Context, name string) error

    // Middleware
    HTTPMiddleware(requiredScopes ...string) func(http.Handler) http.Handler
    GRPCUnaryInterceptor(requiredScopes ...string) grpc.UnaryServerInterceptor
    GRPCStreamInterceptor(requiredScopes ...string) grpc.StreamServerInterceptor
}

// TokenRequest represents an OAuth 2.0 token request
type TokenRequest struct {
    GrantType    string   `json:"grant_type"`    // Must be "client_credentials"
    ClientID     string   `json:"client_id"`
    ClientSecret string   `json:"client_secret"`
    Scope        string   `json:"scope"`         // Space-separated scopes
}

// TokenResponse represents an OAuth 2.0 token response
type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    TokenType    string `json:"token_type"`     // Always "Bearer"
    ExpiresIn    int64  `json:"expires_in"`     // Seconds until expiration
    Scope        string `json:"scope"`          // Granted scopes
}

// TokenClaims represents validated token claims
type TokenClaims struct {
    ClientID  string    `json:"client_id"`
    Scopes    []string  `json:"scopes"`
    ExpiresAt time.Time `json:"exp"`
    IssuedAt  time.Time `json:"iat"`
    Issuer    string    `json:"iss"`
    Audience  string    `json:"aud"`
    JTI       string    `json:"jti"`            // Unique token identifier
}

// IntrospectionResponse represents OAuth 2.0 Token Introspection (RFC 7662)
type IntrospectionResponse struct {
    Active    bool     `json:"active"`
    Scope     string   `json:"scope,omitempty"`
    ClientID  string   `json:"client_id,omitempty"`
    TokenType string   `json:"token_type,omitempty"`
    Exp       int64    `json:"exp,omitempty"`
    Iat       int64    `json:"iat,omitempty"`
    Iss       string   `json:"iss,omitempty"`
    Aud       string   `json:"aud,omitempty"`
}
```

## Repository Interfaces (Storage Agnostic)

```go
package b2boauth

import "context"

// ClientRepository defines persistence operations for OAuth clients.
// Implement this interface with your preferred storage backend.
type ClientRepository interface {
    // Create stores a new OAuth client
    Create(ctx context.Context, client *Client) error

    // GetByID retrieves a client by its unique ID
    GetByID(ctx context.Context, id string) (*Client, error)

    // GetByClientID retrieves a client by its OAuth client_id
    GetByClientID(ctx context.Context, clientID string) (*Client, error)

    // Update modifies an existing client
    Update(ctx context.Context, client *Client) error

    // Delete removes a client (soft or hard delete based on implementation)
    Delete(ctx context.Context, id string) error

    // List returns paginated clients
    List(ctx context.Context, opts ListOptions) ([]*Client, int64, error)

    // IncrementFailedAttempts tracks authentication failures
    IncrementFailedAttempts(ctx context.Context, clientID string) error

    // ResetFailedAttempts clears the failure counter
    ResetFailedAttempts(ctx context.Context, clientID string) error
}

// TokenRepository defines persistence for issued tokens.
// Required for token revocation and introspection.
type TokenRepository interface {
    // Create stores an issued token
    Create(ctx context.Context, token *IssuedToken) error

    // GetByJTI retrieves a token by its unique identifier
    GetByJTI(ctx context.Context, jti string) (*IssuedToken, error)

    // GetByClientID retrieves all tokens for a client
    GetByClientID(ctx context.Context, clientID string) ([]*IssuedToken, error)

    // Revoke marks a token as revoked
    Revoke(ctx context.Context, jti string) error

    // RevokeByClientID revokes all tokens for a client
    RevokeByClientID(ctx context.Context, clientID string) error

    // IsRevoked checks if a token has been revoked
    IsRevoked(ctx context.Context, jti string) (bool, error)

    // DeleteExpired removes expired tokens (cleanup)
    DeleteExpired(ctx context.Context) error
}

// ScopeRepository defines persistence for OAuth scopes.
// Optional - defaults to in-memory if not provided.
type ScopeRepository interface {
    // Create stores a new scope
    Create(ctx context.Context, scope *Scope) error

    // GetByName retrieves a scope by name
    GetByName(ctx context.Context, name string) (*Scope, error)

    // List returns all available scopes
    List(ctx context.Context) ([]*Scope, error)

    // Delete removes a scope
    Delete(ctx context.Context, name string) error

    // ValidateScopes checks if all requested scopes exist
    ValidateScopes(ctx context.Context, scopes []string) error
}

// AuditLogRepository defines persistence for security audit logs.
// Optional - if not provided, audit logging is disabled.
type AuditLogRepository interface {
    // Create stores an audit log entry
    Create(ctx context.Context, log *AuditLog) error

    // GetByClientID retrieves audit logs for a client
    GetByClientID(ctx context.Context, clientID string, opts ListOptions) ([]*AuditLog, error)

    // GetByAction retrieves logs by action type
    GetByAction(ctx context.Context, action string, opts ListOptions) ([]*AuditLog, error)
}

// Repositories aggregates all repository dependencies
type Repositories struct {
    Clients   ClientRepository      // Required
    Tokens    TokenRepository       // Required for revocation/introspection
    Scopes    ScopeRepository       // Optional - uses in-memory default
    AuditLogs AuditLogRepository    // Optional - disables audit if nil
}

// validate checks repository requirements
func (r *Repositories) validate() error {
    if r.Clients == nil {
        return ErrClientRepositoryRequired
    }
    if r.Tokens == nil {
        return ErrTokenRepositoryRequired
    }
    return nil
}
```

## Domain Models

```go
package b2boauth

import "time"

// Client represents an OAuth 2.0 B2B client application
type Client struct {
    ID              string     `json:"id"`                // Internal UUID
    ClientID        string     `json:"client_id"`         // Public OAuth client_id
    ClientSecretHash string    `json:"-"`                 // Hashed secret (never exposed)
    Name            string     `json:"name"`              // Human-readable name
    Description     string     `json:"description"`       // Optional description
    AllowedScopes   []string   `json:"allowed_scopes"`    // Scopes this client can request
    Metadata        Metadata   `json:"metadata"`          // Custom key-value data
    RateLimitRPS    int        `json:"rate_limit_rps"`    // Requests per second limit
    Status          Status     `json:"status"`            // active, suspended, revoked
    FailedAttempts  int        `json:"failed_attempts"`   // Consecutive auth failures
    LockedUntil     *time.Time `json:"locked_until"`      // Account lock expiry
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
    LastUsedAt      *time.Time `json:"last_used_at"`
    CreatedBy       string     `json:"created_by"`        // Admin/system that created
}

// ClientCredentials represents a newly created or rotated client secret
type ClientCredentials struct {
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`  // Only shown once!
}

// Status represents client status
type Status string

const (
    StatusActive    Status = "active"
    StatusSuspended Status = "suspended"
    StatusRevoked   Status = "revoked"
)

// Metadata holds custom key-value pairs
type Metadata map[string]interface{}

// Scope represents an OAuth scope definition
type Scope struct {
    Name        string `json:"name"`         // e.g., "read:users", "write:orders"
    Description string `json:"description"`  // Human-readable description
    System      bool   `json:"system"`       // True if system-defined
}

// IssuedToken tracks a token for revocation/introspection
type IssuedToken struct {
    JTI       string    `json:"jti"`        // Unique token ID
    ClientID  string    `json:"client_id"`
    Scopes    []string  `json:"scopes"`
    IssuedAt  time.Time `json:"issued_at"`
    ExpiresAt time.Time `json:"expires_at"`
    RevokedAt *time.Time `json:"revoked_at"`
    IPAddress string    `json:"ip_address"`
    UserAgent string    `json:"user_agent"`
}

// AuditLog represents a security audit event
type AuditLog struct {
    ID        string                 `json:"id"`
    ClientID  string                 `json:"client_id"`
    Action    string                 `json:"action"`      // e.g., "token_issued", "auth_failed"
    Details   map[string]interface{} `json:"details"`
    IPAddress string                 `json:"ip_address"`
    UserAgent string                 `json:"user_agent"`
    Timestamp time.Time              `json:"timestamp"`
}

// CreateClientRequest for creating a new client
type CreateClientRequest struct {
    Name          string   `json:"name" validate:"required,min=1,max=255"`
    Description   string   `json:"description" validate:"max=1000"`
    AllowedScopes []string `json:"allowed_scopes" validate:"required,min=1"`
    Metadata      Metadata `json:"metadata"`
    RateLimitRPS  int      `json:"rate_limit_rps" validate:"min=0,max=10000"`
}

// UpdateClientRequest for updating a client
type UpdateClientRequest struct {
    Name          *string   `json:"name" validate:"omitempty,min=1,max=255"`
    Description   *string   `json:"description" validate:"omitempty,max=1000"`
    AllowedScopes []string  `json:"allowed_scopes"`
    Metadata      Metadata  `json:"metadata"`
    RateLimitRPS  *int      `json:"rate_limit_rps" validate:"omitempty,min=0,max=10000"`
    Status        *Status   `json:"status"`
}

// ListOptions for pagination
type ListOptions struct {
    Offset int    `json:"offset"`
    Limit  int    `json:"limit"`
    SortBy string `json:"sort_by"`
    Order  string `json:"order"`  // "asc" or "desc"
}

// ClientList represents a paginated list of clients
type ClientList struct {
    Clients []*Client `json:"clients"`
    Total   int64     `json:"total"`
    Offset  int       `json:"offset"`
    Limit   int       `json:"limit"`
}
```

## Configuration

```go
package b2boauth

import (
    "time"
    "os"
    "strconv"
)

// Config holds B2B OAuth configuration
type Config struct {
    // Token Settings
    TokenExpiration      time.Duration `json:"token_expiration"`       // Default: 1 hour
    TokenIssuer          string        `json:"token_issuer"`           // JWT issuer claim
    TokenAudience        string        `json:"token_audience"`         // JWT audience claim
    TokenSigningKey      string        `json:"-"`                      // JWT signing key (never log)
    TokenSigningMethod   string        `json:"token_signing_method"`   // HS256, RS256, ES256

    // Client Settings
    ClientIDLength       int           `json:"client_id_length"`       // Default: 32
    ClientSecretLength   int           `json:"client_secret_length"`   // Default: 64
    ClientSecretHashCost int           `json:"client_secret_hash_cost"` // bcrypt cost

    // Security Settings
    MaxFailedAttempts    int           `json:"max_failed_attempts"`    // Default: 5
    LockoutDuration      time.Duration `json:"lockout_duration"`       // Default: 15 minutes
    RateLimitEnabled     bool          `json:"rate_limit_enabled"`     // Default: true
    DefaultRateLimitRPS  int           `json:"default_rate_limit_rps"` // Default: 100

    // Audit Settings
    AuditEnabled         bool          `json:"audit_enabled"`          // Default: true
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
    cfg := defaultConfig()
    if err := cfg.overrideFromEnv(); err != nil {
        return nil, err
    }
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    return cfg, nil
}

func defaultConfig() *Config {
    return &Config{
        TokenExpiration:      time.Hour,
        TokenIssuer:          "b2b-oauth",
        TokenAudience:        "api",
        TokenSigningMethod:   "HS256",
        ClientIDLength:       32,
        ClientSecretLength:   64,
        ClientSecretHashCost: 12,
        MaxFailedAttempts:    5,
        LockoutDuration:      15 * time.Minute,
        RateLimitEnabled:     true,
        DefaultRateLimitRPS:  100,
        AuditEnabled:         true,
    }
}

func (c *Config) overrideFromEnv() error {
    if v := os.Getenv("B2B_OAUTH_TOKEN_EXPIRATION"); v != "" {
        d, err := time.ParseDuration(v)
        if err != nil {
            return fmt.Errorf("invalid B2B_OAUTH_TOKEN_EXPIRATION: %w", err)
        }
        c.TokenExpiration = d
    }
    if v := os.Getenv("B2B_OAUTH_TOKEN_ISSUER"); v != "" {
        c.TokenIssuer = v
    }
    if v := os.Getenv("B2B_OAUTH_TOKEN_AUDIENCE"); v != "" {
        c.TokenAudience = v
    }
    if v := os.Getenv("B2B_OAUTH_TOKEN_SIGNING_KEY"); v != "" {
        c.TokenSigningKey = v
    }
    if v := os.Getenv("B2B_OAUTH_TOKEN_SIGNING_METHOD"); v != "" {
        c.TokenSigningMethod = v
    }
    if v := os.Getenv("B2B_OAUTH_MAX_FAILED_ATTEMPTS"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil {
            return fmt.Errorf("invalid B2B_OAUTH_MAX_FAILED_ATTEMPTS: %w", err)
        }
        c.MaxFailedAttempts = n
    }
    // ... additional env var overrides
    return nil
}

func (c *Config) Validate() error {
    if c.TokenSigningKey == "" {
        return errors.New("token signing key is required")
    }
    if c.TokenExpiration < time.Minute {
        return errors.New("token expiration must be at least 1 minute")
    }
    if c.ClientIDLength < 16 {
        return errors.New("client ID length must be at least 16")
    }
    if c.ClientSecretLength < 32 {
        return errors.New("client secret length must be at least 32")
    }
    return nil
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `B2B_OAUTH_TOKEN_EXPIRATION` | Token validity duration | `1h` |
| `B2B_OAUTH_TOKEN_ISSUER` | JWT issuer claim | `b2b-oauth` |
| `B2B_OAUTH_TOKEN_AUDIENCE` | JWT audience claim | `api` |
| `B2B_OAUTH_TOKEN_SIGNING_KEY` | JWT signing secret/key | **Required** |
| `B2B_OAUTH_TOKEN_SIGNING_METHOD` | JWT signing algorithm | `HS256` |
| `B2B_OAUTH_CLIENT_ID_LENGTH` | Generated client ID length | `32` |
| `B2B_OAUTH_CLIENT_SECRET_LENGTH` | Generated client secret length | `64` |
| `B2B_OAUTH_MAX_FAILED_ATTEMPTS` | Auth failures before lockout | `5` |
| `B2B_OAUTH_LOCKOUT_DURATION` | Account lockout duration | `15m` |
| `B2B_OAUTH_DEFAULT_RATE_LIMIT` | Default requests/second | `100` |
| `B2B_OAUTH_AUDIT_ENABLED` | Enable audit logging | `true` |

## Error Handling

```go
package b2boauth

import "errors"

// Error codes for API responses
const (
    CodeInvalidRequest       = "invalid_request"
    CodeInvalidClient        = "invalid_client"
    CodeInvalidGrant         = "invalid_grant"
    CodeUnauthorizedClient   = "unauthorized_client"
    CodeUnsupportedGrantType = "unsupported_grant_type"
    CodeInvalidScope         = "invalid_scope"
    CodeAccessDenied         = "access_denied"
    CodeServerError          = "server_error"
    CodeClientLocked         = "client_locked"
    CodeClientSuspended      = "client_suspended"
    CodeTokenRevoked         = "token_revoked"
    CodeTokenExpired         = "token_expired"
    CodeRateLimitExceeded    = "rate_limit_exceeded"
)

// Sentinel errors
var (
    ErrInvalidCredentials       = errors.New("invalid client credentials")
    ErrClientNotFound           = errors.New("client not found")
    ErrClientLocked             = errors.New("client account is locked")
    ErrClientSuspended          = errors.New("client account is suspended")
    ErrClientRevoked            = errors.New("client has been revoked")
    ErrInvalidGrantType         = errors.New("unsupported grant type")
    ErrInvalidScope             = errors.New("requested scope is invalid")
    ErrScopeNotAllowed          = errors.New("scope not allowed for this client")
    ErrTokenExpired             = errors.New("token has expired")
    ErrTokenRevoked             = errors.New("token has been revoked")
    ErrTokenInvalid             = errors.New("token is invalid")
    ErrRateLimitExceeded        = errors.New("rate limit exceeded")
    ErrClientRepositoryRequired = errors.New("client repository is required")
    ErrTokenRepositoryRequired  = errors.New("token repository is required")
)

// OAuthError represents an OAuth 2.0 compliant error response
type OAuthError struct {
    Code        string `json:"error"`
    Description string `json:"error_description,omitempty"`
    URI         string `json:"error_uri,omitempty"`
    StatusCode  int    `json:"-"`
}

func (e *OAuthError) Error() string {
    if e.Description != "" {
        return fmt.Sprintf("%s: %s", e.Code, e.Description)
    }
    return e.Code
}

// NewOAuthError creates a new OAuth error
func NewOAuthError(code string, description string, status int) *OAuthError {
    return &OAuthError{
        Code:        code,
        Description: description,
        StatusCode:  status,
    }
}
```

## Usage Examples

### Basic Service Setup

```go
package main

import (
    "context"
    "log"

    "github.com/user/core-backend/pkg/b2boauth"
)

func main() {
    // Load configuration
    cfg, err := b2boauth.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Initialize repositories (using your storage implementation)
    repos := b2boauth.Repositories{
        Clients:   NewPostgresClientRepository(db),   // Your implementation
        Tokens:    NewPostgresTokenRepository(db),    // Your implementation
        Scopes:    NewPostgresScopeRepository(db),    // Optional
        AuditLogs: NewPostgresAuditLogRepository(db), // Optional
    }

    // Create service
    svc, err := b2boauth.NewService(cfg, repos)
    if err != nil {
        log.Fatal(err)
    }

    // Service is ready to use
    _ = svc
}
```

### Creating a Client

```go
func createClient(ctx context.Context, svc b2boauth.Service) {
    // Create a new B2B client
    client, err := svc.CreateClient(ctx, b2boauth.CreateClientRequest{
        Name:          "Partner API",
        Description:   "Integration for Partner Corp",
        AllowedScopes: []string{"read:orders", "write:orders", "read:products"},
        Metadata: b2boauth.Metadata{
            "partner_id": "PARTNER-001",
            "environment": "production",
        },
        RateLimitRPS: 500, // Custom rate limit
    })
    if err != nil {
        log.Fatal(err)
    }

    // IMPORTANT: ClientSecret is only returned once!
    // Store it securely and provide to the partner
    fmt.Printf("Client ID: %s\n", client.ClientID)
    fmt.Printf("Client Secret: %s\n", client.ClientSecret) // Only shown once!
}
```

### OAuth 2.0 Token Exchange (Client Credentials Flow)

```go
func getAccessToken(ctx context.Context, svc b2boauth.Service) {
    // OAuth 2.0 Client Credentials Grant
    tokenResp, err := svc.TokenExchange(ctx, b2boauth.TokenRequest{
        GrantType:    "client_credentials",
        ClientID:     "abc123...",
        ClientSecret: "xyz789...",
        Scope:        "read:orders write:orders", // Space-separated
    })
    if err != nil {
        var oauthErr *b2boauth.OAuthError
        if errors.As(err, &oauthErr) {
            // Handle OAuth error
            fmt.Printf("OAuth Error: %s - %s\n", oauthErr.Code, oauthErr.Description)
        }
        log.Fatal(err)
    }

    // Use the access token
    fmt.Printf("Access Token: %s\n", tokenResp.AccessToken)
    fmt.Printf("Token Type: %s\n", tokenResp.TokenType)      // "Bearer"
    fmt.Printf("Expires In: %d seconds\n", tokenResp.ExpiresIn)
    fmt.Printf("Granted Scopes: %s\n", tokenResp.Scope)
}
```

### Validating Access Tokens

```go
func validateToken(ctx context.Context, svc b2boauth.Service, token string) {
    // Validate and extract claims
    claims, err := svc.ValidateToken(ctx, token)
    if err != nil {
        switch {
        case errors.Is(err, b2boauth.ErrTokenExpired):
            fmt.Println("Token has expired")
        case errors.Is(err, b2boauth.ErrTokenRevoked):
            fmt.Println("Token has been revoked")
        case errors.Is(err, b2boauth.ErrTokenInvalid):
            fmt.Println("Token is invalid")
        default:
            fmt.Printf("Validation error: %v\n", err)
        }
        return
    }

    // Token is valid
    fmt.Printf("Client ID: %s\n", claims.ClientID)
    fmt.Printf("Scopes: %v\n", claims.Scopes)
    fmt.Printf("Expires At: %s\n", claims.ExpiresAt)
}
```

### HTTP Server with Middleware

```go
package main

import (
    "net/http"

    "github.com/user/core-backend/pkg/b2boauth"
)

func main() {
    svc, _ := b2boauth.NewService(cfg, repos)

    mux := http.NewServeMux()

    // Public endpoints (no auth)
    mux.HandleFunc("/oauth/token", handleTokenExchange(svc))

    // Protected endpoints with scope requirements
    ordersHandler := http.HandlerFunc(handleOrders)
    mux.Handle("/api/orders",
        svc.HTTPMiddleware("read:orders")(ordersHandler))

    writeOrdersHandler := http.HandlerFunc(handleCreateOrder)
    mux.Handle("/api/orders/create",
        svc.HTTPMiddleware("write:orders")(writeOrdersHandler))

    // Multiple scopes required (AND logic)
    adminHandler := http.HandlerFunc(handleAdmin)
    mux.Handle("/api/admin",
        svc.HTTPMiddleware("admin:read", "admin:write")(adminHandler))

    http.ListenAndServe(":8080", mux)
}

// handleTokenExchange implements the /oauth/token endpoint
func handleTokenExchange(svc b2boauth.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // Parse form data (application/x-www-form-urlencoded)
        if err := r.ParseForm(); err != nil {
            writeOAuthError(w, b2boauth.NewOAuthError(
                b2boauth.CodeInvalidRequest,
                "Invalid request body",
                http.StatusBadRequest,
            ))
            return
        }

        req := b2boauth.TokenRequest{
            GrantType:    r.FormValue("grant_type"),
            ClientID:     r.FormValue("client_id"),
            ClientSecret: r.FormValue("client_secret"),
            Scope:        r.FormValue("scope"),
        }

        resp, err := svc.TokenExchange(r.Context(), req)
        if err != nil {
            var oauthErr *b2boauth.OAuthError
            if errors.As(err, &oauthErr) {
                writeOAuthError(w, oauthErr)
                return
            }
            writeOAuthError(w, b2boauth.NewOAuthError(
                b2boauth.CodeServerError,
                "Internal server error",
                http.StatusInternalServerError,
            ))
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Cache-Control", "no-store")
        w.Header().Set("Pragma", "no-cache")
        json.NewEncoder(w).Encode(resp)
    }
}

// handleOrders accesses the validated client from context
func handleOrders(w http.ResponseWriter, r *http.Request) {
    // Get validated claims from context
    claims := b2boauth.ClaimsFromContext(r.Context())

    fmt.Printf("Request from client: %s\n", claims.ClientID)
    fmt.Printf("With scopes: %v\n", claims.Scopes)

    // Handle the request...
    w.Write([]byte(`{"orders": []}`))
}
```

### Rotating Client Secrets

```go
func rotateSecret(ctx context.Context, svc b2boauth.Service, clientID string) {
    // Rotate to a new secret (old secret is invalidated)
    creds, err := svc.RotateClientSecret(ctx, clientID)
    if err != nil {
        log.Fatal(err)
    }

    // New credentials - provide to the client securely
    fmt.Printf("New Client Secret: %s\n", creds.ClientSecret)

    // Note: All existing tokens remain valid until expiration
    // Use RevokeByClientID to invalidate existing tokens if needed
}
```

### Revoking Tokens

```go
func revokeAccess(ctx context.Context, svc b2boauth.Service) {
    // Revoke a specific token
    err := svc.RevokeToken(ctx, "token-jti-here")
    if err != nil {
        log.Fatal(err)
    }

    // Or revoke all tokens for a client
    client, _ := svc.GetClient(ctx, "client-id")
    // Implementation would call: tokenRepo.RevokeByClientID(ctx, client.ClientID)
}
```

### Token Introspection (RFC 7662)

```go
func introspectToken(ctx context.Context, svc b2boauth.Service, token string) {
    resp, err := svc.IntrospectToken(ctx, token)
    if err != nil {
        log.Fatal(err)
    }

    if resp.Active {
        fmt.Println("Token is active")
        fmt.Printf("Client: %s\n", resp.ClientID)
        fmt.Printf("Scopes: %s\n", resp.Scope)
    } else {
        fmt.Println("Token is not active (expired, revoked, or invalid)")
    }
}
```

## Security Considerations

### Client Secret Hashing

```go
// Secrets are hashed using bcrypt with configurable cost
hash, err := bcrypt.GenerateFromPassword([]byte(secret), cfg.ClientSecretHashCost)

// Verification
err := bcrypt.CompareHashAndPassword(client.ClientSecretHash, []byte(providedSecret))
```

### Account Lockout

```go
// After MaxFailedAttempts, account is locked for LockoutDuration
if client.FailedAttempts >= cfg.MaxFailedAttempts {
    if client.LockedUntil != nil && time.Now().Before(*client.LockedUntil) {
        return nil, ErrClientLocked
    }
}
```

### Rate Limiting

```go
// Per-client rate limiting based on RateLimitRPS
// Uses token bucket algorithm
```

### Audit Logging

```go
// All security-relevant events are logged
type AuditAction string

const (
    ActionTokenIssued     AuditAction = "token_issued"
    ActionTokenRevoked    AuditAction = "token_revoked"
    ActionAuthFailed      AuditAction = "auth_failed"
    ActionClientCreated   AuditAction = "client_created"
    ActionClientUpdated   AuditAction = "client_updated"
    ActionClientDeleted   AuditAction = "client_deleted"
    ActionSecretRotated   AuditAction = "secret_rotated"
    ActionAccountLocked   AuditAction = "account_locked"
    ActionAccountUnlocked AuditAction = "account_unlocked"
)
```

## Storage Implementation Examples

### PostgreSQL Schema

```sql
-- OAuth Clients
CREATE TABLE b2b_oauth_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(64) UNIQUE NOT NULL,
    client_secret_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    allowed_scopes TEXT[] NOT NULL,
    metadata JSONB DEFAULT '{}',
    rate_limit_rps INTEGER DEFAULT 100,
    status VARCHAR(20) DEFAULT 'active',
    failed_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255)
);

CREATE INDEX idx_b2b_oauth_clients_client_id ON b2b_oauth_clients(client_id);
CREATE INDEX idx_b2b_oauth_clients_status ON b2b_oauth_clients(status);

-- Issued Tokens (for revocation tracking)
CREATE TABLE b2b_oauth_tokens (
    jti UUID PRIMARY KEY,
    client_id VARCHAR(64) NOT NULL REFERENCES b2b_oauth_clients(client_id),
    scopes TEXT[] NOT NULL,
    issued_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    ip_address VARCHAR(45),
    user_agent TEXT
);

CREATE INDEX idx_b2b_oauth_tokens_client_id ON b2b_oauth_tokens(client_id);
CREATE INDEX idx_b2b_oauth_tokens_expires_at ON b2b_oauth_tokens(expires_at);

-- OAuth Scopes
CREATE TABLE b2b_oauth_scopes (
    name VARCHAR(100) PRIMARY KEY,
    description TEXT,
    system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit Logs
CREATE TABLE b2b_oauth_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(64),
    action VARCHAR(50) NOT NULL,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_b2b_oauth_audit_logs_client_id ON b2b_oauth_audit_logs(client_id);
CREATE INDEX idx_b2b_oauth_audit_logs_action ON b2b_oauth_audit_logs(action);
CREATE INDEX idx_b2b_oauth_audit_logs_timestamp ON b2b_oauth_audit_logs(timestamp);
```

### Redis Token Storage (Alternative)

```go
// For high-performance token storage/revocation
type RedisTokenRepository struct {
    client *redis.Client
    prefix string
}

func (r *RedisTokenRepository) Create(ctx context.Context, token *IssuedToken) error {
    key := r.prefix + token.JTI
    ttl := time.Until(token.ExpiresAt)
    data, _ := json.Marshal(token)
    return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisTokenRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
    key := r.prefix + "revoked:" + jti
    exists, err := r.client.Exists(ctx, key).Result()
    return exists > 0, err
}

func (r *RedisTokenRepository) Revoke(ctx context.Context, jti string) error {
    key := r.prefix + "revoked:" + jti
    // Keep revocation record for token lifetime
    return r.client.Set(ctx, key, "1", 24*time.Hour).Err()
}
```

## Testing

### Mock Repositories

```go
// testutil/mocks.go
package testutil

import "context"

type MockClientRepository struct {
    CreateFunc              func(ctx context.Context, client *Client) error
    GetByIDFunc             func(ctx context.Context, id string) (*Client, error)
    GetByClientIDFunc       func(ctx context.Context, clientID string) (*Client, error)
    UpdateFunc              func(ctx context.Context, client *Client) error
    DeleteFunc              func(ctx context.Context, id string) error
    // ... other functions
}

func (m *MockClientRepository) Create(ctx context.Context, client *Client) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, client)
    }
    return nil
}

// ... implement other methods

type MockTokenRepository struct {
    CreateFunc        func(ctx context.Context, token *IssuedToken) error
    GetByJTIFunc      func(ctx context.Context, jti string) (*IssuedToken, error)
    RevokeFunc        func(ctx context.Context, jti string) error
    IsRevokedFunc     func(ctx context.Context, jti string) (bool, error)
    // ... other functions
}

// ... implement methods
```

### Test Coverage Requirements

- Unit tests for all public functions
- Integration tests with mock repositories
- OAuth 2.0 compliance tests
- Token validation edge cases
- Rate limiting tests
- Account lockout tests
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interfaces & Models
1. Define Service interface
2. Define Repository interfaces
3. Create domain models
4. Implement configuration

### Phase 2: Token Management
1. JWT token generation
2. Token validation
3. Token claims extraction
4. JTI tracking

### Phase 3: Client Management
1. Client CRUD operations
2. Secret hashing & verification
3. Secret rotation
4. Status management

### Phase 4: OAuth 2.0 Token Exchange
1. Client Credentials Grant flow
2. Scope validation
3. Token response formatting
4. OAuth error responses

### Phase 5: Security Features
1. Account lockout
2. Rate limiting
3. Audit logging

### Phase 6: Middleware
1. HTTP middleware
2. gRPC interceptors
3. Context helpers

### Phase 7: Token Revocation & Introspection
1. Token revocation
2. RFC 7662 Token Introspection
3. Bulk revocation

### Phase 8: Testing & Documentation
1. Mock repositories
2. Unit tests
3. Integration tests
4. README documentation
5. Examples

## Dependencies

- **Required:** None (core uses stdlib)
- **Optional:**
  - `github.com/golang-jwt/jwt/v5` for JWT (or stdlib crypto)
  - `golang.org/x/crypto/bcrypt` for secret hashing (stdlib fallback available)

## API Endpoints Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/oauth/token` | POST | Token exchange (Client Credentials Grant) |
| `/oauth/introspect` | POST | Token introspection (RFC 7662) |
| `/oauth/revoke` | POST | Token revocation (RFC 7009) |
| `/clients` | GET | List clients (admin) |
| `/clients` | POST | Create client (admin) |
| `/clients/{id}` | GET | Get client (admin) |
| `/clients/{id}` | PATCH | Update client (admin) |
| `/clients/{id}` | DELETE | Delete client (admin) |
| `/clients/{id}/rotate` | POST | Rotate secret (admin) |
| `/scopes` | GET | List scopes |
| `/scopes` | POST | Create scope (admin) |
