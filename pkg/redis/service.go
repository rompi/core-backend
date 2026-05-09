package redis

import (
	"context"
	"time"
)

// Service provides a high-level interface for Redis operations.
type Service interface {
	Ping(ctx context.Context) error
	Close() error
	IsHealthy(ctx context.Context) bool
	HealthCheck(ctx context.Context) *HealthStatus
	Client() Client

	// String operations
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	GetString(ctx context.Context, key string) (string, error)
	SetString(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Cache operations
	GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, loader func() (interface{}, error)) error
	GetOrSetString(ctx context.Context, key string, ttl time.Duration, loader func() (string, error)) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)

	// TTL operations
	Expire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error)
	Persist(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Counter operations
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	IncrByFloat(ctx context.Context, key string, value float64) (float64, error)
	Decr(ctx context.Context, key string) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)

	// Hash operations
	HSet(ctx context.Context, key string, values map[string]interface{}) error
	HGet(ctx context.Context, key, field string, dest interface{}) error
	HGetString(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error)

	// List operations
	LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	RPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	LPopString(ctx context.Context, key string) (string, error)
	RPopString(ctx context.Context, key string) (string, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LLen(ctx context.Context, key string) (int64, error)

	// Set operations
	SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	SRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)

	// Sorted set operations
	ZAdd(ctx context.Context, key string, members ...ZMember) (int64, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	ZScore(ctx context.Context, key, member string) (float64, error)
	ZRank(ctx context.Context, key, member string) (int64, error)
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMember, error)
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZCard(ctx context.Context, key string) (int64, error)

	// Pub/Sub
	Publish(ctx context.Context, channel string, message interface{}) (int64, error)
	Subscribe(ctx context.Context, handler MessageHandler, channels ...string) (Subscription, error)

	// Distributed locking
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (*DistributedLock, error)
	AcquireLockWithRetry(ctx context.Context, key string, ttl time.Duration, retryCount int, retryDelay time.Duration) (*DistributedLock, error)
	ReleaseLock(ctx context.Context, lock *DistributedLock) error
	ExtendLock(ctx context.Context, lock *DistributedLock, ttl time.Duration) error

	// Rate limiting
	RateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error)

	// Stream operations
	StreamAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error)
	StreamCreateGroup(ctx context.Context, stream, group, start string) error
	StreamReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]StreamMessage, error)
	StreamAck(ctx context.Context, stream, group string, ids ...string) (int64, error)

	// Statistics
	PoolStats() *PoolStats
	Stats() *ServiceStats
}

// MessageHandler handles pub/sub messages.
type MessageHandler func(msg *Message)

// Subscription represents an active subscription.
type Subscription interface {
	Close() error
	Channel() <-chan *Message
}

// Client is the interface for the underlying Redis client.
type Client interface {
	Ping(ctx context.Context) error
	Close() error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl int64) error
	SetNX(ctx context.Context, key string, value interface{}, ttl int64) (bool, error)
	Del(ctx context.Context, keys ...string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, ttl int64) (bool, error)
	ExpireAt(ctx context.Context, key string, timestamp int64) (bool, error)
	Persist(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (int64, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error)
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	IncrByFloat(ctx context.Context, key string, value float64) (float64, error)
	Decr(ctx context.Context, key string) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)
	HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error)
	LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	RPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LLen(ctx context.Context, key string) (int64, error)
	SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	SRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)
	ZAdd(ctx context.Context, key string, members ...ZMember) (int64, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	ZScore(ctx context.Context, key, member string) (float64, error)
	ZRank(ctx context.Context, key, member string) (int64, error)
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMember, error)
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZCard(ctx context.Context, key string) (int64, error)
	ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error)
	Publish(ctx context.Context, channel string, message interface{}) (int64, error)
	Subscribe(ctx context.Context, channels ...string) PubSub
	XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error)
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) error
	XReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]StreamMessage, error)
	XAck(ctx context.Context, stream, group string, ids ...string) (int64, error)
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	Info(ctx context.Context, section ...string) (string, error)
	PoolStats() *PoolStats
}

// PubSub represents a pub/sub connection.
type PubSub interface {
	Close() error
	Channel() <-chan *Message
}
