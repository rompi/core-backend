// Package main demonstrates basic PostgreSQL connection and querying.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rompi/core-backend/pkg/postgres"
)

func main() {
	// Load configuration from environment variables
	cfg, err := postgres.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create the PostgreSQL client
	client, err := postgres.New(*cfg)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Check connection
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping: %v", err)
	}
	fmt.Println("Connected to PostgreSQL!")

	// Example query - get current timestamp
	var now string
	row := client.QueryRow(ctx, "SELECT NOW()::text")
	if err := row.Scan(&now); err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	fmt.Printf("Current time: %s\n", now)

	// Show pool stats
	stats := client.Stats()
	fmt.Printf("Pool stats: total=%d, idle=%d, acquired=%d\n",
		stats.TotalConns, stats.IdleConns, stats.AcquiredConns)
}
