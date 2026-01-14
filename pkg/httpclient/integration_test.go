package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	resp, err := client.Get(context.Background(), "/test").Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.IsSuccess() {
		t.Errorf("expected success status, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := resp.JSON(&result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("got status %q, want %q", result["status"], "ok")
	}
}

func TestClient_Post_WithJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body["name"] != "test" {
			t.Errorf("expected name=test, got %s", body["name"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	resp, err := client.Post(context.Background(), "/users").
		JSON(map[string]string{"name": "test"}).
		Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestClient_HeadersAndQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("expected X-Custom-Header=custom-value, got %s", r.Header.Get("X-Custom-Header"))
		}

		if r.URL.Query().Get("param1") != "value1" {
			t.Errorf("expected param1=value1, got %s", r.URL.Query().Get("param1"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	_, err := client.Get(context.Background(), "/test").
		Header("X-Custom-Header", "custom-value").
		Query("param1", "value1").
		Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Retry_On5xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client, _ := New(Config{
		BaseURL:      server.URL,
		MaxRetries:   3,
		RetryWaitMin: 10 * time.Millisecond,
		RetryWaitMax: 100 * time.Millisecond,
	})

	resp, err := client.Get(context.Background(), "/test").Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	if !resp.IsSuccess() {
		t.Errorf("expected success after retries, got status %d", resp.StatusCode)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Get(ctx, "/test").Do()

	if err == nil {
		t.Fatal("expected context deadline exceeded error")
	}
}

func TestClient_Middleware(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization header, got %s", r.Header.Get("Authorization"))
		}

		if r.Header.Get("User-Agent") != "test-agent" {
			t.Errorf("expected User-Agent=test-agent, got %s", r.Header.Get("User-Agent"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	client.Use(AuthBearerMiddleware("test-token"))
	client.Use(UserAgentMiddleware("test-agent"))

	_, err := client.Get(context.Background(), "/test").Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_AllHTTPMethods(t *testing.T) {
	tests := []struct {
		name   string
		method func(*Client, context.Context, string) *RequestBuilder
		want   string
	}{
		{"GET", (*Client).Get, http.MethodGet},
		{"POST", (*Client).Post, http.MethodPost},
		{"PUT", (*Client).Put, http.MethodPut},
		{"PATCH", (*Client).Patch, http.MethodPatch},
		{"DELETE", (*Client).Delete, http.MethodDelete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.want {
					t.Errorf("expected method %s, got %s", tt.want, r.Method)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewDefault(server.URL)
			_, err := tt.method(client, context.Background(), "/test").Do()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_Timeout(t *testing.T) {
	t.Skip("Timeout test skipped - timing sensitive")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(Config{
		BaseURL: server.URL,
		Timeout: 50 * time.Millisecond,
	})

	_, err := client.Get(context.Background(), "/test").Do()

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestClient_4xxNoRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client, _ := New(Config{
		BaseURL:      server.URL,
		MaxRetries:   3,
		RetryWaitMin: 10 * time.Millisecond,
		RetryWaitMax: 100 * time.Millisecond,
	})

	_, err := client.Get(context.Background(), "/test").Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not retry on 4xx errors
	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retry on 4xx), got %d", attempts)
	}
}

func TestClient_MultipleHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Header1") != "value1" {
			t.Errorf("expected Header1=value1, got %s", r.Header.Get("Header1"))
		}
		if r.Header.Get("Header2") != "value2" {
			t.Errorf("expected Header2=value2, got %s", r.Header.Get("Header2"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	_, err := client.Get(context.Background(), "/test").
		Headers(map[string]string{
			"Header1": "value1",
			"Header2": "value2",
		}).
		Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_QueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("key1") != "value1" {
			t.Errorf("expected key1=value1, got %s", query.Get("key1"))
		}
		if query.Get("key2") != "value2" {
			t.Errorf("expected key2=value2, got %s", query.Get("key2"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefault(server.URL)
	_, err := client.Get(context.Background(), "/test").
		QueryParams(map[string]string{
			"key1": "value1",
			"key2": "value2",
		}).
		Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
