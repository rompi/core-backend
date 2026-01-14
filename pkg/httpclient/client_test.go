package httpclient

import (
	"net/http"
	"testing"
	"time"
)

func TestNew_ValidConfig(t *testing.T) {
	config := Config{
		BaseURL:      "https://api.example.com",
		Timeout:      10 * time.Second,
		MaxRetries:   5,
		RetryWaitMin: 2 * time.Second,
		RetryWaitMax: 20 * time.Second,
	}

	client, err := New(config)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("expected client, got nil")
	}

	if client.baseURL != config.BaseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, config.BaseURL)
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "empty base URL",
			config: Config{
				BaseURL: "",
			},
		},
		{
			name: "negative timeout",
			config: Config{
				BaseURL: "https://api.example.com",
				Timeout: -1 * time.Second,
			},
		},
		{
			name: "negative max retries",
			config: Config{
				BaseURL:    "https://api.example.com",
				MaxRetries: -1,
			},
		},
		{
			name: "negative retry wait min",
			config: Config{
				BaseURL:      "https://api.example.com",
				RetryWaitMin: -1 * time.Second,
			},
		},
		{
			name: "negative retry wait max",
			config: Config{
				BaseURL:      "https://api.example.com",
				RetryWaitMax: -1 * time.Second,
			},
		},
		{
			name: "retry wait min greater than max",
			config: Config{
				BaseURL:      "https://api.example.com",
				RetryWaitMin: 30 * time.Second,
				RetryWaitMax: 10 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)
			if err == nil {
				t.Error("expected error for invalid config, got nil")
			}
		})
	}
}

func TestConfig_ApplyDefaults(t *testing.T) {
	config := Config{
		BaseURL: "https://api.example.com",
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, 30*time.Second)
	}

	if client.retryPolicy.MaxRetries != 3 {
		t.Errorf("max retries = %d, want %d", client.retryPolicy.MaxRetries, 3)
	}

	if client.retryPolicy.RetryWaitMin != 1*time.Second {
		t.Errorf("retry wait min = %v, want %v", client.retryPolicy.RetryWaitMin, 1*time.Second)
	}

	if client.retryPolicy.RetryWaitMax != 30*time.Second {
		t.Errorf("retry wait max = %v, want %v", client.retryPolicy.RetryWaitMax, 30*time.Second)
	}

	if client.logger == nil {
		t.Error("expected default logger, got nil")
	}
}

func TestNewDefault(t *testing.T) {
	baseURL := "https://api.example.com"
	client := NewDefault(baseURL)

	if client == nil {
		t.Fatal("expected client, got nil")
	}

	if client.baseURL != baseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, baseURL)
	}
}

func TestClient_Use(t *testing.T) {
	client := NewDefault("https://api.example.com")

	initialCount := len(client.middleware)

	client.Use(func(next http.RoundTripper) http.RoundTripper {
		return next
	})

	if len(client.middleware) != initialCount+1 {
		t.Errorf("middleware count = %d, want %d", len(client.middleware), initialCount+1)
	}
}

func TestClient_HTTPMethods(t *testing.T) {
	client := NewDefault("https://api.example.com")

	tests := []struct {
		name   string
		method string
	}{
		{"GET", http.MethodGet},
		{"POST", http.MethodPost},
		{"PUT", http.MethodPut},
		{"PATCH", http.MethodPatch},
		{"DELETE", http.MethodDelete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the method creates a builder
			// Actual execution is tested in integration tests
			switch tt.method {
			case http.MethodGet:
				rb := client.Get(nil, "/test")
				if rb == nil {
					t.Error("expected request builder, got nil")
				}
			case http.MethodPost:
				rb := client.Post(nil, "/test")
				if rb == nil {
					t.Error("expected request builder, got nil")
				}
			case http.MethodPut:
				rb := client.Put(nil, "/test")
				if rb == nil {
					t.Error("expected request builder, got nil")
				}
			case http.MethodPatch:
				rb := client.Patch(nil, "/test")
				if rb == nil {
					t.Error("expected request builder, got nil")
				}
			case http.MethodDelete:
				rb := client.Delete(nil, "/test")
				if rb == nil {
					t.Error("expected request builder, got nil")
				}
			}
		})
	}
}

func TestConfig_WithCircuitBreaker(t *testing.T) {
	config := Config{
		BaseURL: "https://api.example.com",
		CircuitBreaker: &CircuitBreakerConfig{
			MaxRequests: 10,
			Interval:    10 * time.Second,
			Timeout:     30 * time.Second,
		},
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.circuitBreaker == nil {
		t.Error("expected circuit breaker to be initialized")
	}
}

func TestConfig_WithCustomTransport(t *testing.T) {
	transport := &http.Transport{}
	config := Config{
		BaseURL:   "https://api.example.com",
		Transport: transport,
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.httpClient.Transport != transport {
		t.Error("custom transport not set")
	}
}

func TestConfig_FollowRedirects(t *testing.T) {
	t.Run("follow redirects disabled", func(t *testing.T) {
		config := Config{
			BaseURL:         "https://api.example.com",
			FollowRedirects: false,
		}

		client, err := New(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if client.httpClient.CheckRedirect == nil {
			t.Error("expected CheckRedirect function to be set")
		}
	})
}
