package config

import (
	"testing"
	"time"
)

func TestBinder_Bind_BasicTypes(t *testing.T) {
	type Config struct {
		Host    string  `config:"host"`
		Port    int     `config:"port"`
		Rate    float64 `config:"rate"`
		Enabled bool    `config:"enabled"`
	}

	values := map[string]any{
		"host":    "localhost",
		"port":    8080,
		"rate":    1.5,
		"enabled": true,
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %v, want %v", cfg.Host, "localhost")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %v, want %v", cfg.Port, 8080)
	}
	if cfg.Rate != 1.5 {
		t.Errorf("Rate = %v, want %v", cfg.Rate, 1.5)
	}
	if !cfg.Enabled {
		t.Errorf("Enabled = %v, want %v", cfg.Enabled, true)
	}
}

func TestBinder_Bind_StringConversion(t *testing.T) {
	type Config struct {
		Port    int     `config:"port"`
		Rate    float64 `config:"rate"`
		Enabled bool    `config:"enabled"`
	}

	values := map[string]any{
		"port":    "8080",
		"rate":    "1.5",
		"enabled": "true",
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %v, want %v", cfg.Port, 8080)
	}
	if cfg.Rate != 1.5 {
		t.Errorf("Rate = %v, want %v", cfg.Rate, 1.5)
	}
	if !cfg.Enabled {
		t.Errorf("Enabled = %v, want %v", cfg.Enabled, true)
	}
}

func TestBinder_Bind_Duration(t *testing.T) {
	type Config struct {
		Timeout time.Duration `config:"timeout"`
	}

	values := map[string]any{
		"timeout": "30s",
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}
}

func TestBinder_Bind_Time(t *testing.T) {
	type Config struct {
		CreatedAt time.Time `config:"created_at"`
	}

	values := map[string]any{
		"created_at": "2024-01-15T10:30:00Z",
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	expected, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
	if !cfg.CreatedAt.Equal(expected) {
		t.Errorf("CreatedAt = %v, want %v", cfg.CreatedAt, expected)
	}
}

func TestBinder_Bind_Slice(t *testing.T) {
	type Config struct {
		Hosts []string `config:"hosts"`
	}

	values := map[string]any{
		"hosts": "host1, host2, host3",
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	expected := []string{"host1", "host2", "host3"}
	if len(cfg.Hosts) != len(expected) {
		t.Errorf("Hosts len = %v, want %v", len(cfg.Hosts), len(expected))
		return
	}

	for i, h := range expected {
		if cfg.Hosts[i] != h {
			t.Errorf("Hosts[%d] = %v, want %v", i, cfg.Hosts[i], h)
		}
	}
}

func TestBinder_Bind_NestedStruct(t *testing.T) {
	type Database struct {
		Host string `config:"host"`
		Port int    `config:"port"`
	}

	type Config struct {
		Database Database `config:"database"`
	}

	values := map[string]any{
		"database.host": "localhost",
		"database.port": 5432,
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %v, want %v", cfg.Database.Host, "localhost")
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %v, want %v", cfg.Database.Port, 5432)
	}
}

func TestBinder_Bind_Default(t *testing.T) {
	type Config struct {
		Host string `config:"host" default:"localhost"`
		Port int    `config:"port" default:"8080"`
	}

	values := map[string]any{}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %v, want %v", cfg.Host, "localhost")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %v, want %v", cfg.Port, 8080)
	}
}

func TestBinder_Bind_Required(t *testing.T) {
	type Config struct {
		Host string `config:"host" required:"true"`
	}

	values := map[string]any{}

	binder := newBinder(values, ".")
	var cfg Config

	err := binder.Bind(&cfg)
	if err == nil {
		t.Fatal("Bind() should return error for missing required field")
	}

	multiErr, ok := err.(*MultiBindError)
	if !ok {
		t.Fatalf("Bind() error should be *MultiBindError, got %T", err)
	}

	if len(multiErr.Errors) == 0 {
		t.Fatal("MultiBindError should contain at least one error")
	}

	if multiErr.Errors[0].Tag != "required" {
		t.Errorf("BindError.Tag = %v, want %v", multiErr.Errors[0].Tag, "required")
	}
}

func TestBinder_Bind_EnvOverride(t *testing.T) {
	type Config struct {
		Port int `config:"server.port" env:"port"` // env key is looked up in lowercase
	}

	// Env override: when both config key and env key are set, env takes precedence
	values := map[string]any{
		"server.port": 8080, // config key value
		"port":        9090, // env override value (stored lowercase)
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	// Env key should override config key
	if cfg.Port != 9090 {
		t.Errorf("Port = %v, want %v (should use env override)", cfg.Port, 9090)
	}
}

func TestBinder_Bind_IgnoreTag(t *testing.T) {
	type Config struct {
		Host    string `config:"host"`
		Ignored string `config:"-"`
	}

	values := map[string]any{
		"host":    "localhost",
		"ignored": "should_not_bind",
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %v, want %v", cfg.Host, "localhost")
	}
	if cfg.Ignored != "" {
		t.Errorf("Ignored = %v, want empty (should not bind)", cfg.Ignored)
	}
}

func TestBinder_Bind_NonPointer(t *testing.T) {
	type Config struct {
		Host string `config:"host"`
	}

	values := map[string]any{}
	binder := newBinder(values, ".")
	var cfg Config

	err := binder.Bind(cfg) // Not a pointer
	if err == nil {
		t.Fatal("Bind() should return error for non-pointer")
	}
}

func TestBinder_Bind_NilPointer(t *testing.T) {
	values := map[string]any{}
	binder := newBinder(values, ".")

	var cfg *struct{}

	err := binder.Bind(cfg) // Nil pointer
	if err == nil {
		t.Fatal("Bind() should return error for nil pointer")
	}
}

func TestBinder_Bind_UnsignedInt(t *testing.T) {
	type Config struct {
		Size uint   `config:"size"`
		Big  uint64 `config:"big"`
	}

	values := map[string]any{
		"size": 100,
		"big":  "1000000",
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Size != 100 {
		t.Errorf("Size = %v, want %v", cfg.Size, 100)
	}
	if cfg.Big != 1000000 {
		t.Errorf("Big = %v, want %v", cfg.Big, 1000000)
	}
}

func TestBinder_Bind_Map(t *testing.T) {
	type Config struct {
		Labels map[string]string `config:"labels"`
	}

	values := map[string]any{
		"labels": map[string]any{
			"app": "myapp",
			"env": "prod",
		},
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Labels["app"] != "myapp" {
		t.Errorf("Labels[app] = %v, want %v", cfg.Labels["app"], "myapp")
	}
	if cfg.Labels["env"] != "prod" {
		t.Errorf("Labels[env] = %v, want %v", cfg.Labels["env"], "prod")
	}
}

func TestBinder_Bind_SliceOfAny(t *testing.T) {
	type Config struct {
		Items []string `config:"items"`
	}

	values := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if len(cfg.Items) != 3 {
		t.Errorf("Items len = %v, want %v", len(cfg.Items), 3)
	}
}

func TestBinder_Bind_PointerField(t *testing.T) {
	type Config struct {
		Host *string `config:"host"`
	}

	values := map[string]any{
		"host": "localhost",
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Host == nil {
		t.Fatal("Host is nil")
	}
	if *cfg.Host != "localhost" {
		t.Errorf("*Host = %v, want %v", *cfg.Host, "localhost")
	}
}

func TestBinder_Bind_IntTypes(t *testing.T) {
	type Config struct {
		Int8Val  int8  `config:"int8"`
		Int16Val int16 `config:"int16"`
		Int32Val int32 `config:"int32"`
	}

	values := map[string]any{
		"int8":  int8(8),
		"int16": int16(16),
		"int32": int32(32),
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Int8Val != 8 {
		t.Errorf("Int8Val = %v, want %v", cfg.Int8Val, 8)
	}
	if cfg.Int16Val != 16 {
		t.Errorf("Int16Val = %v, want %v", cfg.Int16Val, 16)
	}
	if cfg.Int32Val != 32 {
		t.Errorf("Int32Val = %v, want %v", cfg.Int32Val, 32)
	}
}

func TestBinder_Bind_UintTypes(t *testing.T) {
	type Config struct {
		Uint8Val  uint8  `config:"uint8"`
		Uint16Val uint16 `config:"uint16"`
		Uint32Val uint32 `config:"uint32"`
	}

	values := map[string]any{
		"uint8":  uint8(8),
		"uint16": uint16(16),
		"uint32": uint32(32),
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Uint8Val != 8 {
		t.Errorf("Uint8Val = %v, want %v", cfg.Uint8Val, 8)
	}
	if cfg.Uint16Val != 16 {
		t.Errorf("Uint16Val = %v, want %v", cfg.Uint16Val, 16)
	}
	if cfg.Uint32Val != 32 {
		t.Errorf("Uint32Val = %v, want %v", cfg.Uint32Val, 32)
	}
}

func TestBinder_Bind_Float32(t *testing.T) {
	type Config struct {
		Rate float32 `config:"rate"`
	}

	values := map[string]any{
		"rate": float32(1.5),
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Rate != 1.5 {
		t.Errorf("Rate = %v, want %v", cfg.Rate, 1.5)
	}
}

func TestBinder_Bind_DurationInt64(t *testing.T) {
	type Config struct {
		Timeout time.Duration `config:"timeout"`
	}

	values := map[string]any{
		"timeout": int64(30000000000), // 30 seconds in nanoseconds
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}
}

func TestBinder_Bind_TimeObject(t *testing.T) {
	type Config struct {
		CreatedAt time.Time `config:"created_at"`
	}

	now := time.Now()
	values := map[string]any{
		"created_at": now,
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if !cfg.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", cfg.CreatedAt, now)
	}
}

func TestBinder_Bind_BoolFromInt(t *testing.T) {
	type Config struct {
		Enabled bool `config:"enabled"`
	}

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"int 1", 1, true},
		{"int 0", 0, false},
		{"int64 1", int64(1), true},
		{"int64 0", int64(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]any{"enabled": tt.value}
			binder := newBinder(values, ".")
			var cfg Config

			if err := binder.Bind(&cfg); err != nil {
				t.Fatalf("Bind() error = %v", err)
			}

			if cfg.Enabled != tt.want {
				t.Errorf("Enabled = %v, want %v", cfg.Enabled, tt.want)
			}
		})
	}
}

func TestBinder_Bind_NoConfigTag(t *testing.T) {
	type Config struct {
		Host string // No config tag, should use lowercase field name
	}

	values := map[string]any{
		"host": "localhost",
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %v, want %v", cfg.Host, "localhost")
	}
}

func TestBinder_Bind_PointerToStruct(t *testing.T) {
	type Database struct {
		Host string `config:"host"`
		Port int    `config:"port"`
	}

	type Config struct {
		Database *Database `config:"database"`
	}

	values := map[string]any{
		"database.host": "localhost",
		"database.port": 5432,
	}

	binder := newBinder(values, ".")
	var cfg Config

	if err := binder.Bind(&cfg); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	if cfg.Database == nil {
		t.Fatal("Database is nil")
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %v, want %v", cfg.Database.Host, "localhost")
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %v, want %v", cfg.Database.Port, 5432)
	}
}
