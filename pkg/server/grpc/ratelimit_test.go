package grpc

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/rompi/core-backend/pkg/server"
)

func TestDefaultRateLimitConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()

	if cfg.Rate != 20 {
		t.Errorf("Rate = %f, want 20", cfg.Rate)
	}
	if cfg.Burst != 40 {
		t.Errorf("Burst = %d, want 40", cfg.Burst)
	}
	if cfg.KeyFunc != nil {
		t.Error("KeyFunc should be nil by default")
	}
	if cfg.SkipFunc != nil {
		t.Error("SkipFunc should be nil by default")
	}
}

func TestRateLimitInterceptor(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		config := RateLimitConfig{Rate: 100, Burst: 10}
		interceptor := RateLimitInterceptor(config)

		ctx := contextWithPeer("192.168.1.1:1234")
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		for i := 0; i < 5; i++ {
			result, err := interceptor(ctx, "request", info, handler)
			if err != nil {
				t.Fatalf("request %d error = %v", i, err)
			}
			if result != "response" {
				t.Errorf("request %d result = %v, want response", i, result)
			}
		}
	})

	t.Run("denies requests exceeding limit", func(t *testing.T) {
		config := RateLimitConfig{Rate: 1, Burst: 1}
		interceptor := RateLimitInterceptor(config)

		ctx := contextWithPeer("10.0.0.1:5678")
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		// First request should succeed
		_, err := interceptor(ctx, "request", info, handler)
		if err != nil {
			t.Fatalf("first request error = %v", err)
		}

		// Second request should be rate limited
		_, err = interceptor(ctx, "request", info, handler)
		if err == nil {
			t.Fatal("second request should be rate limited")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected gRPC status error, got %v", err)
		}
		if st.Code() != codes.ResourceExhausted {
			t.Errorf("code = %v, want %v", st.Code(), codes.ResourceExhausted)
		}
	})

	t.Run("with custom key function", func(t *testing.T) {
		keyFuncCalled := false
		config := RateLimitConfig{
			Rate:  100,
			Burst: 10,
			KeyFunc: func(ctx context.Context, method string) string {
				keyFuncCalled = true
				return "custom-key"
			},
		}
		interceptor := RateLimitInterceptor(config)

		ctx := context.Background()
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		_, err := interceptor(ctx, "request", info, handler)
		if err != nil {
			t.Fatalf("request error = %v", err)
		}

		if !keyFuncCalled {
			t.Error("custom key function should be called")
		}
	})

	t.Run("with skip function", func(t *testing.T) {
		config := RateLimitConfig{
			Rate:  1,
			Burst: 1,
			SkipFunc: func(ctx context.Context, method string) bool {
				return method == "/test.Service/SkippedMethod"
			},
		}
		interceptor := RateLimitInterceptor(config)

		ctx := contextWithPeer("10.0.0.2:5678")
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/SkippedMethod"}

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		// Multiple requests should all succeed when skipped
		for i := 0; i < 5; i++ {
			_, err := interceptor(ctx, "request", info, handler)
			if err != nil {
				t.Fatalf("request %d should be skipped, got error = %v", i, err)
			}
		}
	})
}

func TestRateLimitPresets(t *testing.T) {
	tests := []struct {
		name       string
		preset     func() grpc.UnaryServerInterceptor
		wantRate   float64
		wantBurst  int
	}{
		{"RateLimitStrict", RateLimitStrict, 5, 10},
		{"RateLimitAuth", RateLimitAuth, 10, 20},
		{"RateLimitNormal", RateLimitNormal, 30, 60},
		{"RateLimitRelaxed", RateLimitRelaxed, 100, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := tt.preset()
			if interceptor == nil {
				t.Errorf("%s() returned nil", tt.name)
			}
		})
	}
}

func TestRateLimitPerMethod(t *testing.T) {
	limits := PerMethodRateLimits{
		"/test.Service/StrictMethod": {Rate: 1, Burst: 1},
		"/test.Service/RelaxedMethod": {Rate: 100, Burst: 100},
	}
	defaultConfig := RateLimitConfig{Rate: 10, Burst: 10}

	interceptor := RateLimitPerMethod(limits, defaultConfig)

	t.Run("applies per-method limit", func(t *testing.T) {
		ctx := contextWithPeer("10.0.0.3:1234")
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/StrictMethod"}

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		// First request succeeds
		_, err := interceptor(ctx, "request", info, handler)
		if err != nil {
			t.Fatalf("first request error = %v", err)
		}

		// Second request should be rate limited
		_, err = interceptor(ctx, "request", info, handler)
		if err == nil {
			t.Fatal("second request should be rate limited")
		}
	})

	t.Run("uses default for unknown methods", func(t *testing.T) {
		ctx := contextWithPeer("10.0.0.4:1234")
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/UnknownMethod"}

		handler := func(ctx context.Context, req any) (any, error) {
			return "response", nil
		}

		// Should allow multiple requests with default config
		for i := 0; i < 5; i++ {
			_, err := interceptor(ctx, "request", info, handler)
			if err != nil {
				t.Fatalf("request %d error = %v", i, err)
			}
		}
	})
}

func TestDefaultKeyFunc(t *testing.T) {
	t.Run("returns peer address", func(t *testing.T) {
		ctx := contextWithPeer("192.168.1.100:9999")

		key := defaultKeyFunc(ctx, "/test.Method")
		if key != "192.168.1.100:9999" {
			t.Errorf("key = %q, want %q", key, "192.168.1.100:9999")
		}
	})

	t.Run("returns unknown when no peer", func(t *testing.T) {
		ctx := context.Background()

		key := defaultKeyFunc(ctx, "/test.Method")
		if key != "unknown" {
			t.Errorf("key = %q, want %q", key, "unknown")
		}
	})
}

func TestMethodKeyFunc(t *testing.T) {
	keyFunc := MethodKeyFunc()

	t.Run("includes method name", func(t *testing.T) {
		ctx := contextWithPeer("10.0.0.1:1234")

		key := keyFunc(ctx, "/test.Service/Method")
		if key != "10.0.0.1:1234:/test.Service/Method" {
			t.Errorf("key = %q", key)
		}
	})

	t.Run("handles missing peer", func(t *testing.T) {
		ctx := context.Background()

		key := keyFunc(ctx, "/test.Service/Method")
		if key != "unknown:/test.Service/Method" {
			t.Errorf("key = %q", key)
		}
	})
}

func TestUserKeyFunc(t *testing.T) {
	keyFunc := UserKeyFunc()

	t.Run("uses user ID when authenticated", func(t *testing.T) {
		user := &server.DefaultUser{ID: "user-123"}
		ctx := server.ContextWithUser(context.Background(), user)

		key := keyFunc(ctx, "/test.Method")
		if key != "user:user-123" {
			t.Errorf("key = %q, want %q", key, "user:user-123")
		}
	})

	t.Run("falls back to peer IP when not authenticated", func(t *testing.T) {
		ctx := contextWithPeer("192.168.1.1:5000")

		key := keyFunc(ctx, "/test.Method")
		if key != "ip:192.168.1.1:5000" {
			t.Errorf("key = %q, want %q", key, "ip:192.168.1.1:5000")
		}
	})

	t.Run("returns unknown when no user and no peer", func(t *testing.T) {
		ctx := context.Background()

		key := keyFunc(ctx, "/test.Method")
		if key != "unknown" {
			t.Errorf("key = %q, want %q", key, "unknown")
		}
	})
}

func TestRateLimiterStore(t *testing.T) {
	t.Run("creates new limiters", func(t *testing.T) {
		store := newRateLimiterStore(10, 20, 0)

		limiter := store.getLimiter("key1")
		if limiter == nil {
			t.Error("getLimiter should return a limiter")
		}
	})

	t.Run("reuses existing limiters", func(t *testing.T) {
		store := newRateLimiterStore(10, 20, 0)

		limiter1 := store.getLimiter("key1")
		limiter2 := store.getLimiter("key1")

		if limiter1 != limiter2 {
			t.Error("should return same limiter for same key")
		}
	})

	t.Run("different keys get different limiters", func(t *testing.T) {
		store := newRateLimiterStore(10, 20, 0)

		limiter1 := store.getLimiter("key1")
		limiter2 := store.getLimiter("key2")

		if limiter1 == limiter2 {
			t.Error("different keys should get different limiters")
		}
	})
}

// Helper functions

func contextWithPeer(addr string) context.Context {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	p := &peer.Peer{
		Addr: tcpAddr,
	}
	return peer.NewContext(context.Background(), p)
}

