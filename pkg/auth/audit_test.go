package auth

import (
	"context"
	"testing"
)

type fakeAuditRepo struct {
	entry *AuditLog
}

func (f *fakeAuditRepo) Create(ctx context.Context, log *AuditLog) error {
	f.entry = log
	return nil
}

func (f *fakeAuditRepo) GetByUserID(ctx context.Context, userID string, limit int) ([]*AuditLog, error) {
	return nil, nil
}

func TestAuditLogger_Log(t *testing.T) {
	repo := &fakeAuditRepo{}
	logger := NewAuditLogger(repo)
	if err := logger.Log(context.Background(), "user-id", "login", "user logged in", nil); err != nil {
		t.Fatalf("Log() error = %v", err)
	}
	if repo.entry == nil {
		t.Fatal("expected audit entry created")
	}
	if repo.entry.UserID != "user-id" {
		t.Fatalf("user id = %s", repo.entry.UserID)
	}
	if repo.entry.Action != "login" {
		t.Fatalf("action = %s", repo.entry.Action)
	}
}
