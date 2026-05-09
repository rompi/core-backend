package redis

import (
	"errors"
	"fmt"
)

// Error codes for Redis operations.
const (
	CodeConnectionFailed      = "connection_failed"
	CodeKeyNotFound           = "key_not_found"
	CodeSerializationFailed   = "serialization_failed"
	CodeDeserializationFailed = "deserialization_failed"
	CodeOperationFailed       = "operation_failed"
	CodeTimeout               = "timeout"
	CodeLockFailed            = "lock_failed"
	CodeLockNotHeld           = "lock_not_held"
)

var (
	ErrKeyNotFound   = errors.New("redis: key not found")
	ErrLockFailed    = errors.New("redis: failed to acquire lock")
	ErrLockNotHeld   = errors.New("redis: lock not held")
	ErrNil           = errors.New("redis: nil")
)

// RedisError is a structured error type.
type RedisError struct {
	Code      string
	Message   string
	Operation string
	Key       string
	Cause     error
}

func (e *RedisError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("redis: %s: %s (key=%s)", e.Code, e.Message, e.Key)
	}
	return fmt.Sprintf("redis: %s: %s", e.Code, e.Message)
}

func (e *RedisError) Unwrap() error { return e.Cause }

func NewErrorWithCause(code, message string, cause error) *RedisError {
	return &RedisError{Code: code, Message: message, Cause: cause}
}

func NewErrorWithKey(code, message, key string) *RedisError {
	return &RedisError{Code: code, Message: message, Key: key}
}

func (e *RedisError) WithKey(key string) *RedisError {
	e.Key = key
	return e
}

// IsKeyNotFound checks if the error indicates a key was not found.
func IsKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrKeyNotFound) || errors.Is(err, ErrNil) {
		return true
	}
	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		return redisErr.Code == CodeKeyNotFound
	}
	return false
}

// IsConnectionError checks if the error is connection-related.
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		return redisErr.Code == CodeConnectionFailed
	}
	return false
}

// IsTimeoutError checks if the error is a timeout.
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	var redisErr *RedisError
	if errors.As(err, &redisErr) {
		return redisErr.Code == CodeTimeout
	}
	return false
}

// IsRetryable checks if the error is retryable.
func IsRetryable(err error) bool {
	return IsConnectionError(err) || IsTimeoutError(err)
}
