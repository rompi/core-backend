package postgres

import (
	"context"
	"time"
)

// HealthStatus represents the health status of the PostgreSQL connection.
type HealthStatus struct {
	Healthy     bool          `json:"healthy"`
	Message     string        `json:"message"`
	Latency     time.Duration `json:"latency"`
	ActiveConns int32         `json:"active_conns"`
	IdleConns   int32         `json:"idle_conns"`
	TotalConns  int32         `json:"total_conns"`
}

// Health checks the health of the PostgreSQL connection.
func (c *Client) Health(ctx context.Context) HealthStatus {
	start := time.Now()

	status := HealthStatus{
		Healthy: true,
		Message: "OK",
	}

	// Get pool stats
	stats := c.pool.Stat()
	status.ActiveConns = stats.AcquiredConns()
	status.IdleConns = stats.IdleConns()
	status.TotalConns = stats.TotalConns()

	// Ping the database
	if err := c.pool.Ping(ctx); err != nil {
		status.Healthy = false
		status.Message = err.Error()
	}

	status.Latency = time.Since(start)

	return status
}
