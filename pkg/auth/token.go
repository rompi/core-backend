package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims encapsulates JWT payload data emitted by the auth package.
type Claims struct {
	jwt.RegisteredClaims
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles,omitempty"`
}

// TokenManager handles JWT creation and validation.
type TokenManager struct {
	secret     []byte
	issuer     string
	expiration time.Duration
}

// NewTokenManager returns a TokenManager configured for the provided settings.
func NewTokenManager(cfg *Config) *TokenManager {
	return &TokenManager{
		secret:     []byte(cfg.JWTSecret),
		issuer:     cfg.JWTIssuer,
		expiration: cfg.JWTExpirationDuration,
	}
}

// Generate creates a signed token for the supplied user and returns the token plus expiration time.
func (m *TokenManager) Generate(user *User) (string, time.Time, error) {
	now := time.Now().UTC()
	expiration := now.Add(m.expiration)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    m.issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
		UserID: user.ID,
		Email:  user.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing token: %w", err)
	}
	return signed, expiration, nil
}

// Validate parses and verifies a JWT token, returning the embedded claims.
func (m *TokenManager) Validate(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %T", t.Method)
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
