package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rompi/core-backend/pkg/server"
)

// LoggingConfig configures the logging interceptor.
type LoggingConfig struct {
	// Logger for logging requests.
	Logger server.Logger

	// SkipMethods are methods to skip logging (e.g., health checks).
	SkipMethods []string

	// LogPayloads logs request/response payloads (careful with sensitive data).
	LogPayloads bool

	// LogLevel sets the log level for successful requests.
	// Errors are always logged at error level.
	LogLevel server.LogLevel
}

// DefaultLoggingConfig returns default logging configuration.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Logger:   server.NoopLogger{},
		LogLevel: server.LogLevelInfo,
	}
}

// LoggingInterceptor creates a logging interceptor with a logger.
func LoggingInterceptor(logger server.Logger) grpc.UnaryServerInterceptor {
	return LoggingInterceptorWithConfig(LoggingConfig{
		Logger: logger,
	})
}

// LoggingInterceptorWithConfig creates a logging interceptor with config.
func LoggingInterceptorWithConfig(config LoggingConfig) grpc.UnaryServerInterceptor {
	if config.Logger == nil {
		config.Logger = server.NoopLogger{}
	}

	skipMethods := make(map[string]bool)
	for _, m := range config.SkipMethods {
		skipMethods[m] = true
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip logging for certain methods
		if skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		start := time.Now()
		requestID := GetRequestID(ctx)

		// Call handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		// Build log fields
		fields := []interface{}{
			"method", info.FullMethod,
			"code", code.String(),
			"duration", duration.String(),
		}

		if requestID != "" {
			fields = append(fields, "request_id", requestID)
		}

		if config.LogPayloads {
			fields = append(fields, "request", req)
			if resp != nil {
				fields = append(fields, "response", resp)
			}
		}

		// Log based on result
		if err != nil {
			fields = append(fields, "error", err.Error())
			config.Logger.Error("gRPC request failed", fields...)
		} else {
			switch config.LogLevel {
			case server.LogLevelDebug:
				config.Logger.Debug("gRPC request completed", fields...)
			default:
				config.Logger.Info("gRPC request completed", fields...)
			}
		}

		return resp, err
	}
}

// LoggingStreamInterceptor creates a streaming logging interceptor.
func LoggingStreamInterceptor(logger server.Logger) grpc.StreamServerInterceptor {
	return LoggingStreamInterceptorWithConfig(LoggingConfig{
		Logger: logger,
	})
}

// LoggingStreamInterceptorWithConfig creates a streaming logging interceptor with config.
func LoggingStreamInterceptorWithConfig(config LoggingConfig) grpc.StreamServerInterceptor {
	if config.Logger == nil {
		config.Logger = server.NoopLogger{}
	}

	skipMethods := make(map[string]bool)
	for _, m := range config.SkipMethods {
		skipMethods[m] = true
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip logging for certain methods
		if skipMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		start := time.Now()
		requestID := GetRequestID(ss.Context())

		// Call handler
		err := handler(srv, ss)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		// Build log fields
		fields := []interface{}{
			"method", info.FullMethod,
			"code", code.String(),
			"duration", duration.String(),
			"stream", true,
		}

		if requestID != "" {
			fields = append(fields, "request_id", requestID)
		}

		// Log based on result
		if err != nil {
			fields = append(fields, "error", err.Error())
			config.Logger.Error("gRPC stream failed", fields...)
		} else {
			switch config.LogLevel {
			case server.LogLevelDebug:
				config.Logger.Debug("gRPC stream completed", fields...)
			default:
				config.Logger.Info("gRPC stream completed", fields...)
			}
		}

		return err
	}
}
