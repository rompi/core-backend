// Package main demonstrates transaction usage.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/rompi/core-backend/pkg/postgres"
)

func main() {
	cfg, err := postgres.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := postgres.New(*cfg)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example: Create a test table and insert data in a transaction
	err = client.Transaction(ctx, func(tx pgx.Tx) error {
		// Create table if not exists
		_, err := tx.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS tx_example (
				id SERIAL PRIMARY KEY,
				name TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT NOW()
			)
		`)
		if err != nil {
			return fmt.Errorf("create table: %w", err)
		}

		// Insert a row
		_, err = tx.Exec(ctx, "INSERT INTO tx_example (name) VALUES ($1)", "TransactionTest")
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}

		// Query count
		var count int
		err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM tx_example").Scan(&count)
		if err != nil {
			return fmt.Errorf("count: %w", err)
		}

		fmt.Printf("Rows in tx_example: %d\n", count)
		return nil
	})

	if err != nil {
		log.Fatalf("Transaction failed: %v", err)
	}

	fmt.Println("Transaction completed successfully!")

	// Example: Transaction with rollback on error
	err = client.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO tx_example (name) VALUES ($1)", "WillRollback")
		if err != nil {
			return err
		}

		// Simulate an error - this will cause rollback
		return fmt.Errorf("simulated error")
	})

	if err != nil {
		fmt.Printf("Transaction rolled back as expected: %v\n", err)
	}
}
