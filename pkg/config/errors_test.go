package config

import (
	"errors"
	"testing"
)

func TestBindError_Error(t *testing.T) {
	err := &BindError{
		Field:   "Host",
		Tag:     "required",
		Value:   nil,
		Message: "required field missing",
	}

	if err.Error() != "required field missing" {
		t.Errorf("Error() = %v, want %v", err.Error(), "required field missing")
	}
}

func TestMultiBindError_Error(t *testing.T) {
	tests := []struct {
		name   string
		errors []BindError
		want   string
	}{
		{
			name:   "no errors",
			errors: nil,
			want:   "config: no binding errors",
		},
		{
			name: "single error",
			errors: []BindError{
				{Field: "Host", Message: "required field missing"},
			},
			want: "required field missing",
		},
		{
			name: "multiple errors",
			errors: []BindError{
				{Field: "Host", Message: "required field missing"},
				{Field: "Port", Message: "invalid type"},
			},
			want: "config: multiple binding errors occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &MultiBindError{Errors: tt.errors}
			if err.Error() != tt.want {
				t.Errorf("Error() = %v, want %v", err.Error(), tt.want)
			}
		})
	}
}

func TestMultiBindError_Add(t *testing.T) {
	err := &MultiBindError{}
	err.Add(BindError{Field: "Host", Message: "test"})

	if len(err.Errors) != 1 {
		t.Errorf("Add() len = %d, want 1", len(err.Errors))
	}

	if err.Errors[0].Field != "Host" {
		t.Errorf("Add() field = %v, want %v", err.Errors[0].Field, "Host")
	}
}

func TestMultiBindError_HasErrors(t *testing.T) {
	err := &MultiBindError{}

	if err.HasErrors() {
		t.Error("HasErrors() = true for empty, want false")
	}

	err.Add(BindError{Field: "Host", Message: "test"})

	if !err.HasErrors() {
		t.Error("HasErrors() = false after Add, want true")
	}
}

func TestErrors_Are(t *testing.T) {
	// Test that errors can be checked with errors.Is
	tests := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrInvalidType", ErrInvalidType},
		{"ErrRequired", ErrRequired},
		{"ErrValidation", ErrValidation},
		{"ErrProviderFailed", ErrProviderFailed},
		{"ErrBindFailed", ErrBindFailed},
		{"ErrInvalidConfig", ErrInvalidConfig},
		{"ErrParseError", ErrParseError},
		{"ErrWatchFailed", ErrWatchFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.err) {
				t.Errorf("errors.Is() = false for %v", tt.name)
			}
		})
	}
}
