// Package main demonstrates multi-source configuration with priority merging.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rompi/core-backend/pkg/config"
	"github.com/rompi/core-backend/pkg/config/provider"
)

func main() {
	// Determine config files based on environment
	profile := os.Getenv("APP_PROFILE")
	if profile == "" {
		profile = "development"
	}

	configDir := "./config"

	// Create config with multiple providers
	// Priority (later overrides earlier): default.yaml < {profile}.yaml < env vars
	cfg := config.New(
		// Base defaults
		config.WithProvider(provider.NewFileProvider(
			filepath.Join(configDir, "default.yaml"),
			provider.WithOptional(),
		)),
		// Profile-specific overrides
		config.WithProvider(provider.NewFileProvider(
			filepath.Join(configDir, fmt.Sprintf("%s.yaml", profile)),
			provider.WithOptional(),
		)),
		// Environment variable overrides (highest priority)
		config.WithProvider(provider.NewEnvProvider(
			provider.WithPrefix("APP"),
		)),
	)

	ctx := context.Background()

	// Load configuration from all sources
	if err := cfg.Load(ctx); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Access configuration
	fmt.Printf("Profile: %s\n", profile)
	fmt.Printf("Server Host: %s\n", cfg.GetString("server.host"))
	fmt.Printf("Server Port: %d\n", cfg.GetInt("server.port"))
	fmt.Printf("Database Host: %s\n", cfg.GetString("database.host"))
	fmt.Printf("Log Level: %s\n", cfg.GetString("logging.level"))

	// Show all settings
	fmt.Printf("\nAll settings:\n")
	for k, v := range cfg.AllSettings() {
		fmt.Printf("  %s: %v\n", k, v)
	}
}

/*
Example config/default.yaml:

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

Example config/production.yaml:

server:
  host: "0.0.0.0"
  port: 80

database:
  host: "db.production.internal"
  ssl_mode: "require"
  max_conns: 50

logging:
  level: "warn"

Environment variables (APP_ prefix):
  APP_SERVER_PORT=9090
  APP_DATABASE_HOST=custom-db.example.com
*/
