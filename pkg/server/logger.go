package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Logger defines the logging interface used by the server package.
// Implement this interface to integrate with your logging framework.
type Logger interface {
	// Debug logs a debug message with optional key-value pairs.
	Debug(msg string, keysAndValues ...interface{})

	// Info logs an info message with optional key-value pairs.
	Info(msg string, keysAndValues ...interface{})

	// Warn logs a warning message with optional key-value pairs.
	Warn(msg string, keysAndValues ...interface{})

	// Error logs an error message with optional key-value pairs.
	Error(msg string, keysAndValues ...interface{})
}

// LogLevel represents the logging level.
type LogLevel int

const (
	// LogLevelDebug logs all messages.
	LogLevelDebug LogLevel = iota
	// LogLevelInfo logs info, warn, and error messages.
	LogLevelInfo
	// LogLevelWarn logs warn and error messages.
	LogLevelWarn
	// LogLevelError logs only error messages.
	LogLevelError
	// LogLevelSilent disables all logging.
	LogLevelSilent
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelSilent:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel.
func ParseLogLevel(s string) LogLevel {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN", "WARNING":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	case "SILENT":
		return LogLevelSilent
	default:
		return LogLevelInfo
	}
}

// NoopLogger is a no-operation logger that discards all log messages.
type NoopLogger struct{}

// Debug implements Logger.
func (NoopLogger) Debug(msg string, keysAndValues ...interface{}) {}

// Info implements Logger.
func (NoopLogger) Info(msg string, keysAndValues ...interface{}) {}

// Warn implements Logger.
func (NoopLogger) Warn(msg string, keysAndValues ...interface{}) {}

// Error implements Logger.
func (NoopLogger) Error(msg string, keysAndValues ...interface{}) {}

// Ensure NoopLogger implements Logger.
var _ Logger = NoopLogger{}

// StdLogger wraps the standard library logger.
type StdLogger struct {
	logger *log.Logger
	level  LogLevel
}

// NewStdLogger creates a new standard library logger.
func NewStdLogger(out io.Writer, level LogLevel) *StdLogger {
	if out == nil {
		out = os.Stdout
	}
	return &StdLogger{
		logger: log.New(out, "", 0),
		level:  level,
	}
}

// Debug implements Logger.
func (l *StdLogger) Debug(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", msg, keysAndValues...)
	}
}

// Info implements Logger.
func (l *StdLogger) Info(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", msg, keysAndValues...)
	}
}

// Warn implements Logger.
func (l *StdLogger) Warn(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", msg, keysAndValues...)
	}
}

// Error implements Logger.
func (l *StdLogger) Error(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelError {
		l.log("ERROR", msg, keysAndValues...)
	}
}

func (l *StdLogger) log(level, msg string, keysAndValues ...interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	kvStr := formatKeyValues(keysAndValues...)
	if kvStr != "" {
		l.logger.Printf("%s [%s] %s %s", timestamp, level, msg, kvStr)
	} else {
		l.logger.Printf("%s [%s] %s", timestamp, level, msg)
	}
}

// Ensure StdLogger implements Logger.
var _ Logger = (*StdLogger)(nil)

// formatKeyValues formats key-value pairs as a string.
func formatKeyValues(keysAndValues ...interface{}) string {
	if len(keysAndValues) == 0 {
		return ""
	}

	var pairs []string
	for i := 0; i < len(keysAndValues); i += 2 {
		key := fmt.Sprintf("%v", keysAndValues[i])
		var value interface{} = ""
		if i+1 < len(keysAndValues) {
			value = keysAndValues[i+1]
		}
		pairs = append(pairs, fmt.Sprintf("%s=%v", key, value))
	}
	return strings.Join(pairs, " ")
}

// DefaultLogger returns a default logger that writes to stdout.
func DefaultLogger() Logger {
	return NewStdLogger(os.Stdout, LogLevelInfo)
}
