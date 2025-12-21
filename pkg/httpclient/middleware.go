package httpclient

import (
	"net/http"
	"time"
)

// Middleware is a function that wraps an http.RoundTripper.
// It can intercept, modify, or log requests and responses.
type Middleware func(next http.RoundTripper) http.RoundTripper

// roundTripperFunc is an adapter to use a function as an http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper.
func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// LoggingMiddleware creates a middleware that logs HTTP requests and responses.
// It logs the method, URL, status code, and duration.
func LoggingMiddleware(logger Logger) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			start := time.Now()

			logger.Info("http request",
				"method", req.Method,
				"url", req.URL.String(),
			)

			resp, err := next.RoundTrip(req)

			duration := time.Since(start)

			if err != nil {
				logger.Error("http request failed",
					"method", req.Method,
					"url", req.URL.String(),
					"duration", duration,
					"error", err,
				)
				return resp, err
			}

			logger.Info("http response",
				"method", req.Method,
				"url", req.URL.String(),
				"status", resp.StatusCode,
				"duration", duration,
			)

			return resp, nil
		})
	}
}

// AuthBearerMiddleware creates a middleware that adds a Bearer token
// to the Authorization header of all requests.
func AuthBearerMiddleware(token string) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("Authorization", "Bearer "+token)
			return next.RoundTrip(req)
		})
	}
}

// AuthAPIKeyMiddleware creates a middleware that adds an API key header
// to all requests.
func AuthAPIKeyMiddleware(headerName, apiKey string) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set(headerName, apiKey)
			return next.RoundTrip(req)
		})
	}
}

// UserAgentMiddleware creates a middleware that sets the User-Agent header
// for all requests.
func UserAgentMiddleware(userAgent string) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("User-Agent", userAgent)
			return next.RoundTrip(req)
		})
	}
}

// HeaderMiddleware creates a middleware that adds custom headers to all requests.
func HeaderMiddleware(headers map[string]string) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			for key, value := range headers {
				req.Header.Set(key, value)
			}
			return next.RoundTrip(req)
		})
	}
}
