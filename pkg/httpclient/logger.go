package httpclient

// Logger defines a minimal logging interface that can be implemented
// by any logging framework (slog, zap, logrus, etc.).
// This allows the httpclient to be logging-framework agnostic.
type Logger interface {
	// Debug logs a debug-level message with optional key-value pairs.
	Debug(msg string, keysAndValues ...interface{})

	// Info logs an info-level message with optional key-value pairs.
	Info(msg string, keysAndValues ...interface{})

	// Warn logs a warning-level message with optional key-value pairs.
	Warn(msg string, keysAndValues ...interface{})

	// Error logs an error-level message with optional key-value pairs.
	Error(msg string, keysAndValues ...interface{})
}

// noopLogger is a logger implementation that does nothing.
// This is used as the default logger when no logger is provided.
type noopLogger struct{}

// Debug does nothing.
func (l *noopLogger) Debug(msg string, keysAndValues ...interface{}) {}

// Info does nothing.
func (l *noopLogger) Info(msg string, keysAndValues ...interface{}) {}

// Warn does nothing.
func (l *noopLogger) Warn(msg string, keysAndValues ...interface{}) {}

// Error does nothing.
func (l *noopLogger) Error(msg string, keysAndValues ...interface{}) {}

// NewNoopLogger returns a logger that discards all log messages.
// This is useful for testing or when logging is not desired.
func NewNoopLogger() Logger {
	return &noopLogger{}
}
