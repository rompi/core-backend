// Package testutil provides testing utilities for the redis package.
package testutil

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrKeyNotFound is returned when a key is not found.
var ErrKeyNotFound = errors.New("redis: key not found")

// ZMember represents a sorted set member.
type ZMember struct {
	Score  float64
	Member interface{}
}

// StreamMessage represents a stream message.
type StreamMessage struct {
	Stream   string
	Messages []StreamEntry
}

// StreamEntry represents a stream entry.
type StreamEntry struct {
	ID     string
	Values map[string]interface{}
}

// PoolStats represents pool statistics.
type PoolStats struct {
	Hits       uint32
	Misses     uint32
	Timeouts   uint32
	TotalConns uint32
	IdleConns  uint32
	StaleConns uint32
}

// Message represents a pub/sub message.
type Message struct {
	Channel      string
	Pattern      string
	Payload      string
	PayloadSlice []string
}

// PubSub interface for pub/sub operations.
type PubSub interface {
	Close() error
	Channel() <-chan *Message
}

// MockClient is a mock implementation of a Redis client.
type MockClient struct {
	mu   sync.RWMutex
	data map[string]string
	ttls map[string]time.Time

	PingFunc  func(ctx context.Context) error
	GetFunc   func(ctx context.Context, key string) (string, error)
	SetFunc   func(ctx context.Context, key string, value interface{}, ttl int64) error
	SetNXFunc func(ctx context.Context, key string, value interface{}, ttl int64) (bool, error)
	DelFunc   func(ctx context.Context, keys ...string) (int64, error)
	EvalFunc  func(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
}

// NewMockClient creates a new mock client.
func NewMockClient() *MockClient {
	return &MockClient{data: make(map[string]string), ttls: make(map[string]time.Time)}
}

func (m *MockClient) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

func (m *MockClient) Close() error { return nil }

func (m *MockClient) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if exp, ok := m.ttls[key]; ok && time.Now().After(exp) {
		delete(m.data, key)
		delete(m.ttls, key)
		return "", ErrKeyNotFound
	}
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", ErrKeyNotFound
}

func (m *MockClient) Set(ctx context.Context, key string, value interface{}, ttl int64) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, ttl)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var strVal string
	switch v := value.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	}
	m.data[key] = strVal
	if ttl > 0 {
		m.ttls[key] = time.Now().Add(time.Duration(ttl) * time.Second)
	}
	return nil
}

func (m *MockClient) SetNX(ctx context.Context, key string, value interface{}, ttl int64) (bool, error) {
	if m.SetNXFunc != nil {
		return m.SetNXFunc(ctx, key, value, ttl)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[key]; ok {
		return false, nil
	}
	var strVal string
	switch v := value.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	}
	m.data[key] = strVal
	if ttl > 0 {
		m.ttls[key] = time.Now().Add(time.Duration(ttl) * time.Second)
	}
	return true, nil
}

func (m *MockClient) Del(ctx context.Context, keys ...string) (int64, error) {
	if m.DelFunc != nil {
		return m.DelFunc(ctx, keys...)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var deleted int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			delete(m.data, key)
			delete(m.ttls, key)
			deleted++
		}
	}
	return deleted, nil
}

func (m *MockClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			count++
		}
	}
	return count, nil
}

func (m *MockClient) Expire(ctx context.Context, key string, ttl int64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[key]; !ok {
		return false, nil
	}
	m.ttls[key] = time.Now().Add(time.Duration(ttl) * time.Second)
	return true, nil
}

func (m *MockClient) ExpireAt(ctx context.Context, key string, timestamp int64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[key]; !ok {
		return false, nil
	}
	m.ttls[key] = time.Unix(timestamp, 0)
	return true, nil
}

func (m *MockClient) Persist(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.ttls[key]; ok {
		delete(m.ttls, key)
		return true, nil
	}
	return false, nil
}

func (m *MockClient) TTL(ctx context.Context, key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.data[key]; !ok {
		return -2, nil
	}
	if exp, ok := m.ttls[key]; ok {
		return int64(time.Until(exp).Seconds()), nil
	}
	return -1, nil
}

func (m *MockClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var keys []string
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *MockClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	keys, err := m.Keys(ctx, match)
	return keys, 0, err
}

func (m *MockClient) Incr(ctx context.Context, key string) (int64, error)                         { return 1, nil }
func (m *MockClient) IncrBy(ctx context.Context, key string, value int64) (int64, error)          { return value, nil }
func (m *MockClient) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) { return value, nil }
func (m *MockClient) Decr(ctx context.Context, key string) (int64, error)                         { return -1, nil }
func (m *MockClient) DecrBy(ctx context.Context, key string, value int64) (int64, error)          { return -value, nil }

func (m *MockClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) { return int64(len(values) / 2), nil }
func (m *MockClient) HGet(ctx context.Context, key, field string) (string, error)                { return "", ErrKeyNotFound }
func (m *MockClient) HGetAll(ctx context.Context, key string) (map[string]string, error)         { return make(map[string]string), nil }
func (m *MockClient) HDel(ctx context.Context, key string, fields ...string) (int64, error)      { return int64(len(fields)), nil }
func (m *MockClient) HExists(ctx context.Context, key, field string) (bool, error)               { return false, nil }
func (m *MockClient) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error)  { return incr, nil }

func (m *MockClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error)  { return int64(len(values)), nil }
func (m *MockClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error)  { return int64(len(values)), nil }
func (m *MockClient) LPop(ctx context.Context, key string) (string, error)                         { return "", ErrKeyNotFound }
func (m *MockClient) RPop(ctx context.Context, key string) (string, error)                         { return "", ErrKeyNotFound }
func (m *MockClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error)  { return nil, nil }
func (m *MockClient) LLen(ctx context.Context, key string) (int64, error)                          { return 0, nil }

func (m *MockClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)  { return int64(len(members)), nil }
func (m *MockClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error)  { return int64(len(members)), nil }
func (m *MockClient) SMembers(ctx context.Context, key string) ([]string, error)                   { return nil, nil }
func (m *MockClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error)  { return false, nil }
func (m *MockClient) SCard(ctx context.Context, key string) (int64, error)                         { return 0, nil }

func (m *MockClient) ZAdd(ctx context.Context, key string, members ...ZMember) (int64, error)                        { return int64(len(members)), nil }
func (m *MockClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)                    { return 0, nil }
func (m *MockClient) ZScore(ctx context.Context, key, member string) (float64, error)                                { return 0, nil }
func (m *MockClient) ZRank(ctx context.Context, key, member string) (int64, error)                                   { return 0, nil }
func (m *MockClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)                    { return nil, nil }
func (m *MockClient) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMember, error)         { return nil, nil }
func (m *MockClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)                 { return nil, nil }
func (m *MockClient) ZCard(ctx context.Context, key string) (int64, error)                                           { return 0, nil }
func (m *MockClient) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error)                      { return 0, nil }

func (m *MockClient) Publish(ctx context.Context, channel string, message interface{}) (int64, error) { return 0, nil }
func (m *MockClient) Subscribe(ctx context.Context, channels ...string) PubSub                        { return &MockPubSub{} }

func (m *MockClient) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) { return "0-0", nil }
func (m *MockClient) XGroupCreateMkStream(ctx context.Context, stream, group, start string) error            { return nil }
func (m *MockClient) XReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]StreamMessage, error) { return nil, nil }
func (m *MockClient) XAck(ctx context.Context, stream, group string, ids ...string) (int64, error)           { return 0, nil }

func (m *MockClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	if m.EvalFunc != nil {
		return m.EvalFunc(ctx, script, keys, args...)
	}
	return int64(1), nil
}

func (m *MockClient) Info(ctx context.Context, section ...string) (string, error) { return "redis_version:7.0.0", nil }
func (m *MockClient) PoolStats() *PoolStats                                       { return &PoolStats{} }

func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]string)
	m.ttls = make(map[string]time.Time)
}

// MockPubSub is a mock PubSub implementation.
type MockPubSub struct{ messages chan *Message }

func (p *MockPubSub) Close() error { return nil }
func (p *MockPubSub) Channel() <-chan *Message {
	if p.messages == nil {
		p.messages = make(chan *Message)
	}
	return p.messages
}
