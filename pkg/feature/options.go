package feature

import "time"

// Option is a functional option for configuring the Client.
type Option func(*clientOptions)

// clientOptions holds all optional settings for the Client.
type clientOptions struct {
	overrides       map[string]interface{}
	defaultContext  *Context
	eventHandler    EventHandler
	refreshInterval time.Duration
	logger          Logger
}

// EventHandler receives flag evaluation events.
type EventHandler func(event Event)

// Event represents a flag evaluation event for analytics.
type Event struct {
	// Type is the event type (e.g., "evaluation", "track").
	Type string

	// Key is the flag key.
	Key string

	// Value is the evaluated value.
	Value interface{}

	// Context is the evaluation context.
	Context *Context

	// Timestamp is when the event occurred.
	Timestamp int64

	// Data holds additional event data.
	Data map[string]interface{}
}

// Logger interface for logging.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// defaultClientOptions returns options with sensible defaults.
func defaultClientOptions() *clientOptions {
	return &clientOptions{
		overrides:       make(map[string]interface{}),
		refreshInterval: 30 * time.Second,
		logger:          &noopLogger{},
	}
}

// WithOverrides sets local flag overrides that take precedence.
func WithOverrides(overrides map[string]interface{}) Option {
	return func(o *clientOptions) {
		o.overrides = overrides
	}
}

// WithDefaultContext sets a default context for evaluations.
func WithDefaultContext(ctx *Context) Option {
	return func(o *clientOptions) {
		o.defaultContext = ctx
	}
}

// WithEventHandler sets an event handler for analytics.
func WithEventHandler(handler EventHandler) Option {
	return func(o *clientOptions) {
		o.eventHandler = handler
	}
}

// WithRefreshInterval sets the refresh interval for remote providers.
func WithRefreshInterval(interval time.Duration) Option {
	return func(o *clientOptions) {
		o.refreshInterval = interval
	}
}

// WithLogger sets the logger for the client.
func WithLogger(logger Logger) Option {
	return func(o *clientOptions) {
		o.logger = logger
	}
}

// noopLogger is a no-op implementation of Logger.
type noopLogger struct{}

func (n *noopLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (n *noopLogger) Info(msg string, keysAndValues ...interface{})  {}
func (n *noopLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (n *noopLogger) Error(msg string, keysAndValues ...interface{}) {}
