package parser

import (
	"testing"
)

func TestNewJSONParser(t *testing.T) {
	p := NewJSONParser()
	if p == nil {
		t.Fatal("NewJSONParser() returned nil")
	}
}

func TestJSONParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result map[string]any)
	}{
		{
			name:    "simple object",
			input:   `{"host": "localhost", "port": 8080}`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				if result["host"] != "localhost" {
					t.Errorf("host = %v, want %v", result["host"], "localhost")
				}
				if result["port"] != float64(8080) {
					t.Errorf("port = %v, want %v", result["port"], 8080)
				}
			},
		},
		{
			name: "nested object",
			input: `{
				"database": {
					"host": "db.example.com",
					"port": 5432
				}
			}`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				db, ok := result["database"].(map[string]any)
				if !ok {
					t.Fatal("database is not a map")
				}
				if db["host"] != "db.example.com" {
					t.Errorf("database.host = %v, want %v", db["host"], "db.example.com")
				}
			},
		},
		{
			name:    "array values",
			input:   `{"hosts": ["a", "b", "c"]}`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				hosts, ok := result["hosts"].([]any)
				if !ok {
					t.Fatal("hosts is not a slice")
				}
				if len(hosts) != 3 {
					t.Errorf("hosts len = %v, want %v", len(hosts), 3)
				}
			},
		},
		{
			name:    "empty object",
			input:   `{}`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				if len(result) != 0 {
					t.Errorf("len = %v, want %v", len(result), 0)
				}
			},
		},
		{
			name:    "invalid json",
			input:   `{"invalid": json}`,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   ``,
			wantErr: true,
		},
	}

	p := NewJSONParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestJSONParser_Extensions(t *testing.T) {
	p := NewJSONParser()
	exts := p.Extensions()

	if len(exts) != 1 {
		t.Errorf("Extensions() len = %v, want %v", len(exts), 1)
	}

	if exts[0] != ".json" {
		t.Errorf("Extensions()[0] = %v, want %v", exts[0], ".json")
	}
}
