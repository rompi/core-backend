package auth

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestTranslator_Message(t *testing.T) {
	translator := NewTranslator("en")
	translator.Register("es", map[string]string{CodeInvalidCredentials: "Credenciales inválidas"})

	if got := translator.Message("es", CodeInvalidCredentials); got != "Credenciales inválidas" {
		t.Fatalf("expected spanish message, got %q", got)
	}

	if got := translator.Message("fr", CodeInvalidCredentials); got != englishMessages[CodeInvalidCredentials] {
		t.Fatalf("expected fallback to english, got %q", got)
	}
}

func TestTranslator_LoadFromFile(t *testing.T) {
	translator := NewTranslator("en")
	dir, err := os.MkdirTemp(".", "translations_test_")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	path := filepath.Join(dir, "fr.json")
	if err := os.WriteFile(path, []byte(`{"invalid_token":"jeton expiré"}`), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := translator.LoadFromFile("fr", path); err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}
	if got := translator.Message("fr", CodeInvalidToken); got != "jeton expiré" {
		t.Fatalf("unexpected message %q", got)
	}
}

func TestNewAuthErrorUsesTranslator(t *testing.T) {
	prev := DefaultTranslator
	translator := NewTranslator("en")
	translator.Register("es", map[string]string{CodeInvalidCredentials: "Credenciales inválidas"})
	DefaultTranslator = translator
	t.Cleanup(func() { DefaultTranslator = prev })

	err := NewAuthError(CodeInvalidCredentials, http.StatusUnauthorized, "es", nil)
	if err.Message != "Credenciales inválidas" {
		t.Fatalf("expected translated error, got %q", err.Message)
	}
}
