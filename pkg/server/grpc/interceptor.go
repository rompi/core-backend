package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// ChainUnaryInterceptors chains multiple unary interceptors into one.
// The first interceptor will be the outermost.
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Build the chain from the end
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chain
			chain = func(ctx context.Context, req interface{}) (interface{}, error) {
				return interceptor(ctx, req, info, next)
			}
		}
		return chain(ctx, req)
	}
}

// ChainStreamInterceptors chains multiple stream interceptors into one.
// The first interceptor will be the outermost.
func ChainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Build the chain from the end
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chain
			chain = func(srv interface{}, ss grpc.ServerStream) error {
				return interceptor(srv, ss, info, next)
			}
		}
		return chain(srv, ss)
	}
}

// UnaryServerInterceptorFunc is a function type for unary interceptors.
type UnaryServerInterceptorFunc = grpc.UnaryServerInterceptor

// StreamServerInterceptorFunc is a function type for stream interceptors.
type StreamServerInterceptorFunc = grpc.StreamServerInterceptor
