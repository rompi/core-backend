package server

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the server configuration.
type Config struct {
	// gRPC Server
	GRPCHost string
	GRPCPort int

	// HTTP Server (Gateway)
	HTTPHost         string
	HTTPPort         int
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPIdleTimeout  time.Duration

	// Shutdown
	ShutdownTimeout time.Duration

	// TLS (applies to both gRPC and HTTP)
	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string

	// Health Checks
	HealthEnabled     bool
	HealthHTTPPath    string
	LivenessHTTPPath  string
	ReadinessHTTPPath string

	// CORS (HTTP only)
	CORSEnabled          bool
	CORSAllowOrigins     []string
	CORSAllowMethods     []string
	CORSAllowHeaders     []string
	CORSExposeHeaders    []string
	CORSAllowCredentials bool
	CORSMaxAge           int

	// Global Rate Limiting
	RateLimitEnabled bool
	RateLimitRate    float64
	RateLimitBurst   int
	RateLimitExpiry  time.Duration

	// Compression (HTTP only)
	CompressionEnabled bool

	// Request ID
	RequestIDEnabled bool
	RequestIDHeader  string

	// Logging
	LogRequests  bool
	LogSkipPaths []string

	// Debug
	Debug bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		// gRPC Server
		GRPCHost: "",
		GRPCPort: 9090,

		// HTTP Server
		HTTPHost:         "",
		HTTPPort:         8080,
		HTTPReadTimeout:  30 * time.Second,
		HTTPWriteTimeout: 30 * time.Second,
		HTTPIdleTimeout:  120 * time.Second,

		// Shutdown
		ShutdownTimeout: 30 * time.Second,

		// TLS
		TLSEnabled:  false,
		TLSCertFile: "",
		TLSKeyFile:  "",

		// Health Checks
		HealthEnabled:     true,
		HealthHTTPPath:    "/health",
		LivenessHTTPPath:  "/health/live",
		ReadinessHTTPPath: "/health/ready",

		// CORS
		CORSEnabled:          false,
		CORSAllowOrigins:     []string{"*"},
		CORSAllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		CORSAllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		CORSExposeHeaders:    []string{},
		CORSAllowCredentials: false,
		CORSMaxAge:           86400,

		// Rate Limiting
		RateLimitEnabled: false,
		RateLimitRate:    20,
		RateLimitBurst:   40,
		RateLimitExpiry:  3 * time.Minute,

		// Compression
		CompressionEnabled: true,

		// Request ID
		RequestIDEnabled: true,
		RequestIDHeader:  "X-Request-ID",

		// Logging
		LogRequests:  true,
		LogSkipPaths: []string{},

		// Debug
		Debug: false,
	}
}

// LoadConfig loads configuration from environment variables with defaults.
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	// gRPC Server
	cfg.GRPCHost = getEnv("GRPC_HOST", cfg.GRPCHost)
	cfg.GRPCPort = getEnvInt("GRPC_PORT", cfg.GRPCPort)

	// HTTP Server
	cfg.HTTPHost = getEnv("HTTP_HOST", cfg.HTTPHost)
	cfg.HTTPPort = getEnvInt("HTTP_PORT", cfg.HTTPPort)
	cfg.HTTPReadTimeout = getEnvDuration("HTTP_READ_TIMEOUT", cfg.HTTPReadTimeout)
	cfg.HTTPWriteTimeout = getEnvDuration("HTTP_WRITE_TIMEOUT", cfg.HTTPWriteTimeout)
	cfg.HTTPIdleTimeout = getEnvDuration("HTTP_IDLE_TIMEOUT", cfg.HTTPIdleTimeout)

	// Shutdown
	cfg.ShutdownTimeout = getEnvDuration("SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout)

	// TLS
	cfg.TLSEnabled = getEnvBool("TLS_ENABLED", cfg.TLSEnabled)
	cfg.TLSCertFile = getEnv("TLS_CERT_FILE", cfg.TLSCertFile)
	cfg.TLSKeyFile = getEnv("TLS_KEY_FILE", cfg.TLSKeyFile)

	// Health Checks
	cfg.HealthEnabled = getEnvBool("HEALTH_ENABLED", cfg.HealthEnabled)
	cfg.HealthHTTPPath = getEnv("HEALTH_HTTP_PATH", cfg.HealthHTTPPath)
	cfg.LivenessHTTPPath = getEnv("LIVENESS_HTTP_PATH", cfg.LivenessHTTPPath)
	cfg.ReadinessHTTPPath = getEnv("READINESS_HTTP_PATH", cfg.ReadinessHTTPPath)

	// CORS
	cfg.CORSEnabled = getEnvBool("CORS_ENABLED", cfg.CORSEnabled)
	cfg.CORSAllowOrigins = getEnvStringSlice("CORS_ALLOW_ORIGINS", cfg.CORSAllowOrigins)
	cfg.CORSAllowMethods = getEnvStringSlice("CORS_ALLOW_METHODS", cfg.CORSAllowMethods)
	cfg.CORSAllowHeaders = getEnvStringSlice("CORS_ALLOW_HEADERS", cfg.CORSAllowHeaders)
	cfg.CORSExposeHeaders = getEnvStringSlice("CORS_EXPOSE_HEADERS", cfg.CORSExposeHeaders)
	cfg.CORSAllowCredentials = getEnvBool("CORS_ALLOW_CREDENTIALS", cfg.CORSAllowCredentials)
	cfg.CORSMaxAge = getEnvInt("CORS_MAX_AGE", cfg.CORSMaxAge)

	// Rate Limiting
	cfg.RateLimitEnabled = getEnvBool("RATE_LIMIT_ENABLED", cfg.RateLimitEnabled)
	cfg.RateLimitRate = getEnvFloat64("RATE_LIMIT_RATE", cfg.RateLimitRate)
	cfg.RateLimitBurst = getEnvInt("RATE_LIMIT_BURST", cfg.RateLimitBurst)
	cfg.RateLimitExpiry = getEnvDuration("RATE_LIMIT_EXPIRY", cfg.RateLimitExpiry)

	// Compression
	cfg.CompressionEnabled = getEnvBool("COMPRESSION_ENABLED", cfg.CompressionEnabled)

	// Request ID
	cfg.RequestIDEnabled = getEnvBool("REQUEST_ID_ENABLED", cfg.RequestIDEnabled)
	cfg.RequestIDHeader = getEnv("REQUEST_ID_HEADER", cfg.RequestIDHeader)

	// Logging
	cfg.LogRequests = getEnvBool("LOG_REQUESTS", cfg.LogRequests)
	cfg.LogSkipPaths = getEnvStringSlice("LOG_SKIP_PATHS", cfg.LogSkipPaths)

	// Debug
	cfg.Debug = getEnvBool("DEBUG", cfg.Debug)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.GRPCPort < 0 || c.GRPCPort > 65535 {
		return fmt.Errorf("invalid GRPC port: %d", c.GRPCPort)
	}

	if c.HTTPPort < 0 || c.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.HTTPPort)
	}

	if c.TLSEnabled {
		if c.TLSCertFile == "" {
			return fmt.Errorf("TLS cert file is required when TLS is enabled")
		}
		if c.TLSKeyFile == "" {
			return fmt.Errorf("TLS key file is required when TLS is enabled")
		}
	}

	if c.RateLimitEnabled {
		if c.RateLimitRate <= 0 {
			return fmt.Errorf("rate limit rate must be positive")
		}
		if c.RateLimitBurst <= 0 {
			return fmt.Errorf("rate limit burst must be positive")
		}
	}

	return nil
}

// GRPCAddr returns the gRPC server address.
func (c *Config) GRPCAddr() string {
	return fmt.Sprintf("%s:%d", c.GRPCHost, c.GRPCPort)
}

// HTTPAddr returns the HTTP server address.
func (c *Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
