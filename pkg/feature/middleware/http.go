package middleware

import (
	"net/http"

	"github.com/rompi/core-backend/pkg/feature"
)

// ContextExtractor extracts feature context from an HTTP request.
type ContextExtractor func(r *http.Request) *feature.Context

// HTTPConfig configures the HTTP middleware.
type HTTPConfig struct {
	// ContextExtractor extracts the feature context from requests.
	ContextExtractor ContextExtractor

	// Client is the feature flag client.
	Client feature.Client
}

// HTTP creates an HTTP middleware that injects feature context into requests.
func HTTP(client feature.Client, opts ...HTTPOption) func(http.Handler) http.Handler {
	cfg := &HTTPConfig{
		Client:           client,
		ContextExtractor: feature.ContextFromHTTPRequest,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract feature context
			fctx := cfg.ContextExtractor(r)

			// Add feature context to request context
			ctx := feature.WithContext(r.Context(), fctx)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// HTTPOption is a functional option for HTTP middleware.
type HTTPOption func(*HTTPConfig)

// WithContextExtractor sets a custom context extractor.
func WithContextExtractor(extractor ContextExtractor) HTTPOption {
	return func(cfg *HTTPConfig) {
		cfg.ContextExtractor = extractor
	}
}

// FeatureGate creates middleware that gates access based on a flag.
func FeatureGate(client feature.Client, flagKey string, fallback http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if client.Bool(r.Context(), flagKey, false) {
				next.ServeHTTP(w, r)
			} else if fallback != nil {
				fallback.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		})
	}
}

// VariantRouter routes requests to different handlers based on flag variants.
type VariantRouter struct {
	client   feature.Client
	flagKey  string
	handlers map[string]http.Handler
	fallback http.Handler
}

// NewVariantRouter creates a new variant router.
func NewVariantRouter(client feature.Client, flagKey string) *VariantRouter {
	return &VariantRouter{
		client:   client,
		flagKey:  flagKey,
		handlers: make(map[string]http.Handler),
	}
}

// Variant registers a handler for a specific variant value.
func (vr *VariantRouter) Variant(value string, handler http.Handler) *VariantRouter {
	vr.handlers[value] = handler
	return vr
}

// Fallback sets the fallback handler when no variant matches.
func (vr *VariantRouter) Fallback(handler http.Handler) *VariantRouter {
	vr.fallback = handler
	return vr
}

// ServeHTTP implements http.Handler.
func (vr *VariantRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	variant := vr.client.String(r.Context(), vr.flagKey, "")

	if handler, ok := vr.handlers[variant]; ok {
		handler.ServeHTTP(w, r)
		return
	}

	if vr.fallback != nil {
		vr.fallback.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

// Handler returns the variant router as an http.Handler.
func (vr *VariantRouter) Handler() http.Handler {
	return vr
}
