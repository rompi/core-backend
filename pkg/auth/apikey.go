package auth

import "time"

// IsValid indicates whether the API key can be used at the provided time.
func (k *APIKey) IsValid(now time.Time) bool {
	if k == nil || k.Revoked {
		return false
	}
	if k.ExpiresAt.IsZero() {
		return true
	}
	return now.Before(k.ExpiresAt)
}
