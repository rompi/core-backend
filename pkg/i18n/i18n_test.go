package i18n

import (
	"context"
	"testing"

	"github.com/rompi/core-backend/pkg/i18n/catalog"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				DefaultLocale:      "en",
				FallbackLocale:     "en",
				CatalogType:        CatalogTypeJSON,
				Path:               "./testdata/locales",
				MissingKeyBehavior: MissingKeyReturnKey,
			},
			wantErr: false,
		},
		{
			name: "empty default locale",
			cfg: Config{
				DefaultLocale:      "",
				FallbackLocale:     "en",
				CatalogType:        CatalogTypeJSON,
				MissingKeyBehavior: MissingKeyReturnKey,
			},
			wantErr: true,
		},
		{
			name: "empty fallback locale",
			cfg: Config{
				DefaultLocale:      "en",
				FallbackLocale:     "",
				CatalogType:        CatalogTypeJSON,
				MissingKeyBehavior: MissingKeyReturnKey,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestI18n_Translation(t *testing.T) {
	// Create in-memory catalog
	cat := catalog.NewInMemoryCatalog()
	cat.AddSimpleMessage("en", "greeting", "Hello, {{.Arg0}}!")
	cat.AddSimpleMessage("en", "welcome", "Welcome")
	cat.AddSimpleMessage("es", "greeting", "¡Hola, {{.Arg0}}!")
	cat.AddSimpleMessage("es", "welcome", "Bienvenido")
	cat.AddPluralMessage("en", "items", "{{.Count}} item", "{{.Count}} items")
	cat.AddPluralMessage("es", "items", "{{.Count}} artículo", "{{.Count}} artículos")

	i, err := New(Config{
		DefaultLocale:      "en",
		FallbackLocale:     "en",
		MissingKeyBehavior: MissingKeyReturnKey,
	}, WithCatalog(&catalogAdapter{cat: cat}))
	if err != nil {
		t.Fatalf("Failed to create i18n: %v", err)
	}

	ctx := context.Background()

	// Test basic translation
	t.Run("T", func(t *testing.T) {
		got := i.T(ctx, "greeting", "World")
		want := "Hello, World!"
		if got != want {
			t.Errorf("T() = %q, want %q", got, want)
		}
	})

	// Test translation with locale switch
	t.Run("T with locale", func(t *testing.T) {
		ctx := i.WithLocale(ctx, "es")
		got := i.T(ctx, "greeting", "Mundo")
		want := "¡Hola, Mundo!"
		if got != want {
			t.Errorf("T() = %q, want %q", got, want)
		}
	})

	// Test pluralization
	t.Run("Tn singular", func(t *testing.T) {
		got := i.Tn(ctx, "items", 1)
		want := "1 item"
		if got != want {
			t.Errorf("Tn(1) = %q, want %q", got, want)
		}
	})

	t.Run("Tn plural", func(t *testing.T) {
		got := i.Tn(ctx, "items", 5)
		want := "5 items"
		if got != want {
			t.Errorf("Tn(5) = %q, want %q", got, want)
		}
	})

	// Test missing key
	t.Run("missing key", func(t *testing.T) {
		got := i.T(ctx, "nonexistent")
		want := "nonexistent"
		if got != want {
			t.Errorf("T() = %q, want %q", got, want)
		}
	})
}

func TestI18n_Tf(t *testing.T) {
	cat := catalog.NewInMemoryCatalog()
	cat.AddSimpleMessage("en", "order", "Order #{{.OrderID}} for {{.Name}}")

	i, err := New(Config{
		DefaultLocale:      "en",
		FallbackLocale:     "en",
		MissingKeyBehavior: MissingKeyReturnKey,
	}, WithCatalog(&catalogAdapter{cat: cat}))
	if err != nil {
		t.Fatalf("Failed to create i18n: %v", err)
	}

	ctx := context.Background()

	got := i.Tf(ctx, "order", map[string]interface{}{
		"OrderID": "12345",
		"Name":    "John",
	})
	want := "Order #12345 for John"
	if got != want {
		t.Errorf("Tf() = %q, want %q", got, want)
	}
}

func TestI18n_Localizer(t *testing.T) {
	cat := catalog.NewInMemoryCatalog()
	cat.AddSimpleMessage("en", "hello", "Hello")
	cat.AddSimpleMessage("es", "hello", "Hola")

	i, err := New(Config{
		DefaultLocale:      "en",
		FallbackLocale:     "en",
		MissingKeyBehavior: MissingKeyReturnKey,
	}, WithCatalog(&catalogAdapter{cat: cat}))
	if err != nil {
		t.Fatalf("Failed to create i18n: %v", err)
	}

	en := i.L("en")
	es := i.L("es")

	if got := en.T("hello"); got != "Hello" {
		t.Errorf("en.T() = %q, want %q", got, "Hello")
	}

	if got := es.T("hello"); got != "Hola" {
		t.Errorf("es.T() = %q, want %q", got, "Hola")
	}

	// Test locale info
	if got := en.Locale(); got != "en" {
		t.Errorf("en.Locale() = %q, want %q", got, "en")
	}

	if got := en.Direction(); got != LTR {
		t.Errorf("en.Direction() = %q, want %q", got, LTR)
	}

	ar := i.L("ar")
	if got := ar.Direction(); got != RTL {
		t.Errorf("ar.Direction() = %q, want %q", got, RTL)
	}
}

func TestI18n_Locales(t *testing.T) {
	cat := catalog.NewInMemoryCatalog()
	cat.AddSimpleMessage("en", "test", "Test")
	cat.AddSimpleMessage("es", "test", "Prueba")
	cat.AddSimpleMessage("fr", "test", "Essai")

	i, err := New(Config{
		DefaultLocale:      "en",
		FallbackLocale:     "en",
		MissingKeyBehavior: MissingKeyReturnKey,
	}, WithCatalog(&catalogAdapter{cat: cat}))
	if err != nil {
		t.Fatalf("Failed to create i18n: %v", err)
	}

	locales := i.Locales()
	if len(locales) != 3 {
		t.Errorf("Locales() returned %d locales, want 3", len(locales))
	}
}

func TestI18n_WithLocale(t *testing.T) {
	cat := catalog.NewInMemoryCatalog()
	cat.AddSimpleMessage("en", "hello", "Hello")
	cat.AddSimpleMessage("es", "hello", "Hola")

	i, err := New(Config{
		DefaultLocale:      "en",
		FallbackLocale:     "en",
		MissingKeyBehavior: MissingKeyReturnKey,
	}, WithCatalog(&catalogAdapter{cat: cat}))
	if err != nil {
		t.Fatalf("Failed to create i18n: %v", err)
	}

	ctx := context.Background()

	// Default locale
	if got := i.Locale(ctx); got != "en" {
		t.Errorf("Locale() = %q, want %q", got, "en")
	}

	// With Spanish locale
	ctx = i.WithLocale(ctx, "es")
	if got := i.Locale(ctx); got != "es" {
		t.Errorf("Locale() = %q, want %q", got, "es")
	}

	// Translation should use Spanish
	if got := i.T(ctx, "hello"); got != "Hola" {
		t.Errorf("T() = %q, want %q", got, "Hola")
	}
}

func TestMissingKeyBehavior(t *testing.T) {
	cat := catalog.NewInMemoryCatalog()
	cat.AddSimpleMessage("en", "exists", "Exists")

	tests := []struct {
		name     string
		behavior MissingKeyBehavior
		want     string
	}{
		{
			name:     "return key",
			behavior: MissingKeyReturnKey,
			want:     "missing",
		},
		{
			name:     "return empty",
			behavior: MissingKeyReturnEmpty,
			want:     "",
		},
		{
			name:     "return error marker",
			behavior: MissingKeyReturnError,
			want:     "[MISSING: missing]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, err := New(Config{
				DefaultLocale:      "en",
				FallbackLocale:     "en",
				MissingKeyBehavior: tt.behavior,
			}, WithCatalog(&catalogAdapter{cat: cat}))
			if err != nil {
				t.Fatalf("Failed to create i18n: %v", err)
			}

			ctx := context.Background()
			got := i.T(ctx, "missing")
			if got != tt.want {
				t.Errorf("T() = %q, want %q", got, tt.want)
			}
		})
	}
}
