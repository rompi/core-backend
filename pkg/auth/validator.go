package auth

import (
	"errors"
	"net/mail"
	"strings"
)

var errInvalidEmail = errors.New("invalid email format")

// ValidateEmail ensures the email string is well-formed.
func ValidateEmail(email string) error {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return errInvalidEmail
	}
	if _, err := mail.ParseAddress(trimmed); err != nil {
		return errInvalidEmail
	}
	return nil
}
