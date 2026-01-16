package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unique violation",
			err:  &pgconn.PgError{Code: "23505"},
			want: true,
		},
		{
			name: "other pg error",
			err:  &pgconn.PgError{Code: "23503"},
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUniqueViolation(tt.err); got != tt.want {
				t.Errorf("IsUniqueViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "foreign key violation",
			err:  &pgconn.PgError{Code: "23503"},
			want: true,
		},
		{
			name: "unique violation",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForeignKeyViolation(tt.err); got != tt.want {
				t.Errorf("IsForeignKeyViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConstraintViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unique violation",
			err:  &pgconn.PgError{Code: "23505"},
			want: true,
		},
		{
			name: "foreign key violation",
			err:  &pgconn.PgError{Code: "23503"},
			want: true,
		},
		{
			name: "check violation",
			err:  &pgconn.PgError{Code: "23514"},
			want: true,
		},
		{
			name: "not null violation",
			err:  &pgconn.PgError{Code: "23502"},
			want: true,
		},
		{
			name: "other error",
			err:  &pgconn.PgError{Code: "42000"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConstraintViolation(tt.err); got != tt.want {
				t.Errorf("IsConstraintViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "pg error",
			err:  &pgconn.PgError{Code: "23505"},
			want: "23505",
		},
		{
			name: "regular error",
			err:  errors.New("some error"),
			want: "",
		},
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorCode(tt.err); got != tt.want {
				t.Errorf("ErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstraintName(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "pg error with constraint",
			err:  &pgconn.PgError{Code: "23505", ConstraintName: "users_email_key"},
			want: "users_email_key",
		},
		{
			name: "pg error without constraint",
			err:  &pgconn.PgError{Code: "23505"},
			want: "",
		},
		{
			name: "regular error",
			err:  errors.New("some error"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConstraintName(tt.err); got != tt.want {
				t.Errorf("ConstraintName() = %v, want %v", got, tt.want)
			}
		})
	}
}
