package gateway

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDHeader is the default header name for request ID.
const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// RequestIDMiddleware adds a request ID to the request context.
// If a request ID is provided in the header, it uses that; otherwise generates a new one.
func RequestIDMiddleware() Middleware {
	return RequestIDMiddlewareWithHeader(RequestIDHeader)
}

// RequestIDMiddlewareWithHeader creates request ID middleware with a custom header name.
func RequestIDMiddlewareWithHeader(header string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(header)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add to response header
			w.Header().Set(header, requestID)

			// Add to context
			ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestID extracts request ID from the request context.
func GetRequestID(r *http.Request) string {
	return GetRequestIDFromContext(r.Context())
}

// GetRequestIDFromContext extracts request ID from context.
func GetRequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
