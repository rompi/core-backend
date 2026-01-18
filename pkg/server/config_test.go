package server

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"GRPCHost", cfg.GRPCHost, ""},
		{"GRPCPort", cfg.GRPCPort, 9090},
		{"HTTPHost", cfg.HTTPHost, ""},
		{"HTTPPort", cfg.HTTPPort, 8080},
		{"HTTPReadTimeout", cfg.HTTPReadTimeout, 30 * time.Second},
		{"HTTPWriteTimeout", cfg.HTTPWriteTimeout, 30 * time.Second},
		{"HTTPIdleTimeout", cfg.HTTPIdleTimeout, 120 * time.Second},
		{"ShutdownTimeout", cfg.ShutdownTimeout, 30 * time.Second},
		{"TLSEnabled", cfg.TLSEnabled, false},
		{"HealthEnabled", cfg.HealthEnabled, true},
		{"HealthHTTPPath", cfg.HealthHTTPPath, "/health"},
		{"LivenessHTTPPath", cfg.LivenessHTTPPath, "/health/live"},
		{"ReadinessHTTPPath", cfg.ReadinessHTTPPath, "/health/ready"},
		{"CORSEnabled", cfg.CORSEnabled, false},
		{"RateLimitEnabled", cfg.RateLimitEnabled, false},
		{"RateLimitRate", cfg.RateLimitRate, 20.0},
		{"RateLimitBurst", cfg.RateLimitBurst, 40},
		{"CompressionEnabled", cfg.CompressionEnabled, true},
		{"RequestIDEnabled", cfg.RequestIDEnabled, true},
		{"RequestIDHeader", cfg.RequestIDHeader, "X-Request-ID"},
		{"LogRequests", cfg.LogRequests, true},
		{"Debug", cfg.Debug, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid default config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name:    "invalid gRPC port negative",
			modify:  func(c *Config) { c.GRPCPort = -1 },
			wantErr: true,
		},
		{
			name:    "invalid gRPC port too high",
			modify:  func(c *Config) { c.GRPCPort = 70000 },
			wantErr: true,
		},
		{
			name:    "invalid HTTP port negative",
			modify:  func(c *Config) { c.HTTPPort = -1 },
			wantErr: true,
		},
		{
			name:    "invalid HTTP port too high",
			modify:  func(c *Config) { c.HTTPPort = 70000 },
			wantErr: true,
		},
		{
			name: "TLS enabled without cert",
			modify: func(c *Config) {
				c.TLSEnabled = true
				c.TLSKeyFile = "key.pem"
			},
			wantErr: true,
		},
		{
			name: "TLS enabled without key",
			modify: func(c *Config) {
				c.TLSEnabled = true
				c.TLSCertFile = "cert.pem"
			},
			wantErr: true,
		},
		{
			name: "TLS enabled with both files",
			modify: func(c *Config) {
				c.TLSEnabled = true
				c.TLSCertFile = "cert.pem"
				c.TLSKeyFile = "key.pem"
			},
			wantErr: false,
		},
		{
			name: "rate limit enabled with invalid rate",
			modify: func(c *Config) {
				c.RateLimitEnabled = true
				c.RateLimitRate = 0
			},
			wantErr: true,
		},
		{
			name: "rate limit enabled with invalid burst",
			modify: func(c *Config) {
				c.RateLimitEnabled = true
				c.RateLimitBurst = 0
			},
			wantErr: true,
		},
		{
			name: "rate limit enabled with valid values",
			modify: func(c *Config) {
				c.RateLimitEnabled = true
				c.RateLimitRate = 10
				c.RateLimitBurst = 20
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GRPCAddr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{"default", "", 9090, ":9090"},
		{"with host", "localhost", 9090, "localhost:9090"},
		{"custom port", "0.0.0.0", 50051, "0.0.0.0:50051"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{GRPCHost: tt.host, GRPCPort: tt.port}
			if got := cfg.GRPCAddr(); got != tt.want {
				t.Errorf("GRPCAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_HTTPAddr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{"default", "", 8080, ":8080"},
		{"with host", "localhost", 8080, "localhost:8080"},
		{"custom port", "0.0.0.0", 3000, "0.0.0.0:3000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{HTTPHost: tt.host, HTTPPort: tt.port}
			if got := cfg.HTTPAddr(); got != tt.want {
				t.Errorf("HTTPAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Save and restore env
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, kv := range origEnv {
			for i := 0; i < len(kv); i++ {
				if kv[i] == '=' {
					os.Setenv(kv[:i], kv[i+1:])
					break
				}
			}
		}
	}()

	t.Run("defaults when no env", func(t *testing.T) {
		os.Clearenv()
		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.GRPCPort != 9090 {
			t.Errorf("GRPCPort = %d, want 9090", cfg.GRPCPort)
		}
		if cfg.HTTPPort != 8080 {
			t.Errorf("HTTPPort = %d, want 8080", cfg.HTTPPort)
		}
	})

	t.Run("reads env vars", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("GRPC_PORT", "50051")
		os.Setenv("HTTP_PORT", "3000")
		os.Setenv("GRPC_HOST", "localhost")
		os.Setenv("HTTP_HOST", "0.0.0.0")
		os.Setenv("HTTP_READ_TIMEOUT", "10s")
		os.Setenv("SHUTDOWN_TIMEOUT", "60s")
		os.Setenv("TLS_ENABLED", "false")
		os.Setenv("CORS_ENABLED", "true")
		os.Setenv("CORS_ALLOW_ORIGINS", "http://localhost:3000, http://example.com")
		os.Setenv("RATE_LIMIT_ENABLED", "false")
		os.Setenv("DEBUG", "true")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if cfg.GRPCPort != 50051 {
			t.Errorf("GRPCPort = %d, want 50051", cfg.GRPCPort)
		}
		if cfg.HTTPPort != 3000 {
			t.Errorf("HTTPPort = %d, want 3000", cfg.HTTPPort)
		}
		if cfg.GRPCHost != "localhost" {
			t.Errorf("GRPCHost = %s, want localhost", cfg.GRPCHost)
		}
		if cfg.HTTPReadTimeout != 10*time.Second {
			t.Errorf("HTTPReadTimeout = %v, want 10s", cfg.HTTPReadTimeout)
		}
		if cfg.ShutdownTimeout != 60*time.Second {
			t.Errorf("ShutdownTimeout = %v, want 60s", cfg.ShutdownTimeout)
		}
		if !cfg.CORSEnabled {
			t.Error("CORSEnabled = false, want true")
		}
		if len(cfg.CORSAllowOrigins) != 2 {
			t.Errorf("CORSAllowOrigins len = %d, want 2", len(cfg.CORSAllowOrigins))
		}
		if !cfg.Debug {
			t.Error("Debug = false, want true")
		}
	})

	t.Run("validation error", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TLS_ENABLED", "true")
		// Missing cert and key files

		_, err := LoadConfig()
		if err == nil {
			t.Fatal("LoadConfig() expected error for TLS without cert/key")
		}
	})
}

func TestGetEnvHelpers(t *testing.T) {
	// Save and restore env
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, kv := range origEnv {
			for i := 0; i < len(kv); i++ {
				if kv[i] == '=' {
					os.Setenv(kv[:i], kv[i+1:])
					break
				}
			}
		}
	}()

	t.Run("getEnv", func(t *testing.T) {
		os.Clearenv()
		if got := getEnv("MISSING", "default"); got != "default" {
			t.Errorf("getEnv() = %s, want default", got)
		}

		os.Setenv("TEST_VAR", "value")
		if got := getEnv("TEST_VAR", "default"); got != "value" {
			t.Errorf("getEnv() = %s, want value", got)
		}
	})

	t.Run("getEnvInt", func(t *testing.T) {
		os.Clearenv()
		if got := getEnvInt("MISSING", 42); got != 42 {
			t.Errorf("getEnvInt() = %d, want 42", got)
		}

		os.Setenv("TEST_INT", "100")
		if got := getEnvInt("TEST_INT", 42); got != 100 {
			t.Errorf("getEnvInt() = %d, want 100", got)
		}

		os.Setenv("TEST_INT", "invalid")
		if got := getEnvInt("TEST_INT", 42); got != 42 {
			t.Errorf("getEnvInt() = %d, want 42 for invalid", got)
		}
	})

	t.Run("getEnvFloat64", func(t *testing.T) {
		os.Clearenv()
		if got := getEnvFloat64("MISSING", 3.14); got != 3.14 {
			t.Errorf("getEnvFloat64() = %f, want 3.14", got)
		}

		os.Setenv("TEST_FLOAT", "2.5")
		if got := getEnvFloat64("TEST_FLOAT", 3.14); got != 2.5 {
			t.Errorf("getEnvFloat64() = %f, want 2.5", got)
		}

		os.Setenv("TEST_FLOAT", "invalid")
		if got := getEnvFloat64("TEST_FLOAT", 3.14); got != 3.14 {
			t.Errorf("getEnvFloat64() = %f, want 3.14 for invalid", got)
		}
	})

	t.Run("getEnvBool", func(t *testing.T) {
		os.Clearenv()
		if got := getEnvBool("MISSING", true); !got {
			t.Error("getEnvBool() = false, want true")
		}

		os.Setenv("TEST_BOOL", "false")
		if got := getEnvBool("TEST_BOOL", true); got {
			t.Error("getEnvBool() = true, want false")
		}

		os.Setenv("TEST_BOOL", "1")
		if got := getEnvBool("TEST_BOOL", false); !got {
			t.Error("getEnvBool() = false, want true")
		}

		os.Setenv("TEST_BOOL", "invalid")
		if got := getEnvBool("TEST_BOOL", true); !got {
			t.Error("getEnvBool() = false, want true for invalid")
		}
	})

	t.Run("getEnvDuration", func(t *testing.T) {
		os.Clearenv()
		if got := getEnvDuration("MISSING", time.Minute); got != time.Minute {
			t.Errorf("getEnvDuration() = %v, want 1m", got)
		}

		os.Setenv("TEST_DUR", "30s")
		if got := getEnvDuration("TEST_DUR", time.Minute); got != 30*time.Second {
			t.Errorf("getEnvDuration() = %v, want 30s", got)
		}

		os.Setenv("TEST_DUR", "invalid")
		if got := getEnvDuration("TEST_DUR", time.Minute); got != time.Minute {
			t.Errorf("getEnvDuration() = %v, want 1m for invalid", got)
		}
	})

	t.Run("getEnvStringSlice", func(t *testing.T) {
		os.Clearenv()
		def := []string{"a", "b"}
		if got := getEnvStringSlice("MISSING", def); len(got) != 2 {
			t.Errorf("getEnvStringSlice() len = %d, want 2", len(got))
		}

		os.Setenv("TEST_SLICE", "one, two, three")
		got := getEnvStringSlice("TEST_SLICE", def)
		if len(got) != 3 {
			t.Errorf("getEnvStringSlice() len = %d, want 3", len(got))
		}
		if got[0] != "one" || got[1] != "two" || got[2] != "three" {
			t.Errorf("getEnvStringSlice() = %v, unexpected values", got)
		}

		os.Setenv("TEST_SLICE", "  ,  ,  ") // empty values after trim
		got = getEnvStringSlice("TEST_SLICE", def)
		if len(got) != 2 {
			t.Errorf("getEnvStringSlice() len = %d, want 2 for empty", len(got))
		}
	})
}
