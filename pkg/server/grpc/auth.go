package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/rompi/core-backend/pkg/server"
)

// AuthInterceptor creates an authentication interceptor.
// Requires valid token, returns Unauthenticated error if invalid.
func AuthInterceptor(auth server.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		token := extractToken(ctx)
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "missing authentication token")
		}

		user, err := auth.ValidateToken(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid authentication token")
		}
		if user == nil {
			return nil, status.Error(codes.Unauthenticated, "invalid authentication token")
		}

		ctx = server.ContextWithUser(ctx, user)
		return handler(ctx, req)
	}
}

// AuthStreamInterceptor creates a streaming auth interceptor.
func AuthStreamInterceptor(auth server.Authenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		token := extractToken(ctx)
		if token == "" {
			return status.Error(codes.Unauthenticated, "missing authentication token")
		}

		user, err := auth.ValidateToken(ctx, token)
		if err != nil {
			return status.Error(codes.Unauthenticated, "invalid authentication token")
		}
		if user == nil {
			return status.Error(codes.Unauthenticated, "invalid authentication token")
		}

		ctx = server.ContextWithUser(ctx, user)
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrapped)
	}
}

// OptionalAuthInterceptor sets user if token present but doesn't require it.
func OptionalAuthInterceptor(auth server.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		token := extractToken(ctx)
		if token != "" {
			user, err := auth.ValidateToken(ctx, token)
			if err == nil && user != nil {
				ctx = server.ContextWithUser(ctx, user)
			}
		}
		return handler(ctx, req)
	}
}

// OptionalAuthStreamInterceptor sets user if token present but doesn't require it (streaming).
func OptionalAuthStreamInterceptor(auth server.Authenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		token := extractToken(ctx)
		if token != "" {
			user, err := auth.ValidateToken(ctx, token)
			if err == nil && user != nil {
				ctx = server.ContextWithUser(ctx, user)
				ss = &wrappedServerStream{
					ServerStream: ss,
					ctx:          ctx,
				}
			}
		}
		return handler(srv, ss)
	}
}

// RequireRoleInterceptor creates an interceptor that requires specific roles.
// Must be used after AuthInterceptor.
func RequireRoleInterceptor(roles ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		user := GetUser(ctx)
		if user == nil {
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}

		for _, role := range roles {
			if user.HasRole(role) {
				return handler(ctx, req)
			}
		}

		return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
	}
}

// RequireRoleStreamInterceptor creates a streaming interceptor that requires specific roles.
func RequireRoleStreamInterceptor(roles ...string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		user := GetUser(ss.Context())
		if user == nil {
			return status.Error(codes.Unauthenticated, "authentication required")
		}

		for _, role := range roles {
			if user.HasRole(role) {
				return handler(srv, ss)
			}
		}

		return status.Error(codes.PermissionDenied, "insufficient permissions")
	}
}

// RequirePermissionInterceptor creates an interceptor that requires specific permissions.
// Must be used after AuthInterceptor.
func RequirePermissionInterceptor(permissions ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		user := GetUser(ctx)
		if user == nil {
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}

		for _, permission := range permissions {
			if !user.HasPermission(permission) {
				return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
			}
		}

		return handler(ctx, req)
	}
}

// RequirePermissionStreamInterceptor creates a streaming interceptor that requires specific permissions.
func RequirePermissionStreamInterceptor(permissions ...string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		user := GetUser(ss.Context())
		if user == nil {
			return status.Error(codes.Unauthenticated, "authentication required")
		}

		for _, permission := range permissions {
			if !user.HasPermission(permission) {
				return status.Error(codes.PermissionDenied, "insufficient permissions")
			}
		}

		return handler(srv, ss)
	}
}

// GetUser extracts authenticated user from context.
func GetUser(ctx context.Context) server.User {
	return server.UserFromContext(ctx)
}

// extractToken extracts the bearer token from gRPC metadata.
func extractToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// Try "authorization" header first
	authHeaders := md.Get("authorization")
	if len(authHeaders) > 0 {
		auth := authHeaders[0]
		// Handle "Bearer <token>" format
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return strings.TrimSpace(auth[7:])
		}
		return auth
	}

	// Try "token" header as fallback
	tokens := md.Get("token")
	if len(tokens) > 0 {
		return tokens[0]
	}

	return ""
}

// SkipAuthMethods creates an interceptor that skips auth for specified methods.
func SkipAuthMethods(auth server.Authenticator, skipMethods ...string) grpc.UnaryServerInterceptor {
	skipMap := make(map[string]bool)
	for _, m := range skipMethods {
		skipMap[m] = true
	}

	authInterceptor := AuthInterceptor(auth)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if skipMap[info.FullMethod] {
			return handler(ctx, req)
		}
		return authInterceptor(ctx, req, info, handler)
	}
}
