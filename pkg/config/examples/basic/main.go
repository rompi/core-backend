// Package main demonstrates basic usage of the config package.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rompi/core-backend/pkg/config"
	"github.com/rompi/core-backend/pkg/config/provider"
)

// AppConfig holds the application configuration.
type AppConfig struct {
	Server   ServerConfig   `config:"server"`
	Database DatabaseConfig `config:"database"`
	Redis    RedisConfig    `config:"redis"`
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Host    string        `config:"host" default:"0.0.0.0"`
	Port    int           `config:"port" default:"8080" env:"PORT"`
	Timeout time.Duration `config:"timeout" default:"30s"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Host     string `config:"host" default:"localhost"`
	Port     int    `config:"port" default:"5432"`
	Name     string `config:"name" required:"true"`
	User     string `config:"user" default:"postgres"`
	Password string `config:"password" sensitive:"true"`
	SSLMode  string `config:"ssl_mode" default:"disable"`
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	URL string `config:"url" env:"REDIS_URL" default:"redis://localhost:6379"`
}

func main() {
	// Create config with environment provider
	cfg := config.New(
		config.WithProvider(provider.NewEnvProvider()),
		config.WithSensitiveKey("database.password"),
	)

	ctx := context.Background()

	// Load configuration
	if err := cfg.Load(ctx); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Bind to struct
	var appCfg AppConfig
	if err := cfg.Bind(&appCfg); err != nil {
		log.Fatalf("Failed to bind config: %v", err)
	}

	// Print configuration
	fmt.Printf("Server: %s:%d (timeout: %s)\n",
		appCfg.Server.Host,
		appCfg.Server.Port,
		appCfg.Server.Timeout,
	)

	fmt.Printf("Database: %s@%s:%d/%s (SSL: %s)\n",
		appCfg.Database.User,
		appCfg.Database.Host,
		appCfg.Database.Port,
		appCfg.Database.Name,
		appCfg.Database.SSLMode,
	)

	fmt.Printf("Redis: %s\n", appCfg.Redis.URL)

	// Direct access
	fmt.Printf("\nDirect access examples:\n")
	fmt.Printf("  server.host: %s\n", cfg.GetString("server.host"))
	fmt.Printf("  server.port: %d\n", cfg.GetInt("server.port"))
	fmt.Printf("  server.timeout: %s\n", cfg.GetDuration("server.timeout"))

	// Sub-configuration
	dbCfg := cfg.Sub("database")
	fmt.Printf("\nSub-configuration:\n")
	fmt.Printf("  database.host: %s\n", dbCfg.GetString("host"))

	// All settings (sensitive values are masked)
	fmt.Printf("\nAll settings: %v\n", cfg.AllSettings())
}
