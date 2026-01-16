package gateway

import (
	"net/http"
)

// Middleware is the standard HTTP middleware type.
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middlewares into one.
// The first middleware will be the outermost.
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Apply applies multiple middlewares to a handler.
func Apply(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// Wrap wraps an http.HandlerFunc with middleware.
func Wrap(fn http.HandlerFunc, middlewares ...Middleware) http.Handler {
	return Apply(fn, middlewares...)
}
