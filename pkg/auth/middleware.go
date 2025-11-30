package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey contextKey = "auth-user"

// UserFromContext extracts the authenticated user stored by Middleware.
func UserFromContext(ctx context.Context) *User {
	if ctx == nil {
		return nil
	}
	if user, ok := ctx.Value(userContextKey).(*User); ok {
		return user
	}
	return nil
}

// Middleware validates JWT bearer tokens and injects the user into the request context.
func (s *service) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" {
				http.Error(w, "authorization header is required", http.StatusUnauthorized)
				return
			}
			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			user, err := s.ValidateToken(r.Context(), parts[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
		})
	}
}

// RequireRole allows requests only for users that have any of the provided roles.
func (s *service) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			userRoles, err := s.GetUserRoles(r.Context(), user.ID)
			if err != nil {
				http.Error(w, "failed to load roles", http.StatusInternalServerError)
				return
			}
			if !hasRole(userRoles, roles) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission allows requests only if the authenticated user has every permission.
func (s *service) RequirePermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			for _, permission := range permissions {
				has, err := s.CheckPermission(r.Context(), user.ID, permission)
				if err != nil {
					http.Error(w, "failed to check permission", http.StatusInternalServerError)
					return
				}
				if !has {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware applies the configured rate limit per request origin.
func (s *service) RateLimitMiddleware() func(http.Handler) http.Handler {
	if s.limiter == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := strings.TrimSpace(r.Header.Get("X-Request-Origin"))
			if key == "" {
				key = r.RemoteAddr
			}
			if err := s.rateLimit(r.Context(), fmt.Sprintf("middleware:%s", key)); err != nil {
				http.Error(w, ErrRateLimitExceeded.Error(), http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasRole(userRoles []Role, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	for _, role := range userRoles {
		for _, candidate := range allowed {
			if strings.EqualFold(role.Name, candidate) {
				return true
			}
		}
	}
	return false
}
