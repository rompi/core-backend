package postgres

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("expected port 5432, got %d", cfg.Port)
	}
	if cfg.Schema != "public" {
		t.Errorf("expected schema public, got %s", cfg.Schema)
	}
	if cfg.SSLMode != "prefer" {
		t.Errorf("expected sslmode prefer, got %s", cfg.SSLMode)
	}
	if cfg.MaxConns != 25 {
		t.Errorf("expected max conns 25, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 5 {
		t.Errorf("expected min conns 5, got %d", cfg.MinConns)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Host:           "localhost",
				Port:           5432,
				User:           "testuser",
				Password:       "testpass",
				Database:       "testdb",
				Schema:         "public",
				SSLMode:        "prefer",
				MaxConns:       25,
				MinConns:       5,
				ConnectTimeout: 10 * time.Second,
				QueryTimeout:   30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing user",
			config: Config{
				Host:           "localhost",
				Port:           5432,
				Password:       "testpass",
				Database:       "testdb",
				SSLMode:        "prefer",
				MaxConns:       25,
				ConnectTimeout: 10 * time.Second,
				QueryTimeout:   30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: Config{
				Host:           "localhost",
				Port:           5432,
				User:           "testuser",
				Database:       "testdb",
				SSLMode:        "prefer",
				MaxConns:       25,
				ConnectTimeout: 10 * time.Second,
				QueryTimeout:   30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing database",
			config: Config{
				Host:           "localhost",
				Port:           5432,
				User:           "testuser",
				Password:       "testpass",
				SSLMode:        "prefer",
				MaxConns:       25,
				ConnectTimeout: 10 * time.Second,
				QueryTimeout:   30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: Config{
				Host:           "localhost",
				Port:           0,
				User:           "testuser",
				Password:       "testpass",
				Database:       "testdb",
				SSLMode:        "prefer",
				MaxConns:       25,
				ConnectTimeout: 10 * time.Second,
				QueryTimeout:   30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid ssl mode",
			config: Config{
				Host:           "localhost",
				Port:           5432,
				User:           "testuser",
				Password:       "testpass",
				Database:       "testdb",
				SSLMode:        "invalid",
				MaxConns:       25,
				ConnectTimeout: 10 * time.Second,
				QueryTimeout:   30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "min conns greater than max conns",
			config: Config{
				Host:           "localhost",
				Port:           5432,
				User:           "testuser",
				Password:       "testpass",
				Database:       "testdb",
				SSLMode:        "prefer",
				MaxConns:       5,
				MinConns:       10,
				ConnectTimeout: 10 * time.Second,
				QueryTimeout:   30 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ConnectionString(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		Schema:   "public",
		SSLMode:  "prefer",
	}

	connStr := cfg.ConnectionString()
	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=prefer"
	if connStr != expected {
		t.Errorf("expected %s, got %s", expected, connStr)
	}

	// Test with custom schema
	cfg.Schema = "myschema"
	connStr = cfg.ConnectionString()
	if connStr != expected+" search_path=myschema" {
		t.Errorf("expected search_path in connection string for custom schema")
	}
}

func TestConfig_ConnectionURL(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		Schema:   "public",
		SSLMode:  "prefer",
	}

	url := cfg.ConnectionURL()
	expected := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=prefer"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}

	// Test with custom schema
	cfg.Schema = "myschema"
	url = cfg.ConnectionURL()
	if url != expected+"&search_path=myschema" {
		t.Errorf("expected search_path in URL for custom schema")
	}
}

func TestLoadConfig_FromEnv(t *testing.T) {
	// Set env vars
	os.Setenv("POSTGRES_HOST", "dbhost")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "envuser")
	os.Setenv("POSTGRES_PASSWORD", "envpass")
	os.Setenv("POSTGRES_DATABASE", "envdb")
	os.Setenv("POSTGRES_SCHEMA", "envschema")
	os.Setenv("POSTGRES_SSL_MODE", "require")
	defer func() {
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DATABASE")
		os.Unsetenv("POSTGRES_SCHEMA")
		os.Unsetenv("POSTGRES_SSL_MODE")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Host != "dbhost" {
		t.Errorf("expected host dbhost, got %s", cfg.Host)
	}
	if cfg.Port != 5433 {
		t.Errorf("expected port 5433, got %d", cfg.Port)
	}
	if cfg.User != "envuser" {
		t.Errorf("expected user envuser, got %s", cfg.User)
	}
	if cfg.Database != "envdb" {
		t.Errorf("expected database envdb, got %s", cfg.Database)
	}
	if cfg.Schema != "envschema" {
		t.Errorf("expected schema envschema, got %s", cfg.Schema)
	}
	if cfg.SSLMode != "require" {
		t.Errorf("expected sslmode require, got %s", cfg.SSLMode)
	}
}
