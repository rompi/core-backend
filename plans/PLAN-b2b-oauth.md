# Package Plan: pkg/b2boauth

## Overview

A **B2B (Business-to-Business) OAuth 2.0 authorization package** designed for secure machine-to-machine authentication. This package enables partner organizations, external services, and third-party integrations to authenticate with your APIs using industry-standard OAuth 2.0 Client Credentials flow.

**Key Value Proposition:** Create a client ID and client secret in seconds, distribute to your partners, and they can immediately start making authenticated API calls - all with enterprise-grade security built-in.

---

## Feature Highlights

### 1. Easy Client Onboarding
| What You Get | Description |
|--------------|-------------|
| **One-Command Client Creation** | Create a new B2B client with a single API call |
| **Auto-Generated Credentials** | Cryptographically secure client ID (32 chars) and secret (64 chars) generated automatically |
| **Instant Distribution** | Share credentials with partners - they can start authenticating immediately |
| **Custom Metadata** | Attach partner-specific data (partner_id, tier, environment) to each client |

### 2. OAuth 2.0 Compliance (RFC 6749)
| What You Get | Description |
|--------------|-------------|
| **Client Credentials Grant** | Industry-standard flow for server-to-server authentication |
| **Bearer Token Format** | JWT access tokens compatible with any HTTP client |
| **Token Introspection** | RFC 7662 compliant endpoint for validating tokens |
| **Token Revocation** | RFC 7009 compliant endpoint for invalidating tokens |
| **Standard Error Responses** | OAuth 2.0 compliant error codes and messages |

### 3. Storage Agnostic Design
| What You Get | Description |
|--------------|-------------|
| **Repository Interfaces** | Plug in PostgreSQL, MongoDB, Redis, DynamoDB, or any storage |
| **No Vendor Lock-in** | Switch storage backends without changing business logic |
| **Mix & Match** | Use PostgreSQL for clients, Redis for tokens - your choice |
| **Mock Implementations** | Ready-to-use mocks for unit testing |

### 4. Scope-Based Authorization
| What You Get | Description |
|--------------|-------------|
| **Fine-Grained Permissions** | Define scopes like `read:orders`, `write:users`, `admin:*` |
| **Per-Client Scope Limits** | Restrict which scopes each client can request |
| **Scope Validation** | Automatic validation that requested scopes are allowed |
| **Hierarchical Scopes** | Support for patterns like `orders:read`, `orders:write`, `orders:*` |

### 5. Enterprise Security
| What You Get | Description |
|--------------|-------------|
| **bcrypt Secret Hashing** | Secrets are never stored in plain text |
| **Account Lockout** | Automatic lockout after N failed authentication attempts |
| **Per-Client Rate Limiting** | Configurable RPS limits per client |
| **Secret Rotation** | Rotate secrets without service interruption |
| **Comprehensive Audit Logs** | Track all authentication events for compliance |

### 6. Ready-to-Use Middleware
| What You Get | Description |
|--------------|-------------|
| **HTTP Middleware** | Drop-in middleware for protected endpoints |
| **gRPC Interceptors** | Unary and streaming interceptors for gRPC services |
| **Context Integration** | Access client claims from request context |
| **Scope Enforcement** | Middleware automatically validates required scopes |

---

## Functional Capabilities

### A. Client Management

```
┌─────────────────────────────────────────────────────────────────┐
│                     CLIENT LIFECYCLE                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   CREATE ──► ACTIVE ──► SUSPENDED ──► ACTIVE ──► REVOKED        │
│                │                         │                       │
│                └─── SECRET ROTATION ─────┘                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

| Function | Description | Use Case |
|----------|-------------|----------|
| `CreateClient` | Register new B2B partner | Onboard a new integration partner |
| `GetClient` | Retrieve client details | View partner configuration |
| `UpdateClient` | Modify client settings | Change allowed scopes, rate limits |
| `DeleteClient` | Remove client | Offboard a partner |
| `RotateClientSecret` | Generate new secret | Periodic security rotation |
| `ListClients` | Paginated client list | Admin dashboard |
| `SuspendClient` | Temporarily disable | Maintenance or investigation |
| `ReactivateClient` | Re-enable suspended client | After issue resolution |

### B. Token Operations

```
┌─────────────────────────────────────────────────────────────────┐
│                  TOKEN EXCHANGE FLOW                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Partner App                        Your API                    │
│       │                                  │                       │
│       │──── POST /oauth/token ──────────►│                       │
│       │     client_id + client_secret    │                       │
│       │     grant_type=client_credentials│                       │
│       │     scope=read:orders            │                       │
│       │                                  │                       │
│       │◄─── { access_token, expires_in } │                       │
│       │                                  │                       │
│       │──── GET /api/orders ────────────►│                       │
│       │     Authorization: Bearer <token>│                       │
│       │                                  │                       │
│       │◄─── { orders: [...] } ──────────│                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

| Function | Description | Use Case |
|----------|-------------|----------|
| `TokenExchange` | Issue access token | Partner authenticates to get token |
| `ValidateToken` | Verify and decode token | Middleware validates incoming requests |
| `RevokeToken` | Invalidate specific token | Security incident response |
| `IntrospectToken` | Check token status | Debug or audit token state |
| `RevokeAllClientTokens` | Invalidate all tokens for client | Emergency client lockout |

### C. Scope Management

```
┌─────────────────────────────────────────────────────────────────┐
│                    SCOPE HIERARCHY EXAMPLE                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   orders:*          ──► Full access to orders                    │
│       ├── orders:read   ──► Read orders                          │
│       ├── orders:write  ──► Create/update orders                 │
│       └── orders:delete ──► Delete orders                        │
│                                                                  │
│   users:*           ──► Full access to users                     │
│       ├── users:read    ──► Read user profiles                   │
│       └── users:write   ──► Modify user profiles                 │
│                                                                  │
│   admin:*           ──► Administrative access                    │
│       ├── admin:clients ──► Manage B2B clients                   │
│       └── admin:scopes  ──► Manage scopes                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

| Function | Description | Use Case |
|----------|-------------|----------|
| `CreateScope` | Define new permission | Add new API capability |
| `GetScope` | Retrieve scope details | View scope description |
| `ListScopes` | List all available scopes | Documentation, admin UI |
| `DeleteScope` | Remove scope | Deprecate API capability |
| `ValidateScopes` | Check scopes exist | Validate client configuration |

### D. Security Features

```
┌─────────────────────────────────────────────────────────────────┐
│                    SECURITY LAYERS                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Layer 1: SECRET HASHING                                        │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │ Client secrets are bcrypt hashed (cost=12)               │   │
│   │ Original secret shown ONCE at creation, never stored     │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│   Layer 2: ACCOUNT LOCKOUT                                       │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │ 5 failed attempts → 15 minute lockout (configurable)     │   │
│   │ Prevents brute force attacks                             │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│   Layer 3: RATE LIMITING                                         │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │ Per-client RPS limits (default: 100/sec)                 │   │
│   │ Token bucket algorithm for smooth limiting               │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│   Layer 4: AUDIT LOGGING                                         │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │ All auth events logged with IP, User-Agent, timestamp    │   │
│   │ Compliance-ready audit trail                             │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│   Layer 5: TOKEN REVOCATION                                      │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │ Instant token invalidation                               │   │
│   │ Bulk revocation per client                               │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### E. Middleware Integration

```
┌─────────────────────────────────────────────────────────────────┐
│                  MIDDLEWARE FLOW                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Request                                                        │
│      │                                                           │
│      ▼                                                           │
│   ┌──────────────────┐                                          │
│   │ Extract Token    │  Authorization: Bearer <token>           │
│   └────────┬─────────┘                                          │
│            │                                                     │
│            ▼                                                     │
│   ┌──────────────────┐                                          │
│   │ Validate JWT     │  Check signature, expiration             │
│   └────────┬─────────┘                                          │
│            │                                                     │
│            ▼                                                     │
│   ┌──────────────────┐                                          │
│   │ Check Revocation │  Is token in revocation list?            │
│   └────────┬─────────┘                                          │
│            │                                                     │
│            ▼                                                     │
│   ┌──────────────────┐                                          │
│   │ Verify Scopes    │  Does token have required scopes?        │
│   └────────┬─────────┘                                          │
│            │                                                     │
│            ▼                                                     │
│   ┌──────────────────┐                                          │
│   │ Inject Claims    │  Add client info to request context      │
│   └────────┬─────────┘                                          │
│            │                                                     │
│            ▼                                                     │
│      Your Handler                                                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Architecture

```
pkg/b2boauth/
├── b2boauth.go           # Core Service interface
├── service.go            # Service implementation
├── repository.go         # Repository interfaces (storage agnostic)
├── models.go             # Domain models (Client, Token, Scope, etc.)
├── config.go             # Configuration with env var support
├── errors.go             # Error definitions with OAuth 2.0 codes
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

---

## Core Interfaces

### Service Interface

```go
package b2boauth

import (
    "context"
    "net/http"
    "time"

    "google.golang.org/grpc"
)

// Service provides complete B2B OAuth 2.0 functionality
type Service interface {
    // ══════════════════════════════════════════════════════════════
    // CLIENT MANAGEMENT
    // ══════════════════════════════════════════════════════════════

    // CreateClient registers a new B2B partner and returns credentials.
    // The client secret is returned ONLY ONCE - store it securely!
    CreateClient(ctx context.Context, req CreateClientRequest) (*ClientWithSecret, error)

    // GetClient retrieves client details by client_id.
    GetClient(ctx context.Context, clientID string) (*Client, error)

    // UpdateClient modifies client settings (scopes, rate limits, metadata).
    UpdateClient(ctx context.Context, clientID string, req UpdateClientRequest) (*Client, error)

    // DeleteClient permanently removes a client and revokes all tokens.
    DeleteClient(ctx context.Context, clientID string) error

    // RotateClientSecret generates a new secret, invalidating the old one.
    // Existing tokens remain valid until expiration.
    RotateClientSecret(ctx context.Context, clientID string) (*ClientCredentials, error)

    // ListClients returns a paginated list of all clients.
    ListClients(ctx context.Context, opts ListOptions) (*ClientList, error)

    // SuspendClient temporarily disables a client (can be reactivated).
    SuspendClient(ctx context.Context, clientID string) error

    // ReactivateClient re-enables a suspended client.
    ReactivateClient(ctx context.Context, clientID string) error

    // ══════════════════════════════════════════════════════════════
    // TOKEN OPERATIONS (OAuth 2.0 Client Credentials Grant)
    // ══════════════════════════════════════════════════════════════

    // TokenExchange implements the OAuth 2.0 token endpoint.
    // Partners exchange client_id + client_secret for an access token.
    TokenExchange(ctx context.Context, req TokenRequest) (*TokenResponse, error)

    // ValidateToken verifies a token and returns the claims.
    // Used by middleware to authenticate requests.
    ValidateToken(ctx context.Context, token string) (*TokenClaims, error)

    // RevokeToken immediately invalidates a specific token.
    RevokeToken(ctx context.Context, tokenJTI string) error

    // RevokeAllClientTokens invalidates all tokens for a client.
    RevokeAllClientTokens(ctx context.Context, clientID string) error

    // IntrospectToken returns token status (RFC 7662).
    IntrospectToken(ctx context.Context, token string) (*IntrospectionResponse, error)

    // ══════════════════════════════════════════════════════════════
    // SCOPE MANAGEMENT
    // ══════════════════════════════════════════════════════════════

    // CreateScope defines a new permission scope.
    CreateScope(ctx context.Context, scope *Scope) error

    // GetScope retrieves a scope by name.
    GetScope(ctx context.Context, name string) (*Scope, error)

    // ListScopes returns all available scopes.
    ListScopes(ctx context.Context) ([]*Scope, error)

    // DeleteScope removes a scope definition.
    DeleteScope(ctx context.Context, name string) error

    // ══════════════════════════════════════════════════════════════
    // MIDDLEWARE (Drop-in protection for your endpoints)
    // ══════════════════════════════════════════════════════════════

    // HTTPMiddleware returns middleware that validates tokens and scopes.
    // Usage: router.Handle("/api/orders", svc.HTTPMiddleware("read:orders")(handler))
    HTTPMiddleware(requiredScopes ...string) func(http.Handler) http.Handler

    // GRPCUnaryInterceptor returns an interceptor for unary gRPC calls.
    GRPCUnaryInterceptor(requiredScopes ...string) grpc.UnaryServerInterceptor

    // GRPCStreamInterceptor returns an interceptor for streaming gRPC calls.
    GRPCStreamInterceptor(requiredScopes ...string) grpc.StreamServerInterceptor
}
```

### Request/Response Types

```go
// ══════════════════════════════════════════════════════════════
// TOKEN EXCHANGE (OAuth 2.0)
// ══════════════════════════════════════════════════════════════

// TokenRequest represents the OAuth 2.0 token request.
// Partners send this to exchange credentials for an access token.
type TokenRequest struct {
    GrantType    string `json:"grant_type"`    // MUST be "client_credentials"
    ClientID     string `json:"client_id"`     // The client's unique identifier
    ClientSecret string `json:"client_secret"` // The client's secret
    Scope        string `json:"scope"`         // Space-separated list of scopes
}

// TokenResponse represents the OAuth 2.0 token response.
// This is returned to partners after successful authentication.
type TokenResponse struct {
    AccessToken string `json:"access_token"`  // JWT access token
    TokenType   string `json:"token_type"`    // Always "Bearer"
    ExpiresIn   int64  `json:"expires_in"`    // Seconds until expiration
    Scope       string `json:"scope"`         // Granted scopes (may differ from requested)
}

// TokenClaims represents the decoded JWT payload.
// Available in request context after middleware validation.
type TokenClaims struct {
    ClientID   string    `json:"client_id"`  // Which client made this request
    ClientName string    `json:"client_name"`// Human-readable client name
    Scopes     []string  `json:"scopes"`     // Granted permissions
    ExpiresAt  time.Time `json:"exp"`        // When token expires
    IssuedAt   time.Time `json:"iat"`        // When token was issued
    Issuer     string    `json:"iss"`        // Token issuer
    Audience   string    `json:"aud"`        // Intended audience
    JTI        string    `json:"jti"`        // Unique token ID (for revocation)
}

// IntrospectionResponse represents RFC 7662 Token Introspection response.
type IntrospectionResponse struct {
    Active    bool   `json:"active"`              // Is token currently valid?
    Scope     string `json:"scope,omitempty"`     // Token scopes
    ClientID  string `json:"client_id,omitempty"` // Token owner
    TokenType string `json:"token_type,omitempty"`// Token type
    Exp       int64  `json:"exp,omitempty"`       // Expiration timestamp
    Iat       int64  `json:"iat,omitempty"`       // Issued at timestamp
    Iss       string `json:"iss,omitempty"`       // Issuer
    Aud       string `json:"aud,omitempty"`       // Audience
}
```

---

## Repository Interfaces (Storage Agnostic)

The package uses **repository interfaces** to abstract storage. Implement these interfaces for your preferred database.

```go
package b2boauth

import "context"

// ══════════════════════════════════════════════════════════════
// CLIENT REPOSITORY (REQUIRED)
// Implement this for your storage backend
// ══════════════════════════════════════════════════════════════

type ClientRepository interface {
    // Create stores a new OAuth client
    Create(ctx context.Context, client *Client) error

    // GetByID retrieves a client by internal UUID
    GetByID(ctx context.Context, id string) (*Client, error)

    // GetByClientID retrieves a client by OAuth client_id
    GetByClientID(ctx context.Context, clientID string) (*Client, error)

    // Update modifies an existing client
    Update(ctx context.Context, client *Client) error

    // Delete removes a client
    Delete(ctx context.Context, id string) error

    // List returns paginated clients
    List(ctx context.Context, opts ListOptions) ([]*Client, int64, error)

    // IncrementFailedAttempts tracks authentication failures
    IncrementFailedAttempts(ctx context.Context, clientID string) error

    // ResetFailedAttempts clears the failure counter
    ResetFailedAttempts(ctx context.Context, clientID string) error
}

// ══════════════════════════════════════════════════════════════
// TOKEN REPOSITORY (REQUIRED for revocation/introspection)
// Can use Redis for high performance
// ══════════════════════════════════════════════════════════════

type TokenRepository interface {
    // Create stores an issued token (for tracking)
    Create(ctx context.Context, token *IssuedToken) error

    // GetByJTI retrieves a token by unique identifier
    GetByJTI(ctx context.Context, jti string) (*IssuedToken, error)

    // GetByClientID retrieves all tokens for a client
    GetByClientID(ctx context.Context, clientID string) ([]*IssuedToken, error)

    // Revoke marks a token as revoked
    Revoke(ctx context.Context, jti string) error

    // RevokeByClientID revokes all tokens for a client
    RevokeByClientID(ctx context.Context, clientID string) error

    // IsRevoked checks if a token has been revoked
    IsRevoked(ctx context.Context, jti string) (bool, error)

    // DeleteExpired removes expired tokens (cleanup job)
    DeleteExpired(ctx context.Context) error
}

// ══════════════════════════════════════════════════════════════
// SCOPE REPOSITORY (OPTIONAL - defaults to in-memory)
// ══════════════════════════════════════════════════════════════

type ScopeRepository interface {
    Create(ctx context.Context, scope *Scope) error
    GetByName(ctx context.Context, name string) (*Scope, error)
    List(ctx context.Context) ([]*Scope, error)
    Delete(ctx context.Context, name string) error
    ValidateScopes(ctx context.Context, scopes []string) error
}

// ══════════════════════════════════════════════════════════════
// AUDIT LOG REPOSITORY (OPTIONAL - disables audit if nil)
// ══════════════════════════════════════════════════════════════

type AuditLogRepository interface {
    Create(ctx context.Context, log *AuditLog) error
    GetByClientID(ctx context.Context, clientID string, opts ListOptions) ([]*AuditLog, error)
    GetByAction(ctx context.Context, action string, opts ListOptions) ([]*AuditLog, error)
}

// ══════════════════════════════════════════════════════════════
// REPOSITORY AGGREGATOR
// ══════════════════════════════════════════════════════════════

type Repositories struct {
    Clients   ClientRepository   // REQUIRED
    Tokens    TokenRepository    // REQUIRED for revocation
    Scopes    ScopeRepository    // OPTIONAL (in-memory default)
    AuditLogs AuditLogRepository // OPTIONAL (disabled if nil)
}
```

---

## Domain Models

```go
package b2boauth

import "time"

// ══════════════════════════════════════════════════════════════
// CLIENT MODEL
// ══════════════════════════════════════════════════════════════

// Client represents a B2B partner/integration
type Client struct {
    // Identity
    ID       string `json:"id"`        // Internal UUID
    ClientID string `json:"client_id"` // Public OAuth client_id (32 chars)
    Name     string `json:"name"`      // Human-readable name ("Partner Corp API")
    Description string `json:"description"` // Optional description

    // Security (secret hash never exposed in JSON)
    ClientSecretHash string `json:"-"`

    // Permissions
    AllowedScopes []string `json:"allowed_scopes"` // Scopes this client can request

    // Configuration
    Metadata     Metadata `json:"metadata"`       // Custom key-value data
    RateLimitRPS int      `json:"rate_limit_rps"` // Requests per second limit

    // Status
    Status         Status     `json:"status"`          // active, suspended, revoked
    FailedAttempts int        `json:"failed_attempts"` // Consecutive auth failures
    LockedUntil    *time.Time `json:"locked_until"`    // Account lock expiry

    // Timestamps
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    LastUsedAt *time.Time `json:"last_used_at"`
    CreatedBy  string     `json:"created_by"` // Admin/system that created
}

// ClientWithSecret is returned ONLY at client creation
type ClientWithSecret struct {
    Client
    ClientSecret string `json:"client_secret"` // SHOWN ONCE - store securely!
}

// ClientCredentials for secret rotation response
type ClientCredentials struct {
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"` // New secret - shown once!
}

// Status represents client lifecycle state
type Status string

const (
    StatusActive    Status = "active"    // Can authenticate
    StatusSuspended Status = "suspended" // Temporarily disabled
    StatusRevoked   Status = "revoked"   // Permanently disabled
)

// Metadata for custom client properties
type Metadata map[string]interface{}

// ══════════════════════════════════════════════════════════════
// SCOPE MODEL
// ══════════════════════════════════════════════════════════════

// Scope defines a permission
type Scope struct {
    Name        string `json:"name"`        // e.g., "read:orders"
    Description string `json:"description"` // Human-readable description
    System      bool   `json:"system"`      // True if system-defined (not deletable)
}

// ══════════════════════════════════════════════════════════════
// TOKEN TRACKING MODEL
// ══════════════════════════════════════════════════════════════

// IssuedToken tracks tokens for revocation/audit
type IssuedToken struct {
    JTI       string     `json:"jti"`        // Unique token identifier
    ClientID  string     `json:"client_id"`  // Token owner
    Scopes    []string   `json:"scopes"`     // Granted scopes
    IssuedAt  time.Time  `json:"issued_at"`  // When issued
    ExpiresAt time.Time  `json:"expires_at"` // When expires
    RevokedAt *time.Time `json:"revoked_at"` // When revoked (if revoked)
    IPAddress string     `json:"ip_address"` // Client IP at issuance
    UserAgent string     `json:"user_agent"` // Client user agent
}

// ══════════════════════════════════════════════════════════════
// AUDIT LOG MODEL
// ══════════════════════════════════════════════════════════════

// AuditLog records security events
type AuditLog struct {
    ID        string                 `json:"id"`
    ClientID  string                 `json:"client_id"`
    Action    AuditAction            `json:"action"`
    Details   map[string]interface{} `json:"details"`
    IPAddress string                 `json:"ip_address"`
    UserAgent string                 `json:"user_agent"`
    Timestamp time.Time              `json:"timestamp"`
}

// AuditAction defines tracked events
type AuditAction string

const (
    ActionTokenIssued     AuditAction = "token_issued"
    ActionTokenRevoked    AuditAction = "token_revoked"
    ActionAuthFailed      AuditAction = "auth_failed"
    ActionAuthLocked      AuditAction = "auth_locked"
    ActionClientCreated   AuditAction = "client_created"
    ActionClientUpdated   AuditAction = "client_updated"
    ActionClientDeleted   AuditAction = "client_deleted"
    ActionClientSuspended AuditAction = "client_suspended"
    ActionClientReactivated AuditAction = "client_reactivated"
    ActionSecretRotated   AuditAction = "secret_rotated"
)
```

---

## Configuration

```go
// Config holds all B2B OAuth settings
type Config struct {
    // ══════════════════════════════════════════════════════════
    // TOKEN SETTINGS
    // ══════════════════════════════════════════════════════════
    TokenExpiration    time.Duration // How long tokens are valid (default: 1h)
    TokenIssuer        string        // JWT "iss" claim (default: "b2b-oauth")
    TokenAudience      string        // JWT "aud" claim (default: "api")
    TokenSigningKey    string        // JWT signing secret (REQUIRED)
    TokenSigningMethod string        // HS256, RS256, ES256 (default: HS256)

    // ══════════════════════════════════════════════════════════
    // CLIENT SETTINGS
    // ══════════════════════════════════════════════════════════
    ClientIDLength       int // Length of generated client_id (default: 32)
    ClientSecretLength   int // Length of generated secret (default: 64)
    ClientSecretHashCost int // bcrypt cost factor (default: 12)

    // ══════════════════════════════════════════════════════════
    // SECURITY SETTINGS
    // ══════════════════════════════════════════════════════════
    MaxFailedAttempts   int           // Failures before lockout (default: 5)
    LockoutDuration     time.Duration // How long lockout lasts (default: 15m)
    RateLimitEnabled    bool          // Enable rate limiting (default: true)
    DefaultRateLimitRPS int           // Default RPS per client (default: 100)

    // ══════════════════════════════════════════════════════════
    // AUDIT SETTINGS
    // ══════════════════════════════════════════════════════════
    AuditEnabled bool // Enable audit logging (default: true)
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `B2B_OAUTH_TOKEN_EXPIRATION` | Token validity duration | `1h` |
| `B2B_OAUTH_TOKEN_ISSUER` | JWT issuer claim | `b2b-oauth` |
| `B2B_OAUTH_TOKEN_AUDIENCE` | JWT audience claim | `api` |
| `B2B_OAUTH_TOKEN_SIGNING_KEY` | JWT signing secret | **REQUIRED** |
| `B2B_OAUTH_TOKEN_SIGNING_METHOD` | JWT algorithm | `HS256` |
| `B2B_OAUTH_CLIENT_ID_LENGTH` | Client ID length | `32` |
| `B2B_OAUTH_CLIENT_SECRET_LENGTH` | Secret length | `64` |
| `B2B_OAUTH_MAX_FAILED_ATTEMPTS` | Failures before lockout | `5` |
| `B2B_OAUTH_LOCKOUT_DURATION` | Lockout duration | `15m` |
| `B2B_OAUTH_DEFAULT_RATE_LIMIT` | Default RPS | `100` |
| `B2B_OAUTH_AUDIT_ENABLED` | Enable audit logs | `true` |

---

## Error Handling

### OAuth 2.0 Compliant Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `invalid_request` | 400 | Malformed request |
| `invalid_client` | 401 | Unknown or invalid client credentials |
| `invalid_grant` | 400 | Invalid grant type |
| `unauthorized_client` | 403 | Client not authorized for this grant |
| `unsupported_grant_type` | 400 | Grant type not supported |
| `invalid_scope` | 400 | Requested scope is invalid |
| `access_denied` | 403 | Resource access denied |
| `client_locked` | 403 | Client account is locked |
| `client_suspended` | 403 | Client account is suspended |
| `token_revoked` | 401 | Token has been revoked |
| `token_expired` | 401 | Token has expired |
| `rate_limit_exceeded` | 429 | Too many requests |

### Error Response Format

```json
{
    "error": "invalid_client",
    "error_description": "Client authentication failed"
}
```

---

## Usage Examples

### 1. Quick Start - Initialize Service

```go
package main

import (
    "log"
    "github.com/user/core-backend/pkg/b2boauth"
)

func main() {
    // Load configuration from environment
    cfg, err := b2boauth.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Plug in your storage implementations
    repos := b2boauth.Repositories{
        Clients: NewPostgresClientRepo(db),  // Your implementation
        Tokens:  NewRedisTokenRepo(redis),   // Your implementation
        // Scopes and AuditLogs are optional
    }

    // Create the service
    svc, err := b2boauth.NewService(cfg, repos)
    if err != nil {
        log.Fatal(err)
    }

    // Ready to use!
}
```

### 2. Onboard a New Partner

```go
// Create a new B2B client for a partner
result, err := svc.CreateClient(ctx, b2boauth.CreateClientRequest{
    Name:          "Acme Corp Integration",
    Description:   "Order sync for Acme Corporation",
    AllowedScopes: []string{"read:orders", "write:orders"},
    Metadata: b2boauth.Metadata{
        "partner_id":  "ACME-001",
        "environment": "production",
        "contact":     "tech@acme.com",
    },
    RateLimitRPS: 500, // Custom rate limit
})
if err != nil {
    log.Fatal(err)
}

// ⚠️  IMPORTANT: Save these credentials securely!
// The secret is shown ONLY ONCE
fmt.Printf("Client ID: %s\n", result.ClientID)
fmt.Printf("Client Secret: %s\n", result.ClientSecret)

// Send credentials to partner via secure channel
```

### 3. Partner Authentication (Token Exchange)

```go
// This is what your partner's application does:

// Step 1: Exchange credentials for token
tokenResp, err := svc.TokenExchange(ctx, b2boauth.TokenRequest{
    GrantType:    "client_credentials",
    ClientID:     "abc123def456...",
    ClientSecret: "xyz789secret...",
    Scope:        "read:orders write:orders",
})
// Response:
// {
//     "access_token": "eyJhbGciOiJIUzI1NiIs...",
//     "token_type": "Bearer",
//     "expires_in": 3600,
//     "scope": "read:orders write:orders"
// }

// Step 2: Use the token in API requests
req.Header.Set("Authorization", "Bearer " + tokenResp.AccessToken)
```

### 4. Protect Your Endpoints (HTTP Middleware)

```go
func main() {
    svc, _ := b2boauth.NewService(cfg, repos)

    mux := http.NewServeMux()

    // Public: Token endpoint (no auth required)
    mux.HandleFunc("/oauth/token", handleTokenExchange(svc))

    // Protected: Require "read:orders" scope
    mux.Handle("/api/orders",
        svc.HTTPMiddleware("read:orders")(http.HandlerFunc(listOrders)))

    // Protected: Require "write:orders" scope
    mux.Handle("/api/orders/create",
        svc.HTTPMiddleware("write:orders")(http.HandlerFunc(createOrder)))

    // Protected: Require MULTIPLE scopes (AND logic)
    mux.Handle("/api/admin/clients",
        svc.HTTPMiddleware("admin:read", "admin:write")(http.HandlerFunc(manageClients)))

    http.ListenAndServe(":8080", mux)
}

// Access client info in your handlers
func listOrders(w http.ResponseWriter, r *http.Request) {
    // Get validated claims from context
    claims := b2boauth.ClaimsFromContext(r.Context())

    log.Printf("Request from: %s (%s)", claims.ClientName, claims.ClientID)
    log.Printf("Scopes: %v", claims.Scopes)

    // ... handle request
}
```

### 5. Protect gRPC Services

```go
func main() {
    svc, _ := b2boauth.NewService(cfg, repos)

    server := grpc.NewServer(
        grpc.UnaryInterceptor(svc.GRPCUnaryInterceptor("read:data")),
        grpc.StreamInterceptor(svc.GRPCStreamInterceptor("read:data")),
    )

    // Register your services...
}
```

### 6. Rotate Client Secret

```go
// Rotate secret (old secret immediately invalidated)
newCreds, err := svc.RotateClientSecret(ctx, "client-id-here")
if err != nil {
    log.Fatal(err)
}

// Send new secret to partner
fmt.Printf("New Secret: %s\n", newCreds.ClientSecret)

// Note: Existing tokens remain valid until they expire
// To invalidate all tokens immediately:
_ = svc.RevokeAllClientTokens(ctx, "client-id-here")
```

### 7. Suspend/Reactivate Client

```go
// Temporarily suspend (investigation, maintenance, etc.)
err := svc.SuspendClient(ctx, "client-id")

// Reactivate when ready
err = svc.ReactivateClient(ctx, "client-id")
```

---

## Storage Implementation Examples

### PostgreSQL Schema

```sql
-- ═══════════════════════════════════════════════════════════════
-- B2B OAuth Clients
-- ═══════════════════════════════════════════════════════════════
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
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    created_by VARCHAR(255)
);

CREATE INDEX idx_clients_client_id ON b2b_oauth_clients(client_id);
CREATE INDEX idx_clients_status ON b2b_oauth_clients(status);

-- ═══════════════════════════════════════════════════════════════
-- Issued Tokens (for revocation tracking)
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE b2b_oauth_tokens (
    jti UUID PRIMARY KEY,
    client_id VARCHAR(64) NOT NULL,
    scopes TEXT[] NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    ip_address VARCHAR(45),
    user_agent TEXT,
    FOREIGN KEY (client_id) REFERENCES b2b_oauth_clients(client_id)
);

CREATE INDEX idx_tokens_client_id ON b2b_oauth_tokens(client_id);
CREATE INDEX idx_tokens_expires_at ON b2b_oauth_tokens(expires_at);

-- ═══════════════════════════════════════════════════════════════
-- Scopes
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE b2b_oauth_scopes (
    name VARCHAR(100) PRIMARY KEY,
    description TEXT,
    system BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ═══════════════════════════════════════════════════════════════
-- Audit Logs
-- ═══════════════════════════════════════════════════════════════
CREATE TABLE b2b_oauth_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(64),
    action VARCHAR(50) NOT NULL,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    timestamp TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_client_id ON b2b_oauth_audit_logs(client_id);
CREATE INDEX idx_audit_action ON b2b_oauth_audit_logs(action);
CREATE INDEX idx_audit_timestamp ON b2b_oauth_audit_logs(timestamp);
```

### Redis Token Repository (High Performance)

```go
// For high-throughput token operations
type RedisTokenRepository struct {
    client *redis.Client
    prefix string // e.g., "b2b:token:"
}

func (r *RedisTokenRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
    key := r.prefix + "revoked:" + jti
    exists, err := r.client.Exists(ctx, key).Result()
    return exists > 0, err
}

func (r *RedisTokenRepository) Revoke(ctx context.Context, jti string) error {
    key := r.prefix + "revoked:" + jti
    // Keep revocation record longer than max token lifetime
    return r.client.Set(ctx, key, "1", 24*time.Hour).Err()
}
```

---

## API Endpoints Reference

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/oauth/token` | POST | None | Exchange credentials for access token |
| `/oauth/introspect` | POST | Bearer | Check token validity (RFC 7662) |
| `/oauth/revoke` | POST | Bearer | Revoke a token (RFC 7009) |
| `/admin/clients` | GET | Admin | List all clients |
| `/admin/clients` | POST | Admin | Create new client |
| `/admin/clients/{id}` | GET | Admin | Get client details |
| `/admin/clients/{id}` | PATCH | Admin | Update client |
| `/admin/clients/{id}` | DELETE | Admin | Delete client |
| `/admin/clients/{id}/rotate` | POST | Admin | Rotate secret |
| `/admin/clients/{id}/suspend` | POST | Admin | Suspend client |
| `/admin/clients/{id}/reactivate` | POST | Admin | Reactivate client |
| `/admin/scopes` | GET | Admin | List scopes |
| `/admin/scopes` | POST | Admin | Create scope |

---

## Implementation Phases

| Phase | Focus | Deliverables |
|-------|-------|--------------|
| **1** | Core Foundation | Service interface, Repository interfaces, Models, Config |
| **2** | Token Engine | JWT generation, Validation, Claims extraction |
| **3** | Client Management | CRUD, Secret hashing, Secret rotation, Status management |
| **4** | OAuth 2.0 Flow | Client Credentials Grant, Scope validation, Error responses |
| **5** | Security | Account lockout, Rate limiting, Audit logging |
| **6** | Middleware | HTTP middleware, gRPC interceptors, Context helpers |
| **7** | Token Lifecycle | Revocation, Introspection, Bulk operations |
| **8** | Quality | Mock repositories, Unit tests, Integration tests, Docs |

---

## Dependencies

| Type | Package | Purpose |
|------|---------|---------|
| **Required** | stdlib | Core functionality |
| **Optional** | `github.com/golang-jwt/jwt/v5` | JWT handling |
| **Optional** | `golang.org/x/crypto/bcrypt` | Secret hashing |

---

## Test Coverage Requirements

- Unit tests for all public functions
- Integration tests with mock repositories
- OAuth 2.0 compliance tests
- Security edge cases (lockout, rate limiting)
- **Target: 80%+ coverage**
