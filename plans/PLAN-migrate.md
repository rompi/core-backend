# Package Plan: pkg/migrate

## Overview

A database migration package for managing schema changes with version control. Supports SQL and Go-based migrations, up/down migrations, and multiple database backends. Designed to integrate seamlessly with `pkg/postgres` while remaining database-agnostic.

## Goals

1. **Version Control** - Track schema changes with sequential versions
2. **Reversible Migrations** - Support up and down migrations
3. **Multiple Formats** - SQL files and Go code migrations
4. **Database Agnostic** - PostgreSQL, MySQL, SQLite support
5. **Embedded Migrations** - Support embed.FS for compiled binaries
6. **CLI & Library** - Use as CLI tool or library
7. **Safe Execution** - Locking to prevent concurrent migrations

## Architecture

```
pkg/migrate/
├── migrate.go            # Core Migrator interface
├── config.go             # Configuration with env support
├── options.go            # Functional options
├── migration.go          # Migration type definition
├── version.go            # Version tracking
├── errors.go             # Custom error types
├── source/
│   ├── source.go         # Source interface
│   ├── file.go           # Filesystem source
│   ├── embed.go          # embed.FS source
│   └── go.go             # Go code migrations
├── driver/
│   ├── driver.go         # Database driver interface
│   ├── postgres.go       # PostgreSQL driver
│   ├── mysql.go          # MySQL driver
│   └── sqlite.go         # SQLite driver
├── lock/
│   ├── lock.go           # Lock interface
│   ├── postgres.go       # PostgreSQL advisory lock
│   └── mysql.go          # MySQL lock
├── cmd/
│   └── migrate/          # CLI tool
│       └── main.go
├── examples/
│   ├── basic/
│   ├── embedded/
│   └── go-migrations/
└── README.md
```

## Core Interfaces

```go
package migrate

import (
    "context"
    "io/fs"
)

// Migrator manages database migrations
type Migrator interface {
    // Up runs all pending migrations
    Up(ctx context.Context) error

    // UpTo runs migrations up to a specific version
    UpTo(ctx context.Context, version int64) error

    // Down rolls back the last migration
    Down(ctx context.Context) error

    // DownTo rolls back to a specific version
    DownTo(ctx context.Context, version int64) error

    // Reset rolls back all migrations and re-runs them
    Reset(ctx context.Context) error

    // Status returns migration status
    Status(ctx context.Context) ([]MigrationStatus, error)

    // Version returns current version
    Version(ctx context.Context) (int64, bool, error)

    // Pending returns pending migrations
    Pending(ctx context.Context) ([]Migration, error)

    // Create creates a new migration file
    Create(name string, opts ...CreateOption) (string, error)

    // Close releases resources
    Close() error
}

// Migration represents a single migration
type Migration struct {
    // Version is the migration version (timestamp or sequence)
    Version int64

    // Name is the migration name
    Name string

    // Up is the up migration SQL or function
    Up MigrationFunc

    // Down is the down migration SQL or function
    Down MigrationFunc
}

// MigrationFunc executes a migration
type MigrationFunc func(ctx context.Context, tx Tx) error

// MigrationStatus represents migration state
type MigrationStatus struct {
    Version   int64
    Name      string
    AppliedAt *time.Time
    Pending   bool
}

// Tx represents a database transaction
type Tx interface {
    Exec(ctx context.Context, sql string, args ...interface{}) error
    Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
    QueryRow(ctx context.Context, sql string, args ...interface{}) Row
}

// Source provides migrations
type Source interface {
    // Migrations returns all available migrations
    Migrations() ([]Migration, error)
}

// Driver provides database operations
type Driver interface {
    // Lock acquires migration lock
    Lock(ctx context.Context) error

    // Unlock releases migration lock
    Unlock(ctx context.Context) error

    // CreateVersionTable creates the version tracking table
    CreateVersionTable(ctx context.Context) error

    // GetAppliedVersions returns applied migration versions
    GetAppliedVersions(ctx context.Context) ([]int64, error)

    // SetVersion marks a version as applied
    SetVersion(ctx context.Context, version int64) error

    // RemoveVersion removes a version record
    RemoveVersion(ctx context.Context, version int64) error

    // Begin starts a transaction
    Begin(ctx context.Context) (Tx, error)

    // Close releases resources
    Close() error
}
```

## Configuration

```go
// Config holds migration configuration
type Config struct {
    // Database URL
    DatabaseURL string `env:"MIGRATE_DATABASE_URL" required:"true"`

    // Migration source path
    SourcePath string `env:"MIGRATE_SOURCE_PATH" default:"./migrations"`

    // Table name for tracking migrations
    TableName string `env:"MIGRATE_TABLE_NAME" default:"schema_migrations"`

    // Schema name (PostgreSQL)
    Schema string `env:"MIGRATE_SCHEMA" default:"public"`

    // Lock timeout
    LockTimeout time.Duration `env:"MIGRATE_LOCK_TIMEOUT" default:"30s"`

    // Statement timeout
    StatementTimeout time.Duration `env:"MIGRATE_STATEMENT_TIMEOUT" default:"0"`

    // Versioning scheme: "timestamp" or "sequence"
    VersionScheme string `env:"MIGRATE_VERSION_SCHEME" default:"timestamp"`
}

// CreateOption configures migration creation
type CreateOption func(*createOptions)

type createOptions struct {
    Format    string // "sql" or "go"
    Directory string
}

// WithFormat sets migration format
func WithFormat(format string) CreateOption

// WithDirectory sets output directory
func WithDirectory(dir string) CreateOption
```

## Migration File Format

### SQL Migrations

```
migrations/
├── 20240115120000_create_users.up.sql
├── 20240115120000_create_users.down.sql
├── 20240115130000_add_email_index.up.sql
├── 20240115130000_add_email_index.down.sql
└── ...
```

**Naming convention:** `{version}_{name}.{direction}.sql`

```sql
-- 20240115120000_create_users.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

```sql
-- 20240115120000_create_users.down.sql
DROP TABLE IF EXISTS users;
```

### Go Migrations

```go
// migrations/20240115120000_create_users.go
package migrations

import (
    "context"
    "github.com/user/core-backend/pkg/migrate"
)

func init() {
    migrate.Register(&migrate.Migration{
        Version: 20240115120000,
        Name:    "create_users",
        Up: func(ctx context.Context, tx migrate.Tx) error {
            _, err := tx.Exec(ctx, `
                CREATE TABLE users (
                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                    email VARCHAR(255) NOT NULL UNIQUE,
                    name VARCHAR(255) NOT NULL,
                    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
                )
            `)
            return err
        },
        Down: func(ctx context.Context, tx migrate.Tx) error {
            _, err := tx.Exec(ctx, `DROP TABLE IF EXISTS users`)
            return err
        },
    })
}
```

## Source Implementations

### Filesystem Source

```go
// FileSource reads migrations from filesystem
type FileSource struct {
    path string
}

// NewFileSource creates a filesystem source
func NewFileSource(path string) *FileSource

// Migrations returns all migrations from files
func (s *FileSource) Migrations() ([]Migration, error)
```

### Embedded Source

```go
// EmbedSource reads migrations from embed.FS
type EmbedSource struct {
    fs   embed.FS
    path string
}

// NewEmbedSource creates an embedded source
func NewEmbedSource(fsys embed.FS, path string) *EmbedSource

// Usage with embed
//go:embed migrations/*.sql
var migrationsFS embed.FS

source := migrate.NewEmbedSource(migrationsFS, "migrations")
```

### Go Source

```go
// GoSource collects registered Go migrations
type GoSource struct {
    migrations []Migration
}

// Register adds a Go migration
func Register(m *Migration)

// NewGoSource creates a Go source from registered migrations
func NewGoSource() *GoSource
```

## Driver Implementations

### PostgreSQL Driver

```go
// PostgresDriver implements Driver for PostgreSQL
type PostgresDriver struct {
    db        *sql.DB
    tableName string
    schema    string
}

// NewPostgresDriver creates a PostgreSQL driver
func NewPostgresDriver(db *sql.DB, opts ...DriverOption) *PostgresDriver

// Integration with pkg/postgres
func NewPostgresDriverFromClient(client *postgres.Client, opts ...DriverOption) *PostgresDriver
```

### MySQL Driver

```go
// MySQLDriver implements Driver for MySQL
type MySQLDriver struct {
    db        *sql.DB
    tableName string
}

// NewMySQLDriver creates a MySQL driver
func NewMySQLDriver(db *sql.DB, opts ...DriverOption) *MySQLDriver
```

## Error Handling

```go
var (
    // ErrNoChange is returned when no migrations to run
    ErrNoChange = errors.New("migrate: no change")

    // ErrLocked is returned when migrations are locked
    ErrLocked = errors.New("migrate: database is locked")

    // ErrDirty is returned when migration state is dirty
    ErrDirty = errors.New("migrate: dirty migration state")

    // ErrVersionNotFound is returned for unknown version
    ErrVersionNotFound = errors.New("migrate: version not found")

    // ErrInvalidVersion is returned for invalid version format
    ErrInvalidVersion = errors.New("migrate: invalid version")

    // ErrNoDown is returned when down migration doesn't exist
    ErrNoDown = errors.New("migrate: down migration not found")
)

// MigrationError provides detailed migration failure info
type MigrationError struct {
    Version int64
    Name    string
    Err     error
    Query   string // Failed SQL statement
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "github.com/user/core-backend/pkg/migrate"
)

func main() {
    // Create migrator
    m, err := migrate.New(migrate.Config{
        DatabaseURL: "postgres://user:pass@localhost/db",
        SourcePath:  "./migrations",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer m.Close()

    ctx := context.Background()

    // Run all pending migrations
    if err := m.Up(ctx); err != nil {
        log.Fatal(err)
    }

    log.Println("Migrations completed!")
}
```

### With Embedded Migrations

```go
import (
    "embed"
    "github.com/user/core-backend/pkg/migrate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
    m, err := migrate.New(
        migrate.Config{
            DatabaseURL: os.Getenv("DATABASE_URL"),
        },
        migrate.WithSource(migrate.NewEmbedSource(migrationsFS, "migrations")),
    )
    if err != nil {
        log.Fatal(err)
    }

    m.Up(ctx)
}
```

### With pkg/postgres Integration

```go
import (
    "github.com/user/core-backend/pkg/postgres"
    "github.com/user/core-backend/pkg/migrate"
)

func main() {
    // Create postgres client
    pgClient, err := postgres.New(postgres.LoadConfig())
    if err != nil {
        log.Fatal(err)
    }

    // Create migrator using postgres client
    m, err := migrate.New(
        migrate.Config{
            SourcePath: "./migrations",
        },
        migrate.WithPostgresClient(pgClient),
    )

    m.Up(ctx)
}
```

### Check Status

```go
func main() {
    m, _ := migrate.New(cfg)

    ctx := context.Background()

    // Get current version
    version, dirty, err := m.Version(ctx)
    fmt.Printf("Current version: %d, dirty: %v\n", version, dirty)

    // Get all migration status
    statuses, err := m.Status(ctx)
    for _, s := range statuses {
        status := "pending"
        if s.AppliedAt != nil {
            status = fmt.Sprintf("applied at %s", s.AppliedAt)
        }
        fmt.Printf("%d_%s: %s\n", s.Version, s.Name, status)
    }

    // Get pending migrations
    pending, err := m.Pending(ctx)
    fmt.Printf("Pending migrations: %d\n", len(pending))
}
```

### Rollback

```go
func main() {
    m, _ := migrate.New(cfg)
    ctx := context.Background()

    // Rollback last migration
    if err := m.Down(ctx); err != nil {
        log.Fatal(err)
    }

    // Rollback to specific version
    if err := m.DownTo(ctx, 20240101000000); err != nil {
        log.Fatal(err)
    }

    // Reset (rollback all, then up all)
    if err := m.Reset(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### Create New Migration

```go
func main() {
    m, _ := migrate.New(cfg)

    // Create SQL migration
    path, err := m.Create("add_user_roles",
        migrate.WithFormat("sql"),
        migrate.WithDirectory("./migrations"),
    )
    fmt.Println("Created:", path)
    // Output: Created: ./migrations/20240115143022_add_user_roles.up.sql

    // Create Go migration
    path, err = m.Create("seed_admin_user",
        migrate.WithFormat("go"),
    )
}
```

### Go Migrations

```go
// migrations/migrations.go
package migrations

import "github.com/user/core-backend/pkg/migrate"

// Import all migration files
import (
    _ "myapp/migrations/20240115_create_users"
    _ "myapp/migrations/20240116_add_roles"
)

// GetSource returns the Go migration source
func GetSource() migrate.Source {
    return migrate.NewGoSource()
}

// main.go
func main() {
    m, _ := migrate.New(cfg,
        migrate.WithSource(migrations.GetSource()),
    )
    m.Up(ctx)
}
```

## CLI Tool

```bash
# Install CLI
go install github.com/user/core-backend/pkg/migrate/cmd/migrate@latest

# Run migrations
migrate -database "postgres://localhost/db" -path ./migrations up

# Rollback
migrate -database "postgres://localhost/db" -path ./migrations down

# Rollback to version
migrate -database "postgres://localhost/db" -path ./migrations down-to 20240101000000

# Check status
migrate -database "postgres://localhost/db" -path ./migrations status

# Get current version
migrate -database "postgres://localhost/db" -path ./migrations version

# Create new migration
migrate -path ./migrations create add_user_roles sql

# Force set version (dangerous!)
migrate -database "postgres://localhost/db" force 20240115120000
```

## Hooks

```go
// Hook interface for observability
type Hook interface {
    BeforeMigration(ctx context.Context, m Migration, direction string)
    AfterMigration(ctx context.Context, m Migration, direction string, err error, duration time.Duration)
    BeforeUp(ctx context.Context, pending []Migration)
    AfterUp(ctx context.Context, applied []Migration, err error)
    BeforeDown(ctx context.Context, m Migration)
    AfterDown(ctx context.Context, m Migration, err error)
}

// WithHook adds observability hooks
func WithHook(hook Hook) Option
```

## Dependencies

- **Required:** `database/sql` (stdlib)
- **Optional:**
  - `github.com/jackc/pgx/v5/stdlib` for PostgreSQL
  - `github.com/go-sql-driver/mysql` for MySQL
  - `github.com/mattn/go-sqlite3` for SQLite

## Test Coverage Requirements

- Unit tests for all public functions
- Migration parsing tests
- Version ordering tests
- Transaction rollback tests
- Lock contention tests
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & PostgreSQL
1. Define Migrator, Migration, Driver interfaces
2. Implement PostgreSQL driver
3. SQL file source
4. Basic up/down operations

### Phase 2: Additional Sources
1. Embedded FS source
2. Go code migrations
3. Migration registration

### Phase 3: Additional Drivers
1. MySQL driver
2. SQLite driver

### Phase 4: CLI Tool
1. Command-line interface
2. Status and version commands
3. Create migration command

### Phase 5: Advanced Features
1. Hooks for observability
2. Statement timeout
3. Dry-run mode
4. Checksum verification

### Phase 6: Documentation & Examples
1. README with full documentation
2. SQL migrations example
3. Embedded migrations example
4. Go migrations example
