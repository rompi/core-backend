package gateway

import (
	"net/http"
	"strings"

	"github.com/rompi/core-backend/pkg/server"
)

// AuthMiddleware creates HTTP auth middleware.
// Requires valid token, returns 401 Unauthorized if invalid.
func AuthMiddleware(auth server.Authenticator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, "missing authentication token", http.StatusUnauthorized)
				return
			}

			user, err := auth.ValidateToken(r.Context(), token)
			if err != nil || user == nil {
				http.Error(w, "invalid authentication token", http.StatusUnauthorized)
				return
			}

			ctx := server.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware sets user if authenticated, but doesn't require it.
func OptionalAuthMiddleware(auth server.Authenticator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token != "" {
				user, err := auth.ValidateToken(r.Context(), token)
				if err == nil && user != nil {
					ctx := server.ContextWithUser(r.Context(), user)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleMiddleware requires specific roles.
// Must be used after AuthMiddleware.
func RequireRoleMiddleware(roles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r)
			if user == nil {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			for _, role := range roles {
				if user.HasRole(role) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "insufficient permissions", http.StatusForbidden)
		})
	}
}

// RequirePermissionMiddleware requires specific permissions.
// Must be used after AuthMiddleware.
func RequirePermissionMiddleware(permissions ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r)
			if user == nil {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			for _, permission := range permissions {
				if !user.HasPermission(permission) {
					http.Error(w, "insufficient permissions", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUser extracts authenticated user from request context.
func GetUser(r *http.Request) server.User {
	return server.UserFromRequest(r)
}

// extractToken extracts the bearer token from the request.
func extractToken(r *http.Request) string {
	// Try Authorization header first
	auth := r.Header.Get("Authorization")
	if auth != "" {
		// Handle "Bearer <token>" format
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return strings.TrimSpace(auth[7:])
		}
		return auth
	}

	// Try query parameter as fallback
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	// Try cookie as fallback
	if cookie, err := r.Cookie("token"); err == nil {
		return cookie.Value
	}

	return ""
}

// SkipAuthPaths creates middleware that skips auth for specified paths.
func SkipAuthPaths(auth server.Authenticator, skipPaths ...string) Middleware {
	skipMap := make(map[string]bool)
	for _, p := range skipPaths {
		skipMap[p] = true
	}

	authMiddleware := AuthMiddleware(auth)

	return func(next http.Handler) http.Handler {
		authHandler := authMiddleware(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipMap[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			authHandler.ServeHTTP(w, r)
		})
	}
}
