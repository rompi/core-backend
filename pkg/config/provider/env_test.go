package provider

import (
	"context"
	"os"
	"testing"
)

func TestNewEnvProvider(t *testing.T) {
	p := NewEnvProvider()
	if p == nil {
		t.Fatal("NewEnvProvider() returned nil")
	}

	if p.Name() != "env" {
		t.Errorf("Name() = %v, want %v", p.Name(), "env")
	}
}

func TestNewEnvProvider_WithOptions(t *testing.T) {
	p := NewEnvProvider(
		WithPrefix("APP"),
		WithDelimiter("__"),
	)

	if p.prefix != "APP" {
		t.Errorf("prefix = %v, want %v", p.prefix, "APP")
	}
	if p.delimiter != "__" {
		t.Errorf("delimiter = %v, want %v", p.delimiter, "__")
	}
}

func TestEnvProvider_Load(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_HOST", "localhost")
	os.Setenv("TEST_PORT", "8080")
	defer os.Unsetenv("TEST_HOST")
	defer os.Unsetenv("TEST_PORT")

	p := NewEnvProvider(WithPrefix("TEST"))
	values, err := p.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if values["host"] != "localhost" {
		t.Errorf("values[host] = %v, want %v", values["host"], "localhost")
	}
	if values["port"] != "8080" {
		t.Errorf("values[port] = %v, want %v", values["port"], "8080")
	}
}

func TestEnvProvider_Load_NoPrefix(t *testing.T) {
	// Set test environment variables with unique names
	os.Setenv("CONFIGTEST_MYHOST", "localhost")
	defer os.Unsetenv("CONFIGTEST_MYHOST")

	p := NewEnvProvider()
	values, err := p.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Without prefix, all env vars are loaded
	if values["configtest.myhost"] != "localhost" {
		t.Errorf("values[configtest.myhost] = %v, want %v", values["configtest.myhost"], "localhost")
	}
}

func TestEnvProvider_Load_NestedKeys(t *testing.T) {
	// Set nested environment variables
	os.Setenv("APP_DATABASE_HOST", "db.example.com")
	os.Setenv("APP_DATABASE_PORT", "5432")
	defer os.Unsetenv("APP_DATABASE_HOST")
	defer os.Unsetenv("APP_DATABASE_PORT")

	p := NewEnvProvider(WithPrefix("APP"))
	values, err := p.Load(context.Background())

	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// DATABASE_HOST -> database.host
	if values["database.host"] != "db.example.com" {
		t.Errorf("values[database.host] = %v, want %v", values["database.host"], "db.example.com")
	}
	if values["database.port"] != "5432" {
		t.Errorf("values[database.port] = %v, want %v", values["database.port"], "5432")
	}
}

func TestEnvProvider_Watch(t *testing.T) {
	p := NewEnvProvider()

	// Watch should return nil (not supported)
	err := p.Watch(context.Background(), func() {})
	if err != nil {
		t.Errorf("Watch() error = %v, want nil", err)
	}
}
