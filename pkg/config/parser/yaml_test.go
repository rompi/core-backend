package parser

import (
	"testing"
)

func TestNewYAMLParser(t *testing.T) {
	p := NewYAMLParser()
	if p == nil {
		t.Fatal("NewYAMLParser() returned nil")
	}
}

func TestYAMLParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result map[string]any)
	}{
		{
			name: "simple object",
			input: `
host: localhost
port: 8080
`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				if result["host"] != "localhost" {
					t.Errorf("host = %v, want %v", result["host"], "localhost")
				}
				if result["port"] != 8080 {
					t.Errorf("port = %v, want %v", result["port"], 8080)
				}
			},
		},
		{
			name: "nested object",
			input: `
database:
  host: db.example.com
  port: 5432
`,
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
			name: "array values",
			input: `
hosts:
  - a
  - b
  - c
`,
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
			name:    "empty document",
			input:   ``,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				// Empty YAML returns nil map
			},
		},
		{
			name: "invalid yaml - bad indentation",
			input: `
host: localhost
  invalid: bad
`,
			wantErr: true,
		},
		{
			name: "boolean values",
			input: `
enabled: true
disabled: false
`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				if result["enabled"] != true {
					t.Errorf("enabled = %v, want %v", result["enabled"], true)
				}
				if result["disabled"] != false {
					t.Errorf("disabled = %v, want %v", result["disabled"], false)
				}
			},
		},
		{
			name: "duration-like strings",
			input: `
timeout: 30s
interval: 5m
`,
			wantErr: false,
			check: func(t *testing.T, result map[string]any) {
				if result["timeout"] != "30s" {
					t.Errorf("timeout = %v, want %v", result["timeout"], "30s")
				}
				if result["interval"] != "5m" {
					t.Errorf("interval = %v, want %v", result["interval"], "5m")
				}
			},
		},
	}

	p := NewYAMLParser()

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

func TestYAMLParser_Extensions(t *testing.T) {
	p := NewYAMLParser()
	exts := p.Extensions()

	if len(exts) != 2 {
		t.Errorf("Extensions() len = %v, want %v", len(exts), 2)
	}

	hasYaml := false
	hasYml := false
	for _, ext := range exts {
		if ext == ".yaml" {
			hasYaml = true
		}
		if ext == ".yml" {
			hasYml = true
		}
	}

	if !hasYaml {
		t.Error("Extensions() should contain .yaml")
	}
	if !hasYml {
		t.Error("Extensions() should contain .yml")
	}
}
