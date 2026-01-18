package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestRequestIDInterceptor(t *testing.T) {
	interceptor := RequestIDInterceptor()

	t.Run("generates request ID when not provided", func(t *testing.T) {
		ctx := context.Background()

		var capturedCtx context.Context
		handler := func(ctx context.Context, req any) (any, error) {
			capturedCtx = ctx
			return "response", nil
		}

		_, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{}, handler)
		if err != nil {
			t.Fatalf("interceptor error = %v", err)
		}

		requestID := GetRequestID(capturedCtx)
		if requestID == "" {
			t.Error("request ID should be generated")
		}

		// Verify UUID format (36 characters)
		if len(requestID) != 36 {
			t.Errorf("request ID length = %d, want 36 (UUID format)", len(requestID))
		}
	})

	t.Run("uses provided request ID from metadata", func(t *testing.T) {
		providedID := "custom-request-id-123"
		md := metadata.New(map[string]string{RequestIDKey: providedID})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		var capturedCtx context.Context
		handler := func(ctx context.Context, req any) (any, error) {
			capturedCtx = ctx
			return "response", nil
		}

		_, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{}, handler)
		if err != nil {
			t.Fatalf("interceptor error = %v", err)
		}

		requestID := GetRequestID(capturedCtx)
		if requestID != providedID {
			t.Errorf("request ID = %q, want %q", requestID, providedID)
		}
	})

	t.Run("passes through handler result", func(t *testing.T) {
		ctx := context.Background()

		handler := func(ctx context.Context, req any) (any, error) {
			return "expected-response", nil
		}

		result, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{}, handler)
		if err != nil {
			t.Fatalf("interceptor error = %v", err)
		}

		if result != "expected-response" {
			t.Errorf("result = %v, want expected-response", result)
		}
	})
}

func TestGetRequestID(t *testing.T) {
	t.Run("returns empty string for empty context", func(t *testing.T) {
		id := GetRequestID(context.Background())
		if id != "" {
			t.Errorf("GetRequestID() = %q, want empty string", id)
		}
	})

	t.Run("returns ID from context", func(t *testing.T) {
		expectedID := "test-id-123"
		ctx := context.WithValue(context.Background(), requestIDKey{}, expectedID)

		id := GetRequestID(ctx)
		if id != expectedID {
			t.Errorf("GetRequestID() = %q, want %q", id, expectedID)
		}
	})

	t.Run("returns empty string for wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), requestIDKey{}, 12345)

		id := GetRequestID(ctx)
		if id != "" {
			t.Errorf("GetRequestID() = %q, want empty string for wrong type", id)
		}
	})
}

func TestExtractOrGenerateRequestID(t *testing.T) {
	t.Run("extracts ID from metadata", func(t *testing.T) {
		expectedID := "metadata-id"
		md := metadata.New(map[string]string{RequestIDKey: expectedID})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		id := extractOrGenerateRequestID(ctx)
		if id != expectedID {
			t.Errorf("extractOrGenerateRequestID() = %q, want %q", id, expectedID)
		}
	})

	t.Run("generates ID when not in metadata", func(t *testing.T) {
		ctx := context.Background()

		id := extractOrGenerateRequestID(ctx)
		if id == "" {
			t.Error("should generate a request ID")
		}
		if len(id) != 36 {
			t.Errorf("generated ID length = %d, want 36", len(id))
		}
	})

	t.Run("generates ID when metadata value is empty", func(t *testing.T) {
		md := metadata.New(map[string]string{RequestIDKey: ""})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		id := extractOrGenerateRequestID(ctx)
		if id == "" {
			t.Error("should generate a request ID when metadata is empty")
		}
	})
}

func TestRequestIDKey_Constant(t *testing.T) {
	if RequestIDKey != "x-request-id" {
		t.Errorf("RequestIDKey = %q, want %q", RequestIDKey, "x-request-id")
	}
}

func TestWrappedServerStream(t *testing.T) {
	// Create a mock server stream
	mockStream := &mockServerStream{
		ctx: context.Background(),
	}

	customCtx := context.WithValue(context.Background(), requestIDKey{}, "test-id")
	wrapped := &wrappedServerStream{
		ServerStream: mockStream,
		ctx:          customCtx,
	}

	if wrapped.Context() != customCtx {
		t.Error("wrapped stream should return custom context")
	}

	// Verify request ID is accessible from wrapped context
	id := GetRequestID(wrapped.Context())
	if id != "test-id" {
		t.Errorf("GetRequestID() = %q, want %q", id, "test-id")
	}
}

// mockServerStream is a minimal mock for testing
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}
