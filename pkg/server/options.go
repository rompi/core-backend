package server

import (
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// Option is a functional option for configuring the Server.
type Option func(*Server) error

// Middleware is the standard HTTP middleware type.
type Middleware func(http.Handler) http.Handler

// ShutdownHook is called during graceful shutdown.
type ShutdownHook func() error

// --- Configuration Options ---

// WithConfig sets the server configuration.
func WithConfig(cfg *Config) Option {
	return func(s *Server) error {
		if cfg == nil {
			return nil
		}
		s.config = cfg
		return nil
	}
}

// WithLogger sets the server logger.
func WithLogger(logger Logger) Option {
	return func(s *Server) error {
		if logger == nil {
			logger = NoopLogger{}
		}
		s.logger = logger
		return nil
	}
}

// --- gRPC Options ---

// WithGRPCAddr sets the gRPC server address.
func WithGRPCAddr(addr string) Option {
	return func(s *Server) error {
		s.grpcAddr = addr
		return nil
	}
}

// WithUnaryInterceptor adds unary interceptors to the gRPC server.
func WithUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(s *Server) error {
		s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
		return nil
	}
}

// WithStreamInterceptor adds stream interceptors to the gRPC server.
func WithStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(s *Server) error {
		s.streamInterceptors = append(s.streamInterceptors, interceptors...)
		return nil
	}
}

// WithGRPCServerOption adds gRPC server options.
func WithGRPCServerOption(opts ...grpc.ServerOption) Option {
	return func(s *Server) error {
		s.grpcServerOptions = append(s.grpcServerOptions, opts...)
		return nil
	}
}

// --- HTTP/Gateway Options ---

// WithHTTPAddr sets the HTTP server address.
func WithHTTPAddr(addr string) Option {
	return func(s *Server) error {
		s.httpAddr = addr
		return nil
	}
}

// WithHTTPMiddleware adds HTTP middleware to the server.
func WithHTTPMiddleware(middleware ...Middleware) Option {
	return func(s *Server) error {
		s.httpMiddleware = append(s.httpMiddleware, middleware...)
		return nil
	}
}

// WithGatewayOption adds gRPC-Gateway ServeMux options.
func WithGatewayOption(opts ...runtime.ServeMuxOption) Option {
	return func(s *Server) error {
		s.gatewayOptions = append(s.gatewayOptions, opts...)
		return nil
	}
}

// --- Feature Options ---

// WithTLS enables TLS for both gRPC and HTTP servers.
func WithTLS(certFile, keyFile string) Option {
	return func(s *Server) error {
		s.config.TLSEnabled = true
		s.config.TLSCertFile = certFile
		s.config.TLSKeyFile = keyFile
		return nil
	}
}

// WithCORS enables CORS with the specified origins.
func WithCORS(origins ...string) Option {
	return func(s *Server) error {
		s.config.CORSEnabled = true
		if len(origins) > 0 {
			s.config.CORSAllowOrigins = origins
		}
		return nil
	}
}

// WithRateLimit enables global rate limiting.
func WithRateLimit(rate float64, burst int) Option {
	return func(s *Server) error {
		s.config.RateLimitEnabled = true
		s.config.RateLimitRate = rate
		s.config.RateLimitBurst = burst
		return nil
	}
}

// WithShutdownTimeout sets the graceful shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) error {
		s.config.ShutdownTimeout = timeout
		return nil
	}
}

// WithHealthEnabled enables or disables health checks.
func WithHealthEnabled(enabled bool) Option {
	return func(s *Server) error {
		s.config.HealthEnabled = enabled
		return nil
	}
}

// WithHealthPaths sets custom health check paths.
func WithHealthPaths(health, liveness, readiness string) Option {
	return func(s *Server) error {
		if health != "" {
			s.config.HealthHTTPPath = health
		}
		if liveness != "" {
			s.config.LivenessHTTPPath = liveness
		}
		if readiness != "" {
			s.config.ReadinessHTTPPath = readiness
		}
		return nil
	}
}

// WithCompression enables or disables HTTP compression.
func WithCompression(enabled bool) Option {
	return func(s *Server) error {
		s.config.CompressionEnabled = enabled
		return nil
	}
}

// WithRequestID enables or disables request ID generation.
func WithRequestID(enabled bool) Option {
	return func(s *Server) error {
		s.config.RequestIDEnabled = enabled
		return nil
	}
}

// WithRequestIDHeader sets the request ID header name.
func WithRequestIDHeader(header string) Option {
	return func(s *Server) error {
		s.config.RequestIDHeader = header
		return nil
	}
}

// WithLogging enables or disables request logging.
func WithLogging(enabled bool) Option {
	return func(s *Server) error {
		s.config.LogRequests = enabled
		return nil
	}
}

// WithLogSkipPaths sets paths to skip from logging.
func WithLogSkipPaths(paths ...string) Option {
	return func(s *Server) error {
		s.config.LogSkipPaths = paths
		return nil
	}
}

// WithDebug enables or disables debug mode.
func WithDebug(enabled bool) Option {
	return func(s *Server) error {
		s.config.Debug = enabled
		return nil
	}
}

// --- Auth Options ---

// WithAuthenticator sets the authenticator for the server.
func WithAuthenticator(auth Authenticator) Option {
	return func(s *Server) error {
		s.authenticator = auth
		return nil
	}
}

// --- Shutdown Options ---

// WithShutdownHook adds a shutdown hook to be called during graceful shutdown.
func WithShutdownHook(hook ShutdownHook) Option {
	return func(s *Server) error {
		s.shutdownHooks = append(s.shutdownHooks, hook)
		return nil
	}
}

// --- HTTP Server Options ---

// WithHTTPReadTimeout sets the HTTP read timeout.
func WithHTTPReadTimeout(timeout time.Duration) Option {
	return func(s *Server) error {
		s.config.HTTPReadTimeout = timeout
		return nil
	}
}

// WithHTTPWriteTimeout sets the HTTP write timeout.
func WithHTTPWriteTimeout(timeout time.Duration) Option {
	return func(s *Server) error {
		s.config.HTTPWriteTimeout = timeout
		return nil
	}
}

// WithHTTPIdleTimeout sets the HTTP idle timeout.
func WithHTTPIdleTimeout(timeout time.Duration) Option {
	return func(s *Server) error {
		s.config.HTTPIdleTimeout = timeout
		return nil
	}
}
