package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileProvider(t *testing.T) {
	p := NewFileProvider("/path/to/config.yaml")
	if p == nil {
		t.Fatal("NewFileProvider() returned nil")
	}

	if p.Name() != "file:config.yaml" {
		t.Errorf("Name() = %v, want %v", p.Name(), "file:config.yaml")
	}
}

func TestNewFileProvider_WithOptions(t *testing.T) {
	parser := NewJSONParser()
	p := NewFileProvider("/path/to/config.json",
		WithParser(parser),
		WithOptional(),
	)

	if p.parser != parser {
		t.Error("WithParser() did not set the parser")
	}
	if !p.optional {
		t.Error("WithOptional() did not set optional flag")
	}
}

func TestFileProvider_Load_JSON(t *testing.T) {
	// Create a temporary JSON file
	content := `{"host": "localhost", "port": 8080}`
	tmpFile := createTempFile(t, "config*.json", content)
	defer os.Remove(tmpFile)

	p := NewFileProvider(tmpFile)
	values, err := p.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if values["host"] != "localhost" {
		t.Errorf("values[host] = %v, want %v", values["host"], "localhost")
	}

	port, ok := values["port"].(float64) // JSON unmarshals numbers as float64
	if !ok {
		t.Fatalf("values[port] is not float64")
	}
	if int(port) != 8080 {
		t.Errorf("values[port] = %v, want %v", port, 8080)
	}
}

func TestFileProvider_Load_YAML(t *testing.T) {
	// Create a temporary YAML file
	content := `
host: localhost
port: 8080
database:
  name: mydb
  ssl: true
`
	tmpFile := createTempFile(t, "config*.yaml", content)
	defer os.Remove(tmpFile)

	p := NewFileProvider(tmpFile)
	values, err := p.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if values["host"] != "localhost" {
		t.Errorf("values[host] = %v, want %v", values["host"], "localhost")
	}
	if values["port"] != 8080 {
		t.Errorf("values[port] = %v, want %v", values["port"], 8080)
	}

	db, ok := values["database"].(map[string]any)
	if !ok {
		t.Fatalf("values[database] is not map[string]any")
	}
	if db["name"] != "mydb" {
		t.Errorf("values[database][name] = %v, want %v", db["name"], "mydb")
	}
	if db["ssl"] != true {
		t.Errorf("values[database][ssl] = %v, want %v", db["ssl"], true)
	}
}

func TestFileProvider_Load_FileNotFound(t *testing.T) {
	p := NewFileProvider("/nonexistent/config.yaml")
	_, err := p.Load(context.Background())

	if err == nil {
		t.Fatal("Load() should return error for non-existent file")
	}
}

func TestFileProvider_Load_Optional(t *testing.T) {
	p := NewFileProvider("/nonexistent/config.yaml", WithOptional())
	values, err := p.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v, should not error for optional file", err)
	}

	if len(values) != 0 {
		t.Errorf("Load() should return empty map for missing optional file, got len = %d", len(values))
	}
}

func TestFileProvider_Load_InvalidJSON(t *testing.T) {
	content := `{"invalid": json`
	tmpFile := createTempFile(t, "config*.json", content)
	defer os.Remove(tmpFile)

	p := NewFileProvider(tmpFile)
	_, err := p.Load(context.Background())

	if err == nil {
		t.Fatal("Load() should return error for invalid JSON")
	}
}

func TestFileProvider_Load_InvalidYAML(t *testing.T) {
	content := `
host: localhost
  invalid: yaml
`
	tmpFile := createTempFile(t, "config*.yaml", content)
	defer os.Remove(tmpFile)

	p := NewFileProvider(tmpFile)
	_, err := p.Load(context.Background())

	if err == nil {
		t.Fatal("Load() should return error for invalid YAML")
	}
}

func TestFileProvider_Load_UnsupportedExtension(t *testing.T) {
	content := `some content`
	tmpFile := createTempFile(t, "config*.unknown", content)
	defer os.Remove(tmpFile)

	p := NewFileProvider(tmpFile)
	_, err := p.Load(context.Background())

	if err == nil {
		t.Fatal("Load() should return error for unsupported file extension")
	}
}

func TestFileProvider_Watch(t *testing.T) {
	content := `{"host": "localhost"}`
	tmpFile := createTempFile(t, "config*.json", content)
	defer os.Remove(tmpFile)

	p := NewFileProvider(tmpFile)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	callback := func() {
		// Callback for file changes
	}

	err := p.Watch(ctx, callback)
	if err != nil {
		t.Fatalf("Watch() error = %v", err)
	}

	// Note: We can't easily test the actual file change detection
	// without adding delays and modifying the file, which would make
	// the test slow and flaky. The Watch function is tested for
	// basic setup only.
}

// NewJSONParser creates a JSON parser for testing.
func NewJSONParser() *jsonParser {
	return &jsonParser{}
}

// createTempFile creates a temporary file with the given content.
func createTempFile(t *testing.T, pattern, content string) string {
	t.Helper()

	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return filepath.Clean(tmpFile.Name())
}
