package auth

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds authentication service configuration loaded from the environment.
type Config struct {
	JWTSecret             string        `json:"jwt_secret"`
	JWTExpirationDuration time.Duration `json:"jwt_expiration_duration"`
	JWTIssuer             string        `json:"jwt_issuer"`

	PasswordMinLength      int  `json:"password_min_length"`
	PasswordRequireUpper   bool `json:"password_require_upper"`
	PasswordRequireLower   bool `json:"password_require_lower"`
	PasswordRequireNumber  bool `json:"password_require_number"`
	PasswordRequireSpecial bool `json:"password_require_special"`
	BcryptCost             int  `json:"bcrypt_cost"`

	MaxFailedAttempts int           `json:"max_failed_attempts"`
	LockoutDuration   time.Duration `json:"lockout_duration"`

	RateLimitWindow      time.Duration `json:"rate_limit_window"`
	RateLimitMaxRequests int           `json:"rate_limit_max_requests"`

	ResetTokenLength     int           `json:"reset_token_length"`
	ResetTokenExpiration time.Duration `json:"reset_token_expiration"`

	DefaultLanguage string `json:"default_language"`
}

// LoadConfig reads configuration from environment variables and validates it.
func LoadConfig() (*Config, error) {
	cfg := defaultConfig()
	if err := cfg.overrideFromEnv(); err != nil {
		return nil, fmt.Errorf("loading auth config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating auth config: %w", err)
	}
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		JWTExpirationDuration:  24 * time.Hour,
		JWTIssuer:              "rompi-auth",
		PasswordMinLength:      8,
		PasswordRequireUpper:   true,
		PasswordRequireLower:   true,
		PasswordRequireNumber:  true,
		PasswordRequireSpecial: true,
		BcryptCost:             12,
		MaxFailedAttempts:      5,
		LockoutDuration:        15 * time.Minute,
		RateLimitWindow:        time.Minute,
		RateLimitMaxRequests:   5,
		ResetTokenLength:       32,
		ResetTokenExpiration:   time.Hour,
		DefaultLanguage:        "en",
	}
}

func (c *Config) overrideFromEnv() error {
	if v := strings.TrimSpace(os.Getenv("AUTH_JWT_SECRET")); v != "" {
		c.JWTSecret = v
	}
	if v := strings.TrimSpace(os.Getenv("AUTH_JWT_ISSUER")); v != "" {
		c.JWTIssuer = v
	}
	if d, err := parseDurationEnv("AUTH_JWT_EXPIRATION"); err != nil {
		return err
	} else if d != nil {
		c.JWTExpirationDuration = *d
	}
	if ints, err := parseIntEnv("AUTH_PASSWORD_MIN_LENGTH"); err != nil {
		return err
	} else if ints != nil {
		c.PasswordMinLength = *ints
	}
	if b, err := parseBoolEnv("AUTH_PASSWORD_REQUIRE_UPPER"); err != nil {
		return err
	} else if b != nil {
		c.PasswordRequireUpper = *b
	}
	if b, err := parseBoolEnv("AUTH_PASSWORD_REQUIRE_LOWER"); err != nil {
		return err
	} else if b != nil {
		c.PasswordRequireLower = *b
	}
	if b, err := parseBoolEnv("AUTH_PASSWORD_REQUIRE_NUMBER"); err != nil {
		return err
	} else if b != nil {
		c.PasswordRequireNumber = *b
	}
	if b, err := parseBoolEnv("AUTH_PASSWORD_REQUIRE_SPECIAL"); err != nil {
		return err
	} else if b != nil {
		c.PasswordRequireSpecial = *b
	}
	if ints, err := parseIntEnv("AUTH_BCRYPT_COST"); err != nil {
		return err
	} else if ints != nil {
		c.BcryptCost = *ints
	}
	if ints, err := parseIntEnv("AUTH_MAX_FAILED_ATTEMPTS"); err != nil {
		return err
	} else if ints != nil {
		c.MaxFailedAttempts = *ints
	}
	if d, err := parseDurationEnv("AUTH_LOCKOUT_DURATION"); err != nil {
		return err
	} else if d != nil {
		c.LockoutDuration = *d
	}
	if d, err := parseDurationEnv("AUTH_RATE_LIMIT_WINDOW"); err != nil {
		return err
	} else if d != nil {
		c.RateLimitWindow = *d
	}
	if ints, err := parseIntEnv("AUTH_RATE_LIMIT_MAX_REQUESTS"); err != nil {
		return err
	} else if ints != nil {
		c.RateLimitMaxRequests = *ints
	}
	if ints, err := parseIntEnv("AUTH_RESET_TOKEN_LENGTH"); err != nil {
		return err
	} else if ints != nil {
		c.ResetTokenLength = *ints
	}
	if d, err := parseDurationEnv("AUTH_RESET_TOKEN_EXPIRATION"); err != nil {
		return err
	} else if d != nil {
		c.ResetTokenExpiration = *d
	}
	if v := strings.TrimSpace(os.Getenv("AUTH_DEFAULT_LANGUAGE")); v != "" {
		c.DefaultLanguage = v
	}
	return nil
}

func parseBoolEnv(key string) (*bool, error) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		return &parsed, nil
	}
	return nil, nil
}

func parseIntEnv(key string) (*int, error) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		return &parsed, nil
	}
	return nil, nil
}

func parseDurationEnv(key string) (*time.Duration, error) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		return &parsed, nil
	}
	return nil, nil
}

// Validate ensures the configuration contains valid and secure values.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.JWTSecret) == "" {
		return fmt.Errorf("AUTH_JWT_SECRET is required")
	}
	if c.JWTExpirationDuration <= 0 {
		return fmt.Errorf("AUTH_JWT_EXPIRATION must be positive")
	}
	if c.PasswordMinLength < 8 {
		return fmt.Errorf("AUTH_PASSWORD_MIN_LENGTH must be at least 8")
	}
	if c.BcryptCost < 4 || c.BcryptCost > 31 {
		return fmt.Errorf("AUTH_BCRYPT_COST must be between 4 and 31")
	}
	if c.MaxFailedAttempts < 1 {
		return fmt.Errorf("AUTH_MAX_FAILED_ATTEMPTS must be at least 1")
	}
	if c.LockoutDuration < time.Minute {
		return fmt.Errorf("AUTH_LOCKOUT_DURATION must be at least 1m")
	}
	if c.RateLimitWindow < time.Second {
		return fmt.Errorf("AUTH_RATE_LIMIT_WINDOW must be at least 1s")
	}
	if c.RateLimitMaxRequests < 1 {
		return fmt.Errorf("AUTH_RATE_LIMIT_MAX_REQUESTS must be at least 1")
	}
	if c.ResetTokenLength < 16 {
		return fmt.Errorf("AUTH_RESET_TOKEN_LENGTH must be at least 16")
	}
	if c.ResetTokenExpiration < time.Minute {
		return fmt.Errorf("AUTH_RESET_TOKEN_EXPIRATION must be at least 1m")
	}
	if strings.TrimSpace(c.DefaultLanguage) == "" {
		return fmt.Errorf("AUTH_DEFAULT_LANGUAGE is required")
	}
	return nil
}
