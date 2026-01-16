package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// RequestIDKey is the metadata key for request ID.
const RequestIDKey = "x-request-id"

type requestIDKey struct{}

// RequestIDInterceptor adds a request ID to context.
// If a request ID is provided in metadata, it uses that; otherwise generates a new one.
func RequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		requestID := extractOrGenerateRequestID(ctx)
		ctx = context.WithValue(ctx, requestIDKey{}, requestID)

		// Add request ID to outgoing metadata
		ctx = metadata.AppendToOutgoingContext(ctx, RequestIDKey, requestID)

		// Set response header
		if err := grpc.SetHeader(ctx, metadata.Pairs(RequestIDKey, requestID)); err != nil {
			// Log but don't fail
		}

		return handler(ctx, req)
	}
}

// RequestIDStreamInterceptor adds a request ID to streaming context.
func RequestIDStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		requestID := extractOrGenerateRequestID(ctx)
		ctx = context.WithValue(ctx, requestIDKey{}, requestID)

		// Set response header
		if err := grpc.SetHeader(ctx, metadata.Pairs(RequestIDKey, requestID)); err != nil {
			// Log but don't fail
		}

		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrapped)
	}
}

// GetRequestID extracts request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// extractOrGenerateRequestID extracts request ID from metadata or generates a new one.
func extractOrGenerateRequestID(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ids := md.Get(RequestIDKey); len(ids) > 0 && ids[0] != "" {
			return ids[0]
		}
	}
	return uuid.New().String()
}

// wrappedServerStream wraps grpc.ServerStream to provide custom context.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context.
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
