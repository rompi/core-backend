package gateway

import (
	"net/http"
	"runtime/debug"

	"github.com/rompi/core-backend/pkg/server"
)

// RecoveryConfig configures the recovery middleware.
type RecoveryConfig struct {
	// Logger for logging panics.
	Logger server.Logger

	// EnableStack includes stack trace in logs.
	EnableStack bool

	// ErrorHandler handles the error response.
	// If nil, returns 500 Internal Server Error.
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err interface{})
}

// DefaultRecoveryConfig returns default recovery configuration.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		Logger:      server.NoopLogger{},
		EnableStack: true,
	}
}

// RecoveryMiddleware creates a panic recovery middleware with a logger.
func RecoveryMiddleware(logger server.Logger) Middleware {
	return RecoveryMiddlewareWithConfig(RecoveryConfig{
		Logger:      logger,
		EnableStack: true,
	})
}

// RecoveryMiddlewareWithConfig creates a panic recovery middleware with config.
func RecoveryMiddlewareWithConfig(config RecoveryConfig) Middleware {
	if config.Logger == nil {
		config.Logger = server.NoopLogger{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic
					if config.EnableStack {
						config.Logger.Error("panic recovered",
							"panic", err,
							"method", r.Method,
							"path", r.URL.Path,
							"stack", string(debug.Stack()),
						)
					} else {
						config.Logger.Error("panic recovered",
							"panic", err,
							"method", r.Method,
							"path", r.URL.Path,
						)
					}

					// Handle error response
					if config.ErrorHandler != nil {
						config.ErrorHandler(w, r, err)
					} else {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
