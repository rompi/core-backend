package auth

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutator func(*Config)
		wantErr bool
	}{
		{
			name: "valid configuration",
			mutator: func(c *Config) {
				c.JWTSecret = "secret"
			},
		},
		{
			name: "missing secret",
			mutator: func(c *Config) {
				c.JWTSecret = ""
			},
			wantErr: true,
		},
		{
			name: "negative expiration",
			mutator: func(c *Config) {
				c.JWTSecret = "secret"
				c.JWTExpirationDuration = -time.Hour
			},
			wantErr: true,
		},
		{
			name: "low bcrypt cost",
			mutator: func(c *Config) {
				c.JWTSecret = "secret"
				c.BcryptCost = 3
			},
			wantErr: true,
		},
		{
			name: "short reset token",
			mutator: func(c *Config) {
				c.JWTSecret = "secret"
				c.ResetTokenLength = 4
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			if tt.mutator != nil {
				tt.mutator(cfg)
			}
			if err := cfg.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
