package auth

import (
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the plain text password using bcrypt.
func HashPassword(password string, cost int) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hashed), nil
}

// ComparePassword validates a password against a stored bcrypt hash.
func ComparePassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("compare password: %w", err)
	}
	return nil
}

// ValidatePassword ensures the password meets the configured complexity rules.
func ValidatePassword(password string, cfg *Config) error {
	if len(password) < cfg.PasswordMinLength {
		return fmt.Errorf("%w: minimum length %d", ErrWeakPassword, cfg.PasswordMinLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if cfg.PasswordRequireUpper && !hasUpper {
		return fmt.Errorf("%w: uppercase required", ErrWeakPassword)
	}
	if cfg.PasswordRequireLower && !hasLower {
		return fmt.Errorf("%w: lowercase required", ErrWeakPassword)
	}
	if cfg.PasswordRequireNumber && !hasNumber {
		return fmt.Errorf("%w: number required", ErrWeakPassword)
	}
	if cfg.PasswordRequireSpecial && !hasSpecial {
		return fmt.Errorf("%w: special character required", ErrWeakPassword)
	}
	return nil
}
