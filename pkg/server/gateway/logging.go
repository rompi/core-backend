package gateway

import (
	"net/http"
	"time"

	"github.com/rompi/core-backend/pkg/server"
)

// LoggingConfig configures the logging middleware.
type LoggingConfig struct {
	// Logger for logging requests.
	Logger server.Logger

	// SkipPaths are paths to skip logging (e.g., health checks).
	SkipPaths []string

	// LogLevel sets the log level for successful requests.
	LogLevel server.LogLevel
}

// DefaultLoggingConfig returns default logging configuration.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Logger:   server.NoopLogger{},
		LogLevel: server.LogLevelInfo,
	}
}

// LoggingMiddleware creates a logging middleware with a logger.
func LoggingMiddleware(logger server.Logger) Middleware {
	return LoggingMiddlewareWithConfig(LoggingConfig{
		Logger: logger,
	})
}

// LoggingMiddlewareWithConfig creates a logging middleware with config.
func LoggingMiddlewareWithConfig(config LoggingConfig) Middleware {
	if config.Logger == nil {
		config.Logger = server.NoopLogger{}
	}

	skipPaths := make(map[string]bool)
	for _, p := range config.SkipPaths {
		skipPaths[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for certain paths
			if skipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			requestID := GetRequestID(r)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call handler
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Build log fields
			fields := []interface{}{
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", duration.String(),
				"size", wrapped.size,
			}

			if requestID != "" {
				fields = append(fields, "request_id", requestID)
			}

			if r.URL.RawQuery != "" {
				fields = append(fields, "query", r.URL.RawQuery)
			}

			// Log based on status code
			if wrapped.statusCode >= 500 {
				config.Logger.Error("HTTP request failed", fields...)
			} else if wrapped.statusCode >= 400 {
				config.Logger.Warn("HTTP request client error", fields...)
			} else {
				switch config.LogLevel {
				case server.LogLevelDebug:
					config.Logger.Debug("HTTP request completed", fields...)
				default:
					config.Logger.Info("HTTP request completed", fields...)
				}
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and size.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader captures the status code.
func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write captures the response size.
func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// Unwrap returns the original ResponseWriter for http.ResponseController.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
