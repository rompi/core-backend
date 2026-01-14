package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestResponse_JSON(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "valid JSON",
			body: `{"key":"value","foo":"bar"}`,
			want: map[string]string{"key": "value", "foo": "bar"},
		},
		{
			name:    "invalid JSON",
			body:    `{invalid}`,
			wantErr: true,
		},
		{
			name: "empty JSON object",
			body: `{}`,
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					Body: io.NopCloser(bytes.NewBufferString(tt.body)),
				},
			}

			var got map[string]string
			err := resp.JSON(&got)

			if (err != nil) != tt.wantErr {
				t.Errorf("Response.JSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("Response.JSON() got %v items, want %v", len(got), len(tt.want))
					return
				}

				for k, v := range tt.want {
					if got[k] != v {
						t.Errorf("Response.JSON()[%q] = %q, want %q", k, got[k], v)
					}
				}
			}
		})
	}
}

func TestResponse_String(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr bool
	}{
		{
			name: "simple string",
			body: "hello world",
			want: "hello world",
		},
		{
			name: "empty string",
			body: "",
			want: "",
		},
		{
			name: "multiline string",
			body: "line1\nline2\nline3",
			want: "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					Body: io.NopCloser(bytes.NewBufferString(tt.body)),
				},
			}

			got, err := resp.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("Response.String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Response.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResponse_Bytes(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		want    []byte
		wantErr bool
	}{
		{
			name: "byte slice",
			body: []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}, // "Hello"
			want: []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name: "empty bytes",
			body: []byte{},
			want: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					Body: io.NopCloser(bytes.NewBuffer(tt.body)),
				},
			}

			got, err := resp.Bytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("Response.Bytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !bytes.Equal(got, tt.want) {
				t.Errorf("Response.Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, true},
		{"201 Created", 201, true},
		{"299 Custom Success", 299, true},
		{"199 Below 2xx", 199, false},
		{"300 Redirect", 300, false},
		{"400 Bad Request", 400, false},
		{"500 Server Error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					StatusCode: tt.statusCode,
				},
			}

			got := resp.IsSuccess()
			if got != tt.want {
				t.Errorf("Response.IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_IsClientError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"400 Bad Request", 400, true},
		{"404 Not Found", 404, true},
		{"499 Custom Client Error", 499, true},
		{"399 Below 4xx", 399, false},
		{"500 Server Error", 500, false},
		{"200 Success", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					StatusCode: tt.statusCode,
				},
			}

			got := resp.IsClientError()
			if got != tt.want {
				t.Errorf("Response.IsClientError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_IsServerError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"599 Custom Server Error", 599, true},
		{"499 Below 5xx", 499, false},
		{"600 Above 5xx", 600, false},
		{"200 Success", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					StatusCode: tt.statusCode,
				},
			}

			got := resp.IsServerError()
			if got != tt.want {
				t.Errorf("Response.IsServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_CacheBody(t *testing.T) {
	t.Run("multiple reads", func(t *testing.T) {
		resp := &Response{
			Response: &http.Response{
				Body: io.NopCloser(bytes.NewBufferString("test data")),
			},
		}

		// First read
		str1, err := resp.String()
		if err != nil {
			t.Fatalf("first read failed: %v", err)
		}

		// Second read (should use cached body)
		str2, err := resp.String()
		if err != nil {
			t.Fatalf("second read failed: %v", err)
		}

		if str1 != str2 {
			t.Errorf("cached read returned different data: %q vs %q", str1, str2)
		}

		if str1 != "test data" {
			t.Errorf("got %q, want %q", str1, "test data")
		}
	})

	t.Run("nil body", func(t *testing.T) {
		resp := &Response{
			Response: &http.Response{
				Body: nil,
			},
		}

		bytes, err := resp.Bytes()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if bytes != nil {
			t.Errorf("expected nil, got %v", bytes)
		}
	})
}
