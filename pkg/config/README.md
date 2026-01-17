# pkg/config

A unified configuration management package supporting multiple sources with type-safe binding, validation, and hot reloading.

## Features

- **Multiple Sources**: Environment variables, YAML, JSON files
- **Priority Merging**: Layer configurations with defined precedence
- **Type-Safe Binding**: Bind to structs with struct tags
- **Hot Reloading**: Watch for configuration changes
- **Sensitive Values**: Mask sensitive fields in logs/output
- **Minimal Dependencies**: Core uses stdlib; YAML is the only external dependency

## Installation

```go
import "github.com/rompi/core-backend/pkg/config"
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/rompi/core-backend/pkg/config"
    "github.com/rompi/core-backend/pkg/config/provider"
)

type AppConfig struct {
    Server struct {
        Host string `config:"host" default:"localhost"`
        Port int    `config:"port" default:"8080" env:"PORT"`
    } `config:"server"`

    Database struct {
        URL string `config:"url" env:"DATABASE_URL" required:"true"`
    } `config:"database"`
}

func main() {
    cfg := config.New(
        config.WithProvider(provider.NewEnvProvider()),
    )

    ctx := context.Background()
    if err := cfg.Load(ctx); err != nil {
        log.Fatal(err)
    }

    var appCfg AppConfig
    if err := cfg.Bind(&appCfg); err != nil {
        log.Fatal(err)
    }

    log.Printf("Server: %s:%d", appCfg.Server.Host, appCfg.Server.Port)
}
```

## Struct Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `config` | Configuration key | `config:"host"` |
| `default` | Default value | `default:"localhost"` |
| `env` | Environment variable override | `env:"PORT"` |
| `required` | Mark as required | `required:"true"` |
| `sensitive` | Mark as sensitive (masked in output) | `sensitive:"true"` |

## Providers

### Environment Provider

```go
// Read all environment variables
provider.NewEnvProvider()

// With prefix (APP_DATABASE_HOST -> database.host)
provider.NewEnvProvider(
    provider.WithPrefix("APP"),
)
```

### File Provider

```go
// Auto-detect format from extension
provider.NewFileProvider("config.yaml")
provider.NewFileProvider("config.json")

// Optional file (no error if missing)
provider.NewFileProvider("config.local.yaml", provider.WithOptional())
```

### Memory Provider (Testing)

```go
provider.NewMemoryProvider()
provider.NewMemoryProviderWithValues(map[string]any{
    "host": "localhost",
    "port": 8080,
})
```

## Multi-Source Configuration

Providers are merged in order, with later providers overriding earlier ones:

```go
cfg := config.New(
    // Base defaults (lowest priority)
    config.WithProvider(provider.NewFileProvider("config/default.yaml")),

    // Profile-specific overrides
    config.WithProvider(provider.NewFileProvider("config/production.yaml", provider.WithOptional())),

    // Environment variables (highest priority)
    config.WithProvider(provider.NewEnvProvider(provider.WithPrefix("APP"))),
)
```

## Configuration Access

### Direct Access

```go
host := cfg.GetString("server.host")
port := cfg.GetInt("server.port")
enabled := cfg.GetBool("feature.enabled")
timeout := cfg.GetDuration("server.timeout")
hosts := cfg.GetStringSlice("allowed.hosts")
```

### Sub-Configuration

```go
dbCfg := cfg.Sub("database")
host := dbCfg.GetString("host")  // Same as cfg.GetString("database.host")
port := dbCfg.GetInt("port")
```

### Check If Set

```go
if cfg.IsSet("optional.key") {
    value := cfg.GetString("optional.key")
}
```

## Hot Reloading

```go
cfg.Watch(ctx, func(newCfg config.Config) {
    var updated AppConfig
    newCfg.Bind(&updated)
    log.Println("Configuration updated!")
    applyConfig(updated)
})
```

## Sensitive Values

```go
cfg := config.New(
    config.WithProvider(provider.NewEnvProvider()),
    config.WithSensitiveKey("database.password"),
    config.WithSensitiveKeys("api.key", "secret.token"),
)

// AllSettings() masks sensitive values
settings := cfg.AllSettings()
// {"database.password": "***", "api.key": "***", ...}
```

## Validation

Integrate with a validator to validate bound structs:

```go
type myValidator struct{}

func (v *myValidator) Validate(cfg any) error {
    // Custom validation logic
    return nil
}

cfg := config.New(
    config.WithProvider(provider.NewEnvProvider()),
    config.WithValidator(&myValidator{}),
)
```

## Example Configuration Files

### YAML (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  timeout: "30s"

database:
  host: "localhost"
  port: 5432
  name: "myapp"
  ssl_mode: "disable"

logging:
  level: "info"
  format: "json"
```

### JSON (config.json)

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "timeout": "30s"
  },
  "database": {
    "host": "localhost",
    "port": 5432
  }
}
```

## Environment Variable Mapping

With prefix `APP`:

| Environment Variable | Config Key |
|---------------------|------------|
| `APP_SERVER_HOST` | `server.host` |
| `APP_DATABASE_PORT` | `database.port` |
| `APP_LOGGING_LEVEL` | `logging.level` |

## Error Handling

```go
import "github.com/rompi/core-backend/pkg/config"

// Check for specific errors
if errors.Is(err, config.ErrRequired) {
    // Required field missing
}

if errors.Is(err, config.ErrProviderFailed) {
    // Provider failed to load
}

// Handle binding errors
if bindErr, ok := err.(*config.MultiBindError); ok {
    for _, e := range bindErr.Errors {
        log.Printf("Field %s: %s", e.Field, e.Message)
    }
}
```

## Best Practices

1. **Use struct binding** for type safety and documentation
2. **Set defaults** in struct tags for resilient configuration
3. **Mark secrets as sensitive** to prevent accidental logging
4. **Use environment variables** for deployment-specific overrides
5. **Validate on startup** to fail fast on misconfiguration
6. **Use profiles** for environment-specific configuration (dev, staging, prod)

## License

MIT
