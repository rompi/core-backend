package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultCORSConfig(t *testing.T) {
	cfg := DefaultCORSConfig()

	if len(cfg.AllowOrigins) != 1 || cfg.AllowOrigins[0] != "*" {
		t.Errorf("AllowOrigins = %v, want [*]", cfg.AllowOrigins)
	}

	expectedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	if len(cfg.AllowMethods) != len(expectedMethods) {
		t.Errorf("AllowMethods len = %d, want %d", len(cfg.AllowMethods), len(expectedMethods))
	}

	if cfg.MaxAge != 86400 {
		t.Errorf("MaxAge = %d, want 86400", cfg.MaxAge)
	}

	if cfg.AllowCredentials {
		t.Error("AllowCredentials should be false by default")
	}
}

func TestCORSMiddleware_AllowAllOrigins(t *testing.T) {
	middleware := CORSMiddleware(DefaultCORSConfig())

	t.Run("allows any origin", func(t *testing.T) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("ACAO = %q, want %q", rr.Header().Get("Access-Control-Allow-Origin"), "*")
		}
	})

	t.Run("handles preflight OPTIONS request", func(t *testing.T) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("should not reach here"))
		}))

		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusNoContent)
		}

		if rr.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("ACAM header should be set")
		}
		if rr.Header().Get("Access-Control-Allow-Headers") == "" {
			t.Error("ACAH header should be set")
		}
		if rr.Header().Get("Access-Control-Max-Age") != "86400" {
			t.Errorf("ACMA = %q, want %q", rr.Header().Get("Access-Control-Max-Age"), "86400")
		}
		if rr.Body.String() != "" {
			t.Error("body should be empty for OPTIONS")
		}
	})
}

func TestCORSMiddleware_SpecificOrigins(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"http://example.com", "http://test.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type"},
		MaxAge:       3600,
	}
	middleware := CORSMiddleware(config)

	t.Run("allows listed origin", func(t *testing.T) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
			t.Errorf("ACAO = %q, want %q", rr.Header().Get("Access-Control-Allow-Origin"), "http://example.com")
		}
	})

	t.Run("denies unlisted origin", func(t *testing.T) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://malicious.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("ACAO = %q, should be empty for unlisted origin", rr.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("no origin header", func(t *testing.T) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("ACAO should be empty when no origin header")
		}
	})
}

func TestCORSMiddleware_WithCredentials(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}
	middleware := CORSMiddleware(config)

	t.Run("sets credentials header with origin echo", func(t *testing.T) {
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// When allow credentials is true with *, it should echo back the origin
		if rr.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
			t.Errorf("ACAO = %q, want origin echo", rr.Header().Get("Access-Control-Allow-Origin"))
		}
		if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Errorf("ACAC = %q, want %q", rr.Header().Get("Access-Control-Allow-Credentials"), "true")
		}
	})
}

func TestCORSMiddleware_ExposeHeaders(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET"},
		ExposeHeaders: []string{"X-Custom-Header", "X-Another-Header"},
	}
	middleware := CORSMiddleware(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	exposedHeaders := rr.Header().Get("Access-Control-Expose-Headers")
	if exposedHeaders == "" {
		t.Error("ACEH should be set")
	}
	if exposedHeaders != "X-Custom-Header, X-Another-Header" {
		t.Errorf("ACEH = %q", exposedHeaders)
	}
}

func TestCORSMiddleware_NoMaxAge(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET"},
		MaxAge:       0,
	}
	middleware := CORSMiddleware(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Max-Age") != "" {
		t.Error("ACMA should not be set when MaxAge is 0")
	}
}

func TestCORSWithOrigins(t *testing.T) {
	t.Run("with specified origins", func(t *testing.T) {
		middleware := CORSWithOrigins("http://example.com", "http://test.com")

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
			t.Errorf("ACAO = %q, want %q", rr.Header().Get("Access-Control-Allow-Origin"), "http://example.com")
		}
	})

	t.Run("with no origins uses default", func(t *testing.T) {
		middleware := CORSWithOrigins()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://any.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("ACAO = %q, want %q", rr.Header().Get("Access-Control-Allow-Origin"), "*")
		}
	})
}

func TestCORSAllowAll(t *testing.T) {
	middleware := CORSAllowAll()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("ACAO = %q, want %q", rr.Header().Get("Access-Control-Allow-Origin"), "*")
	}
}

func TestCORSMiddleware_NonCORSRequest(t *testing.T) {
	middleware := CORSMiddleware(DefaultCORSConfig())

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Origin header
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "response" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "response")
	}
}

func TestCORSMiddleware_ContentLength(t *testing.T) {
	middleware := CORSMiddleware(DefaultCORSConfig())

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Length") != "0" {
		t.Errorf("Content-Length = %q, want %q", rr.Header().Get("Content-Length"), "0")
	}
}
