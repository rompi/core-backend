package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client is the main PostgreSQL client with connection pooling.
// It is safe for concurrent use.
type Client struct {
	pool      *pgxpool.Pool
	config    *Config
	logger    Logger
	queryHook QueryHook
}

// PoolStats contains connection pool statistics.
type PoolStats struct {
	AcquireCount         int64
	AcquireDuration      time.Duration
	AcquiredConns        int32
	CanceledAcquireCount int64
	ConstructingConns    int32
	EmptyAcquireCount    int64
	IdleConns            int32
	MaxConns             int32
	TotalConns           int32
}

// New creates a new PostgreSQL client with the provided configuration.
func New(cfg Config, opts ...Option) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := &Client{
		config: &cfg,
		logger: NewNoopLogger(),
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Configure pool
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionURL())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = time.Minute

	// Set connect timeout
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	// Create pool with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	client.pool = pool
	client.logger.Info("postgres client connected",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database,
		"schema", cfg.Schema,
	)

	return client, nil
}

// NewFromURL creates a new PostgreSQL client from a connection URL.
func NewFromURL(url string, opts ...Option) (*Client, error) {
	client := &Client{
		logger: NewNoopLogger(),
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	client.pool = pool
	client.logger.Info("postgres client connected from URL")

	return client, nil
}

// Pool returns the underlying connection pool.
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// Query executes a query that returns rows.
func (c *Client) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if c.queryHook != nil {
		c.queryHook.BeforeQuery(sql, args)
	}

	rows, err := c.pool.Query(ctx, sql, args...)

	if c.queryHook != nil {
		c.queryHook.AfterQuery(sql, args, err)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	return rows, nil
}

// QueryRow executes a query that returns at most one row.
func (c *Client) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if c.queryHook != nil {
		c.queryHook.BeforeQuery(sql, args)
		defer c.queryHook.AfterQuery(sql, args, nil)
	}

	return c.pool.QueryRow(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows.
func (c *Client) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if c.queryHook != nil {
		c.queryHook.BeforeQuery(sql, args)
	}

	tag, err := c.pool.Exec(ctx, sql, args...)

	if c.queryHook != nil {
		c.queryHook.AfterQuery(sql, args, err)
	}

	if err != nil {
		return tag, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	return tag, nil
}

// Transaction executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (c *Client) Transaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return c.TransactionWithOptions(ctx, pgx.TxOptions{}, fn)
}

// TransactionWithOptions executes a function within a transaction with custom options.
func (c *Client) TransactionWithOptions(ctx context.Context, opts pgx.TxOptions, fn func(tx pgx.Tx) error) error {
	tx, err := c.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("%w: begin transaction: %v", ErrQueryFailed, err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			c.logger.Error("rollback failed", "error", rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: commit transaction: %v", ErrQueryFailed, err)
	}

	return nil
}

// Ping checks if the database is reachable.
func (c *Client) Ping(ctx context.Context) error {
	if err := c.pool.Ping(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	return nil
}

// Close closes the connection pool.
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
		c.logger.Info("postgres client closed")
	}
}

// Stats returns connection pool statistics.
func (c *Client) Stats() *PoolStats {
	stat := c.pool.Stat()
	return &PoolStats{
		AcquireCount:         stat.AcquireCount(),
		AcquireDuration:      stat.AcquireDuration(),
		AcquiredConns:        stat.AcquiredConns(),
		CanceledAcquireCount: stat.CanceledAcquireCount(),
		ConstructingConns:    stat.ConstructingConns(),
		EmptyAcquireCount:    stat.EmptyAcquireCount(),
		IdleConns:            stat.IdleConns(),
		MaxConns:             stat.MaxConns(),
		TotalConns:           stat.TotalConns(),
	}
}
