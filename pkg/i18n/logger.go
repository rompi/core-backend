package i18n

// Logger defines the logging interface used by the i18n package.
type Logger interface {
	// Debug logs a debug message with optional key-value pairs.
	Debug(msg string, keysAndValues ...interface{})

	// Info logs an informational message with optional key-value pairs.
	Info(msg string, keysAndValues ...interface{})

	// Warn logs a warning message with optional key-value pairs.
	Warn(msg string, keysAndValues ...interface{})

	// Error logs an error message with optional key-value pairs.
	Error(msg string, keysAndValues ...interface{})
}

// NoopLogger is a logger that discards all log messages.
type NoopLogger struct{}

// NewNoopLogger creates a new NoopLogger instance.
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

// Debug does nothing.
func (l *NoopLogger) Debug(msg string, keysAndValues ...interface{}) {}

// Info does nothing.
func (l *NoopLogger) Info(msg string, keysAndValues ...interface{}) {}

// Warn does nothing.
func (l *NoopLogger) Warn(msg string, keysAndValues ...interface{}) {}

// Error does nothing.
func (l *NoopLogger) Error(msg string, keysAndValues ...interface{}) {}
