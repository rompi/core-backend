package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Repositories groups the persistence contracts required by the auth service.
type Repositories struct {
	Users               UserRepository
	Sessions            SessionRepository
	Roles               RoleRepository
	AuditLogs           AuditLogRepository
	PasswordResetTokens PasswordResetTokenRepository
	APIKeys             APIKeyRepository
}

func (r Repositories) validate() error {
	if r.Users == nil {
		return errors.New("user repository is required")
	}
	return nil
}

// service implements the Service interface.
type service struct {
	cfg          *Config
	repos        Repositories
	tokenManager *TokenManager
	limiter      *RateLimiter
	audit        *AuditLogger
	now          func() time.Time
}

// NewService wires the auth service with configuration and persistence layers.
func NewService(cfg *Config, repos Repositories) (Service, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if err := repos.validate(); err != nil {
		return nil, err
	}
	return &service{
		cfg:          cfg,
		repos:        repos,
		tokenManager: NewTokenManager(cfg),
		limiter:      NewRateLimiter(cfg),
		audit:        NewAuditLogger(repos.AuditLogs),
		now:          time.Now,
	}, nil
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}
	if err := s.rateLimit(ctx, fmt.Sprintf("register:%s", email)); err != nil {
		return nil, err
	}
	if err := ValidatePassword(req.Password, s.cfg); err != nil {
		return nil, err
	}

	existing, err := s.repos.Users.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrUserAlreadyExists
	}
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("checking user existence: %w", err)
	}

	hash, err := HashPassword(req.Password, s.cfg.BcryptCost)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	language := strings.TrimSpace(req.Language)
	if language == "" {
		language = s.cfg.DefaultLanguage
	}

	user := &User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: hash,
		Language:     language,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repos.Users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	s.logEvent(ctx, user.ID, "register", "user registered", map[string]interface{}{"language": language})
	return user, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := s.rateLimit(ctx, fmt.Sprintf("login:%s", email)); err != nil {
		return nil, err
	}
	user, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}

	now := s.now()
	if user.LockedUntil.After(now) {
		return nil, ErrAccountLocked
	}

	if err := ComparePassword(user.PasswordHash, req.Password); err != nil {
		s.handleFailedAttempt(ctx, user)
		return nil, ErrInvalidCredentials
	}

	if err := s.repos.Users.ResetFailedAttempts(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("reset failed attempts: %w", err)
	}

	token, expiresAt, err := s.tokenManager.Generate(user)
	if err != nil {
		return nil, err
	}

	if s.repos.Sessions != nil {
		session := &Session{
			Token:     token,
			UserID:    user.ID,
			IssuedAt:  now,
			ExpiresAt: expiresAt,
		}
		if err := s.repos.Sessions.Create(ctx, session); err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}
	}

	s.logEvent(ctx, user.ID, "login", "user logged in", map[string]interface{}{"expires_at": expiresAt})
	return &LoginResponse{Token: token, ExpiresAt: expiresAt, User: user}, nil
}

func (s *service) Logout(ctx context.Context, token string) error {
	if s.repos.Sessions == nil || token == "" {
		return nil
	}
	return s.repos.Sessions.Delete(ctx, token)
}

func (s *service) ValidateToken(ctx context.Context, token string) (*User, error) {
	claims, err := s.tokenManager.Validate(token)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}
	user, err := s.repos.Users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	return user, nil
}

func (s *service) RefreshToken(ctx context.Context, token string) (*LoginResponse, error) {
	user, err := s.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}
	newToken, expiresAt, err := s.tokenManager.Generate(user)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{Token: newToken, ExpiresAt: expiresAt, User: user}, nil
}

func (s *service) InitiatePasswordReset(ctx context.Context, email string) (*PasswordResetToken, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if err := ValidateEmail(normalized); err != nil {
		return nil, err
	}
	if err := s.rateLimit(ctx, fmt.Sprintf("password_reset:%s", normalized)); err != nil {
		return nil, err
	}
	if s.repos.PasswordResetTokens == nil {
		return nil, errors.New("password reset token repository is required")
	}

	user, err := s.repos.Users.GetByEmail(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}

	token, err := generateRandomToken(s.cfg.ResetTokenLength)
	if err != nil {
		return nil, err
	}

	now := s.now()
	reset := &PasswordResetToken{
		Token:     token,
		UserID:    user.ID,
		IssuedAt:  now,
		ExpiresAt: now.Add(s.cfg.ResetTokenExpiration),
	}

	if err := s.repos.PasswordResetTokens.Create(ctx, reset); err != nil {
		return nil, fmt.Errorf("store reset token: %w", err)
	}
	s.logEvent(ctx, user.ID, "password_reset_initiated", "password reset requested", map[string]interface{}{"expires_at": reset.ExpiresAt})
	return reset, nil
}

func (s *service) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	if token == "" {
		return fmt.Errorf("%w: token is required", ErrInvalidResetToken)
	}
	if s.repos.PasswordResetTokens == nil {
		return errors.New("password reset token repository is required")
	}
	if err := ValidatePassword(newPassword, s.cfg); err != nil {
		return err
	}

	reset, err := s.repos.PasswordResetTokens.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("fetch reset token: %w", err)
	}
	if reset == nil || reset.Used || s.now().After(reset.ExpiresAt) {
		return ErrInvalidResetToken
	}

	user, err := s.repos.Users.GetByID(ctx, reset.UserID)
	if err != nil {
		return fmt.Errorf("fetch user: %w", err)
	}

	hash, err := HashPassword(newPassword, s.cfg.BcryptCost)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.UpdatedAt = s.now().UTC()

	if err := s.repos.Users.Update(ctx, user); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	if err := s.repos.PasswordResetTokens.Delete(ctx, token); err != nil {
		return fmt.Errorf("delete reset token: %w", err)
	}
	_ = s.repos.Users.ResetFailedAttempts(ctx, user.ID)
	_ = s.repos.Users.UnlockAccount(ctx, user.ID)
	s.logEvent(ctx, user.ID, "password_reset_completed", "password reset completed", nil)
	return nil
}

func (s *service) ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetch user: %w", err)
	}
	if err := ComparePassword(user.PasswordHash, oldPassword); err != nil {
		s.handleFailedAttempt(ctx, user)
		return ErrInvalidCredentials
	}
	if err := ValidatePassword(newPassword, s.cfg); err != nil {
		return err
	}

	hash, err := HashPassword(newPassword, s.cfg.BcryptCost)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.UpdatedAt = s.now().UTC()

	if err := s.repos.Users.Update(ctx, user); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	_ = s.repos.Users.ResetFailedAttempts(ctx, user.ID)
	_ = s.repos.Users.UnlockAccount(ctx, user.ID)
	s.logEvent(ctx, user.ID, "password_changed", "password changed", nil)
	return nil
}

func (s *service) ValidateAPIKey(ctx context.Context, apiKey string) (*User, error) {
	if s.repos.APIKeys == nil {
		return nil, errors.New("API key repository is required")
	}
	key, err := s.repos.APIKeys.GetByKey(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("fetch api key: %w", err)
	}
	if key == nil || !key.IsValid(s.now()) {
		return nil, ErrInvalidToken
	}

	user, err := s.repos.Users.GetByID(ctx, key.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	return user, nil
}

func (s *service) GetUserRoles(ctx context.Context, userID string) ([]Role, error) {
	if s.repos.Roles == nil {
		return nil, errors.New("role repository is required")
	}
	return s.repos.Roles.GetByUserID(ctx, userID)
}

func (s *service) CheckPermission(ctx context.Context, userID string, permission string) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, role := range roles {
		for _, perm := range role.Permissions {
			if perm == permission {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *service) handleFailedAttempt(ctx context.Context, user *User) {
	if err := s.repos.Users.IncrementFailedAttempts(ctx, user.ID); err != nil {
		return
	}
	user.FailedAttempts++
	if user.FailedAttempts >= s.cfg.MaxFailedAttempts {
		if err := s.repos.Users.LockAccount(ctx, user.ID); err == nil {
			s.logEvent(ctx, user.ID, "account_locked", "account locked due to failed login attempts", nil)
		}
	}
}

func (s *service) rateLimit(ctx context.Context, key string) error {
	if s.limiter == nil || key == "" {
		return nil
	}
	if !s.limiter.Allow(key) {
		s.logEvent(ctx, "", "rate_limit_exceeded", "rate limit exceeded", map[string]interface{}{"key": key})
		return ErrRateLimitExceeded
	}
	return nil
}

func (s *service) logEvent(ctx context.Context, userID, action, message string, metadata map[string]interface{}) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Log(ctx, userID, action, message, metadata)
}

func generateRandomToken(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("token length must be positive")
	}
	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}
	token := hex.EncodeToString(buffer)
	if len(token) > length {
		token = token[:length]
	}
	return token, nil
}
