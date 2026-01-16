package grpc

import (
	"context"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rompi/core-backend/pkg/server"
)

// RecoveryConfig configures the recovery interceptor.
type RecoveryConfig struct {
	// Logger for logging panics.
	Logger server.Logger

	// RecoveryFunc is a custom recovery handler.
	// If nil, a default handler is used.
	RecoveryFunc func(ctx context.Context, p interface{}) error

	// EnableStack includes stack trace in logs.
	EnableStack bool
}

// DefaultRecoveryConfig returns default recovery configuration.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		Logger:      server.NoopLogger{},
		EnableStack: true,
	}
}

// RecoveryInterceptor creates a panic recovery interceptor with a logger.
func RecoveryInterceptor(logger server.Logger) grpc.UnaryServerInterceptor {
	return RecoveryInterceptorWithConfig(RecoveryConfig{
		Logger:      logger,
		EnableStack: true,
	})
}

// RecoveryInterceptorWithConfig creates a panic recovery interceptor with config.
func RecoveryInterceptorWithConfig(config RecoveryConfig) grpc.UnaryServerInterceptor {
	if config.Logger == nil {
		config.Logger = server.NoopLogger{}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if p := recover(); p != nil {
				// Log the panic
				if config.EnableStack {
					config.Logger.Error("panic recovered",
						"panic", p,
						"method", info.FullMethod,
						"stack", string(debug.Stack()),
					)
				} else {
					config.Logger.Error("panic recovered",
						"panic", p,
						"method", info.FullMethod,
					)
				}

				// Use custom recovery function if provided
				if config.RecoveryFunc != nil {
					err = config.RecoveryFunc(ctx, p)
				} else {
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}
		}()

		return handler(ctx, req)
	}
}

// RecoveryStreamInterceptor creates a streaming panic recovery interceptor.
func RecoveryStreamInterceptor(logger server.Logger) grpc.StreamServerInterceptor {
	return RecoveryStreamInterceptorWithConfig(RecoveryConfig{
		Logger:      logger,
		EnableStack: true,
	})
}

// RecoveryStreamInterceptorWithConfig creates a streaming panic recovery interceptor with config.
func RecoveryStreamInterceptorWithConfig(config RecoveryConfig) grpc.StreamServerInterceptor {
	if config.Logger == nil {
		config.Logger = server.NoopLogger{}
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if p := recover(); p != nil {
				// Log the panic
				if config.EnableStack {
					config.Logger.Error("panic recovered",
						"panic", p,
						"method", info.FullMethod,
						"stack", string(debug.Stack()),
					)
				} else {
					config.Logger.Error("panic recovered",
						"panic", p,
						"method", info.FullMethod,
					)
				}

				// Use custom recovery function if provided
				if config.RecoveryFunc != nil {
					err = config.RecoveryFunc(ss.Context(), p)
				} else {
					err = status.Errorf(codes.Internal, "internal server error")
				}
			}
		}()

		return handler(srv, ss)
	}
}
