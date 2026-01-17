# Package Plan: pkg/tenant

## Overview

A multi-tenancy package for building SaaS applications with tenant isolation. Supports multiple isolation strategies (schema, row-level, database), tenant resolution, and cross-tenant operations with proper data isolation.

## Goals

1. **Multiple Strategies** - Schema, row-level, database isolation
2. **Tenant Resolution** - Domain, header, path, JWT-based
3. **Data Isolation** - Automatic query scoping
4. **Cross-Tenant** - Admin/superuser operations
5. **Middleware** - HTTP and gRPC middleware
6. **Context Propagation** - Tenant context throughout request
7. **Postgres Integration** - Schema-based with pkg/postgres

## Architecture

```
pkg/tenant/
├── tenant.go             # Core interfaces
├── config.go             # Configuration
├── options.go            # Functional options
├── context.go            # Context management
├── resolver/
│   ├── resolver.go       # Resolver interface
│   ├── domain.go         # Domain-based resolution
│   ├── header.go         # Header-based resolution
│   ├── path.go           # Path-based resolution
│   ├── jwt.go            # JWT claim-based resolution
│   └── composite.go      # Multiple resolvers
├── strategy/
│   ├── strategy.go       # Isolation strategy interface
│   ├── schema.go         # Schema-per-tenant
│   ├── row.go            # Row-level isolation
│   └── database.go       # Database-per-tenant
├── middleware/
│   ├── http.go           # HTTP middleware
│   └── grpc.go           # gRPC interceptor
├── postgres/
│   ├── schema.go         # Schema management
│   └── scoped.go         # Scoped queries
├── examples/
│   ├── basic/
│   ├── schema-isolation/
│   └── row-level/
└── README.md
```

## Core Interfaces

```go
package tenant

import (
    "context"
    "time"
)

// Manager manages tenants
type Manager interface {
    // Create creates a new tenant
    Create(ctx context.Context, tenant *Tenant) error

    // Get retrieves a tenant by ID
    Get(ctx context.Context, id string) (*Tenant, error)

    // GetByDomain retrieves by domain
    GetByDomain(ctx context.Context, domain string) (*Tenant, error)

    // Update updates a tenant
    Update(ctx context.Context, tenant *Tenant) error

    // Delete deletes a tenant
    Delete(ctx context.Context, id string) error

    // List lists all tenants
    List(ctx context.Context, filter Filter) ([]*Tenant, error)

    // Provision provisions resources for a tenant
    Provision(ctx context.Context, id string) error

    // Deprovision removes tenant resources
    Deprovision(ctx context.Context, id string) error
}

// Tenant represents a tenant
type Tenant struct {
    // ID is the unique identifier
    ID string

    // Name is the display name
    Name string

    // Slug is the URL-safe identifier
    Slug string

    // Domain is the custom domain (optional)
    Domain string

    // Subdomain for subdomain-based resolution
    Subdomain string

    // Plan is the subscription plan
    Plan string

    // Status: "active", "suspended", "pending"
    Status string

    // Settings holds tenant-specific settings
    Settings map[string]interface{}

    // Metadata for additional data
    Metadata map[string]interface{}

    // CreatedAt timestamp
    CreatedAt time.Time

    // UpdatedAt timestamp
    UpdatedAt time.Time
}

// Resolver resolves tenant from request
type Resolver interface {
    // Resolve returns the tenant ID from context/request
    Resolve(ctx context.Context, r interface{}) (string, error)
}

// Strategy defines the isolation strategy
type Strategy interface {
    // Name returns the strategy name
    Name() string

    // Provision sets up resources for a tenant
    Provision(ctx context.Context, tenant *Tenant) error

    // Deprovision removes resources
    Deprovision(ctx context.Context, tenant *Tenant) error

    // Scope returns a scoped database connection
    Scope(ctx context.Context, tenantID string) (ScopedDB, error)
}

// ScopedDB is a tenant-scoped database interface
type ScopedDB interface {
    // All queries are automatically scoped to tenant
    Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
    Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
    // ... standard database methods
}
```

## Configuration

```go
// Config holds tenant configuration
type Config struct {
    // Strategy: "schema", "row", "database"
    Strategy string `env:"TENANT_STRATEGY" default:"schema"`

    // Default resolver: "domain", "header", "path", "jwt"
    Resolver string `env:"TENANT_RESOLVER" default:"header"`

    // Header name for header-based resolution
    HeaderName string `env:"TENANT_HEADER" default:"X-Tenant-ID"`

    // JWT claim for JWT-based resolution
    JWTClaim string `env:"TENANT_JWT_CLAIM" default:"tenant_id"`

    // Path prefix for path-based resolution
    PathPrefix string `env:"TENANT_PATH_PREFIX" default:"/t/"`

    // Domain suffix for subdomain resolution
    DomainSuffix string `env:"TENANT_DOMAIN_SUFFIX" default:".example.com"`

    // Schema prefix for schema-based isolation
    SchemaPrefix string `env:"TENANT_SCHEMA_PREFIX" default:"tenant_"`

    // Tenant ID column for row-level isolation
    TenantColumn string `env:"TENANT_COLUMN" default:"tenant_id"`
}
```

## Resolver Implementations

```go
// DomainResolver resolves from request domain
type DomainResolver struct {
    suffix string // e.g., ".example.com"
}

// acme.example.com -> "acme"

// HeaderResolver resolves from HTTP header
type HeaderResolver struct {
    headerName string // e.g., "X-Tenant-ID"
}

// PathResolver resolves from URL path
type PathResolver struct {
    prefix string // e.g., "/t/"
}

// /t/acme/api/users -> "acme"

// JWTResolver resolves from JWT claim
type JWTResolver struct {
    claimName string // e.g., "tenant_id"
}

// CompositeResolver tries multiple resolvers
type CompositeResolver struct {
    resolvers []Resolver
}
```

## Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "github.com/user/core-backend/pkg/tenant"
)

func main() {
    mgr, err := tenant.New(tenant.Config{
        Strategy: "schema",
        Resolver: "header",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Create a tenant
    t := &tenant.Tenant{
        ID:   "acme",
        Name: "ACME Corporation",
        Slug: "acme",
        Plan: "enterprise",
    }

    if err := mgr.Create(ctx, t); err != nil {
        log.Fatal(err)
    }

    // Provision resources (create schema, tables, etc.)
    if err := mgr.Provision(ctx, t.ID); err != nil {
        log.Fatal(err)
    }
}
```

### HTTP Middleware

```go
import (
    "github.com/user/core-backend/pkg/tenant"
    "github.com/user/core-backend/pkg/tenant/middleware"
)

func main() {
    mgr, _ := tenant.New(cfg)

    // Middleware resolves and sets tenant context
    mw := middleware.HTTP(mgr,
        middleware.WithResolver(tenant.NewHeaderResolver("X-Tenant-ID")),
        middleware.WithNotFoundHandler(func(w http.ResponseWriter, r *http.Request) {
            http.Error(w, "Tenant not found", http.StatusNotFound)
        }),
    )

    mux := http.NewServeMux()
    http.ListenAndServe(":8080", mw(mux))
}

// In handlers, access tenant from context
func handler(w http.ResponseWriter, r *http.Request) {
    t := tenant.FromContext(r.Context())
    if t == nil {
        http.Error(w, "No tenant", http.StatusBadRequest)
        return
    }

    fmt.Printf("Request from tenant: %s\n", t.Name)
}
```

### Schema-Based Isolation (PostgreSQL)

```go
import (
    "github.com/user/core-backend/pkg/tenant"
    "github.com/user/core-backend/pkg/tenant/strategy"
    "github.com/user/core-backend/pkg/postgres"
)

func main() {
    db, _ := postgres.New(pgConfig)

    schemaStrategy := strategy.NewSchema(db,
        strategy.WithSchemaPrefix("tenant_"),
        strategy.WithMigrations(migrationsFS),
    )

    mgr, _ := tenant.New(tenant.Config{},
        tenant.WithStrategy(schemaStrategy),
    )

    // Create tenant (creates schema "tenant_acme")
    mgr.Create(ctx, &tenant.Tenant{ID: "acme"})
    mgr.Provision(ctx, "acme")

    // In request handler
    t := tenant.FromContext(ctx)
    scopedDB, _ := schemaStrategy.Scope(ctx, t.ID)

    // Queries are scoped to tenant schema
    // SET search_path = tenant_acme
    scopedDB.Query(ctx, "SELECT * FROM users")
}
```

### Row-Level Isolation

```go
import (
    "github.com/user/core-backend/pkg/tenant/strategy"
)

func main() {
    rowStrategy := strategy.NewRow(db,
        strategy.WithTenantColumn("tenant_id"),
    )

    mgr, _ := tenant.New(tenant.Config{},
        tenant.WithStrategy(rowStrategy),
    )

    // In request handler
    t := tenant.FromContext(ctx)
    scopedDB, _ := rowStrategy.Scope(ctx, t.ID)

    // Queries automatically include tenant_id filter
    // SELECT * FROM users WHERE tenant_id = 'acme'
    scopedDB.Query(ctx, "SELECT * FROM users")

    // Inserts automatically include tenant_id
    // INSERT INTO users (name, tenant_id) VALUES ('John', 'acme')
    scopedDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
}
```

### Subdomain Resolution

```go
func main() {
    mgr, _ := tenant.New(tenant.Config{
        Resolver:     "domain",
        DomainSuffix: ".myapp.com",
    })

    // Resolves:
    // acme.myapp.com -> tenant "acme"
    // bigcorp.myapp.com -> tenant "bigcorp"

    mw := middleware.HTTP(mgr)
    http.ListenAndServe(":8080", mw(mux))
}
```

### Cross-Tenant Operations (Admin)

```go
func adminHandler(w http.ResponseWriter, r *http.Request) {
    // Check if user is admin
    if !isAdmin(r.Context()) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Use bypass context for cross-tenant queries
    adminCtx := tenant.WithBypass(r.Context())

    // Query all tenants
    allUsers, _ := db.Query(adminCtx, "SELECT * FROM users")

    // Or iterate tenants
    tenants, _ := mgr.List(adminCtx, tenant.Filter{})
    for _, t := range tenants {
        ctx := tenant.WithTenant(adminCtx, t)
        users, _ := scopedDB.Query(ctx, "SELECT * FROM users")
        // Process users for each tenant
    }
}
```

### Tenant-Aware Repository

```go
type UserRepository struct {
    strategy tenant.Strategy
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    t := tenant.FromContext(ctx)
    if t == nil {
        return nil, tenant.ErrNoTenant
    }

    db, _ := r.strategy.Scope(ctx, t.ID)

    var user User
    err := db.QueryRow(ctx,
        "SELECT id, name, email FROM users WHERE id = $1",
        id,
    ).Scan(&user.ID, &user.Name, &user.Email)

    return &user, err
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    t := tenant.FromContext(ctx)
    if t == nil {
        return tenant.ErrNoTenant
    }

    db, _ := r.strategy.Scope(ctx, t.ID)

    _, err := db.Exec(ctx,
        "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)",
        user.ID, user.Name, user.Email,
    )

    return err
}
```

### Tenant Provisioning

```go
func (s *SchemaStrategy) Provision(ctx context.Context, tenant *Tenant) error {
    schemaName := s.schemaName(tenant.ID)

    // Create schema
    _, err := s.db.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))
    if err != nil {
        return err
    }

    // Run migrations in tenant schema
    for _, migration := range s.migrations {
        _, err := s.db.Exec(ctx, fmt.Sprintf("SET search_path = %s; %s", schemaName, migration))
        if err != nil {
            return err
        }
    }

    return nil
}

func (s *SchemaStrategy) Deprovision(ctx context.Context, tenant *Tenant) error {
    schemaName := s.schemaName(tenant.ID)

    // Drop schema and all objects
    _, err := s.db.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
    return err
}
```

## Error Handling

```go
var (
    // ErrNoTenant is returned when tenant context is missing
    ErrNoTenant = errors.New("tenant: no tenant in context")

    // ErrTenantNotFound is returned when tenant doesn't exist
    ErrTenantNotFound = errors.New("tenant: tenant not found")

    // ErrTenantSuspended is returned for suspended tenants
    ErrTenantSuspended = errors.New("tenant: tenant is suspended")

    // ErrProvisionFailed is returned when provisioning fails
    ErrProvisionFailed = errors.New("tenant: provisioning failed")
)
```

## Dependencies

- **Required:** None (memory store)
- **Optional:**
  - `github.com/jackc/pgx/v5` for PostgreSQL strategies

## Implementation Phases

### Phase 1: Core Interface
1. Define Tenant, Manager interfaces
2. Context management
3. Basic resolver

### Phase 2: Resolvers
1. Header resolver
2. Domain resolver
3. Path resolver
4. JWT resolver

### Phase 3: Schema Strategy
1. PostgreSQL schema isolation
2. Schema provisioning
3. Migration support

### Phase 4: Row-Level Strategy
1. Query scoping
2. Automatic tenant_id injection

### Phase 5: Middleware
1. HTTP middleware
2. gRPC interceptor
3. Cross-tenant bypass

### Phase 6: Documentation
1. README
2. Examples
3. Migration guide
