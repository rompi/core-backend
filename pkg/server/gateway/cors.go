package gateway

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig configures the CORS middleware.
type CORSConfig struct {
	// AllowOrigins is a list of origins that may access the resource.
	AllowOrigins []string

	// AllowMethods is a list of allowed HTTP methods.
	AllowMethods []string

	// AllowHeaders is a list of allowed request headers.
	AllowHeaders []string

	// ExposeHeaders is a list of headers exposed to the client.
	ExposeHeaders []string

	// AllowCredentials indicates whether credentials are allowed.
	AllowCredentials bool

	// MaxAge is the max age for preflight cache in seconds.
	MaxAge int
}

// DefaultCORSConfig returns default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		MaxAge:       86400,
	}
}

// CORSMiddleware creates CORS middleware with the given config.
func CORSMiddleware(config CORSConfig) Middleware {
	// Pre-compute header values
	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(config.MaxAge)

	// Build origin checker
	allowAllOrigins := len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*"
	originMap := make(map[string]bool)
	for _, o := range config.AllowOrigins {
		originMap[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			var allowedOrigin string
			if allowAllOrigins {
				if config.AllowCredentials {
					allowedOrigin = origin
				} else {
					allowedOrigin = "*"
				}
			} else if originMap[origin] {
				allowedOrigin = origin
			}

			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)

				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if exposeHeaders != "" {
					w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
				}
			}

			// Handle preflight request
			if r.Method == http.MethodOptions {
				if allowedOrigin != "" {
					w.Header().Set("Access-Control-Allow-Methods", allowMethods)
					w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
					if config.MaxAge > 0 {
						w.Header().Set("Access-Control-Max-Age", maxAge)
					}
				}
				w.Header().Set("Content-Length", "0")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSWithOrigins creates a simple CORS middleware with specified origins.
func CORSWithOrigins(origins ...string) Middleware {
	config := DefaultCORSConfig()
	if len(origins) > 0 {
		config.AllowOrigins = origins
	}
	return CORSMiddleware(config)
}

// CORSAllowAll creates a CORS middleware that allows all origins.
func CORSAllowAll() Middleware {
	return CORSWithOrigins("*")
}
