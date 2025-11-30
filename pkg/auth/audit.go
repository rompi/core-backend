package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditLogger writes audit events to the configured repository.
type AuditLogger struct {
	repo AuditLogRepository
}

// NewAuditLogger returns an AuditLogger backed by repo. If repo is nil, it returns nil.
func NewAuditLogger(repo AuditLogRepository) *AuditLogger {
	if repo == nil {
		return nil
	}
	return &AuditLogger{repo: repo}
}

// Log creates a new audit entry when repo is configured.
func (l *AuditLogger) Log(ctx context.Context, userID, action, message string, metadata map[string]interface{}) error {
	if l == nil || l.repo == nil {
		return nil
	}
	entry := &AuditLog{
		ID:        uuid.NewString(),
		UserID:    userID,
		Action:    action,
		Message:   message,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC(),
	}
	return l.repo.Create(ctx, entry)
}
