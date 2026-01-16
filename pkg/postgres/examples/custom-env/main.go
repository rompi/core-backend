// Package main demonstrates connecting with custom environment variables.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rompi/core-backend/pkg/postgres"
)

func main() {
	// Option 1: Create config manually from custom env vars
	cfg := postgres.Config{
		Host:     getEnv("MYAPP_DB_HOST", "localhost"),
		Port:     5432,
		User:     os.Getenv("MYAPP_DB_USER"),
		Password: os.Getenv("MYAPP_DB_PASSWORD"),
		Database: os.Getenv("MYAPP_DB_NAME"),
		Schema:   getEnv("MYAPP_DB_SCHEMA", "public"),
		SSLMode:  getEnv("MYAPP_DB_SSL_MODE", "prefer"),
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	client, err := postgres.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Option 2: Use connection URL from custom env var
	// url := os.Getenv("MYAPP_DATABASE_URL")
	// client, err := postgres.NewFromURL(url)

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping: %v", err)
	}

	fmt.Println("Connected successfully with custom env vars!")

	// Show the connection details (without password)
	fmt.Printf("Connected to: %s:%d/%s (schema: %s)\n",
		cfg.Host, cfg.Port, cfg.Database, cfg.Schema)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
