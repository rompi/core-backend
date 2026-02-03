# Pluggable Role-Based Access Control (RBAC) System Design

## Overview

This document outlines the design for a **pluggable, provider-agnostic, database-agnostic RBAC system** with **multi-tenant support**. The design follows the established patterns in the core-backend codebase:

- Interface-based dependency injection
- Functional options pattern
- Repository pattern for storage abstraction
- Middleware/interceptor composability

## Table of Contents

1. [Core Concepts](#core-concepts)
2. [Architecture Overview](#architecture-overview)
3. [Domain Models](#domain-models)
4. [Interface Definitions](#interface-definitions)
5. [Multi-Tenancy Design](#multi-tenancy-design)
6. [Provider Abstraction](#provider-abstraction)
7. [Storage Abstraction](#storage-abstraction)
8. [Middleware & Interceptors](#middleware--interceptors)
9. [Policy Engine](#policy-engine)
10. [Implementation Plan](#implementation-plan)
11. [Usage Examples](#usage-examples)

---

## Core Concepts

### RBAC Hierarchy

```
┌─────────────────────────────────────────────────────────────────┐
│                          TENANT                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                      ORGANIZATION                        │    │
│  │  ┌─────────────────────────────────────────────────┐    │    │
│  │  │                    RESOURCE                      │    │    │
│  │  │  ┌─────────────────────────────────────────┐    │    │    │
│  │  │  │              PERMISSION                  │    │    │    │
│  │  │  │  (resource:action - e.g., users:read)   │    │    │    │
│  │  │  └─────────────────────────────────────────┘    │    │    │
│  │  └─────────────────────────────────────────────────┘    │    │
│  │                         │                                │    │
│  │                    assigned to                           │    │
│  │                         ▼                                │    │
│  │  ┌─────────────────────────────────────────────────┐    │    │
│  │  │                     ROLE                         │    │    │
│  │  │  (collection of permissions + inheritance)       │    │    │
│  │  └─────────────────────────────────────────────────┘    │    │
│  │                         │                                │    │
│  │                    assigned to                           │    │
│  │                         ▼                                │    │
│  │  ┌─────────────────────────────────────────────────┐    │    │
│  │  │                   SUBJECT                        │    │    │
│  │  │  (user, service account, API key)                │    │    │
│  │  └─────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### Key Principles

1. **Provider Agnostic**: Works with any identity provider (OAuth2, SAML, LDAP, custom JWT, API keys)
2. **Database Agnostic**: Repository interfaces allow any storage backend
3. **Multi-Tenant First**: Tenant isolation is built into every layer
4. **Hierarchical RBAC**: Support for role inheritance and permission cascading
5. **Resource-Based**: Permissions tied to specific resources with fine-grained actions
6. **Policy-Driven**: Optional attribute-based policies for complex access rules

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                              APPLICATION                                  │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐  │
│  │  HTTP Middleware │  │ gRPC Interceptor│  │    Direct Service Call   │  │
│  └────────┬────────┘  └────────┬────────┘  └────────────┬────────────┘  │
│           │                    │                         │               │
│           └────────────────────┼─────────────────────────┘               │
│                                ▼                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                         RBAC SERVICE                              │   │
│  │  ┌────────────────────────────────────────────────────────────┐  │   │
│  │  │                    Enforcement Engine                       │  │   │
│  │  │  - Permission checking                                      │  │   │
│  │  │  - Role resolution with inheritance                         │  │   │
│  │  │  - Multi-tenant context validation                          │  │   │
│  │  │  - Policy evaluation (ABAC extension)                       │  │   │
│  │  └────────────────────────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                │                                         │
│           ┌────────────────────┼────────────────────┐                   │
│           ▼                    ▼                    ▼                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │
│  │ Identity Provider│  │  Policy Provider │  │ Storage Provider │         │
│  │    Interface     │  │    Interface     │  │    Interface     │         │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘         │
│           │                    │                    │                   │
└───────────┼────────────────────┼────────────────────┼───────────────────┘
            │                    │                    │
            ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐
│  Implementations │  │  Implementations │  │      Implementations        │
│  - OAuth2/OIDC   │  │  - OPA/Rego     │  │  - PostgreSQL               │
│  - SAML          │  │  - Cedar        │  │  - MySQL                    │
│  - LDAP          │  │  - Custom       │  │  - MongoDB                  │
│  - Custom JWT    │  │  - In-memory    │  │  - Redis (cache)            │
│  - API Key       │  │                 │  │  - In-memory                │
└─────────────────┘  └─────────────────┘  └─────────────────────────────┘
```

---

## Domain Models

### Core Entities

```go
// pkg/rbac/models.go

// Tenant represents an isolated environment for multi-tenancy
type Tenant struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Slug        string            `json:"slug"`         // URL-friendly identifier
    Status      TenantStatus      `json:"status"`       // active, suspended, deleted
    Settings    TenantSettings    `json:"settings"`
    Metadata    map[string]string `json:"metadata"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type TenantSettings struct {
    MaxUsers            int           `json:"max_users"`
    MaxRoles            int           `json:"max_roles"`
    SessionTimeout      time.Duration `json:"session_timeout"`
    AllowedAuthMethods  []string      `json:"allowed_auth_methods"`
    RequireMFA          bool          `json:"require_mfa"`
    CustomPermissions   bool          `json:"custom_permissions"`  // Allow tenant-specific permissions
}

type TenantStatus string

const (
    TenantStatusActive    TenantStatus = "active"
    TenantStatusSuspended TenantStatus = "suspended"
    TenantStatusDeleted   TenantStatus = "deleted"
)

// Subject represents any entity that can be granted permissions
type Subject struct {
    ID         string            `json:"id"`
    TenantID   string            `json:"tenant_id"`
    Type       SubjectType       `json:"type"`
    ExternalID string            `json:"external_id"`   // ID from identity provider
    Email      string            `json:"email"`
    Name       string            `json:"name"`
    Status     SubjectStatus     `json:"status"`
    Attributes map[string]any    `json:"attributes"`    // For ABAC policies
    Metadata   map[string]string `json:"metadata"`
    CreatedAt  time.Time         `json:"created_at"`
    UpdatedAt  time.Time         `json:"updated_at"`
}

type SubjectType string

const (
    SubjectTypeUser           SubjectType = "user"
    SubjectTypeServiceAccount SubjectType = "service_account"
    SubjectTypeAPIKey         SubjectType = "api_key"
)

type SubjectStatus string

const (
    SubjectStatusActive    SubjectStatus = "active"
    SubjectStatusInactive  SubjectStatus = "inactive"
    SubjectStatusSuspended SubjectStatus = "suspended"
)

// Resource represents a protected resource type
type Resource struct {
    ID          string            `json:"id"`
    TenantID    string            `json:"tenant_id"`    // Empty for global resources
    Name        string            `json:"name"`         // e.g., "users", "orders", "documents"
    DisplayName string            `json:"display_name"`
    Description string            `json:"description"`
    Actions     []Action          `json:"actions"`      // Available actions on this resource
    IsGlobal    bool              `json:"is_global"`    // Available across all tenants
    Metadata    map[string]string `json:"metadata"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

// Action represents an operation that can be performed on a resource
type Action struct {
    Name        string `json:"name"`         // e.g., "read", "write", "delete", "admin"
    DisplayName string `json:"display_name"`
    Description string `json:"description"`
}

// Permission represents a specific access right
type Permission struct {
    ID          string            `json:"id"`
    TenantID    string            `json:"tenant_id"`    // Empty for global permissions
    ResourceID  string            `json:"resource_id"`
    Resource    string            `json:"resource"`     // Resource name (denormalized)
    Action      string            `json:"action"`       // Action name
    Scope       PermissionScope   `json:"scope"`        // Scope of the permission
    Conditions  []Condition       `json:"conditions"`   // Optional ABAC conditions
    Description string            `json:"description"`
    IsGlobal    bool              `json:"is_global"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

// String returns the permission string (e.g., "users:read", "orders:write")
func (p *Permission) String() string {
    return fmt.Sprintf("%s:%s", p.Resource, p.Action)
}

type PermissionScope string

const (
    PermissionScopeGlobal PermissionScope = "global"    // Access all instances
    PermissionScopeTenant PermissionScope = "tenant"    // Access within tenant
    PermissionScopeOwn    PermissionScope = "own"       // Access own resources only
)

// Condition for ABAC-style policies
type Condition struct {
    Attribute string      `json:"attribute"`  // e.g., "subject.department", "resource.owner_id"
    Operator  Operator    `json:"operator"`   // eq, neq, in, not_in, gt, lt, etc.
    Value     interface{} `json:"value"`
}

type Operator string

const (
    OperatorEquals         Operator = "eq"
    OperatorNotEquals      Operator = "neq"
    OperatorIn             Operator = "in"
    OperatorNotIn          Operator = "not_in"
    OperatorGreaterThan    Operator = "gt"
    OperatorLessThan       Operator = "lt"
    OperatorContains       Operator = "contains"
    OperatorStartsWith     Operator = "starts_with"
    OperatorMatches        Operator = "matches"  // Regex
)

// Role represents a collection of permissions with optional inheritance
type Role struct {
    ID           string            `json:"id"`
    TenantID     string            `json:"tenant_id"`    // Empty for global roles
    Name         string            `json:"name"`
    DisplayName  string            `json:"display_name"`
    Description  string            `json:"description"`
    Type         RoleType          `json:"type"`
    ParentID     string            `json:"parent_id"`    // For role inheritance
    Permissions  []string          `json:"permissions"`  // Permission IDs
    IsSystem     bool              `json:"is_system"`    // Cannot be modified/deleted
    IsGlobal     bool              `json:"is_global"`    // Available across all tenants
    Priority     int               `json:"priority"`     // For conflict resolution
    Metadata     map[string]string `json:"metadata"`
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
}

type RoleType string

const (
    RoleTypeSystem  RoleType = "system"   // Built-in, cannot be modified
    RoleTypeCustom  RoleType = "custom"   // Tenant-created
    RoleTypeDerived RoleType = "derived"  // Inherits from parent
)

// RoleAssignment links a subject to a role within a tenant context
type RoleAssignment struct {
    ID         string            `json:"id"`
    TenantID   string            `json:"tenant_id"`
    SubjectID  string            `json:"subject_id"`
    RoleID     string            `json:"role_id"`
    ResourceID string            `json:"resource_id"`  // Optional: scope to specific resource instance
    ExpiresAt  *time.Time        `json:"expires_at"`   // Optional: time-limited access
    GrantedBy  string            `json:"granted_by"`   // Subject who granted this
    Reason     string            `json:"reason"`       // Audit trail
    Metadata   map[string]string `json:"metadata"`
    CreatedAt  time.Time         `json:"created_at"`
}

// AccessContext contains all information needed for access decisions
type AccessContext struct {
    Subject     *Subject          `json:"subject"`
    TenantID    string            `json:"tenant_id"`
    Resource    string            `json:"resource"`
    ResourceID  string            `json:"resource_id"`   // Specific instance
    Action      string            `json:"action"`
    Environment map[string]any    `json:"environment"`   // IP, time, device, etc.
}

// AccessDecision represents the result of an access check
type AccessDecision struct {
    Allowed     bool              `json:"allowed"`
    Reason      string            `json:"reason"`
    MatchedRole string            `json:"matched_role"`      // Which role granted access
    MatchedPerm string            `json:"matched_permission"` // Which permission matched
    Conditions  []Condition       `json:"conditions"`        // Any conditions applied
    ExpiresAt   *time.Time        `json:"expires_at"`        // If access is time-limited
}
```

---

## Interface Definitions

### Core Service Interface

```go
// pkg/rbac/service.go

// Service defines the main RBAC service interface
type Service interface {
    // Access Control
    CheckAccess(ctx context.Context, req *AccessRequest) (*AccessDecision, error)
    CheckPermission(ctx context.Context, subjectID, tenantID, permission string) (bool, error)
    CheckRole(ctx context.Context, subjectID, tenantID, role string) (bool, error)

    // Batch Operations
    CheckPermissions(ctx context.Context, subjectID, tenantID string, permissions []string) (map[string]bool, error)
    FilterAuthorized(ctx context.Context, subjectID, tenantID string, resources []ResourceAction) ([]ResourceAction, error)

    // Subject Management
    GetSubjectPermissions(ctx context.Context, subjectID, tenantID string) ([]Permission, error)
    GetSubjectRoles(ctx context.Context, subjectID, tenantID string) ([]Role, error)
    GetEffectivePermissions(ctx context.Context, subjectID, tenantID string) ([]string, error)

    // Role Assignment
    AssignRole(ctx context.Context, assignment *RoleAssignment) error
    RevokeRole(ctx context.Context, subjectID, tenantID, roleID string) error

    // Cache Management
    InvalidateSubjectCache(ctx context.Context, subjectID string) error
    InvalidateTenantCache(ctx context.Context, tenantID string) error
}

// AccessRequest contains all information for an access check
type AccessRequest struct {
    SubjectID   string         `json:"subject_id"`
    TenantID    string         `json:"tenant_id"`
    Resource    string         `json:"resource"`
    ResourceID  string         `json:"resource_id"`   // Optional: specific instance
    Action      string         `json:"action"`
    Context     map[string]any `json:"context"`       // Additional ABAC attributes
}

type ResourceAction struct {
    Resource   string `json:"resource"`
    ResourceID string `json:"resource_id"`
    Action     string `json:"action"`
}
```

### Management Interface

```go
// pkg/rbac/management.go

// Management provides administrative operations for RBAC
type Management interface {
    // Tenant Management
    CreateTenant(ctx context.Context, tenant *Tenant) error
    GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
    UpdateTenant(ctx context.Context, tenant *Tenant) error
    DeleteTenant(ctx context.Context, tenantID string) error
    ListTenants(ctx context.Context, filter TenantFilter) (*TenantList, error)

    // Subject Management
    CreateSubject(ctx context.Context, subject *Subject) error
    GetSubject(ctx context.Context, subjectID string) (*Subject, error)
    GetSubjectByExternalID(ctx context.Context, tenantID, externalID string) (*Subject, error)
    UpdateSubject(ctx context.Context, subject *Subject) error
    DeleteSubject(ctx context.Context, subjectID string) error
    ListSubjects(ctx context.Context, tenantID string, filter SubjectFilter) (*SubjectList, error)

    // Role Management
    CreateRole(ctx context.Context, role *Role) error
    GetRole(ctx context.Context, roleID string) (*Role, error)
    UpdateRole(ctx context.Context, role *Role) error
    DeleteRole(ctx context.Context, roleID string) error
    ListRoles(ctx context.Context, tenantID string, filter RoleFilter) (*RoleList, error)

    // Permission Management
    CreatePermission(ctx context.Context, permission *Permission) error
    GetPermission(ctx context.Context, permissionID string) (*Permission, error)
    DeletePermission(ctx context.Context, permissionID string) error
    ListPermissions(ctx context.Context, tenantID string, filter PermissionFilter) (*PermissionList, error)

    // Resource Management
    RegisterResource(ctx context.Context, resource *Resource) error
    GetResource(ctx context.Context, resourceID string) (*Resource, error)
    ListResources(ctx context.Context, tenantID string) ([]*Resource, error)

    // Role Assignment Management
    ListRoleAssignments(ctx context.Context, filter AssignmentFilter) (*AssignmentList, error)
    BulkAssignRoles(ctx context.Context, assignments []*RoleAssignment) error
    BulkRevokeRoles(ctx context.Context, subjectID, tenantID string, roleIDs []string) error
}
```

---

## Multi-Tenancy Design

### Tenant Isolation Strategies

```go
// pkg/rbac/tenant.go

// TenantResolver extracts tenant context from various sources
type TenantResolver interface {
    // ResolveTenant extracts tenant ID from the request context
    ResolveTenant(ctx context.Context) (string, error)
}

// TenantResolverFunc is a function adapter for TenantResolver
type TenantResolverFunc func(ctx context.Context) (string, error)

func (f TenantResolverFunc) ResolveTenant(ctx context.Context) (string, error) {
    return f(ctx)
}

// Built-in tenant resolution strategies
type TenantResolutionStrategy string

const (
    // TenantFromHeader extracts tenant from X-Tenant-ID header
    TenantFromHeader TenantResolutionStrategy = "header"

    // TenantFromSubdomain extracts tenant from subdomain (tenant.example.com)
    TenantFromSubdomain TenantResolutionStrategy = "subdomain"

    // TenantFromPath extracts tenant from URL path (/tenants/{tenant_id}/...)
    TenantFromPath TenantResolutionStrategy = "path"

    // TenantFromJWT extracts tenant from JWT claims
    TenantFromJWT TenantResolutionStrategy = "jwt"

    // TenantFromSubject uses subject's default tenant
    TenantFromSubject TenantResolutionStrategy = "subject"
)

// NewTenantResolver creates a resolver based on strategy
func NewTenantResolver(strategy TenantResolutionStrategy, opts ...TenantResolverOption) TenantResolver {
    // Implementation based on strategy
}

// TenantContext is stored in context for downstream use
type TenantContext struct {
    TenantID string
    Tenant   *Tenant  // Loaded tenant info (optional, for performance)
}

// Context keys
type contextKey string

const (
    tenantContextKey  contextKey = "rbac_tenant"
    subjectContextKey contextKey = "rbac_subject"
    accessContextKey  contextKey = "rbac_access"
)

// TenantFromContext retrieves tenant context
func TenantFromContext(ctx context.Context) *TenantContext {
    if tc, ok := ctx.Value(tenantContextKey).(*TenantContext); ok {
        return tc
    }
    return nil
}

// ContextWithTenant adds tenant context
func ContextWithTenant(ctx context.Context, tc *TenantContext) context.Context {
    return context.WithValue(ctx, tenantContextKey, tc)
}
```

### Cross-Tenant Access Control

```go
// pkg/rbac/cross_tenant.go

// CrossTenantPolicy defines rules for cross-tenant access
type CrossTenantPolicy struct {
    // AllowCrossTenant enables cross-tenant access for super admins
    AllowCrossTenant bool `json:"allow_cross_tenant"`

    // SuperAdminRoles that can access any tenant
    SuperAdminRoles []string `json:"super_admin_roles"`

    // TenantGroups allow access between grouped tenants
    TenantGroups map[string][]string `json:"tenant_groups"`

    // DelegatedAccess rules for service-to-service calls
    DelegatedAccess []DelegatedAccessRule `json:"delegated_access"`
}

type DelegatedAccessRule struct {
    SourceTenant string   `json:"source_tenant"`
    TargetTenant string   `json:"target_tenant"`
    Permissions  []string `json:"permissions"`
    ExpiresAt    *time.Time `json:"expires_at"`
}
```

---

## Provider Abstraction

### Identity Provider Interface

```go
// pkg/rbac/provider/identity.go

// IdentityProvider abstracts the source of identity/authentication
type IdentityProvider interface {
    // ValidateToken validates an authentication token and returns subject info
    ValidateToken(ctx context.Context, token string) (*IdentityInfo, error)

    // GetProviderID returns unique identifier for this provider
    GetProviderID() string

    // GetProviderType returns the type (oauth2, saml, ldap, etc.)
    GetProviderType() ProviderType
}

type ProviderType string

const (
    ProviderTypeOAuth2  ProviderType = "oauth2"
    ProviderTypeOIDC    ProviderType = "oidc"
    ProviderTypeSAML    ProviderType = "saml"
    ProviderTypeLDAP    ProviderType = "ldap"
    ProviderTypeAPIKey  ProviderType = "api_key"
    ProviderTypeJWT     ProviderType = "jwt"
    ProviderTypeCustom  ProviderType = "custom"
)

// IdentityInfo contains identity information extracted from auth token
type IdentityInfo struct {
    ProviderID   string            `json:"provider_id"`
    ProviderType ProviderType      `json:"provider_type"`
    ExternalID   string            `json:"external_id"`   // ID from the provider
    Email        string            `json:"email"`
    Name         string            `json:"name"`
    Groups       []string          `json:"groups"`        // Provider groups (for group mapping)
    Claims       map[string]any    `json:"claims"`        // Raw claims/attributes
    ExpiresAt    *time.Time        `json:"expires_at"`
}

// IdentityProviderRegistry manages multiple identity providers
type IdentityProviderRegistry interface {
    // Register adds a new identity provider
    Register(provider IdentityProvider) error

    // Get retrieves a provider by ID
    Get(providerID string) (IdentityProvider, error)

    // List returns all registered providers
    List() []IdentityProvider

    // ValidateToken tries all providers until one succeeds
    ValidateToken(ctx context.Context, token string) (*IdentityInfo, IdentityProvider, error)
}
```

### Provider Implementations (Interfaces)

```go
// pkg/rbac/provider/oauth2.go

// OAuth2ProviderConfig configuration for OAuth2/OIDC providers
type OAuth2ProviderConfig struct {
    ProviderID     string   `json:"provider_id"`
    ClientID       string   `json:"client_id"`
    ClientSecret   string   `json:"client_secret"`
    IssuerURL      string   `json:"issuer_url"`       // OIDC discovery URL
    JWKSURL        string   `json:"jwks_url"`         // If not using discovery
    Audience       string   `json:"audience"`
    Scopes         []string `json:"scopes"`
    ClaimMappings  ClaimMappings `json:"claim_mappings"`
    GroupMappings  map[string]string `json:"group_mappings"` // Provider group -> role
}

type ClaimMappings struct {
    SubjectClaim string `json:"subject_claim"` // Default: "sub"
    EmailClaim   string `json:"email_claim"`   // Default: "email"
    NameClaim    string `json:"name_claim"`    // Default: "name"
    GroupsClaim  string `json:"groups_claim"`  // Default: "groups"
}

// pkg/rbac/provider/saml.go

// SAMLProviderConfig configuration for SAML providers
type SAMLProviderConfig struct {
    ProviderID       string `json:"provider_id"`
    EntityID         string `json:"entity_id"`
    MetadataURL      string `json:"metadata_url"`
    ACSURL           string `json:"acs_url"`
    Certificate      string `json:"certificate"`
    AttributeMappings map[string]string `json:"attribute_mappings"`
}

// pkg/rbac/provider/ldap.go

// LDAPProviderConfig configuration for LDAP/AD providers
type LDAPProviderConfig struct {
    ProviderID     string   `json:"provider_id"`
    Host           string   `json:"host"`
    Port           int      `json:"port"`
    UseTLS         bool     `json:"use_tls"`
    BindDN         string   `json:"bind_dn"`
    BindPassword   string   `json:"bind_password"`
    BaseDN         string   `json:"base_dn"`
    UserFilter     string   `json:"user_filter"`
    GroupFilter    string   `json:"group_filter"`
    AttributeMappings map[string]string `json:"attribute_mappings"`
}

// pkg/rbac/provider/apikey.go

// APIKeyProvider validates API keys
type APIKeyProvider interface {
    IdentityProvider

    // CreateAPIKey creates a new API key for a subject
    CreateAPIKey(ctx context.Context, subjectID string, opts APIKeyOptions) (*APIKey, error)

    // RevokeAPIKey revokes an API key
    RevokeAPIKey(ctx context.Context, keyID string) error

    // ListAPIKeys lists API keys for a subject
    ListAPIKeys(ctx context.Context, subjectID string) ([]*APIKey, error)
}

type APIKey struct {
    ID          string    `json:"id"`
    SubjectID   string    `json:"subject_id"`
    TenantID    string    `json:"tenant_id"`
    Name        string    `json:"name"`
    KeyPrefix   string    `json:"key_prefix"`  // First 8 chars for identification
    Scopes      []string  `json:"scopes"`      // Limited permissions
    ExpiresAt   *time.Time `json:"expires_at"`
    LastUsedAt  *time.Time `json:"last_used_at"`
    CreatedAt   time.Time `json:"created_at"`
}

type APIKeyOptions struct {
    Name      string
    Scopes    []string
    ExpiresIn time.Duration
}
```

---

## Storage Abstraction

### Repository Interfaces

```go
// pkg/rbac/repository.go

// Repositories bundles all repository interfaces
type Repositories struct {
    Tenants         TenantRepository
    Subjects        SubjectRepository
    Roles           RoleRepository
    Permissions     PermissionRepository
    Resources       ResourceRepository
    RoleAssignments RoleAssignmentRepository
    AuditLogs       AuditLogRepository       // Optional
    Cache           CacheRepository          // Optional
}

// TenantRepository handles tenant persistence
type TenantRepository interface {
    Create(ctx context.Context, tenant *Tenant) error
    GetByID(ctx context.Context, id string) (*Tenant, error)
    GetBySlug(ctx context.Context, slug string) (*Tenant, error)
    Update(ctx context.Context, tenant *Tenant) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter TenantFilter, pagination Pagination) ([]*Tenant, int64, error)
}

// SubjectRepository handles subject persistence
type SubjectRepository interface {
    Create(ctx context.Context, subject *Subject) error
    GetByID(ctx context.Context, id string) (*Subject, error)
    GetByExternalID(ctx context.Context, tenantID, externalID string) (*Subject, error)
    GetByEmail(ctx context.Context, tenantID, email string) (*Subject, error)
    Update(ctx context.Context, subject *Subject) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, tenantID string, filter SubjectFilter, pagination Pagination) ([]*Subject, int64, error)
}

// RoleRepository handles role persistence
type RoleRepository interface {
    Create(ctx context.Context, role *Role) error
    GetByID(ctx context.Context, id string) (*Role, error)
    GetByName(ctx context.Context, tenantID, name string) (*Role, error)
    Update(ctx context.Context, role *Role) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, tenantID string, filter RoleFilter) ([]*Role, error)
    GetWithParents(ctx context.Context, id string) ([]*Role, error) // For inheritance
}

// PermissionRepository handles permission persistence
type PermissionRepository interface {
    Create(ctx context.Context, permission *Permission) error
    GetByID(ctx context.Context, id string) (*Permission, error)
    GetByResourceAction(ctx context.Context, tenantID, resource, action string) (*Permission, error)
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, tenantID string, filter PermissionFilter) ([]*Permission, error)
    ListByRole(ctx context.Context, roleID string) ([]*Permission, error)
}

// ResourceRepository handles resource registration
type ResourceRepository interface {
    Create(ctx context.Context, resource *Resource) error
    GetByID(ctx context.Context, id string) (*Resource, error)
    GetByName(ctx context.Context, tenantID, name string) (*Resource, error)
    Update(ctx context.Context, resource *Resource) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, tenantID string) ([]*Resource, error)
}

// RoleAssignmentRepository handles role assignments
type RoleAssignmentRepository interface {
    Create(ctx context.Context, assignment *RoleAssignment) error
    Delete(ctx context.Context, id string) error
    DeleteBySubjectRole(ctx context.Context, subjectID, tenantID, roleID string) error
    GetBySubject(ctx context.Context, subjectID, tenantID string) ([]*RoleAssignment, error)
    GetByRole(ctx context.Context, roleID string) ([]*RoleAssignment, error)
    List(ctx context.Context, filter AssignmentFilter, pagination Pagination) ([]*RoleAssignment, int64, error)
    CleanupExpired(ctx context.Context) (int64, error)
}

// CacheRepository provides caching for RBAC data
type CacheRepository interface {
    // Subject permissions cache
    GetSubjectPermissions(ctx context.Context, subjectID, tenantID string) ([]string, error)
    SetSubjectPermissions(ctx context.Context, subjectID, tenantID string, permissions []string, ttl time.Duration) error
    InvalidateSubject(ctx context.Context, subjectID string) error

    // Role cache
    GetRole(ctx context.Context, roleID string) (*Role, error)
    SetRole(ctx context.Context, role *Role, ttl time.Duration) error
    InvalidateRole(ctx context.Context, roleID string) error

    // Tenant cache
    GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
    SetTenant(ctx context.Context, tenant *Tenant, ttl time.Duration) error
    InvalidateTenant(ctx context.Context, tenantID string) error

    // Bulk invalidation
    InvalidateAll(ctx context.Context) error
}

// AuditLogRepository for compliance and debugging
type AuditLogRepository interface {
    Log(ctx context.Context, entry *AuditEntry) error
    Query(ctx context.Context, filter AuditFilter, pagination Pagination) ([]*AuditEntry, int64, error)
}

type AuditEntry struct {
    ID          string         `json:"id"`
    Timestamp   time.Time      `json:"timestamp"`
    TenantID    string         `json:"tenant_id"`
    SubjectID   string         `json:"subject_id"`
    Action      string         `json:"action"`       // check_access, assign_role, etc.
    Resource    string         `json:"resource"`
    ResourceID  string         `json:"resource_id"`
    Decision    string         `json:"decision"`     // allowed, denied
    Reason      string         `json:"reason"`
    Context     map[string]any `json:"context"`
    IPAddress   string         `json:"ip_address"`
    UserAgent   string         `json:"user_agent"`
}
```

### Filter and Pagination Types

```go
// pkg/rbac/filter.go

type Pagination struct {
    Page     int `json:"page"`
    PageSize int `json:"page_size"`
}

type TenantFilter struct {
    Status    *TenantStatus `json:"status"`
    Search    string        `json:"search"`
    CreatedAfter *time.Time `json:"created_after"`
}

type SubjectFilter struct {
    Type       *SubjectType   `json:"type"`
    Status     *SubjectStatus `json:"status"`
    Search     string         `json:"search"`
    HasRole    string         `json:"has_role"`
}

type RoleFilter struct {
    Type       *RoleType `json:"type"`
    IsSystem   *bool     `json:"is_system"`
    IsGlobal   *bool     `json:"is_global"`
    Search     string    `json:"search"`
}

type PermissionFilter struct {
    Resource   string           `json:"resource"`
    Action     string           `json:"action"`
    Scope      *PermissionScope `json:"scope"`
    IsGlobal   *bool            `json:"is_global"`
}

type AssignmentFilter struct {
    SubjectID  string     `json:"subject_id"`
    RoleID     string     `json:"role_id"`
    TenantID   string     `json:"tenant_id"`
    ResourceID string     `json:"resource_id"`
    ExpiresBefore *time.Time `json:"expires_before"`
}

type AuditFilter struct {
    TenantID   string     `json:"tenant_id"`
    SubjectID  string     `json:"subject_id"`
    Action     string     `json:"action"`
    Resource   string     `json:"resource"`
    Decision   string     `json:"decision"`
    StartTime  *time.Time `json:"start_time"`
    EndTime    *time.Time `json:"end_time"`
}
```

---

## Middleware & Interceptors

### HTTP Middleware

```go
// pkg/rbac/middleware/http.go

// Middleware creates the base RBAC middleware that:
// 1. Resolves tenant context
// 2. Validates authentication
// 3. Loads subject into context
func Middleware(svc Service, opts ...MiddlewareOption) func(http.Handler) http.Handler {
    cfg := defaultMiddlewareConfig()
    for _, opt := range opts {
        opt(cfg)
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Implementation
        })
    }
}

// RequirePermission middleware checks for specific permission
func RequirePermission(svc Service, permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            subject := SubjectFromContext(r.Context())
            tenant := TenantFromContext(r.Context())

            if subject == nil || tenant == nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            allowed, err := svc.CheckPermission(r.Context(), subject.ID, tenant.TenantID, permission)
            if err != nil || !allowed {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// RequireRole middleware checks for specific role
func RequireRole(svc Service, roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            subject := SubjectFromContext(r.Context())
            tenant := TenantFromContext(r.Context())

            if subject == nil || tenant == nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            for _, role := range roles {
                hasRole, err := svc.CheckRole(r.Context(), subject.ID, tenant.TenantID, role)
                if err == nil && hasRole {
                    next.ServeHTTP(w, r)
                    return
                }
            }

            http.Error(w, "forbidden", http.StatusForbidden)
        })
    }
}

// RequireResourcePermission checks permission for specific resource instance
func RequireResourcePermission(svc Service, resource, action string, resourceIDFunc func(*http.Request) string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            subject := SubjectFromContext(r.Context())
            tenant := TenantFromContext(r.Context())
            resourceID := resourceIDFunc(r)

            decision, err := svc.CheckAccess(r.Context(), &AccessRequest{
                SubjectID:  subject.ID,
                TenantID:   tenant.TenantID,
                Resource:   resource,
                ResourceID: resourceID,
                Action:     action,
            })

            if err != nil || !decision.Allowed {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Middleware options
type MiddlewareConfig struct {
    TenantResolver    TenantResolver
    IdentityRegistry  IdentityProviderRegistry
    TokenExtractor    TokenExtractor
    ErrorHandler      ErrorHandler
    SkipPaths         []string
    AuditEnabled      bool
}

type TokenExtractor func(*http.Request) (string, error)
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// Built-in token extractors
func BearerTokenExtractor() TokenExtractor {
    return func(r *http.Request) (string, error) {
        auth := r.Header.Get("Authorization")
        if !strings.HasPrefix(auth, "Bearer ") {
            return "", ErrNoToken
        }
        return strings.TrimPrefix(auth, "Bearer "), nil
    }
}

func APIKeyExtractor(headerName string) TokenExtractor {
    return func(r *http.Request) (string, error) {
        key := r.Header.Get(headerName)
        if key == "" {
            return "", ErrNoToken
        }
        return key, nil
    }
}

func CookieExtractor(cookieName string) TokenExtractor {
    return func(r *http.Request) (string, error) {
        cookie, err := r.Cookie(cookieName)
        if err != nil {
            return "", ErrNoToken
        }
        return cookie.Value, nil
    }
}

func ChainExtractor(extractors ...TokenExtractor) TokenExtractor {
    return func(r *http.Request) (string, error) {
        for _, ext := range extractors {
            token, err := ext(r)
            if err == nil && token != "" {
                return token, nil
            }
        }
        return "", ErrNoToken
    }
}
```

### gRPC Interceptors

```go
// pkg/rbac/middleware/grpc.go

// UnaryAuthInterceptor validates authentication for unary calls
func UnaryAuthInterceptor(svc Service, registry IdentityProviderRegistry, resolver TenantResolver) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // Extract token from metadata
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, status.Error(codes.Unauthenticated, "missing metadata")
        }

        tokens := md.Get("authorization")
        if len(tokens) == 0 {
            return nil, status.Error(codes.Unauthenticated, "missing authorization")
        }

        token := strings.TrimPrefix(tokens[0], "Bearer ")

        // Validate token
        identity, provider, err := registry.ValidateToken(ctx, token)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "invalid token")
        }

        // Resolve tenant
        tenantID, err := resolver.ResolveTenant(ctx)
        if err != nil {
            return nil, status.Error(codes.InvalidArgument, "cannot resolve tenant")
        }

        // Load or create subject
        subject, err := svc.GetOrCreateSubject(ctx, tenantID, identity)
        if err != nil {
            return nil, status.Error(codes.Internal, "failed to load subject")
        }

        // Add to context
        ctx = ContextWithSubject(ctx, subject)
        ctx = ContextWithTenant(ctx, &TenantContext{TenantID: tenantID})

        return handler(ctx, req)
    }
}

// UnaryPermissionInterceptor checks permission for each method
func UnaryPermissionInterceptor(svc Service, methodPermissions map[string]string) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        permission, ok := methodPermissions[info.FullMethod]
        if !ok {
            // No permission required for this method
            return handler(ctx, req)
        }

        subject := SubjectFromContext(ctx)
        tenant := TenantFromContext(ctx)

        if subject == nil || tenant == nil {
            return nil, status.Error(codes.Unauthenticated, "not authenticated")
        }

        allowed, err := svc.CheckPermission(ctx, subject.ID, tenant.TenantID, permission)
        if err != nil {
            return nil, status.Error(codes.Internal, "permission check failed")
        }

        if !allowed {
            return nil, status.Error(codes.PermissionDenied, "permission denied")
        }

        return handler(ctx, req)
    }
}

// UnaryRoleInterceptor checks role for each method
func UnaryRoleInterceptor(svc Service, methodRoles map[string][]string) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        roles, ok := methodRoles[info.FullMethod]
        if !ok {
            return handler(ctx, req)
        }

        subject := SubjectFromContext(ctx)
        tenant := TenantFromContext(ctx)

        for _, role := range roles {
            hasRole, err := svc.CheckRole(ctx, subject.ID, tenant.TenantID, role)
            if err == nil && hasRole {
                return handler(ctx, req)
            }
        }

        return nil, status.Error(codes.PermissionDenied, "insufficient role")
    }
}

// Stream interceptors follow similar pattern
func StreamAuthInterceptor(svc Service, registry IdentityProviderRegistry, resolver TenantResolver) grpc.StreamServerInterceptor {
    // Similar implementation for streaming RPCs
}
```

---

## Policy Engine

### Policy Interface (for ABAC extension)

```go
// pkg/rbac/policy/policy.go

// PolicyEngine evaluates complex access policies
type PolicyEngine interface {
    // Evaluate checks if access should be granted based on policies
    Evaluate(ctx context.Context, request *PolicyRequest) (*PolicyDecision, error)

    // LoadPolicies loads policies from a source
    LoadPolicies(ctx context.Context, source PolicySource) error

    // AddPolicy adds a single policy
    AddPolicy(ctx context.Context, policy *Policy) error

    // RemovePolicy removes a policy
    RemovePolicy(ctx context.Context, policyID string) error
}

type PolicyRequest struct {
    Subject    map[string]any `json:"subject"`    // Subject attributes
    Resource   map[string]any `json:"resource"`   // Resource attributes
    Action     string         `json:"action"`
    Context    map[string]any `json:"context"`    // Environment attributes
}

type PolicyDecision struct {
    Effect     PolicyEffect `json:"effect"`
    Reason     string       `json:"reason"`
    PolicyID   string       `json:"policy_id"`
    Obligations []Obligation `json:"obligations"` // Post-decision actions
}

type PolicyEffect string

const (
    PolicyEffectAllow PolicyEffect = "allow"
    PolicyEffectDeny  PolicyEffect = "deny"
)

type Obligation struct {
    Type   string         `json:"type"`
    Params map[string]any `json:"params"`
}

// Policy defines an access control policy
type Policy struct {
    ID          string           `json:"id"`
    TenantID    string           `json:"tenant_id"`
    Name        string           `json:"name"`
    Description string           `json:"description"`
    Effect      PolicyEffect     `json:"effect"`
    Subjects    []PolicyMatcher  `json:"subjects"`
    Resources   []PolicyMatcher  `json:"resources"`
    Actions     []string         `json:"actions"`
    Conditions  []PolicyCondition `json:"conditions"`
    Priority    int              `json:"priority"`
}

type PolicyMatcher struct {
    Type   string `json:"type"`   // any, match, equals
    Value  string `json:"value"`
    Field  string `json:"field"`  // For attribute matching
}

type PolicyCondition struct {
    Expression string `json:"expression"` // CEL or custom expression
}

// PolicySource for loading policies
type PolicySource interface {
    Load(ctx context.Context) ([]*Policy, error)
}

// Built-in policy sources
type FilePolicySource struct {
    Path string
}

type DatabasePolicySource struct {
    Repository PolicyRepository
}
```

---

## Implementation Plan

### Package Structure

```
pkg/rbac/
├── rbac.go              # Package entry point, NewService()
├── service.go           # Service interface
├── service_impl.go      # Service implementation
├── management.go        # Management interface
├── management_impl.go   # Management implementation
├── config.go            # Configuration with LoadConfig()
├── errors.go            # Custom error types
├── models.go            # Domain models (Tenant, Subject, Role, etc.)
├── repository.go        # Repository interfaces
├── filter.go            # Filter and pagination types
├── context.go           # Context helpers
├── tenant.go            # Tenant resolution
├── options.go           # Functional options
│
├── provider/            # Identity provider abstractions
│   ├── provider.go      # IdentityProvider interface
│   ├── registry.go      # IdentityProviderRegistry
│   ├── oauth2.go        # OAuth2/OIDC implementation
│   ├── saml.go          # SAML implementation (interface)
│   ├── ldap.go          # LDAP implementation (interface)
│   ├── apikey.go        # API key implementation
│   └── jwt.go           # Custom JWT implementation
│
├── middleware/          # HTTP and gRPC middleware
│   ├── http.go          # HTTP middleware
│   ├── grpc.go          # gRPC interceptors
│   ├── extractors.go    # Token extractors
│   └── handlers.go      # Error handlers
│
├── policy/              # Policy engine (ABAC)
│   ├── policy.go        # Policy interface
│   ├── engine.go        # Default policy engine
│   ├── cel.go           # CEL expression evaluator
│   └── sources.go       # Policy sources
│
├── cache/               # Caching implementations
│   ├── cache.go         # Cache interface
│   ├── memory.go        # In-memory cache
│   └── redis.go         # Redis cache (interface)
│
├── testutil/            # Testing utilities
│   ├── mocks.go         # Mock implementations
│   └── fixtures.go      # Test fixtures
│
└── examples/            # Usage examples
    ├── basic/           # Basic RBAC usage
    ├── multi_tenant/    # Multi-tenant setup
    ├── oauth2/          # OAuth2 integration
    └── custom_policy/   # Custom policy engine
```

### Implementation Phases

#### Phase 1: Core Foundation
1. Domain models (`models.go`)
2. Repository interfaces (`repository.go`)
3. Configuration (`config.go`)
4. Error types (`errors.go`)
5. Context helpers (`context.go`)

#### Phase 2: Service Layer
1. Service interface (`service.go`)
2. Service implementation (`service_impl.go`)
3. Management interface (`management.go`)
4. Management implementation (`management_impl.go`)

#### Phase 3: Provider Layer
1. Identity provider interface (`provider/provider.go`)
2. Provider registry (`provider/registry.go`)
3. JWT provider implementation (`provider/jwt.go`)
4. API key provider (`provider/apikey.go`)

#### Phase 4: Middleware Layer
1. HTTP middleware (`middleware/http.go`)
2. gRPC interceptors (`middleware/grpc.go`)
3. Token extractors (`middleware/extractors.go`)

#### Phase 5: Multi-Tenancy
1. Tenant resolver (`tenant.go`)
2. Cross-tenant policies
3. Tenant-scoped operations

#### Phase 6: Policy Engine (Optional)
1. Policy interface (`policy/policy.go`)
2. Default engine (`policy/engine.go`)
3. CEL evaluator (`policy/cel.go`)

#### Phase 7: Caching
1. Cache interface (`cache/cache.go`)
2. In-memory implementation (`cache/memory.go`)

#### Phase 8: Testing & Examples
1. Mock implementations (`testutil/mocks.go`)
2. Example applications (`examples/`)

---

## Usage Examples

### Basic Setup

```go
package main

import (
    "github.com/org/core-backend/pkg/rbac"
    "github.com/org/core-backend/pkg/rbac/provider"
    "github.com/org/core-backend/pkg/rbac/middleware"
)

func main() {
    // Load configuration
    cfg, err := rbac.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Create repositories (implement for your database)
    repos := rbac.Repositories{
        Tenants:         myPostgresTenantRepo,
        Subjects:        myPostgresSubjectRepo,
        Roles:           myPostgresRoleRepo,
        Permissions:     myPostgresPermissionRepo,
        Resources:       myPostgresResourceRepo,
        RoleAssignments: myPostgresAssignmentRepo,
        Cache:           myRedisCache,  // Optional
        AuditLogs:       myAuditRepo,   // Optional
    }

    // Create identity provider registry
    identityRegistry := provider.NewRegistry()

    // Register OAuth2 provider (e.g., Auth0, Okta)
    oauth2Provider, err := provider.NewOAuth2Provider(provider.OAuth2ProviderConfig{
        ProviderID:   "auth0",
        ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
        ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
        IssuerURL:    "https://your-tenant.auth0.com/",
    })
    if err != nil {
        log.Fatal(err)
    }
    identityRegistry.Register(oauth2Provider)

    // Register API key provider
    apiKeyProvider := provider.NewAPIKeyProvider(repos.Subjects)
    identityRegistry.Register(apiKeyProvider)

    // Create RBAC service
    svc, err := rbac.NewService(cfg, repos,
        rbac.WithIdentityRegistry(identityRegistry),
        rbac.WithLogger(myLogger),
        rbac.WithCacheTTL(5*time.Minute),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create tenant resolver
    tenantResolver := rbac.NewTenantResolver(rbac.TenantFromHeader)

    // Setup HTTP router with middleware
    mux := http.NewServeMux()

    // Apply base RBAC middleware
    handler := middleware.Middleware(svc,
        middleware.WithTenantResolver(tenantResolver),
        middleware.WithIdentityRegistry(identityRegistry),
        middleware.WithTokenExtractor(middleware.ChainExtractor(
            middleware.BearerTokenExtractor(),
            middleware.APIKeyExtractor("X-API-Key"),
        )),
    )(mux)

    // Protected endpoint with permission check
    mux.Handle("/api/users",
        middleware.RequirePermission(svc, "users:read")(
            http.HandlerFunc(listUsersHandler),
        ),
    )

    // Protected endpoint with role check
    mux.Handle("/api/admin",
        middleware.RequireRole(svc, "admin", "super_admin")(
            http.HandlerFunc(adminHandler),
        ),
    )

    http.ListenAndServe(":8080", handler)
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
    subject := rbac.SubjectFromContext(r.Context())
    tenant := rbac.TenantFromContext(r.Context())

    // subject and tenant are guaranteed to be non-nil here
    fmt.Fprintf(w, "Hello %s from tenant %s", subject.Email, tenant.TenantID)
}
```

### Multi-Tenant with Resource-Level Permissions

```go
// Check access to specific document
decision, err := svc.CheckAccess(ctx, &rbac.AccessRequest{
    SubjectID:  subject.ID,
    TenantID:   tenant.ID,
    Resource:   "documents",
    ResourceID: "doc-123",
    Action:     "edit",
    Context: map[string]any{
        "ip_address": r.RemoteAddr,
        "user_agent": r.UserAgent(),
    },
})

if !decision.Allowed {
    http.Error(w, decision.Reason, http.StatusForbidden)
    return
}

// Proceed with edit operation
```

### Programmatic Role Management

```go
// Create a custom role for a tenant
role := &rbac.Role{
    TenantID:    tenantID,
    Name:        "document_editor",
    DisplayName: "Document Editor",
    Description: "Can view and edit documents",
    Type:        rbac.RoleTypeCustom,
    Permissions: []string{
        "documents:read",
        "documents:write",
        "documents:delete",
    },
}

err := management.CreateRole(ctx, role)

// Assign role to user
assignment := &rbac.RoleAssignment{
    TenantID:  tenantID,
    SubjectID: userID,
    RoleID:    role.ID,
    GrantedBy: adminID,
    Reason:    "Promoted to editor",
}

err = svc.AssignRole(ctx, assignment)
```

### Integration with Existing Auth Package

```go
// Adapter to integrate with pkg/auth
type AuthServiceAdapter struct {
    authSvc auth.Service
}

func (a *AuthServiceAdapter) ValidateToken(ctx context.Context, token string) (*rbac.IdentityInfo, error) {
    // Validate using existing auth service
    user, err := a.authSvc.ValidateSession(ctx, token)
    if err != nil {
        return nil, err
    }

    return &rbac.IdentityInfo{
        ProviderID:   "internal",
        ProviderType: rbac.ProviderTypeJWT,
        ExternalID:   user.ID,
        Email:        user.Email,
        Name:         user.Name,
    }, nil
}

// Use the adapter
adapter := &AuthServiceAdapter{authSvc: existingAuthService}
identityRegistry.Register(provider.NewCustomProvider("internal", adapter))
```

---

## Design Decisions & Trade-offs

### 1. Subject vs User Separation
**Decision**: Introduced `Subject` as a separate entity from identity provider's user.

**Rationale**: Allows the same external user to have different permissions across tenants, and supports service accounts and API keys uniformly.

### 2. Role Inheritance
**Decision**: Single-parent inheritance with `ParentID`.

**Trade-off**: Simpler than multiple inheritance but may require role duplication in some cases. Priority field helps resolve conflicts.

### 3. Permission Scope
**Decision**: Three levels - Global, Tenant, Own.

**Rationale**: Covers most common access patterns. "Own" scope is crucial for user-generated content.

### 4. Policy Engine as Optional
**Decision**: Core RBAC works without policy engine; ABAC is an extension.

**Rationale**: Keeps the core simple. Complex policies can be added when needed without impacting basic usage.

### 5. Cache Repository
**Decision**: Optional cache repository with invalidation methods.

**Rationale**: Caching is crucial for performance but implementation varies (Redis, memcached, in-memory). Interface allows flexibility.

---

## Security Considerations

1. **Tenant Isolation**: All queries must include tenant ID. Use database-level row security where possible.

2. **Permission Caching**: Cache invalidation is critical. When roles change, invalidate affected subjects immediately.

3. **Audit Logging**: Log all access decisions for compliance. Include enough context for forensics.

4. **Rate Limiting**: Apply rate limits to permission checks to prevent abuse.

5. **Token Validation**: Always validate tokens server-side. Don't trust client-provided tenant IDs without verification.

6. **Least Privilege**: Default deny. Only grant explicit permissions.

---

## Next Steps

1. Review and approve this design
2. Create the package structure
3. Implement Phase 1 (Core Foundation)
4. Create in-memory repository implementations for testing
5. Implement service layer
6. Add middleware
7. Create PostgreSQL repository implementations
8. Write comprehensive tests
9. Document with examples
