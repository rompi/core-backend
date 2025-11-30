package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rompi/core-backend/pkg/auth"
	"github.com/rompi/core-backend/pkg/auth/testutil"
)

func buildMiddlewareService(t *testing.T, configure func(*auth.Config, *auth.Repositories)) (auth.Service, *auth.TokenManager, *auth.User) {
	t.Helper()

	cfg := newTestConfig()
	cfg.JWTSecret = "secret"
	cfg.RateLimitMaxRequests = 100
	cfg.ResetTokenLength = 16
	cfg.RateLimitWindow = time.Minute

	user := &auth.User{ID: "user-1", Email: "user@example.com"}
	users := &testutil.MockUserRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*auth.User, error) {
			if id != user.ID {
				return nil, auth.ErrUserNotFound
			}
			return user, nil
		},
	}

	repos := auth.Repositories{Users: users}
	if configure != nil {
		configure(cfg, &repos)
	}

	svc, err := auth.NewService(cfg, repos)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return svc, auth.NewTokenManager(cfg), user
}

func TestMiddleware_AllowsValidToken(t *testing.T) {
	svc, manager, user := buildMiddlewareService(t, nil)
	token, _, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	handler := svc.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUser := auth.UserFromContext(r.Context())
		if ctxUser == nil {
			http.Error(w, "missing user", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(ctxUser.Email))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != user.Email {
		t.Fatalf("unexpected body %q", rr.Body.String())
	}
}

func TestMiddleware_DeniesWithoutToken(t *testing.T) {
	svc, _, _ := buildMiddlewareService(t, nil)

	handler := svc.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRequireRole(t *testing.T) {
	svc, manager, user := buildMiddlewareService(t, func(cfg *auth.Config, repos *auth.Repositories) {
		repos.Roles = &testutil.MockRoleRepository{
			GetByUserIDFunc: func(ctx context.Context, id string) ([]auth.Role, error) {
				return []auth.Role{{Name: "admin"}}, nil
			},
		}
	})

	token, _, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	protected := svc.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	handler := svc.Middleware()(protected)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	forbidden := svc.RequireRole("owner")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	handler = svc.Middleware()(forbidden)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestRequirePermission(t *testing.T) {
	svc, manager, user := buildMiddlewareService(t, func(cfg *auth.Config, repos *auth.Repositories) {
		repos.Roles = &testutil.MockRoleRepository{
			GetByUserIDFunc: func(ctx context.Context, id string) ([]auth.Role, error) {
				return []auth.Role{{Permissions: []string{"read", "write"}}}, nil
			},
		}
	})

	token, _, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	handler := svc.Middleware()(svc.RequirePermission("write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	handler = svc.Middleware()(svc.RequirePermission("delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	svc, _, _ := buildMiddlewareService(t, func(cfg *auth.Config, repos *auth.Repositories) {
		cfg.RateLimitMaxRequests = 1
		cfg.RateLimitWindow = time.Minute
	})

	handler := svc.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Origin", "origin")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
}
