package postgres

// Logger defines the logging interface for the postgres package.
// It follows a structured logging approach with key-value pairs.
type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

// NoopLogger is a logger that does nothing.
type NoopLogger struct{}

// NewNoopLogger creates a new no-op logger.
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

// Debug does nothing.
func (l *NoopLogger) Debug(msg string, keysAndValues ...any) {}

// Info does nothing.
func (l *NoopLogger) Info(msg string, keysAndValues ...any) {}

// Warn does nothing.
func (l *NoopLogger) Warn(msg string, keysAndValues ...any) {}

// Error does nothing.
func (l *NoopLogger) Error(msg string, keysAndValues ...any) {}
