package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	middleware := RequestIDMiddleware()

	t.Run("generates request ID when not provided", func(t *testing.T) {
		var capturedID string

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedID = GetRequestID(r)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if capturedID == "" {
			t.Error("request ID should be generated")
		}

		// Check response header
		responseID := rr.Header().Get(RequestIDHeader)
		if responseID == "" {
			t.Error("response should have X-Request-ID header")
		}
		if responseID != capturedID {
			t.Errorf("response ID %q != captured ID %q", responseID, capturedID)
		}
	})

	t.Run("uses provided request ID", func(t *testing.T) {
		providedID := "custom-request-id-123"
		var capturedID string

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedID = GetRequestID(r)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(RequestIDHeader, providedID)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if capturedID != providedID {
			t.Errorf("capturedID = %q, want %q", capturedID, providedID)
		}

		responseID := rr.Header().Get(RequestIDHeader)
		if responseID != providedID {
			t.Errorf("responseID = %q, want %q", responseID, providedID)
		}
	})
}

func TestRequestIDMiddlewareWithHeader(t *testing.T) {
	customHeader := "X-Custom-Request-ID"
	middleware := RequestIDMiddlewareWithHeader(customHeader)

	t.Run("generates request ID with custom header", func(t *testing.T) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		responseID := rr.Header().Get(customHeader)
		if responseID == "" {
			t.Errorf("response should have %s header", customHeader)
		}

		// Default header should not be set
		defaultID := rr.Header().Get(RequestIDHeader)
		if defaultID != "" && customHeader != RequestIDHeader {
			t.Errorf("default header should not be set")
		}
	})

	t.Run("uses provided custom header", func(t *testing.T) {
		providedID := "my-custom-id"

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(customHeader, providedID)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		responseID := rr.Header().Get(customHeader)
		if responseID != providedID {
			t.Errorf("responseID = %q, want %q", responseID, providedID)
		}
	})
}

func TestGetRequestID(t *testing.T) {
	t.Run("returns empty string when not set", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		id := GetRequestID(req)
		if id != "" {
			t.Errorf("GetRequestID() = %q, want empty string", id)
		}
	})

	t.Run("returns ID from context", func(t *testing.T) {
		expectedID := "test-id-123"
		ctx := context.WithValue(context.Background(), requestIDKey{}, expectedID)
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

		id := GetRequestID(req)
		if id != expectedID {
			t.Errorf("GetRequestID() = %q, want %q", id, expectedID)
		}
	})
}

func TestGetRequestIDFromContext(t *testing.T) {
	t.Run("returns empty string for empty context", func(t *testing.T) {
		id := GetRequestIDFromContext(context.Background())
		if id != "" {
			t.Errorf("GetRequestIDFromContext() = %q, want empty string", id)
		}
	})

	t.Run("returns ID from context", func(t *testing.T) {
		expectedID := "ctx-id-456"
		ctx := context.WithValue(context.Background(), requestIDKey{}, expectedID)

		id := GetRequestIDFromContext(ctx)
		if id != expectedID {
			t.Errorf("GetRequestIDFromContext() = %q, want %q", id, expectedID)
		}
	})

	t.Run("returns empty for wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), requestIDKey{}, 12345)

		id := GetRequestIDFromContext(ctx)
		if id != "" {
			t.Errorf("GetRequestIDFromContext() = %q, want empty string for wrong type", id)
		}
	})
}

func TestRequestIDHeader_Constant(t *testing.T) {
	if RequestIDHeader != "X-Request-ID" {
		t.Errorf("RequestIDHeader = %q, want %q", RequestIDHeader, "X-Request-ID")
	}
}

func TestRequestID_UUID_Format(t *testing.T) {
	middleware := RequestIDMiddleware()

	var generatedID string
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		generatedID = GetRequestID(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	if len(generatedID) != 36 {
		t.Errorf("generated ID length = %d, want 36 (UUID format)", len(generatedID))
	}

	// Check UUID has hyphens at correct positions
	if generatedID[8] != '-' || generatedID[13] != '-' || generatedID[18] != '-' || generatedID[23] != '-' {
		t.Errorf("generated ID %q is not in UUID format", generatedID)
	}
}
