package feature

import (
	"context"
	"testing"
)

func TestClient_Bool(t *testing.T) {
	tests := []struct {
		name         string
		flag         *Flag
		context      *Context
		defaultValue bool
		want         bool
	}{
		{
			name: "enabled flag returns true variant",
			flag: &Flag{
				Key:          "test-flag",
				Type:         FlagTypeBool,
				DefaultValue: false,
				Enabled:      true,
				Variants: []Variant{
					{Name: "off", Value: false},
					{Name: "on", Value: true},
				},
				Rules: []Rule{
					{Variation: 1},
				},
			},
			context:      NewContext("user-1"),
			defaultValue: false,
			want:         true,
		},
		{
			name: "disabled flag returns default",
			flag: &Flag{
				Key:          "test-flag",
				Type:         FlagTypeBool,
				DefaultValue: false,
				Enabled:      false,
			},
			context:      NewContext("user-1"),
			defaultValue: true,
			want:         false,
		},
		{
			name:         "missing flag returns default",
			flag:         nil,
			context:      NewContext("user-1"),
			defaultValue: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{
				flags: make(map[string]*Flag),
			}
			if tt.flag != nil {
				provider.flags[tt.flag.Key] = tt.flag
			}

			client := NewWithProvider(provider)
			ctx := WithContext(context.Background(), tt.context)

			key := "test-flag"
			if tt.flag != nil {
				key = tt.flag.Key
			}

			got := client.Bool(ctx, key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_String(t *testing.T) {
	tests := []struct {
		name         string
		flag         *Flag
		context      *Context
		defaultValue string
		want         string
	}{
		{
			name: "returns string variant",
			flag: &Flag{
				Key:          "theme-flag",
				Type:         FlagTypeString,
				DefaultValue: "light",
				Enabled:      true,
				Variants: []Variant{
					{Name: "light", Value: "light"},
					{Name: "dark", Value: "dark"},
				},
				Rules: []Rule{
					{Variation: 1},
				},
			},
			context:      NewContext("user-1"),
			defaultValue: "light",
			want:         "dark",
		},
		{
			name: "disabled flag returns default",
			flag: &Flag{
				Key:          "theme-flag",
				Type:         FlagTypeString,
				DefaultValue: "default-theme",
				Enabled:      false,
			},
			context:      NewContext("user-1"),
			defaultValue: "fallback",
			want:         "default-theme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{
				flags: make(map[string]*Flag),
			}
			if tt.flag != nil {
				provider.flags[tt.flag.Key] = tt.flag
			}

			client := NewWithProvider(provider)
			ctx := WithContext(context.Background(), tt.context)

			got := client.String(ctx, tt.flag.Key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Int(t *testing.T) {
	tests := []struct {
		name         string
		flag         *Flag
		defaultValue int
		want         int
	}{
		{
			name: "returns int variant",
			flag: &Flag{
				Key:          "limit-flag",
				Type:         FlagTypeInt,
				DefaultValue: 10,
				Enabled:      true,
				Variants: []Variant{
					{Name: "low", Value: 10},
					{Name: "high", Value: 100},
				},
				Rules: []Rule{
					{Variation: 1},
				},
			},
			defaultValue: 10,
			want:         100,
		},
		{
			name: "returns float64 as int",
			flag: &Flag{
				Key:          "limit-flag",
				Type:         FlagTypeInt,
				DefaultValue: 10,
				Enabled:      true,
				Variants: []Variant{
					{Name: "value", Value: float64(50)},
				},
				Rules: []Rule{
					{Variation: 0},
				},
			},
			defaultValue: 10,
			want:         50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{
				flags: map[string]*Flag{tt.flag.Key: tt.flag},
			}

			client := NewWithProvider(provider)
			ctx := WithContext(context.Background(), NewContext("user-1"))

			got := client.Int(ctx, tt.flag.Key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Float(t *testing.T) {
	flag := &Flag{
		Key:          "rate-flag",
		Type:         FlagTypeFloat,
		DefaultValue: 1.0,
		Enabled:      true,
		Variants: []Variant{
			{Name: "low", Value: 1.0},
			{Name: "high", Value: 2.5},
		},
		Rules: []Rule{
			{Variation: 1},
		},
	}

	provider := &mockProvider{
		flags: map[string]*Flag{flag.Key: flag},
	}

	client := NewWithProvider(provider)
	ctx := WithContext(context.Background(), NewContext("user-1"))

	got := client.Float(ctx, flag.Key, 1.0)
	want := 2.5
	if got != want {
		t.Errorf("Float() = %v, want %v", got, want)
	}
}

func TestClient_WithOverrides(t *testing.T) {
	flag := &Flag{
		Key:          "test-flag",
		Type:         FlagTypeBool,
		DefaultValue: false,
		Enabled:      true,
		Variants: []Variant{
			{Name: "off", Value: false},
			{Name: "on", Value: true},
		},
		Rules: []Rule{
			{Variation: 0}, // Would return false normally
		},
	}

	provider := &mockProvider{
		flags: map[string]*Flag{flag.Key: flag},
	}

	// Override to true
	client := NewWithProvider(provider, WithOverrides(map[string]interface{}{
		"test-flag": true,
	}))

	ctx := context.Background()
	got := client.Bool(ctx, "test-flag", false)
	if !got {
		t.Errorf("Bool() with override = %v, want true", got)
	}
}

func TestClient_Variation(t *testing.T) {
	flag := &Flag{
		Key:          "experiment",
		Type:         FlagTypeString,
		DefaultValue: "control",
		Enabled:      true,
		Variants: []Variant{
			{Name: "control", Value: "control"},
			{Name: "variant-a", Value: "variant-a"},
			{Name: "variant-b", Value: "variant-b"},
		},
		Rules: []Rule{
			{
				ID:        "rule-1",
				Variation: 1,
			},
		},
	}

	provider := &mockProvider{
		flags: map[string]*Flag{flag.Key: flag},
	}

	client := NewWithProvider(provider)
	ctx := WithContext(context.Background(), NewContext("user-1"))

	eval, err := client.Variation(ctx, flag.Key)
	if err != nil {
		t.Fatalf("Variation() error = %v", err)
	}

	if eval.Key != flag.Key {
		t.Errorf("Evaluation.Key = %v, want %v", eval.Key, flag.Key)
	}
	if eval.Value != "variant-a" {
		t.Errorf("Evaluation.Value = %v, want variant-a", eval.Value)
	}
	if eval.VariationIdx != 1 {
		t.Errorf("Evaluation.VariationIdx = %v, want 1", eval.VariationIdx)
	}
	if eval.Reason != ReasonRuleMatch {
		t.Errorf("Evaluation.Reason = %v, want %v", eval.Reason, ReasonRuleMatch)
	}
	if eval.RuleID != "rule-1" {
		t.Errorf("Evaluation.RuleID = %v, want rule-1", eval.RuleID)
	}
}

func TestClient_AllFlags(t *testing.T) {
	flags := map[string]*Flag{
		"flag-1": {
			Key:          "flag-1",
			Type:         FlagTypeBool,
			DefaultValue: true,
			Enabled:      true,
		},
		"flag-2": {
			Key:          "flag-2",
			Type:         FlagTypeString,
			DefaultValue: "value",
			Enabled:      true,
		},
	}

	provider := &mockProvider{flags: flags}
	client := NewWithProvider(provider)
	ctx := WithContext(context.Background(), NewContext("user-1"))

	allFlags := client.AllFlags(ctx)
	if len(allFlags) != 2 {
		t.Errorf("AllFlags() returned %d flags, want 2", len(allFlags))
	}
}

func TestClient_DisabledFlag(t *testing.T) {
	flag := &Flag{
		Key:          "disabled-flag",
		Type:         FlagTypeBool,
		DefaultValue: true,
		Enabled:      false,
		Variants: []Variant{
			{Name: "off", Value: false},
			{Name: "on", Value: true},
		},
		Rules: []Rule{
			{Variation: 0}, // Would return false
		},
	}

	provider := &mockProvider{
		flags: map[string]*Flag{flag.Key: flag},
	}

	client := NewWithProvider(provider)
	ctx := WithContext(context.Background(), NewContext("user-1"))

	eval, err := client.Variation(ctx, flag.Key)
	if err != nil {
		t.Fatalf("Variation() error = %v", err)
	}

	if eval.Reason != ReasonOff {
		t.Errorf("Evaluation.Reason = %v, want %v", eval.Reason, ReasonOff)
	}
	if eval.Value != true { // Should return DefaultValue when disabled
		t.Errorf("Evaluation.Value = %v, want true (default)", eval.Value)
	}
}

// mockProvider is a simple mock provider for testing.
type mockProvider struct {
	flags map[string]*Flag
}

func (m *mockProvider) GetFlag(ctx context.Context, key string) (*Flag, error) {
	flag, ok := m.flags[key]
	if !ok {
		return nil, ErrFlagNotFound
	}
	return flag, nil
}

func (m *mockProvider) GetAllFlags(ctx context.Context) (map[string]*Flag, error) {
	return m.flags, nil
}

func (m *mockProvider) SetFlag(ctx context.Context, flag *Flag) error {
	m.flags[flag.Key] = flag
	return nil
}

func (m *mockProvider) DeleteFlag(ctx context.Context, key string) error {
	delete(m.flags, key)
	return nil
}

func (m *mockProvider) Close() error {
	return nil
}
