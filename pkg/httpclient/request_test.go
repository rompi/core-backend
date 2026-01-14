package httpclient

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestBuilder_Body(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("expected Content-Type 'text/plain', got %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	body := bytes.NewBufferString("test body content")

	_, err := client.Post(context.Background(), "/test").
		Body(body, "text/plain").
		Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestBuilder_JSONError(t *testing.T) {
	client := NewDefault("http://localhost")

	// Create a channel which cannot be marshaled to JSON
	ch := make(chan int)

	_, err := client.Post(context.Background(), "/test").
		JSON(ch).
		Do()

	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestRequestBuilder_Header(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "value" {
			t.Errorf("expected X-Test 'value', got %q", r.Header.Get("X-Test"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)

	_, err := client.Get(context.Background(), "/test").
		Header("X-Test", "value").
		Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestBuilder_FullURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/test" {
			t.Errorf("expected path '/api/test', got %q", r.URL.Path)
		}
		if r.URL.Query().Get("foo") != "bar" {
			t.Errorf("expected query param foo=bar, got %q", r.URL.Query().Get("foo"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)

	_, err := client.Get(context.Background(), "/api/test").
		Query("foo", "bar").
		Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestErrorReader_Read(t *testing.T) {
	testErr := &errorReader{err: http.ErrBodyReadAfterClose}

	buf := make([]byte, 10)
	n, err := testErr.Read(buf)

	if n != 0 {
		t.Errorf("expected n=0, got %d", n)
	}

	if err != http.ErrBodyReadAfterClose {
		t.Errorf("expected ErrBodyReadAfterClose, got %v", err)
	}
}
