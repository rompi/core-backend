package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Sentinel errors for PostgreSQL operations.
var (
	ErrConnectionFailed    = errors.New("postgres: connection failed")
	ErrQueryFailed         = errors.New("postgres: query failed")
	ErrNoRows              = errors.New("postgres: no rows")
	ErrConstraintViolation = errors.New("postgres: constraint violation")
	ErrDuplicateKey        = errors.New("postgres: duplicate key")
	ErrForeignKeyViolation = errors.New("postgres: foreign key violation")
	ErrInvalidConfig       = errors.New("postgres: invalid configuration")
	ErrTimeout             = errors.New("postgres: timeout")
	ErrPoolExhausted       = errors.New("postgres: connection pool exhausted")
	ErrTxAlreadyClosed     = errors.New("postgres: transaction already closed")
)

// PostgreSQL error codes
// See: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	uniqueViolationCode     = "23505"
	foreignKeyViolationCode = "23503"
	checkViolationCode      = "23514"
	notNullViolationCode    = "23502"
)

// IsUniqueViolation checks if the error is a unique constraint violation.
func IsUniqueViolation(err error) bool {
	return hasErrorCode(err, uniqueViolationCode)
}

// IsForeignKeyViolation checks if the error is a foreign key violation.
func IsForeignKeyViolation(err error) bool {
	return hasErrorCode(err, foreignKeyViolationCode)
}

// IsCheckViolation checks if the error is a check constraint violation.
func IsCheckViolation(err error) bool {
	return hasErrorCode(err, checkViolationCode)
}

// IsNotNullViolation checks if the error is a not null constraint violation.
func IsNotNullViolation(err error) bool {
	return hasErrorCode(err, notNullViolationCode)
}

// IsConstraintViolation checks if the error is any constraint violation.
func IsConstraintViolation(err error) bool {
	return IsUniqueViolation(err) ||
		IsForeignKeyViolation(err) ||
		IsCheckViolation(err) ||
		IsNotNullViolation(err)
}

// hasErrorCode checks if the error contains a specific PostgreSQL error code.
func hasErrorCode(err error, code string) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}
	return false
}

// ErrorCode extracts the PostgreSQL error code from an error, if present.
func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// ConstraintName extracts the constraint name from a PostgreSQL error, if present.
func ConstraintName(err error) string {
	if err == nil {
		return ""
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}
