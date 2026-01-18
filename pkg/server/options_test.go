package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func TestWithConfig(t *testing.T) {
	t.Run("sets config", func(t *testing.T) {
		s := &Server{config: DefaultConfig()}
		cfg := &Config{GRPCPort: 50051, HTTPPort: 3000}

		opt := WithConfig(cfg)
		if err := opt(s); err != nil {
			t.Fatalf("WithConfig() error = %v", err)
		}

		if s.config.GRPCPort != 50051 {
			t.Errorf("GRPCPort = %d, want 50051", s.config.GRPCPort)
		}
		if s.config.HTTPPort != 3000 {
			t.Errorf("HTTPPort = %d, want 3000", s.config.HTTPPort)
		}
	})

	t.Run("nil config is ignored", func(t *testing.T) {
		originalConfig := &Config{GRPCPort: 9090}
		s := &Server{config: originalConfig}

		opt := WithConfig(nil)
		if err := opt(s); err != nil {
			t.Fatalf("WithConfig(nil) error = %v", err)
		}

		// Config should remain unchanged
		if s.config != originalConfig {
			t.Error("config should not be changed when nil is passed")
		}
	})
}

func TestWithLogger(t *testing.T) {
	t.Run("sets logger", func(t *testing.T) {
		s := &Server{}
		logger := NoopLogger{}

		opt := WithLogger(logger)
		if err := opt(s); err != nil {
			t.Fatalf("WithLogger() error = %v", err)
		}

		if s.logger == nil {
			t.Error("logger should be set")
		}
	})

	t.Run("nil logger uses noop", func(t *testing.T) {
		s := &Server{}

		opt := WithLogger(nil)
		if err := opt(s); err != nil {
			t.Fatalf("WithLogger(nil) error = %v", err)
		}

		if s.logger == nil {
			t.Error("logger should default to NoopLogger")
		}
	})
}

func TestWithGRPCAddr(t *testing.T) {
	s := &Server{}

	opt := WithGRPCAddr(":50051")
	if err := opt(s); err != nil {
		t.Fatalf("WithGRPCAddr() error = %v", err)
	}

	if s.grpcAddr != ":50051" {
		t.Errorf("grpcAddr = %q, want %q", s.grpcAddr, ":50051")
	}
}

func TestWithHTTPAddr(t *testing.T) {
	s := &Server{}

	opt := WithHTTPAddr(":8080")
	if err := opt(s); err != nil {
		t.Fatalf("WithHTTPAddr() error = %v", err)
	}

	if s.httpAddr != ":8080" {
		t.Errorf("httpAddr = %q, want %q", s.httpAddr, ":8080")
	}
}

func TestWithUnaryInterceptor(t *testing.T) {
	s := &Server{}

	interceptor1 := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}
	interceptor2 := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ctx, req)
	}

	opt := WithUnaryInterceptor(interceptor1, interceptor2)
	if err := opt(s); err != nil {
		t.Fatalf("WithUnaryInterceptor() error = %v", err)
	}

	if len(s.unaryInterceptors) != 2 {
		t.Errorf("unaryInterceptors len = %d, want 2", len(s.unaryInterceptors))
	}
}

func TestWithStreamInterceptor(t *testing.T) {
	s := &Server{}

	interceptor := func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	}

	opt := WithStreamInterceptor(interceptor)
	if err := opt(s); err != nil {
		t.Fatalf("WithStreamInterceptor() error = %v", err)
	}

	if len(s.streamInterceptors) != 1 {
		t.Errorf("streamInterceptors len = %d, want 1", len(s.streamInterceptors))
	}
}

func TestWithGRPCServerOption(t *testing.T) {
	s := &Server{}

	opt := WithGRPCServerOption(grpc.MaxRecvMsgSize(1024))
	if err := opt(s); err != nil {
		t.Fatalf("WithGRPCServerOption() error = %v", err)
	}

	if len(s.grpcServerOptions) != 1 {
		t.Errorf("grpcServerOptions len = %d, want 1", len(s.grpcServerOptions))
	}
}

func TestWithHTTPMiddleware(t *testing.T) {
	s := &Server{}

	middleware1 := func(next http.Handler) http.Handler {
		return next
	}
	middleware2 := func(next http.Handler) http.Handler {
		return next
	}

	opt := WithHTTPMiddleware(middleware1, middleware2)
	if err := opt(s); err != nil {
		t.Fatalf("WithHTTPMiddleware() error = %v", err)
	}

	if len(s.httpMiddleware) != 2 {
		t.Errorf("httpMiddleware len = %d, want 2", len(s.httpMiddleware))
	}
}

func TestWithGatewayOption(t *testing.T) {
	s := &Server{}

	opt := WithGatewayOption(runtime.WithOutgoingHeaderMatcher(func(s string) (string, bool) {
		return s, true
	}))
	if err := opt(s); err != nil {
		t.Fatalf("WithGatewayOption() error = %v", err)
	}

	if len(s.gatewayOptions) != 1 {
		t.Errorf("gatewayOptions len = %d, want 1", len(s.gatewayOptions))
	}
}

func TestWithTLS(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithTLS("cert.pem", "key.pem")
	if err := opt(s); err != nil {
		t.Fatalf("WithTLS() error = %v", err)
	}

	if !s.config.TLSEnabled {
		t.Error("TLSEnabled should be true")
	}
	if s.config.TLSCertFile != "cert.pem" {
		t.Errorf("TLSCertFile = %q, want %q", s.config.TLSCertFile, "cert.pem")
	}
	if s.config.TLSKeyFile != "key.pem" {
		t.Errorf("TLSKeyFile = %q, want %q", s.config.TLSKeyFile, "key.pem")
	}
}

func TestWithCORS(t *testing.T) {
	t.Run("with origins", func(t *testing.T) {
		s := &Server{config: DefaultConfig()}

		opt := WithCORS("http://example.com", "http://test.com")
		if err := opt(s); err != nil {
			t.Fatalf("WithCORS() error = %v", err)
		}

		if !s.config.CORSEnabled {
			t.Error("CORSEnabled should be true")
		}
		if len(s.config.CORSAllowOrigins) != 2 {
			t.Errorf("CORSAllowOrigins len = %d, want 2", len(s.config.CORSAllowOrigins))
		}
	})

	t.Run("without origins uses default", func(t *testing.T) {
		s := &Server{config: DefaultConfig()}

		opt := WithCORS()
		if err := opt(s); err != nil {
			t.Fatalf("WithCORS() error = %v", err)
		}

		if !s.config.CORSEnabled {
			t.Error("CORSEnabled should be true")
		}
		// Should keep default origins
		if len(s.config.CORSAllowOrigins) != 1 || s.config.CORSAllowOrigins[0] != "*" {
			t.Errorf("CORSAllowOrigins = %v, want [*]", s.config.CORSAllowOrigins)
		}
	})
}

func TestWithRateLimit(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithRateLimit(100.0, 200)
	if err := opt(s); err != nil {
		t.Fatalf("WithRateLimit() error = %v", err)
	}

	if !s.config.RateLimitEnabled {
		t.Error("RateLimitEnabled should be true")
	}
	if s.config.RateLimitRate != 100.0 {
		t.Errorf("RateLimitRate = %f, want 100.0", s.config.RateLimitRate)
	}
	if s.config.RateLimitBurst != 200 {
		t.Errorf("RateLimitBurst = %d, want 200", s.config.RateLimitBurst)
	}
}

func TestWithShutdownTimeout(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithShutdownTimeout(60 * time.Second)
	if err := opt(s); err != nil {
		t.Fatalf("WithShutdownTimeout() error = %v", err)
	}

	if s.config.ShutdownTimeout != 60*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 60s", s.config.ShutdownTimeout)
	}
}

func TestWithHealthEnabled(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithHealthEnabled(false)
	if err := opt(s); err != nil {
		t.Fatalf("WithHealthEnabled() error = %v", err)
	}

	if s.config.HealthEnabled {
		t.Error("HealthEnabled should be false")
	}
}

func TestWithHealthPaths(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithHealthPaths("/custom/health", "/custom/live", "/custom/ready")
	if err := opt(s); err != nil {
		t.Fatalf("WithHealthPaths() error = %v", err)
	}

	if s.config.HealthHTTPPath != "/custom/health" {
		t.Errorf("HealthHTTPPath = %q, want %q", s.config.HealthHTTPPath, "/custom/health")
	}
	if s.config.LivenessHTTPPath != "/custom/live" {
		t.Errorf("LivenessHTTPPath = %q, want %q", s.config.LivenessHTTPPath, "/custom/live")
	}
	if s.config.ReadinessHTTPPath != "/custom/ready" {
		t.Errorf("ReadinessHTTPPath = %q, want %q", s.config.ReadinessHTTPPath, "/custom/ready")
	}
}

func TestWithHealthPaths_EmptyValues(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithHealthPaths("", "", "")
	if err := opt(s); err != nil {
		t.Fatalf("WithHealthPaths() error = %v", err)
	}

	// Empty values should not change defaults
	if s.config.HealthHTTPPath != "/health" {
		t.Errorf("HealthHTTPPath = %q, should remain default", s.config.HealthHTTPPath)
	}
}

func TestWithCompression(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithCompression(false)
	if err := opt(s); err != nil {
		t.Fatalf("WithCompression() error = %v", err)
	}

	if s.config.CompressionEnabled {
		t.Error("CompressionEnabled should be false")
	}
}

func TestWithRequestID(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithRequestID(false)
	if err := opt(s); err != nil {
		t.Fatalf("WithRequestID() error = %v", err)
	}

	if s.config.RequestIDEnabled {
		t.Error("RequestIDEnabled should be false")
	}
}

func TestWithRequestIDHeader(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithRequestIDHeader("X-Custom-ID")
	if err := opt(s); err != nil {
		t.Fatalf("WithRequestIDHeader() error = %v", err)
	}

	if s.config.RequestIDHeader != "X-Custom-ID" {
		t.Errorf("RequestIDHeader = %q, want %q", s.config.RequestIDHeader, "X-Custom-ID")
	}
}

func TestWithLogging(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithLogging(false)
	if err := opt(s); err != nil {
		t.Fatalf("WithLogging() error = %v", err)
	}

	if s.config.LogRequests {
		t.Error("LogRequests should be false")
	}
}

func TestWithLogSkipPaths(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithLogSkipPaths("/health", "/metrics")
	if err := opt(s); err != nil {
		t.Fatalf("WithLogSkipPaths() error = %v", err)
	}

	if len(s.config.LogSkipPaths) != 2 {
		t.Errorf("LogSkipPaths len = %d, want 2", len(s.config.LogSkipPaths))
	}
	if s.config.LogSkipPaths[0] != "/health" || s.config.LogSkipPaths[1] != "/metrics" {
		t.Errorf("LogSkipPaths = %v", s.config.LogSkipPaths)
	}
}

func TestWithDebug(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithDebug(true)
	if err := opt(s); err != nil {
		t.Fatalf("WithDebug() error = %v", err)
	}

	if !s.config.Debug {
		t.Error("Debug should be true")
	}
}

func TestWithAuthenticator(t *testing.T) {
	s := &Server{}
	auth := NoopAuthenticator{}

	opt := WithAuthenticator(auth)
	if err := opt(s); err != nil {
		t.Fatalf("WithAuthenticator() error = %v", err)
	}

	if s.authenticator == nil {
		t.Error("authenticator should be set")
	}
}

func TestWithShutdownHook(t *testing.T) {
	s := &Server{}

	hook1 := func() error { return nil }
	hook2 := func() error { return nil }

	opt1 := WithShutdownHook(hook1)
	if err := opt1(s); err != nil {
		t.Fatalf("WithShutdownHook() error = %v", err)
	}

	opt2 := WithShutdownHook(hook2)
	if err := opt2(s); err != nil {
		t.Fatalf("WithShutdownHook() error = %v", err)
	}

	if len(s.shutdownHooks) != 2 {
		t.Errorf("shutdownHooks len = %d, want 2", len(s.shutdownHooks))
	}
}

func TestWithHTTPReadTimeout(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithHTTPReadTimeout(10 * time.Second)
	if err := opt(s); err != nil {
		t.Fatalf("WithHTTPReadTimeout() error = %v", err)
	}

	if s.config.HTTPReadTimeout != 10*time.Second {
		t.Errorf("HTTPReadTimeout = %v, want 10s", s.config.HTTPReadTimeout)
	}
}

func TestWithHTTPWriteTimeout(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithHTTPWriteTimeout(15 * time.Second)
	if err := opt(s); err != nil {
		t.Fatalf("WithHTTPWriteTimeout() error = %v", err)
	}

	if s.config.HTTPWriteTimeout != 15*time.Second {
		t.Errorf("HTTPWriteTimeout = %v, want 15s", s.config.HTTPWriteTimeout)
	}
}

func TestWithHTTPIdleTimeout(t *testing.T) {
	s := &Server{config: DefaultConfig()}

	opt := WithHTTPIdleTimeout(60 * time.Second)
	if err := opt(s); err != nil {
		t.Fatalf("WithHTTPIdleTimeout() error = %v", err)
	}

	if s.config.HTTPIdleTimeout != 60*time.Second {
		t.Errorf("HTTPIdleTimeout = %v, want 60s", s.config.HTTPIdleTimeout)
	}
}
