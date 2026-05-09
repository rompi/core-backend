// Package redis provides a configurable Redis integration with full functionality support.
package redis

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the Redis client configuration.
type Config struct {
	Host            string
	Port            int
	Password        string
	DB              int
	Username        string
	PoolSize        int
	MinIdleConns    int
	ConnectTimeout  time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	IdleTimeout     time.Duration
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	TLSEnabled            bool
	TLSInsecureSkipVerify bool
	TLSCertFile           string
	TLSKeyFile            string
	TLSCAFile             string
	ClusterEnabled bool
	ClusterNodes   []string
	SentinelEnabled    bool
	SentinelMasterName string
	SentinelAddresses  []string
	SentinelPassword   string
	DefaultTTL       time.Duration
	KeyPrefix        string
	EnableTracing    bool
	EnableMetrics    bool
	CompressionLevel int
	Serializer       SerializerType
}

// SerializerType defines the serialization format.
type SerializerType string

const (
	SerializerJSON    SerializerType = "json"
	SerializerMsgPack SerializerType = "msgpack"
	SerializerGob     SerializerType = "gob"
)

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            6379,
		DB:              0,
		PoolSize:        10,
		MinIdleConns:    2,
		ConnectTimeout:  5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		IdleTimeout:     5 * time.Minute,
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		Serializer:      SerializerJSON,
	}
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Host = v
	}
	if v, err := parseIntEnv("REDIS_PORT"); err != nil {
		return nil, err
	} else if v != nil {
		cfg.Port = *v
	}
	cfg.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Username = os.Getenv("REDIS_USERNAME")
	if v, err := parseIntEnv("REDIS_DB"); err != nil {
		return nil, err
	} else if v != nil {
		cfg.DB = *v
	}
	if v, err := parseIntEnv("REDIS_POOL_SIZE"); err != nil {
		return nil, err
	} else if v != nil {
		cfg.PoolSize = *v
	}
	if v, err := parseDurationEnv("REDIS_CONNECT_TIMEOUT"); err != nil {
		return nil, err
	} else if v != nil {
		cfg.ConnectTimeout = *v
	}
	if v, err := parseBoolEnv("REDIS_TLS_ENABLED"); err != nil {
		return nil, err
	} else if v != nil {
		cfg.TLSEnabled = *v
	}
	if v, err := parseBoolEnv("REDIS_CLUSTER_ENABLED"); err != nil {
		return nil, err
	} else if v != nil {
		cfg.ClusterEnabled = *v
	}
	if v := os.Getenv("REDIS_CLUSTER_NODES"); v != "" {
		cfg.ClusterNodes = parseStringSlice(v)
	}
	if v, err := parseBoolEnv("REDIS_SENTINEL_ENABLED"); err != nil {
		return nil, err
	} else if v != nil {
		cfg.SentinelEnabled = *v
	}
	cfg.SentinelMasterName = os.Getenv("REDIS_SENTINEL_MASTER_NAME")
	if v := os.Getenv("REDIS_SENTINEL_ADDRESSES"); v != "" {
		cfg.SentinelAddresses = parseStringSlice(v)
	}
	cfg.KeyPrefix = os.Getenv("REDIS_KEY_PREFIX")
	if v := os.Getenv("REDIS_SERIALIZER"); v != "" {
		cfg.Serializer = SerializerType(strings.ToLower(v))
	}
	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("REDIS_PORT must be between 1 and 65535")
	}
	if c.DB < 0 || c.DB > 15 {
		return fmt.Errorf("REDIS_DB must be between 0 and 15")
	}
	if c.PoolSize < 1 {
		return fmt.Errorf("REDIS_POOL_SIZE must be at least 1")
	}
	if c.MinIdleConns > c.PoolSize {
		return fmt.Errorf("REDIS_MIN_IDLE_CONNS cannot exceed REDIS_POOL_SIZE")
	}
	if c.CompressionLevel < 0 || c.CompressionLevel > 9 {
		return fmt.Errorf("REDIS_COMPRESSION_LEVEL must be between 0 and 9")
	}
	switch c.Serializer {
	case SerializerJSON, SerializerMsgPack, SerializerGob, "":
	default:
		return fmt.Errorf("REDIS_SERIALIZER must be one of: json, msgpack, gob")
	}
	if c.TLSEnabled && c.TLSCertFile != "" && c.TLSKeyFile == "" {
		return fmt.Errorf("REDIS_TLS_KEY_FILE is required when REDIS_TLS_CERT_FILE is set")
	}
	if c.ClusterEnabled && len(c.ClusterNodes) == 0 {
		return fmt.Errorf("REDIS_CLUSTER_NODES is required when cluster mode is enabled")
	}
	if c.SentinelEnabled && c.SentinelMasterName == "" {
		return fmt.Errorf("REDIS_SENTINEL_MASTER_NAME is required when sentinel mode is enabled")
	}
	if c.SentinelEnabled && len(c.SentinelAddresses) == 0 {
		return fmt.Errorf("REDIS_SENTINEL_ADDRESSES is required when sentinel mode is enabled")
	}
	if c.ClusterEnabled && c.SentinelEnabled {
		return fmt.Errorf("cannot enable both cluster and sentinel mode simultaneously")
	}
	return nil
}

// Address returns the Redis server address.
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// TLSConfig returns a TLS configuration.
func (c *Config) TLSConfig() (*tls.Config, error) {
	if !c.TLSEnabled {
		return nil, nil
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: c.TLSInsecureSkipVerify}
	if c.TLSCertFile != "" && c.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.TLSCertFile, c.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return tlsConfig, nil
}

// WithKeyPrefix returns the key with the configured prefix.
func (c *Config) WithKeyPrefix(key string) string {
	if c.KeyPrefix == "" {
		return key
	}
	return c.KeyPrefix + key
}

func parseIntEnv(key string) (*int, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid integer value %q: %w", key, v, err)
	}
	return &parsed, nil
}

func parseBoolEnv(key string) (*bool, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid boolean value %q: %w", key, v, err)
	}
	return &parsed, nil
}

func parseDurationEnv(key string) (*time.Duration, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil, nil
	}
	parsed, err := time.ParseDuration(v)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid duration value %q: %w", key, v, err)
	}
	return &parsed, nil
}

func parseStringSlice(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
