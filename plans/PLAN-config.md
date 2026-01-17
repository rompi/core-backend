# Package Plan: pkg/config

## Overview

A unified configuration management package supporting multiple sources (environment variables, files, remote stores). Provides type-safe configuration with validation, hot reloading, and struct tag-based binding.

## Goals

1. **Multiple Sources** - Environment, YAML, JSON, TOML, Consul, etcd
2. **Priority Merging** - Layer configurations with defined precedence
3. **Type-Safe Binding** - Bind to structs with struct tags
4. **Validation** - Integrate with pkg/validator
5. **Hot Reloading** - Watch for configuration changes
6. **Secret Integration** - Mark sensitive fields, integrate with pkg/secrets
7. **Minimal Dependencies** - Core functionality uses stdlib

## Architecture

```
pkg/config/
├── config.go             # Core Config interface
├── options.go            # Functional options
├── errors.go             # Custom error types
├── binder.go             # Struct binding
├── watcher.go            # Hot reload support
├── provider/
│   ├── provider.go       # Provider interface
│   ├── env.go            # Environment variables
│   ├── file.go           # File-based (YAML, JSON, TOML)
│   ├── consul.go         # HashiCorp Consul
│   ├── etcd.go           # etcd
│   └── memory.go         # In-memory (testing)
├── parser/
│   ├── parser.go         # Parser interface
│   ├── yaml.go           # YAML parser
│   ├── json.go           # JSON parser
│   └── toml.go           # TOML parser
├── examples/
│   ├── basic/
│   ├── multi-source/
│   ├── hot-reload/
│   └── with-consul/
└── README.md
```

## Core Interfaces

```go
package config

import (
    "context"
    "time"
)

// Config manages configuration loading and binding
type Config interface {
    // Load loads configuration from all providers
    Load(ctx context.Context) error

    // Bind binds configuration to a struct
    Bind(v interface{}) error

    // Get returns a value by key
    Get(key string) interface{}

    // GetString returns a string value
    GetString(key string) string

    // GetInt returns an int value
    GetInt(key string) int

    // GetInt64 returns an int64 value
    GetInt64(key string) int64

    // GetFloat64 returns a float64 value
    GetFloat64(key string) float64

    // GetBool returns a bool value
    GetBool(key string) bool

    // GetDuration returns a time.Duration value
    GetDuration(key string) time.Duration

    // GetTime returns a time.Time value
    GetTime(key string) time.Time

    // GetStringSlice returns a string slice
    GetStringSlice(key string) []string

    // GetStringMap returns a string map
    GetStringMap(key string) map[string]interface{}

    // IsSet checks if a key is set
    IsSet(key string) bool

    // Set sets a value
    Set(key string, value interface{})

    // Watch watches for configuration changes
    Watch(ctx context.Context, callback func(Config)) error

    // Sub returns a sub-configuration
    Sub(key string) Config

    // AllSettings returns all settings as a map
    AllSettings() map[string]interface{}
}

// Provider provides configuration from a source
type Provider interface {
    // Name returns the provider name
    Name() string

    // Load loads configuration
    Load(ctx context.Context) (map[string]interface{}, error)

    // Watch watches for changes (optional)
    Watch(ctx context.Context, callback func()) error
}

// Parser parses configuration data
type Parser interface {
    // Parse parses data into a map
    Parse(data []byte) (map[string]interface{}, error)

    // Extensions returns supported file extensions
    Extensions() []string
}
```

## Struct Tags

```go
type ServerConfig struct {
    // Basic binding
    Host string `config:"host" default:"localhost"`
    Port int    `config:"port" default:"8080"`

    // Environment variable override
    APIKey string `config:"api_key" env:"API_KEY"`

    // Required field
    DatabaseURL string `config:"database_url" required:"true"`

    // Sensitive field (masked in logs)
    Password string `config:"password" sensitive:"true"`

    // Validation (requires pkg/validator)
    Email string `config:"email" validate:"email"`

    // Duration parsing
    Timeout time.Duration `config:"timeout" default:"30s"`

    // Nested configuration
    Database DatabaseConfig `config:"database"`

    // Slice binding
    AllowedOrigins []string `config:"allowed_origins" default:"http://localhost:3000"`
}

type DatabaseConfig struct {
    Host     string `config:"host" default:"localhost"`
    Port     int    `config:"port" default:"5432"`
    Name     string `config:"name" required:"true"`
    SSLMode  string `config:"ssl_mode" default:"disable"`
    MaxConns int    `config:"max_conns" default:"10"`
}
```

## Provider Configurations

### Environment Provider

```go
// EnvProvider reads from environment variables
type EnvProvider struct {
    prefix string
}

// NewEnvProvider creates an environment provider
func NewEnvProvider(opts ...EnvOption) *EnvProvider

// Options
func WithPrefix(prefix string) EnvOption     // e.g., "APP_"
func WithDelimiter(delim string) EnvOption   // e.g., "_" for APP_DATABASE_HOST

// Key mapping:
// database.host -> APP_DATABASE_HOST (with prefix "APP_")
// database.host -> DATABASE_HOST (without prefix)
```

### File Provider

```go
// FileProvider reads from configuration files
type FileProvider struct {
    path   string
    parser Parser
}

// NewFileProvider creates a file provider
func NewFileProvider(path string, opts ...FileOption) *FileProvider

// Options
func WithParser(parser Parser) FileOption
func WithOptional() FileOption // Don't error if file missing

// Supported formats: YAML, JSON, TOML
// Auto-detected from file extension
```

### Consul Provider

```go
// ConsulProvider reads from HashiCorp Consul
type ConsulProvider struct {
    client *api.Client
    prefix string
}

// NewConsulProvider creates a Consul provider
func NewConsulProvider(opts ...ConsulOption) (*ConsulProvider, error)

// Options
func WithConsulAddress(addr string) ConsulOption
func WithConsulToken(token string) ConsulOption
func WithConsulPrefix(prefix string) ConsulOption
func WithConsulDatacenter(dc string) ConsulOption

// Configuration
type ConsulConfig struct {
    Address    string `env:"CONSUL_HTTP_ADDR" default:"localhost:8500"`
    Token      string `env:"CONSUL_HTTP_TOKEN"`
    Prefix     string `env:"CONSUL_CONFIG_PREFIX" default:"config/"`
    Datacenter string `env:"CONSUL_DATACENTER"`
}
```

### etcd Provider

```go
// EtcdProvider reads from etcd
type EtcdProvider struct {
    client *clientv3.Client
    prefix string
}

// NewEtcdProvider creates an etcd provider
func NewEtcdProvider(opts ...EtcdOption) (*EtcdProvider, error)

// Options
func WithEtcdEndpoints(endpoints []string) EtcdOption
func WithEtcdPrefix(prefix string) EtcdOption
func WithEtcdUsername(user string) EtcdOption
func WithEtcdPassword(pass string) EtcdOption

// Configuration
type EtcdConfig struct {
    Endpoints []string `env:"ETCD_ENDPOINTS" default:"localhost:2379"`
    Prefix    string   `env:"ETCD_CONFIG_PREFIX" default:"/config/"`
    Username  string   `env:"ETCD_USERNAME"`
    Password  string   `env:"ETCD_PASSWORD"`
}
```

## Error Handling

```go
var (
    // ErrNotFound is returned when key doesn't exist
    ErrNotFound = errors.New("config: key not found")

    // ErrInvalidType is returned for type conversion errors
    ErrInvalidType = errors.New("config: invalid type")

    // ErrRequired is returned for missing required fields
    ErrRequired = errors.New("config: required field missing")

    // ErrValidation is returned for validation failures
    ErrValidation = errors.New("config: validation failed")

    // ErrProviderFailed is returned when provider fails
    ErrProviderFailed = errors.New("config: provider failed")
)

// BindError provides detailed binding failure info
type BindError struct {
    Field   string
    Tag     string
    Value   interface{}
    Message string
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "github.com/user/core-backend/pkg/config"
)

type AppConfig struct {
    Server   ServerConfig   `config:"server"`
    Database DatabaseConfig `config:"database"`
    Redis    RedisConfig    `config:"redis"`
}

type ServerConfig struct {
    Host string `config:"host" default:"0.0.0.0"`
    Port int    `config:"port" default:"8080" env:"PORT"`
}

type DatabaseConfig struct {
    URL string `config:"url" env:"DATABASE_URL" required:"true"`
}

type RedisConfig struct {
    URL string `config:"url" env:"REDIS_URL" default:"redis://localhost:6379"`
}

func main() {
    // Create config with providers
    cfg := config.New(
        config.WithProvider(config.NewEnvProvider()),
    )

    ctx := context.Background()

    // Load configuration
    if err := cfg.Load(ctx); err != nil {
        log.Fatal(err)
    }

    // Bind to struct
    var appCfg AppConfig
    if err := cfg.Bind(&appCfg); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Server: %s:%d\n", appCfg.Server.Host, appCfg.Server.Port)
}
```

### Multi-Source Configuration

```go
func main() {
    cfg := config.New(
        // Providers are merged in order (later overrides earlier)
        config.WithProvider(config.NewFileProvider("config/default.yaml")),
        config.WithProvider(config.NewFileProvider("config/local.yaml", config.WithOptional())),
        config.WithProvider(config.NewEnvProvider(config.WithPrefix("APP_"))),
    )

    // Priority: env > local.yaml > default.yaml
    cfg.Load(ctx)
}
```

### Configuration Files

```yaml
# config/default.yaml
server:
  host: "0.0.0.0"
  port: 8080
  timeout: "30s"

database:
  host: "localhost"
  port: 5432
  name: "myapp"
  ssl_mode: "disable"
  max_conns: 10

redis:
  url: "redis://localhost:6379"

logging:
  level: "info"
  format: "json"
```

```yaml
# config/production.yaml
server:
  host: "0.0.0.0"
  port: 80

database:
  host: "db.production.internal"
  ssl_mode: "require"
  max_conns: 50

logging:
  level: "warn"
```

### Environment Variable Mapping

```go
// With prefix "APP_"
type Config struct {
    Database struct {
        Host     string `config:"host"`     // APP_DATABASE_HOST
        Port     int    `config:"port"`     // APP_DATABASE_PORT
        Password string `config:"password"` // APP_DATABASE_PASSWORD
    } `config:"database"`
}

// Direct env override
type Config struct {
    Port int `config:"port" env:"PORT"` // Uses PORT directly
}
```

### Hot Reloading

```go
func main() {
    cfg := config.New(
        config.WithProvider(config.NewFileProvider("config.yaml")),
        config.WithProvider(config.NewConsulProvider()),
        config.WithWatchInterval(30 * time.Second),
    )

    cfg.Load(ctx)

    var appCfg AppConfig
    cfg.Bind(&appCfg)

    // Watch for changes
    cfg.Watch(ctx, func(newCfg config.Config) {
        var updated AppConfig
        newCfg.Bind(&updated)

        log.Println("Configuration updated!")

        // Apply changes
        applyConfig(updated)
    })

    // Application continues running...
}
```

### With Consul

```go
import (
    "github.com/user/core-backend/pkg/config"
    "github.com/user/core-backend/pkg/config/provider"
)

func main() {
    consulProvider, err := provider.NewConsulProvider(
        provider.WithConsulAddress("consul:8500"),
        provider.WithConsulPrefix("myapp/config/"),
    )
    if err != nil {
        log.Fatal(err)
    }

    cfg := config.New(
        config.WithProvider(config.NewEnvProvider()),
        config.WithProvider(consulProvider),
    )

    cfg.Load(ctx)

    // Consul KV structure:
    // myapp/config/database/host = "db.internal"
    // myapp/config/database/port = "5432"
    // myapp/config/server/port = "8080"
}
```

### With Validation

```go
import (
    "github.com/user/core-backend/pkg/config"
    "github.com/user/core-backend/pkg/validator"
)

type Config struct {
    Server struct {
        Host string `config:"host" validate:"required,hostname"`
        Port int    `config:"port" validate:"required,gte=1,lte=65535"`
    } `config:"server"`

    Database struct {
        URL string `config:"url" validate:"required,url"`
    } `config:"database"`

    Email string `config:"admin_email" validate:"required,email"`
}

func main() {
    v := validator.New()

    cfg := config.New(
        config.WithProvider(config.NewEnvProvider()),
        config.WithValidator(v), // Integrate validator
    )

    cfg.Load(ctx)

    var appCfg Config
    if err := cfg.Bind(&appCfg); err != nil {
        // Returns validation errors
        log.Fatal(err)
    }
}
```

### Sensitive Values

```go
type Config struct {
    Database struct {
        Password string `config:"password" sensitive:"true"`
    } `config:"database"`

    APIKey string `config:"api_key" sensitive:"true"`
}

func main() {
    cfg := config.New(
        config.WithProvider(config.NewEnvProvider()),
    )

    cfg.Load(ctx)

    // AllSettings() masks sensitive values
    settings := cfg.AllSettings()
    // {
    //   "database": {"password": "***"},
    //   "api_key": "***"
    // }

    // Actual values still accessible
    password := cfg.GetString("database.password") // Returns real value
}
```

### Sub-Configuration

```go
func main() {
    cfg := config.New(/*...*/)
    cfg.Load(ctx)

    // Get sub-configuration for database
    dbCfg := cfg.Sub("database")

    host := dbCfg.GetString("host")     // Instead of cfg.GetString("database.host")
    port := dbCfg.GetInt("port")

    // Bind sub-configuration
    var dbConfig DatabaseConfig
    dbCfg.Bind(&dbConfig)
}
```

### Default Values

```go
type Config struct {
    // Default from tag
    Port int `config:"port" default:"8080"`

    // Default duration
    Timeout time.Duration `config:"timeout" default:"30s"`

    // Default slice (comma-separated)
    Hosts []string `config:"hosts" default:"localhost,127.0.0.1"`

    // Default bool
    Debug bool `config:"debug" default:"false"`
}
```

### Profile-Based Configuration

```go
func main() {
    profile := os.Getenv("APP_PROFILE") // "development", "staging", "production"
    if profile == "" {
        profile = "development"
    }

    cfg := config.New(
        config.WithProvider(config.NewFileProvider("config/default.yaml")),
        config.WithProvider(config.NewFileProvider(fmt.Sprintf("config/%s.yaml", profile))),
        config.WithProvider(config.NewEnvProvider()),
    )

    cfg.Load(ctx)
}
```

### Integration with pkg/secrets

```go
import (
    "github.com/user/core-backend/pkg/config"
    "github.com/user/core-backend/pkg/secrets"
)

type Config struct {
    Database struct {
        // Value fetched from secrets manager
        Password string `config:"password" secret:"database/password"`
    } `config:"database"`
}

func main() {
    secretsClient, _ := secrets.New(/*...*/)

    cfg := config.New(
        config.WithProvider(config.NewEnvProvider()),
        config.WithSecrets(secretsClient), // Resolve secret:// references
    )

    cfg.Load(ctx)
}
```

## Printing Configuration

```go
func main() {
    cfg := config.New(/*...*/)
    cfg.Load(ctx)

    // Debug print (masks sensitive)
    cfg.Print(os.Stdout)

    // Output:
    // server:
    //   host: "0.0.0.0"
    //   port: 8080
    // database:
    //   host: "localhost"
    //   password: "***"
}
```

## Dependencies

- **Required:** None (uses stdlib for env, JSON)
- **Optional:**
  - `gopkg.in/yaml.v3` for YAML
  - `github.com/BurntSushi/toml` for TOML
  - `github.com/hashicorp/consul/api` for Consul
  - `go.etcd.io/etcd/client/v3` for etcd

## Test Coverage Requirements

- Unit tests for all providers
- Binding tests with various types
- Precedence/merging tests
- Hot reload tests
- Validation integration tests
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & Environment Provider
1. Define Config interface
2. Implement environment provider
3. Basic struct binding
4. Default values support

### Phase 2: File Providers
1. YAML parser and provider
2. JSON parser and provider
3. TOML parser and provider
4. Auto-detection by extension

### Phase 3: Struct Binding
1. Complete tag parsing
2. Nested struct binding
3. Slice and map binding
4. Duration and time parsing

### Phase 4: Remote Providers
1. Consul provider
2. etcd provider
3. Watch/reload support

### Phase 5: Advanced Features
1. Validator integration
2. Sensitive value masking
3. Sub-configuration
4. Profile support

### Phase 6: Documentation & Examples
1. README with full documentation
2. Basic example
3. Multi-source example
4. Hot reload example
