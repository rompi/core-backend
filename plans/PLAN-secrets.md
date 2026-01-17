# Package Plan: pkg/secrets

## Overview

A unified secrets management package supporting multiple backends (HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager, Azure Key Vault). Provides secure retrieval, caching, automatic rotation handling, and integration with pkg/config.

## Goals

1. **Multiple Backends** - Vault, AWS, GCP, Azure, local file
2. **Unified Interface** - Single API for all secret stores
3. **Caching** - In-memory caching with TTL
4. **Rotation Support** - Handle secret rotation gracefully
5. **Version Support** - Access specific secret versions
6. **Encryption** - Local encryption for file-based secrets
7. **Zero Secrets in Logs** - Never expose secrets in logs/errors

## Architecture

```
pkg/secrets/
├── secrets.go            # Core Secrets interface
├── config.go             # Configuration
├── options.go            # Functional options
├── errors.go             # Custom error types
├── cache.go              # Secret caching
├── provider/
│   ├── provider.go       # Provider interface
│   ├── vault.go          # HashiCorp Vault
│   ├── aws.go            # AWS Secrets Manager
│   ├── gcp.go            # GCP Secret Manager
│   ├── azure.go          # Azure Key Vault
│   ├── env.go            # Environment variables
│   └── file.go           # Encrypted local file
├── crypto/
│   ├── encrypt.go        # Encryption utilities
│   └── kdf.go            # Key derivation
├── examples/
│   ├── basic/
│   ├── vault/
│   ├── aws/
│   └── rotation/
└── README.md
```

## Core Interfaces

```go
package secrets

import (
    "context"
    "time"
)

// Manager manages secret retrieval and caching
type Manager interface {
    // Get retrieves a secret by path/name
    Get(ctx context.Context, path string) (*Secret, error)

    // GetString retrieves a secret value as string
    GetString(ctx context.Context, path string) (string, error)

    // GetVersion retrieves a specific version of a secret
    GetVersion(ctx context.Context, path string, version string) (*Secret, error)

    // List lists secrets at a path
    List(ctx context.Context, path string) ([]string, error)

    // Put stores a secret (if supported by provider)
    Put(ctx context.Context, path string, value map[string]interface{}) error

    // Delete deletes a secret (if supported by provider)
    Delete(ctx context.Context, path string) error

    // Watch watches for secret changes
    Watch(ctx context.Context, path string, callback func(*Secret)) error

    // Refresh forces a cache refresh for a path
    Refresh(ctx context.Context, path string) error

    // Close releases resources
    Close() error
}

// Secret represents a retrieved secret
type Secret struct {
    // Path is the secret path/name
    Path string

    // Data holds the secret key-value pairs
    Data map[string]interface{}

    // Version is the secret version
    Version string

    // CreatedAt is when the secret was created
    CreatedAt time.Time

    // ExpiresAt is when the secret expires (if applicable)
    ExpiresAt *time.Time

    // Metadata holds additional metadata
    Metadata map[string]string
}

// Value returns a string value from the secret
func (s *Secret) Value(key string) string

// ValueBytes returns a byte slice value
func (s *Secret) ValueBytes(key string) []byte

// Has checks if a key exists
func (s *Secret) Has(key string) bool

// Provider provides secrets from a backend
type Provider interface {
    // Name returns the provider name
    Name() string

    // Get retrieves a secret
    Get(ctx context.Context, path string, opts ...GetOption) (*Secret, error)

    // List lists secrets at a path
    List(ctx context.Context, path string) ([]string, error)

    // Put stores a secret (optional)
    Put(ctx context.Context, path string, value map[string]interface{}) error

    // Delete deletes a secret (optional)
    Delete(ctx context.Context, path string) error

    // Watch watches for changes (optional)
    Watch(ctx context.Context, path string, callback func(*Secret)) error

    // Close releases resources
    Close() error
}
```

## Configuration

```go
// Config holds secrets manager configuration
type Config struct {
    // Provider type: "vault", "aws", "gcp", "azure", "env", "file"
    Provider string `env:"SECRETS_PROVIDER" default:"env"`

    // Cache configuration
    Cache CacheConfig

    // Retry configuration
    Retry RetryConfig
}

type CacheConfig struct {
    // Enable caching
    Enabled bool `env:"SECRETS_CACHE_ENABLED" default:"true"`

    // Default TTL for cached secrets
    TTL time.Duration `env:"SECRETS_CACHE_TTL" default:"5m"`

    // Maximum cached entries
    MaxEntries int `env:"SECRETS_CACHE_MAX_ENTRIES" default:"1000"`
}

type RetryConfig struct {
    // Maximum retry attempts
    MaxAttempts int `env:"SECRETS_RETRY_MAX_ATTEMPTS" default:"3"`

    // Initial delay
    InitialDelay time.Duration `env:"SECRETS_RETRY_INITIAL_DELAY" default:"100ms"`

    // Maximum delay
    MaxDelay time.Duration `env:"SECRETS_RETRY_MAX_DELAY" default:"5s"`
}
```

## Provider Configurations

### HashiCorp Vault

```go
type VaultConfig struct {
    // Vault address
    Address string `env:"VAULT_ADDR" default:"http://localhost:8200"`

    // Authentication token
    Token string `env:"VAULT_TOKEN"`

    // Mount path
    MountPath string `env:"VAULT_MOUNT_PATH" default:"secret"`

    // Namespace (enterprise)
    Namespace string `env:"VAULT_NAMESPACE"`

    // TLS configuration
    TLS struct {
        CACert     string `env:"VAULT_CACERT"`
        ClientCert string `env:"VAULT_CLIENT_CERT"`
        ClientKey  string `env:"VAULT_CLIENT_KEY"`
        SkipVerify bool   `env:"VAULT_SKIP_VERIFY" default:"false"`
    }

    // Authentication method: "token", "approle", "kubernetes", "aws"
    AuthMethod string `env:"VAULT_AUTH_METHOD" default:"token"`

    // AppRole authentication
    AppRole struct {
        RoleID   string `env:"VAULT_ROLE_ID"`
        SecretID string `env:"VAULT_SECRET_ID"`
    }

    // Kubernetes authentication
    Kubernetes struct {
        Role          string `env:"VAULT_K8S_ROLE"`
        TokenPath     string `env:"VAULT_K8S_TOKEN_PATH" default:"/var/run/secrets/kubernetes.io/serviceaccount/token"`
        MountPath     string `env:"VAULT_K8S_MOUNT_PATH" default:"kubernetes"`
    }
}
```

### AWS Secrets Manager

```go
type AWSConfig struct {
    // Region
    Region string `env:"AWS_REGION" default:"us-east-1"`

    // Access key (if not using IAM/environment)
    AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
    SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`

    // Role to assume
    RoleARN string `env:"AWS_ROLE_ARN"`

    // Endpoint (for localstack)
    Endpoint string `env:"AWS_SECRETS_ENDPOINT"`

    // Prefix for secret names
    Prefix string `env:"AWS_SECRETS_PREFIX" default:""`
}
```

### GCP Secret Manager

```go
type GCPConfig struct {
    // Project ID
    ProjectID string `env:"GCP_PROJECT_ID" required:"true"`

    // Credentials JSON file path
    CredentialsFile string `env:"GOOGLE_APPLICATION_CREDENTIALS"`

    // Credentials JSON string
    CredentialsJSON string `env:"GCP_CREDENTIALS_JSON"`

    // Prefix for secret names
    Prefix string `env:"GCP_SECRETS_PREFIX" default:""`
}
```

### Azure Key Vault

```go
type AzureConfig struct {
    // Vault URL
    VaultURL string `env:"AZURE_VAULT_URL" required:"true"`

    // Tenant ID
    TenantID string `env:"AZURE_TENANT_ID"`

    // Client ID
    ClientID string `env:"AZURE_CLIENT_ID"`

    // Client Secret
    ClientSecret string `env:"AZURE_CLIENT_SECRET"`

    // Use managed identity
    UseManagedIdentity bool `env:"AZURE_USE_MANAGED_IDENTITY" default:"false"`
}
```

### Environment Provider

```go
type EnvConfig struct {
    // Prefix for environment variables
    Prefix string `env:"SECRETS_ENV_PREFIX" default:"SECRET_"`

    // Delimiter for nested keys
    Delimiter string `env:"SECRETS_ENV_DELIMITER" default:"_"`
}

// Mapping:
// database/password -> SECRET_DATABASE_PASSWORD
// api/keys/stripe -> SECRET_API_KEYS_STRIPE
```

### File Provider

```go
type FileConfig struct {
    // Path to secrets file
    Path string `env:"SECRETS_FILE_PATH" default:".secrets.enc"`

    // Encryption key (or derive from password)
    Key string `env:"SECRETS_FILE_KEY"`

    // Password for key derivation
    Password string `env:"SECRETS_FILE_PASSWORD"`

    // Key derivation salt
    Salt string `env:"SECRETS_FILE_SALT"`
}
```

## Error Handling

```go
var (
    // ErrNotFound is returned when secret doesn't exist
    ErrNotFound = errors.New("secrets: secret not found")

    // ErrAccessDenied is returned on permission errors
    ErrAccessDenied = errors.New("secrets: access denied")

    // ErrVersionNotFound is returned for unknown version
    ErrVersionNotFound = errors.New("secrets: version not found")

    // ErrProviderError is returned on provider failures
    ErrProviderError = errors.New("secrets: provider error")

    // ErrExpired is returned when secret has expired
    ErrExpired = errors.New("secrets: secret expired")

    // ErrReadOnly is returned for read-only providers
    ErrReadOnly = errors.New("secrets: provider is read-only")
)

// SecretError wraps errors without exposing secret data
type SecretError struct {
    Path string
    Op   string
    Err  error
}

// Error never includes secret values
func (e *SecretError) Error() string {
    return fmt.Sprintf("secrets: %s %s: %v", e.Op, e.Path, e.Err)
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "github.com/user/core-backend/pkg/secrets"
)

func main() {
    // Create secrets manager
    mgr, err := secrets.New(secrets.Config{
        Provider: "env",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer mgr.Close()

    ctx := context.Background()

    // Get secret
    secret, err := mgr.Get(ctx, "database/credentials")
    if err != nil {
        log.Fatal(err)
    }

    username := secret.Value("username")
    password := secret.Value("password")

    // Or get single value
    apiKey, err := mgr.GetString(ctx, "api/stripe/key")
}
```

### With HashiCorp Vault

```go
import (
    "github.com/user/core-backend/pkg/secrets"
    "github.com/user/core-backend/pkg/secrets/provider"
)

func main() {
    vaultProvider, err := provider.NewVault(provider.VaultConfig{
        Address:   "https://vault.example.com:8200",
        AuthMethod: "approle",
        AppRole: provider.AppRoleConfig{
            RoleID:   os.Getenv("VAULT_ROLE_ID"),
            SecretID: os.Getenv("VAULT_SECRET_ID"),
        },
        MountPath: "secret",
    })
    if err != nil {
        log.Fatal(err)
    }

    mgr, err := secrets.New(secrets.Config{},
        secrets.WithProvider(vaultProvider),
        secrets.WithCache(5 * time.Minute),
    )

    ctx := context.Background()

    // Get secret from Vault path: secret/data/database/credentials
    secret, err := mgr.Get(ctx, "database/credentials")
    password := secret.Value("password")
}
```

### With AWS Secrets Manager

```go
import (
    "github.com/user/core-backend/pkg/secrets/provider"
)

func main() {
    awsProvider, err := provider.NewAWS(provider.AWSConfig{
        Region: "us-west-2",
    })
    if err != nil {
        log.Fatal(err)
    }

    mgr, _ := secrets.New(secrets.Config{},
        secrets.WithProvider(awsProvider),
    )

    // Get secret by name
    secret, err := mgr.Get(ctx, "prod/database/credentials")

    // Get specific version
    secret, err = mgr.GetVersion(ctx, "prod/database/credentials", "AWSCURRENT")
}
```

### With Caching

```go
func main() {
    mgr, _ := secrets.New(secrets.Config{
        Cache: secrets.CacheConfig{
            Enabled:    true,
            TTL:        5 * time.Minute,
            MaxEntries: 100,
        },
    })

    // First call fetches from provider
    secret, _ := mgr.Get(ctx, "database/password")

    // Second call returns cached value
    secret, _ = mgr.Get(ctx, "database/password")

    // Force refresh
    mgr.Refresh(ctx, "database/password")
}
```

### Secret Rotation Handling

```go
func main() {
    mgr, _ := secrets.New(cfg)

    // Watch for changes
    mgr.Watch(ctx, "database/credentials", func(secret *secrets.Secret) {
        log.Println("Database credentials rotated!")

        // Update connection pool with new credentials
        newPassword := secret.Value("password")
        updateDatabasePool(newPassword)
    })

    // Alternative: Use short TTL and handle rotation on next Get
    mgr, _ := secrets.New(cfg,
        secrets.WithCache(30 * time.Second), // Short TTL
    )
}
```

### Integration with pkg/config

```go
import (
    "github.com/user/core-backend/pkg/config"
    "github.com/user/core-backend/pkg/secrets"
)

type AppConfig struct {
    Database struct {
        Host     string `config:"host"`
        Port     int    `config:"port"`
        // Secret reference - resolved automatically
        Password string `config:"password" secret:"database/password"`
    } `config:"database"`

    // Direct secret reference
    APIKey string `secret:"api/stripe/key"`
}

func main() {
    secretsMgr, _ := secrets.New(secrets.Config{
        Provider: "vault",
    })

    cfg := config.New(
        config.WithProvider(config.NewEnvProvider()),
        config.WithSecrets(secretsMgr), // Resolve secret references
    )

    cfg.Load(ctx)

    var appCfg AppConfig
    cfg.Bind(&appCfg)

    // appCfg.Database.Password is resolved from secrets
}
```

### Storing Secrets

```go
func main() {
    mgr, _ := secrets.New(cfg)

    // Store secret (if provider supports write)
    err := mgr.Put(ctx, "database/new-credentials", map[string]interface{}{
        "username": "admin",
        "password": "secure-password-123",
    })

    // Delete secret
    err = mgr.Delete(ctx, "database/old-credentials")
}
```

### Local Encrypted File

```go
func main() {
    // For development/testing with encrypted local file
    fileProvider, err := provider.NewFile(provider.FileConfig{
        Path:     ".secrets.enc",
        Password: os.Getenv("SECRETS_PASSWORD"),
    })

    mgr, _ := secrets.New(secrets.Config{},
        secrets.WithProvider(fileProvider),
    )

    // Secrets are encrypted at rest using AES-256-GCM
}
```

### Multi-Provider (Fallback)

```go
func main() {
    // Try Vault first, fall back to environment
    mgr, _ := secrets.New(cfg,
        secrets.WithProvider(vaultProvider),
        secrets.WithFallbackProvider(envProvider),
    )

    // If Vault fails, tries environment
    secret, err := mgr.Get(ctx, "database/password")
}
```

### Listing Secrets

```go
func main() {
    mgr, _ := secrets.New(cfg)

    // List secrets at path
    paths, err := mgr.List(ctx, "database/")
    // ["database/primary", "database/replica", "database/readonly"]

    for _, path := range paths {
        secret, _ := mgr.Get(ctx, path)
        // ...
    }
}
```

## Secret Struct Helpers

```go
// Parse common secret formats
func (s *Secret) AsConnectionString() string {
    return fmt.Sprintf("%s:%s@%s:%d/%s",
        s.Value("username"),
        s.Value("password"),
        s.Value("host"),
        s.ValueInt("port"),
        s.Value("database"),
    )
}

func (s *Secret) ValueInt(key string) int
func (s *Secret) ValueBool(key string) bool
func (s *Secret) ValueDuration(key string) time.Duration
```

## Health Check

```go
// HealthCheck verifies secrets backend connectivity
func (m *Manager) HealthCheck() func(ctx context.Context) error {
    return func(ctx context.Context) error {
        // Try to access a test secret or perform a simple operation
        return m.provider.Ping(ctx)
    }
}
```

## Dependencies

- **Required:** None (env provider uses stdlib)
- **Optional:**
  - `github.com/hashicorp/vault/api` for Vault
  - `github.com/aws/aws-sdk-go-v2/service/secretsmanager` for AWS
  - `cloud.google.com/go/secretmanager` for GCP
  - `github.com/Azure/azure-sdk-for-go` for Azure

## Test Coverage Requirements

- Unit tests for all providers
- Caching tests
- Rotation handling tests
- Encryption/decryption tests
- Integration tests with localstack/mock
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & Environment Provider
1. Define Manager, Secret interfaces
2. Implement environment provider
3. Basic caching layer
4. Error handling

### Phase 2: Vault Provider
1. Implement Vault provider
2. Multiple auth methods (token, approle, k8s)
3. KV v2 support
4. Token renewal

### Phase 3: AWS Provider
1. Implement AWS Secrets Manager
2. IAM authentication
3. Version support
4. Rotation handling

### Phase 4: Additional Providers
1. GCP Secret Manager
2. Azure Key Vault
3. Encrypted file provider

### Phase 5: Advanced Features
1. Watch/rotation callbacks
2. Multi-provider with fallback
3. Config integration
4. Secret struct helpers

### Phase 6: Documentation & Examples
1. README with full documentation
2. Vault example
3. AWS example
4. Rotation handling example
