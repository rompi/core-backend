package feature

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContext_NewContext(t *testing.T) {
	ctx := NewContext("user-123")

	if ctx.Key != "user-123" {
		t.Errorf("Key = %v, want user-123", ctx.Key)
	}
	if ctx.Custom == nil {
		t.Error("Custom should be initialized")
	}
	if ctx.Groups == nil {
		t.Error("Groups should be initialized")
	}
}

func TestContext_WithMethods(t *testing.T) {
	ctx := NewContext("user-123").
		WithName("John Doe").
		WithEmail("john@example.com").
		WithIP("192.168.1.1").
		WithCountry("US").
		WithAnonymous(false).
		WithGroup("beta-testers").
		WithAttribute("plan", "pro")

	if ctx.Name != "John Doe" {
		t.Errorf("Name = %v, want John Doe", ctx.Name)
	}
	if ctx.Email != "john@example.com" {
		t.Errorf("Email = %v, want john@example.com", ctx.Email)
	}
	if ctx.IP != "192.168.1.1" {
		t.Errorf("IP = %v, want 192.168.1.1", ctx.IP)
	}
	if ctx.Country != "US" {
		t.Errorf("Country = %v, want US", ctx.Country)
	}
	if ctx.Anonymous != false {
		t.Errorf("Anonymous = %v, want false", ctx.Anonymous)
	}
	if len(ctx.Groups) != 1 || ctx.Groups[0] != "beta-testers" {
		t.Errorf("Groups = %v, want [beta-testers]", ctx.Groups)
	}
	if ctx.Custom["plan"] != "pro" {
		t.Errorf("Custom[plan] = %v, want pro", ctx.Custom["plan"])
	}
}

func TestContext_WithGroups(t *testing.T) {
	ctx := NewContext("user-123").WithGroups([]string{"group1", "group2", "group3"})

	if len(ctx.Groups) != 3 {
		t.Errorf("Groups length = %d, want 3", len(ctx.Groups))
	}
	if ctx.Groups[0] != "group1" {
		t.Errorf("Groups[0] = %v, want group1", ctx.Groups[0])
	}
}

func TestContext_GetAttribute(t *testing.T) {
	ctx := NewContext("user-123").
		WithName("John").
		WithEmail("john@example.com").
		WithIP("192.168.1.1").
		WithCountry("US").
		WithAnonymous(true).
		WithGroups([]string{"admin"}).
		WithAttribute("plan", "pro")

	tests := []struct {
		name string
		attr string
		want interface{}
	}{
		{"key", "key", "user-123"},
		{"name", "name", "John"},
		{"email", "email", "john@example.com"},
		{"ip", "ip", "192.168.1.1"},
		{"country", "country", "US"},
		{"anonymous", "anonymous", true},
		{"groups", "groups", []string{"admin"}},
		{"custom attribute", "plan", "pro"},
		{"non-existent", "unknown", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ctx.getAttribute(tt.attr)

			// Special handling for slices
			if gotSlice, ok := got.([]string); ok {
				wantSlice := tt.want.([]string)
				if len(gotSlice) != len(wantSlice) {
					t.Errorf("getAttribute(%s) len = %d, want %d", tt.attr, len(gotSlice), len(wantSlice))
					return
				}
				for i, v := range gotSlice {
					if v != wantSlice[i] {
						t.Errorf("getAttribute(%s)[%d] = %v, want %v", tt.attr, i, v, wantSlice[i])
					}
				}
				return
			}

			if got != tt.want {
				t.Errorf("getAttribute(%s) = %v, want %v", tt.attr, got, tt.want)
			}
		})
	}
}

func TestContext_GetAttribute_NilContext(t *testing.T) {
	var ctx *Context
	got := ctx.getAttribute("key")
	if got != nil {
		t.Errorf("getAttribute() on nil context = %v, want nil", got)
	}
}

func TestContext_GetAttribute_NilCustom(t *testing.T) {
	ctx := &Context{Key: "user-1"}
	got := ctx.getAttribute("unknown")
	if got != nil {
		t.Errorf("getAttribute(unknown) with nil Custom = %v, want nil", got)
	}
}

func TestWithContext(t *testing.T) {
	fctx := NewContext("user-123").WithEmail("test@example.com")
	ctx := WithContext(context.Background(), fctx)

	retrieved := FromContext(ctx)
	if retrieved == nil {
		t.Fatal("FromContext() returned nil")
	}
	if retrieved.Key != "user-123" {
		t.Errorf("Key = %v, want user-123", retrieved.Key)
	}
	if retrieved.Email != "test@example.com" {
		t.Errorf("Email = %v, want test@example.com", retrieved.Email)
	}
}

func TestFromContext_NilContext(t *testing.T) {
	result := FromContext(nil)
	if result != nil {
		t.Errorf("FromContext(nil) = %v, want nil", result)
	}
}

func TestFromContext_NoFeatureContext(t *testing.T) {
	ctx := context.Background()
	result := FromContext(ctx)
	if result != nil {
		t.Errorf("FromContext() without feature context = %v, want nil", result)
	}
}

func TestContextFromHTTPRequest(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		wantIP  string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
			},
			wantIP: "203.0.113.195",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "198.51.100.178",
			},
			wantIP: "198.51.100.178",
		},
		{
			name: "X-User-ID header",
			headers: map[string]string{
				"X-User-ID": "user-456",
			},
			wantIP: "", // Will be RemoteAddr
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ctx := ContextFromHTTPRequest(req)

			if tt.wantIP != "" && ctx.IP != tt.wantIP {
				t.Errorf("IP = %v, want %v", ctx.IP, tt.wantIP)
			}

			if tt.headers["X-User-ID"] != "" && ctx.Key != tt.headers["X-User-ID"] {
				t.Errorf("Key = %v, want %v", ctx.Key, tt.headers["X-User-ID"])
			}
		})
	}
}

func TestContextFromHTTPRequest_NilRequest(t *testing.T) {
	ctx := ContextFromHTTPRequest(nil)
	if ctx == nil {
		t.Error("ContextFromHTTPRequest(nil) should return non-nil context")
	}
	if ctx.Key != "" {
		t.Errorf("Key = %v, want empty string", ctx.Key)
	}
}

func TestContextFromHTTPRequest_CloudflareCountry(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("CF-IPCountry", "DE")

	ctx := ContextFromHTTPRequest(req)
	if ctx.Country != "DE" {
		t.Errorf("Country = %v, want DE", ctx.Country)
	}
}
