// Package main demonstrates health check integration.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rompi/core-backend/pkg/postgres"
)

var client *postgres.Client

func main() {
	cfg, err := postgres.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err = postgres.New(*cfg)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Register health endpoint
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/health/db", dbHealthHandler)

	fmt.Println("Server starting on :8080")
	fmt.Println("Endpoints:")
	fmt.Println("  GET /health    - Overall health")
	fmt.Println("  GET /health/db - Database health")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbHealth := client.Health(ctx)

	response := map[string]any{
		"status":   "ok",
		"database": dbHealth,
	}

	if !dbHealth.Healthy {
		response["status"] = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func dbHealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health := client.Health(ctx)

	if !health.Healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
