package postgres

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds PostgreSQL connection configuration loaded from the environment.
type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Schema   string `json:"schema"` // Database schema (default: "public")
	SSLMode  string `json:"ssl_mode"`

	MaxConns        int32         `json:"max_conns"`
	MinConns        int32         `json:"min_conns"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time"`
	ConnectTimeout  time.Duration `json:"connect_timeout"`
	QueryTimeout    time.Duration `json:"query_timeout"`
}

// LoadConfig reads configuration from environment variables and validates it.
func LoadConfig() (*Config, error) {
	cfg := defaultConfig()
	if err := cfg.overrideFromEnv(); err != nil {
		return nil, fmt.Errorf("loading postgres config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating postgres config: %w", err)
	}
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		Schema:          "public",
		SSLMode:         "prefer",
		MaxConns:        25,
		MinConns:        5,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
		ConnectTimeout:  10 * time.Second,
		QueryTimeout:    30 * time.Second,
	}
}

func (c *Config) overrideFromEnv() error {
	if v := strings.TrimSpace(os.Getenv("POSTGRES_HOST")); v != "" {
		c.Host = v
	}
	if v := strings.TrimSpace(os.Getenv("POSTGRES_USER")); v != "" {
		c.User = v
	}
	if v := strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD")); v != "" {
		c.Password = v
	}
	if v := strings.TrimSpace(os.Getenv("POSTGRES_DATABASE")); v != "" {
		c.Database = v
	}
	if v := strings.TrimSpace(os.Getenv("POSTGRES_SCHEMA")); v != "" {
		c.Schema = v
	}
	if v := strings.TrimSpace(os.Getenv("POSTGRES_SSL_MODE")); v != "" {
		c.SSLMode = v
	}

	if port, err := parseIntEnv("POSTGRES_PORT"); err != nil {
		return err
	} else if port != nil {
		c.Port = *port
	}

	if maxConns, err := parseInt32Env("POSTGRES_MAX_CONNS"); err != nil {
		return err
	} else if maxConns != nil {
		c.MaxConns = *maxConns
	}

	if minConns, err := parseInt32Env("POSTGRES_MIN_CONNS"); err != nil {
		return err
	} else if minConns != nil {
		c.MinConns = *minConns
	}

	if d, err := parseDurationEnv("POSTGRES_MAX_CONN_LIFETIME"); err != nil {
		return err
	} else if d != nil {
		c.MaxConnLifetime = *d
	}

	if d, err := parseDurationEnv("POSTGRES_MAX_CONN_IDLE_TIME"); err != nil {
		return err
	} else if d != nil {
		c.MaxConnIdleTime = *d
	}

	if d, err := parseDurationEnv("POSTGRES_CONNECT_TIMEOUT"); err != nil {
		return err
	} else if d != nil {
		c.ConnectTimeout = *d
	}

	if d, err := parseDurationEnv("POSTGRES_QUERY_TIMEOUT"); err != nil {
		return err
	} else if d != nil {
		c.QueryTimeout = *d
	}

	return nil
}

// Validate ensures the configuration contains valid values.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.User) == "" {
		return fmt.Errorf("%w: POSTGRES_USER is required", ErrInvalidConfig)
	}
	if strings.TrimSpace(c.Password) == "" {
		return fmt.Errorf("%w: POSTGRES_PASSWORD is required", ErrInvalidConfig)
	}
	if strings.TrimSpace(c.Database) == "" {
		return fmt.Errorf("%w: POSTGRES_DATABASE is required", ErrInvalidConfig)
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("%w: POSTGRES_PORT must be between 1 and 65535", ErrInvalidConfig)
	}
	if c.MaxConns < 1 {
		return fmt.Errorf("%w: POSTGRES_MAX_CONNS must be at least 1", ErrInvalidConfig)
	}
	if c.MinConns < 0 {
		return fmt.Errorf("%w: POSTGRES_MIN_CONNS cannot be negative", ErrInvalidConfig)
	}
	if c.MinConns > c.MaxConns {
		return fmt.Errorf("%w: POSTGRES_MIN_CONNS cannot exceed POSTGRES_MAX_CONNS", ErrInvalidConfig)
	}
	if c.ConnectTimeout < time.Second {
		return fmt.Errorf("%w: POSTGRES_CONNECT_TIMEOUT must be at least 1s", ErrInvalidConfig)
	}
	if c.QueryTimeout < time.Second {
		return fmt.Errorf("%w: POSTGRES_QUERY_TIMEOUT must be at least 1s", ErrInvalidConfig)
	}

	validSSLModes := map[string]bool{
		"disable": true, "allow": true, "prefer": true,
		"require": true, "verify-ca": true, "verify-full": true,
	}
	if !validSSLModes[c.SSLMode] {
		return fmt.Errorf("%w: invalid POSTGRES_SSL_MODE: %s", ErrInvalidConfig, c.SSLMode)
	}

	return nil
}

// ConnectionString returns a PostgreSQL connection string.
func (c *Config) ConnectionString() string {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
	if c.Schema != "" && c.Schema != "public" {
		connStr += fmt.Sprintf(" search_path=%s", c.Schema)
	}
	return connStr
}

// ConnectionURL returns a PostgreSQL connection URL.
func (c *Config) ConnectionURL() string {
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode,
	)
	if c.Schema != "" && c.Schema != "public" {
		url += fmt.Sprintf("&search_path=%s", c.Schema)
	}
	return url
}

func parseIntEnv(key string) (*int, error) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		return &parsed, nil
	}
	return nil, nil
}

func parseInt32Env(key string) (*int32, error) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		result := int32(parsed)
		return &result, nil
	}
	return nil, nil
}

func parseDurationEnv(key string) (*time.Duration, error) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		return &parsed, nil
	}
	return nil, nil
}
