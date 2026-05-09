package redis

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.Host)
	}
	if cfg.Port != 6379 {
		t.Errorf("expected 6379, got %d", cfg.Port)
	}
	if cfg.PoolSize != 10 {
		t.Errorf("expected 10, got %d", cfg.PoolSize)
	}
	if cfg.ConnectTimeout != 5*time.Second {
		t.Errorf("expected 5s, got %v", cfg.ConnectTimeout)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{"valid", func(c *Config) {}, false},
		{"empty host", func(c *Config) { c.Host = "" }, true},
		{"invalid port", func(c *Config) { c.Port = 0 }, true},
		{"invalid db", func(c *Config) { c.DB = 20 }, true},
		{"invalid pool", func(c *Config) { c.PoolSize = 0 }, true},
		{"cluster without nodes", func(c *Config) { c.ClusterEnabled = true }, true},
		{"sentinel without master", func(c *Config) { c.SentinelEnabled = true }, true},
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

func TestConfigAddress(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Host = "redis.example.com"
	cfg.Port = 6380
	if addr := cfg.Address(); addr != "redis.example.com:6380" {
		t.Errorf("expected redis.example.com:6380, got %s", addr)
	}
}

func TestConfigWithKeyPrefix(t *testing.T) {
	cfg := DefaultConfig()
	if got := cfg.WithKeyPrefix("key"); got != "key" {
		t.Errorf("expected key, got %s", got)
	}
	cfg.KeyPrefix = "app:"
	if got := cfg.WithKeyPrefix("key"); got != "app:key" {
		t.Errorf("expected app:key, got %s", got)
	}
}

func TestLoadConfig(t *testing.T) {
	orig := os.Getenv("REDIS_HOST")
	defer os.Setenv("REDIS_HOST", orig)

	os.Setenv("REDIS_HOST", "testhost")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Host != "testhost" {
		t.Errorf("expected testhost, got %s", cfg.Host)
	}
}
