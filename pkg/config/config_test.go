package config

import (
	"context"
	"testing"
	"time"
)

// mockProvider is a test provider that returns predefined values.
type mockProvider struct {
	name   string
	values map[string]any
	err    error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Load(_ context.Context) (map[string]any, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.values, nil
}

func (m *mockProvider) Watch(_ context.Context, _ func()) error {
	return nil
}

func TestNew(t *testing.T) {
	cfg := New()
	if cfg == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNew_WithOptions(t *testing.T) {
	provider := &mockProvider{name: "test"}
	cfg := New(
		WithProvider(provider),
		WithWatchInterval(10*time.Second),
		WithKeyDelimiter("_"),
		WithSensitiveKey("password"),
	)

	if cfg == nil {
		t.Fatal("New() returned nil")
	}
}

func TestConfig_Load(t *testing.T) {
	tests := []struct {
		name      string
		providers []*mockProvider
		wantErr   bool
	}{
		{
			name:      "no providers",
			providers: nil,
			wantErr:   false,
		},
		{
			name: "single provider",
			providers: []*mockProvider{
				{name: "test", values: map[string]any{"key": "value"}},
			},
			wantErr: false,
		},
		{
			name: "multiple providers",
			providers: []*mockProvider{
				{name: "first", values: map[string]any{"key1": "value1"}},
				{name: "second", values: map[string]any{"key2": "value2"}},
			},
			wantErr: false,
		},
		{
			name: "provider with error",
			providers: []*mockProvider{
				{name: "error", err: ErrProviderFailed},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []Option{}
			for _, p := range tt.providers {
				opts = append(opts, WithProvider(p))
			}

			cfg := New(opts...)
			err := cfg.Load(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetString(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		key    string
		want   string
	}{
		{
			name:   "existing string key",
			values: map[string]any{"host": "localhost"},
			key:    "host",
			want:   "localhost",
		},
		{
			name:   "non-existing key",
			values: map[string]any{},
			key:    "missing",
			want:   "",
		},
		{
			name:   "case insensitive",
			values: map[string]any{"HOST": "localhost"},
			key:    "host",
			want:   "localhost",
		},
		{
			name:   "int to string",
			values: map[string]any{"port": 8080},
			key:    "port",
			want:   "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetString(tt.key)
			if got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetInt(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		key    string
		want   int
	}{
		{
			name:   "existing int key",
			values: map[string]any{"port": 8080},
			key:    "port",
			want:   8080,
		},
		{
			name:   "non-existing key",
			values: map[string]any{},
			key:    "missing",
			want:   0,
		},
		{
			name:   "string to int",
			values: map[string]any{"port": "8080"},
			key:    "port",
			want:   8080,
		},
		{
			name:   "float to int",
			values: map[string]any{"port": 8080.5},
			key:    "port",
			want:   8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetInt(tt.key)
			if got != tt.want {
				t.Errorf("GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetBool(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		key    string
		want   bool
	}{
		{
			name:   "true bool",
			values: map[string]any{"enabled": true},
			key:    "enabled",
			want:   true,
		},
		{
			name:   "false bool",
			values: map[string]any{"enabled": false},
			key:    "enabled",
			want:   false,
		},
		{
			name:   "string true",
			values: map[string]any{"enabled": "true"},
			key:    "enabled",
			want:   true,
		},
		{
			name:   "string yes",
			values: map[string]any{"enabled": "yes"},
			key:    "enabled",
			want:   true,
		},
		{
			name:   "string 1",
			values: map[string]any{"enabled": "1"},
			key:    "enabled",
			want:   true,
		},
		{
			name:   "int non-zero",
			values: map[string]any{"enabled": 1},
			key:    "enabled",
			want:   true,
		},
		{
			name:   "non-existing key",
			values: map[string]any{},
			key:    "missing",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetBool(tt.key)
			if got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetDuration(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		key    string
		want   time.Duration
	}{
		{
			name:   "duration string",
			values: map[string]any{"timeout": "30s"},
			key:    "timeout",
			want:   30 * time.Second,
		},
		{
			name:   "duration minutes",
			values: map[string]any{"timeout": "5m"},
			key:    "timeout",
			want:   5 * time.Minute,
		},
		{
			name:   "non-existing key",
			values: map[string]any{},
			key:    "missing",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetDuration(tt.key)
			if got != tt.want {
				t.Errorf("GetDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetStringSlice(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		key    string
		want   []string
	}{
		{
			name:   "string slice",
			values: map[string]any{"hosts": []string{"a", "b", "c"}},
			key:    "hosts",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "comma-separated string",
			values: map[string]any{"hosts": "a, b, c"},
			key:    "hosts",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "interface slice",
			values: map[string]any{"hosts": []any{"a", "b", "c"}},
			key:    "hosts",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "non-existing key",
			values: map[string]any{},
			key:    "missing",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetStringSlice(tt.key)

			if len(got) != len(tt.want) {
				t.Errorf("GetStringSlice() len = %v, want len %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetStringSlice()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestConfig_IsSet(t *testing.T) {
	provider := &mockProvider{
		name:   "test",
		values: map[string]any{"existing": "value"},
	}
	cfg := New(WithProvider(provider))
	_ = cfg.Load(context.Background())

	if !cfg.IsSet("existing") {
		t.Error("IsSet() = false for existing key, want true")
	}

	if cfg.IsSet("missing") {
		t.Error("IsSet() = true for missing key, want false")
	}
}

func TestConfig_Set(t *testing.T) {
	cfg := New()
	_ = cfg.Load(context.Background())

	cfg.Set("new_key", "new_value")

	if !cfg.IsSet("new_key") {
		t.Error("Set() did not set the key")
	}

	got := cfg.GetString("new_key")
	if got != "new_value" {
		t.Errorf("GetString() = %v, want %v", got, "new_value")
	}
}

func TestConfig_Sub(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		values: map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		},
	}
	cfg := New(WithProvider(provider))
	_ = cfg.Load(context.Background())

	dbCfg := cfg.Sub("database")

	host := dbCfg.GetString("host")
	if host != "localhost" {
		t.Errorf("Sub().GetString() = %v, want %v", host, "localhost")
	}

	port := dbCfg.GetInt("port")
	if port != 5432 {
		t.Errorf("Sub().GetInt() = %v, want %v", port, 5432)
	}
}

func TestConfig_AllSettings(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		values: map[string]any{
			"host":     "localhost",
			"password": "secret",
		},
	}
	cfg := New(
		WithProvider(provider),
		WithSensitiveKey("password"),
	)
	_ = cfg.Load(context.Background())

	settings := cfg.AllSettings()

	if settings["host"] != "localhost" {
		t.Errorf("AllSettings()[host] = %v, want %v", settings["host"], "localhost")
	}

	if settings["password"] != "***" {
		t.Errorf("AllSettings()[password] = %v, want %v (should be masked)", settings["password"], "***")
	}
}

func TestConfig_NestedValues(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		values: map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
				"ssl": map[string]any{
					"enabled": true,
					"mode":    "require",
				},
			},
		},
	}
	cfg := New(WithProvider(provider))
	_ = cfg.Load(context.Background())

	// Test nested access with dot notation
	host := cfg.GetString("database.host")
	if host != "localhost" {
		t.Errorf("GetString(database.host) = %v, want %v", host, "localhost")
	}

	port := cfg.GetInt("database.port")
	if port != 5432 {
		t.Errorf("GetInt(database.port) = %v, want %v", port, 5432)
	}

	sslEnabled := cfg.GetBool("database.ssl.enabled")
	if !sslEnabled {
		t.Errorf("GetBool(database.ssl.enabled) = %v, want %v", sslEnabled, true)
	}

	sslMode := cfg.GetString("database.ssl.mode")
	if sslMode != "require" {
		t.Errorf("GetString(database.ssl.mode) = %v, want %v", sslMode, "require")
	}
}

func TestConfig_ProviderPrecedence(t *testing.T) {
	// Later providers should override earlier ones
	first := &mockProvider{
		name:   "first",
		values: map[string]any{"key": "first_value", "first_only": "from_first"},
	}
	second := &mockProvider{
		name:   "second",
		values: map[string]any{"key": "second_value", "second_only": "from_second"},
	}

	cfg := New(
		WithProvider(first),
		WithProvider(second),
	)
	_ = cfg.Load(context.Background())

	// "key" should have second's value (override)
	key := cfg.GetString("key")
	if key != "second_value" {
		t.Errorf("GetString(key) = %v, want %v", key, "second_value")
	}

	// first_only should still exist
	firstOnly := cfg.GetString("first_only")
	if firstOnly != "from_first" {
		t.Errorf("GetString(first_only) = %v, want %v", firstOnly, "from_first")
	}

	// second_only should exist
	secondOnly := cfg.GetString("second_only")
	if secondOnly != "from_second" {
		t.Errorf("GetString(second_only) = %v, want %v", secondOnly, "from_second")
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{"on", true},
		{"ON", true},
		{"false", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseBool(tt.input)
			if got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfig_GetFloat64(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		key    string
		want   float64
	}{
		{
			name:   "existing float64 key",
			values: map[string]any{"rate": 1.5},
			key:    "rate",
			want:   1.5,
		},
		{
			name:   "non-existing key",
			values: map[string]any{},
			key:    "missing",
			want:   0,
		},
		{
			name:   "int to float",
			values: map[string]any{"rate": 10},
			key:    "rate",
			want:   10.0,
		},
		{
			name:   "string to float",
			values: map[string]any{"rate": "3.14"},
			key:    "rate",
			want:   3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetFloat64(tt.key)
			if got != tt.want {
				t.Errorf("GetFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetInt64(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		key    string
		want   int64
	}{
		{
			name:   "existing int64 key",
			values: map[string]any{"count": int64(100)},
			key:    "count",
			want:   100,
		},
		{
			name:   "int32 key",
			values: map[string]any{"count": int32(50)},
			key:    "count",
			want:   50,
		},
		{
			name:   "non-existing key",
			values: map[string]any{},
			key:    "missing",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetInt64(tt.key)
			if got != tt.want {
				t.Errorf("GetInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetTime(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string]any
		key     string
		wantStr string
	}{
		{
			name:    "RFC3339 format",
			values:  map[string]any{"created": "2024-01-15T10:30:00Z"},
			key:     "created",
			wantStr: "2024-01-15T10:30:00Z",
		},
		{
			name:    "date only format",
			values:  map[string]any{"created": "2024-01-15"},
			key:     "created",
			wantStr: "2024-01-15",
		},
		{
			name:    "non-existing key",
			values:  map[string]any{},
			key:     "missing",
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetTime(tt.key)

			if tt.wantStr == "" {
				if !got.IsZero() {
					t.Errorf("GetTime() should be zero for missing key")
				}
			} else {
				// Just verify it parsed something
				if got.IsZero() {
					t.Errorf("GetTime() returned zero time for %v", tt.name)
				}
			}
		})
	}
}

func TestConfig_GetStringMap(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		values: map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		},
	}
	cfg := New(WithProvider(provider))
	_ = cfg.Load(context.Background())

	// Test via buildNestedMap (when direct map not found)
	dbMap := cfg.GetStringMap("database")
	if dbMap == nil {
		t.Fatal("GetStringMap() returned nil")
	}

	if dbMap["host"] != "localhost" {
		t.Errorf("GetStringMap()[host] = %v, want %v", dbMap["host"], "localhost")
	}
}

func TestSubConfig_AllMethods(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		values: map[string]any{
			"app": map[string]any{
				"server": map[string]any{
					"host":    "localhost",
					"port":    8080,
					"enabled": true,
					"timeout": "30s",
				},
			},
		},
	}
	cfg := New(WithProvider(provider))
	_ = cfg.Load(context.Background())

	// Get sub-config
	appCfg := cfg.Sub("app")
	serverCfg := appCfg.Sub("server")

	// Test various getters
	if serverCfg.GetString("host") != "localhost" {
		t.Errorf("Sub().GetString() = %v, want localhost", serverCfg.GetString("host"))
	}

	if serverCfg.GetInt("port") != 8080 {
		t.Errorf("Sub().GetInt() = %v, want 8080", serverCfg.GetInt("port"))
	}

	if !serverCfg.GetBool("enabled") {
		t.Errorf("Sub().GetBool() = false, want true")
	}

	if !serverCfg.IsSet("host") {
		t.Errorf("Sub().IsSet() = false for existing key")
	}

	// Test Set
	serverCfg.Set("new_key", "new_value")
	if serverCfg.GetString("new_key") != "new_value" {
		t.Errorf("Sub().Set() did not work")
	}

	// Test AllSettings
	settings := serverCfg.AllSettings()
	if len(settings) == 0 {
		t.Error("Sub().AllSettings() returned empty map")
	}
}

func TestConfig_Watch(t *testing.T) {
	provider := &mockProvider{
		name:   "test",
		values: map[string]any{"key": "value"},
	}
	cfg := New(WithProvider(provider))
	_ = cfg.Load(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test that Watch doesn't error
	err := cfg.Watch(ctx, func(c Config) {
		// Callback
	})

	if err != nil {
		t.Errorf("Watch() error = %v", err)
	}
}

func TestConfig_Bind_WithValidator(t *testing.T) {
	type Config struct {
		Host string `config:"host"`
	}

	provider := &mockProvider{
		name:   "test",
		values: map[string]any{"host": "localhost"},
	}

	validator := &testValidator{err: ErrValidation}

	cfg := New(
		WithProvider(provider),
		WithValidator(validator),
	)
	_ = cfg.Load(context.Background())

	var appCfg Config
	err := cfg.Bind(&appCfg)

	if err == nil {
		t.Error("Bind() should return error when validator fails")
	}
}

type testValidator struct {
	err error
}

func (v *testValidator) Validate(_ any) error {
	return v.err
}

func TestConfig_GetDuration_Types(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		key    string
		want   time.Duration
	}{
		{
			name:   "int value",
			values: map[string]any{"timeout": 1000000000}, // 1 second in nanoseconds
			key:    "timeout",
			want:   time.Second,
		},
		{
			name:   "int64 value",
			values: map[string]any{"timeout": int64(2000000000)},
			key:    "timeout",
			want:   2 * time.Second,
		},
		{
			name:   "float64 value",
			values: map[string]any{"timeout": float64(3000000000)},
			key:    "timeout",
			want:   3 * time.Second,
		},
		{
			name:   "duration value",
			values: map[string]any{"timeout": 5 * time.Second},
			key:    "timeout",
			want:   5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetDuration(tt.key)
			if got != tt.want {
				t.Errorf("GetDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetFloat64_MoreTypes(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]any
		key    string
		want   float64
	}{
		{
			name:   "float32 value",
			values: map[string]any{"rate": float32(1.5)},
			key:    "rate",
			want:   float64(float32(1.5)), // Account for float32 precision
		},
		{
			name:   "int64 value",
			values: map[string]any{"rate": int64(100)},
			key:    "rate",
			want:   100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetFloat64(tt.key)
			if got != tt.want {
				t.Errorf("GetFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetTime_MoreFormats(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string]any
		key     string
		wantErr bool
	}{
		{
			name:    "RFC3339Nano format",
			values:  map[string]any{"ts": "2024-01-15T10:30:00.123456789Z"},
			key:     "ts",
			wantErr: false,
		},
		{
			name:    "datetime format",
			values:  map[string]any{"ts": "2024-01-15 10:30:00"},
			key:     "ts",
			wantErr: false,
		},
		{
			name:    "time.Time value",
			values:  map[string]any{"ts": time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
			key:     "ts",
			wantErr: false,
		},
		{
			name:    "invalid format",
			values:  map[string]any{"ts": "not-a-date"},
			key:     "ts",
			wantErr: true, // Returns zero time
		},
		{
			name:    "unsupported type",
			values:  map[string]any{"ts": 12345},
			key:     "ts",
			wantErr: true, // Returns zero time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{name: "test", values: tt.values}
			cfg := New(WithProvider(provider))
			_ = cfg.Load(context.Background())

			got := cfg.GetTime(tt.key)
			if tt.wantErr && !got.IsZero() {
				t.Errorf("GetTime() should return zero time for invalid format")
			}
			if !tt.wantErr && got.IsZero() {
				t.Errorf("GetTime() returned zero time for valid format")
			}
		})
	}
}

func TestSubConfig_MoreMethods(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		values: map[string]any{
			"app": map[string]any{
				"rate":    1.5,
				"count":   int64(100),
				"timeout": "30s",
				"created": "2024-01-15T10:30:00Z",
				"hosts":   "a, b, c",
			},
		},
	}
	cfg := New(WithProvider(provider))
	_ = cfg.Load(context.Background())

	appCfg := cfg.Sub("app")

	// Test GetFloat64
	if appCfg.GetFloat64("rate") != 1.5 {
		t.Errorf("Sub().GetFloat64() = %v, want 1.5", appCfg.GetFloat64("rate"))
	}

	// Test GetInt64
	if appCfg.GetInt64("count") != 100 {
		t.Errorf("Sub().GetInt64() = %v, want 100", appCfg.GetInt64("count"))
	}

	// Test GetDuration
	if appCfg.GetDuration("timeout") != 30*time.Second {
		t.Errorf("Sub().GetDuration() = %v, want 30s", appCfg.GetDuration("timeout"))
	}

	// Test GetTime
	got := appCfg.GetTime("created")
	if got.IsZero() {
		t.Error("Sub().GetTime() returned zero time")
	}

	// Test GetStringSlice
	hosts := appCfg.GetStringSlice("hosts")
	if len(hosts) != 3 {
		t.Errorf("Sub().GetStringSlice() len = %v, want 3", len(hosts))
	}

	// Test GetStringMap
	dbMap := appCfg.GetStringMap("nested")
	// Should return nil for non-existent nested map
	_ = dbMap

	// Test Load (delegates to parent)
	err := appCfg.Load(context.Background())
	if err != nil {
		t.Errorf("Sub().Load() error = %v", err)
	}

	// Test Watch (delegates to parent)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = appCfg.Watch(ctx, func(c Config) {})
	if err != nil {
		t.Errorf("Sub().Watch() error = %v", err)
	}
}
