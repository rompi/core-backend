package auth_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rompi/core-backend/pkg/auth"
	_ "modernc.org/sqlite"
)

func TestIntegration_RegisterLoginPasswordReset(t *testing.T) {
	db := newSQLiteDB(t)

	users := &sqliteUserRepo{db: db}
	sessions := &sqliteSessionRepo{db: db}
	resets := &sqlitePasswordResetRepo{db: db}

	cfg := newTestConfig()
	cfg.JWTSecret = "integration"
	cfg.ResetTokenLength = 32
	cfg.ResetTokenExpiration = time.Minute

	svc, err := auth.NewService(cfg, auth.Repositories{
		Users:               users,
		Sessions:            sessions,
		PasswordResetTokens: resets,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	ctx := context.Background()
	email := fmt.Sprintf("integration+%d@rompi.com", time.Now().UnixNano())
	req := auth.RegisterRequest{Email: email, Password: "IntPass1!", Language: "en"}
	user, err := svc.Register(ctx, req)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Email != email {
		t.Fatalf("unexpected email %s", user.Email)
	}

	login, err := svc.Login(ctx, auth.LoginRequest{Email: email, Password: req.Password})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if login.Token == "" {
		t.Fatalf("expected token")
	}

	resetToken, err := svc.InitiatePasswordReset(ctx, email)
	if err != nil {
		t.Fatalf("initiate reset: %v", err)
	}

	if err := svc.CompletePasswordReset(ctx, resetToken.Token, "NewPass1!"); err != nil {
		t.Fatalf("complete reset: %v", err)
	}

	if _, err := svc.Login(ctx, auth.LoginRequest{Email: email, Password: "NewPass1!"}); err != nil {
		t.Fatalf("login after reset: %v", err)
	}
}

func newSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := createTables(db); err != nil {
		t.Fatalf("create tables: %v", err)
	}
	return db
}

func createTables(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            email TEXT UNIQUE,
            password_hash TEXT,
            created_at INTEGER,
            updated_at INTEGER,
            failed_attempts INTEGER,
            locked_until INTEGER,
            language TEXT,
            metadata TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS sessions (
            token TEXT PRIMARY KEY,
            user_id TEXT,
            issued_at INTEGER,
            expires_at INTEGER,
            revoked INTEGER,
            metadata TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS password_reset_tokens (
            token TEXT PRIMARY KEY,
            user_id TEXT,
            issued_at INTEGER,
            expires_at INTEGER,
            used INTEGER
        )`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func encodeMetadata(meta map[string]interface{}) (string, error) {
	if len(meta) == 0 {
		return "", nil
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeMetadata(raw sql.NullString) map[string]interface{} {
	if !raw.Valid || raw.String == "" {
		return nil
	}
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(raw.String), &meta); err != nil {
		return nil
	}
	return meta
}

type sqliteUserRepo struct {
	db *sql.DB
}

func (r *sqliteUserRepo) Create(ctx context.Context, user *auth.User) error {
	meta, err := encodeMetadata(user.Metadata)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`INSERT INTO users (id, email, password_hash, created_at, updated_at, failed_attempts, locked_until, language, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.PasswordHash, user.CreatedAt.UnixNano(), user.UpdatedAt.UnixNano(), user.FailedAttempts, user.LockedUntil.UnixNano(), user.Language, meta)
	return err
}

func (r *sqliteUserRepo) GetByID(ctx context.Context, id string) (*auth.User, error) {
	row := r.db.QueryRow(`SELECT id, email, password_hash, created_at, updated_at, failed_attempts, locked_until, language, metadata FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func (r *sqliteUserRepo) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	row := r.db.QueryRow(`SELECT id, email, password_hash, created_at, updated_at, failed_attempts, locked_until, language, metadata FROM users WHERE email = ?`, email)
	return scanUser(row)
}

func (r *sqliteUserRepo) Update(ctx context.Context, user *auth.User) error {
	meta, err := encodeMetadata(user.Metadata)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`UPDATE users SET password_hash=?, updated_at=?, failed_attempts=?, locked_until=?, language=?, metadata=? WHERE id=?`,
		user.PasswordHash, user.UpdatedAt.UnixNano(), user.FailedAttempts, user.LockedUntil.UnixNano(), user.Language, meta, user.ID)
	return err
}

func (r *sqliteUserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

func (r *sqliteUserRepo) IncrementFailedAttempts(ctx context.Context, userID string) error {
	_, err := r.db.Exec(`UPDATE users SET failed_attempts = failed_attempts + 1 WHERE id = ?`, userID)
	return err
}

func (r *sqliteUserRepo) ResetFailedAttempts(ctx context.Context, userID string) error {
	_, err := r.db.Exec(`UPDATE users SET failed_attempts = 0, locked_until = 0 WHERE id = ?`, userID)
	return err
}

func (r *sqliteUserRepo) LockAccount(ctx context.Context, userID string) error {
	_, err := r.db.Exec(`UPDATE users SET locked_until = ? WHERE id = ?`, time.Now().Add(15*time.Minute).UnixNano(), userID)
	return err
}

func (r *sqliteUserRepo) UnlockAccount(ctx context.Context, userID string) error {
	_, err := r.db.Exec(`UPDATE users SET locked_until = 0 WHERE id = ?`, userID)
	return err
}

func scanUser(scanner interface{ Scan(...interface{}) error }) (*auth.User, error) {
	var (
		id                   string
		email                string
		hash                 string
		createdAt, updatedAt int64
		failedAttempts       int
		lockedUntil          int64
		language             string
		metadata             sql.NullString
	)
	err := scanner.Scan(&id, &email, &hash, &createdAt, &updatedAt, &failedAttempts, &lockedUntil, &language, &metadata)
	if err == sql.ErrNoRows {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &auth.User{
		ID:             id,
		Email:          email,
		PasswordHash:   hash,
		CreatedAt:      time.Unix(0, createdAt),
		UpdatedAt:      time.Unix(0, updatedAt),
		FailedAttempts: failedAttempts,
		LockedUntil:    time.Unix(0, lockedUntil),
		Language:       language,
		Metadata:       decodeMetadata(metadata),
	}, nil
}

// sqliteSessionRepo stores sessions for integration tests.
type sqliteSessionRepo struct {
	db *sql.DB
}

func (r *sqliteSessionRepo) Create(ctx context.Context, session *auth.Session) error {
	_, err := r.db.Exec(`INSERT OR REPLACE INTO sessions (token, user_id, issued_at, expires_at, revoked, metadata) VALUES (?, ?, ?, ?, ?, ?)`,
		session.Token, session.UserID, session.IssuedAt.UnixNano(), session.ExpiresAt.UnixNano(), boolToInt(session.Revoked), session.Metadata)
	return err
}

func (r *sqliteSessionRepo) GetByToken(ctx context.Context, token string) (*auth.Session, error) {
	row := r.db.QueryRow(`SELECT token, user_id, issued_at, expires_at, revoked, metadata FROM sessions WHERE token = ?`, token)
	var s auth.Session
	var issuedAt, expiresAt int64
	var revoked int
	if err := row.Scan(&s.Token, &s.UserID, &issuedAt, &expiresAt, &revoked, &s.Metadata); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.IssuedAt = time.Unix(0, issuedAt)
	s.ExpiresAt = time.Unix(0, expiresAt)
	s.Revoked = revoked != 0
	return &s, nil
}

func (r *sqliteSessionRepo) GetByUserID(ctx context.Context, userID string) ([]*auth.Session, error) {
	rows, err := r.db.Query(`SELECT token, user_id, issued_at, expires_at, revoked, metadata FROM sessions WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []*auth.Session
	for rows.Next() {
		var s auth.Session
		var issuedAt, expiresAt int64
		var revoked int
		if err := rows.Scan(&s.Token, &s.UserID, &issuedAt, &expiresAt, &revoked, &s.Metadata); err != nil {
			return nil, err
		}
		s.IssuedAt = time.Unix(0, issuedAt)
		s.ExpiresAt = time.Unix(0, expiresAt)
		s.Revoked = revoked != 0
		sessions = append(sessions, &s)
	}
	return sessions, nil
}

func (r *sqliteSessionRepo) Delete(ctx context.Context, token string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (r *sqliteSessionRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at <= ?`, time.Now().UnixNano())
	return err
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// sqlitePasswordResetRepo handles reset tokens.
type sqlitePasswordResetRepo struct {
	db *sql.DB
}

func (r *sqlitePasswordResetRepo) Create(ctx context.Context, token *auth.PasswordResetToken) error {
	_, err := r.db.Exec(`INSERT INTO password_reset_tokens (token, user_id, issued_at, expires_at, used) VALUES (?, ?, ?, ?, ?)`,
		token.Token, token.UserID, token.IssuedAt.UnixNano(), token.ExpiresAt.UnixNano(), 0)
	return err
}

func (r *sqlitePasswordResetRepo) GetByToken(ctx context.Context, token string) (*auth.PasswordResetToken, error) {
	row := r.db.QueryRow(`SELECT token, user_id, issued_at, expires_at, used FROM password_reset_tokens WHERE token = ?`, token)
	var t auth.PasswordResetToken
	var issuedAt, expiresAt int64
	var used int
	if err := row.Scan(&t.Token, &t.UserID, &issuedAt, &expiresAt, &used); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t.IssuedAt = time.Unix(0, issuedAt)
	t.ExpiresAt = time.Unix(0, expiresAt)
	t.Used = used != 0
	return &t, nil
}

func (r *sqlitePasswordResetRepo) Delete(ctx context.Context, token string) error {
	_, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE token = ?`, token)
	return err
}

func (r *sqlitePasswordResetRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(`DELETE FROM password_reset_tokens WHERE expires_at <= ?`, time.Now().UnixNano())
	return err
}
