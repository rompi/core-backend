package feature

import (
	"fmt"
	"time"
)

// Config holds feature flag configuration.
type Config struct {
	// Provider specifies the backend: "memory", "file", "postgres", "redis", "launchdarkly".
	Provider string

	// RefreshInterval is how often to refresh flags from remote providers.
	RefreshInterval time.Duration

	// OfflineMode enables evaluation without remote provider connectivity.
	OfflineMode bool

	// SendEvents enables sending analytics events.
	SendEvents bool

	// File configuration (when Provider is "file").
	File FileConfig

	// Postgres configuration (when Provider is "postgres").
	Postgres PostgresConfig
}

// FileConfig for file-based provider.
type FileConfig struct {
	// Path is the path to the features file.
	Path string

	// WatchFile enables file watching for live updates.
	WatchFile bool
}

// PostgresConfig for database provider.
type PostgresConfig struct {
	// ConnectionString is the database connection string.
	ConnectionString string

	// TableName is the name of the feature flags table.
	TableName string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Provider:        "memory",
		RefreshInterval: 30 * time.Second,
		OfflineMode:     false,
		SendEvents:      true,
		File: FileConfig{
			Path:      "features.yaml",
			WatchFile: true,
		},
		Postgres: PostgresConfig{
			TableName: "feature_flags",
		},
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	validProviders := map[string]bool{
		"memory":   true,
		"file":     true,
		"postgres": true,
		"redis":    true,
	}

	if c.Provider != "" && !validProviders[c.Provider] {
		return fmt.Errorf("%w: invalid provider %q", ErrInvalidConfig, c.Provider)
	}

	if c.Provider == "file" && c.File.Path == "" {
		return fmt.Errorf("%w: file path is required for file provider", ErrInvalidConfig)
	}

	if c.Provider == "postgres" && c.Postgres.ConnectionString == "" {
		return fmt.Errorf("%w: connection string is required for postgres provider", ErrInvalidConfig)
	}

	if c.RefreshInterval < 0 {
		return fmt.Errorf("%w: refresh interval cannot be negative", ErrInvalidConfig)
	}

	return nil
}

// applyDefaults sets default values for unset fields.
func (c *Config) applyDefaults() {
	if c.Provider == "" {
		c.Provider = "memory"
	}

	if c.RefreshInterval == 0 {
		c.RefreshInterval = 30 * time.Second
	}

	if c.File.Path == "" {
		c.File.Path = "features.yaml"
	}

	if c.Postgres.TableName == "" {
		c.Postgres.TableName = "feature_flags"
	}
}
