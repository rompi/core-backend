package redis

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockClient is an internal mock for testing.
type mockClient struct {
	mu   sync.RWMutex
	data map[string]string
	ttls map[string]time.Time
}

func newMockClient() *mockClient {
	return &mockClient{data: make(map[string]string), ttls: make(map[string]time.Time)}
}

func (m *mockClient) Ping(ctx context.Context) error { return nil }
func (m *mockClient) Close() error                   { return nil }

func (m *mockClient) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", ErrKeyNotFound
}

func (m *mockClient) Set(ctx context.Context, key string, value interface{}, ttl int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch v := value.(type) {
	case string:
		m.data[key] = v
	case []byte:
		m.data[key] = string(v)
	}
	if ttl > 0 {
		m.ttls[key] = time.Now().Add(time.Duration(ttl) * time.Second)
	}
	return nil
}

func (m *mockClient) SetNX(ctx context.Context, key string, value interface{}, ttl int64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[key]; ok {
		return false, nil
	}
	switch v := value.(type) {
	case string:
		m.data[key] = v
	case []byte:
		m.data[key] = string(v)
	}
	return true, nil
}

func (m *mockClient) Del(ctx context.Context, keys ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var deleted int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			delete(m.data, key)
			deleted++
		}
	}
	return deleted, nil
}

func (m *mockClient) Exists(ctx context.Context, keys ...string) (int64, error) {
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

func (m *mockClient) Expire(ctx context.Context, key string, ttl int64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[key]; !ok {
		return false, nil
	}
	m.ttls[key] = time.Now().Add(time.Duration(ttl) * time.Second)
	return true, nil
}

func (m *mockClient) ExpireAt(ctx context.Context, key string, timestamp int64) (bool, error) { return true, nil }
func (m *mockClient) Persist(ctx context.Context, key string) (bool, error)                   { return true, nil }
func (m *mockClient) TTL(ctx context.Context, key string) (int64, error)                      { return 60, nil }
func (m *mockClient) Keys(ctx context.Context, pattern string) ([]string, error)              { return nil, nil }
func (m *mockClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) { return nil, 0, nil }
func (m *mockClient) Incr(ctx context.Context, key string) (int64, error)                     { return 1, nil }
func (m *mockClient) IncrBy(ctx context.Context, key string, value int64) (int64, error)      { return value, nil }
func (m *mockClient) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) { return value, nil }
func (m *mockClient) Decr(ctx context.Context, key string) (int64, error)                     { return -1, nil }
func (m *mockClient) DecrBy(ctx context.Context, key string, value int64) (int64, error)      { return -value, nil }
func (m *mockClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) { return int64(len(values) / 2), nil }
func (m *mockClient) HGet(ctx context.Context, key, field string) (string, error)             { return "", ErrKeyNotFound }
func (m *mockClient) HGetAll(ctx context.Context, key string) (map[string]string, error)      { return make(map[string]string), nil }
func (m *mockClient) HDel(ctx context.Context, key string, fields ...string) (int64, error)   { return int64(len(fields)), nil }
func (m *mockClient) HExists(ctx context.Context, key, field string) (bool, error)            { return false, nil }
func (m *mockClient) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) { return incr, nil }
func (m *mockClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) { return int64(len(values)), nil }
func (m *mockClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) { return int64(len(values)), nil }
func (m *mockClient) LPop(ctx context.Context, key string) (string, error)                    { return "", ErrKeyNotFound }
func (m *mockClient) RPop(ctx context.Context, key string) (string, error)                    { return "", ErrKeyNotFound }
func (m *mockClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) { return nil, nil }
func (m *mockClient) LLen(ctx context.Context, key string) (int64, error)                     { return 0, nil }
func (m *mockClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) { return int64(len(members)), nil }
func (m *mockClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) { return int64(len(members)), nil }
func (m *mockClient) SMembers(ctx context.Context, key string) ([]string, error)              { return nil, nil }
func (m *mockClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) { return false, nil }
func (m *mockClient) SCard(ctx context.Context, key string) (int64, error)                    { return 0, nil }
func (m *mockClient) ZAdd(ctx context.Context, key string, members ...ZMember) (int64, error) { return int64(len(members)), nil }
func (m *mockClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) { return 0, nil }
func (m *mockClient) ZScore(ctx context.Context, key, member string) (float64, error)         { return 0, nil }
func (m *mockClient) ZRank(ctx context.Context, key, member string) (int64, error)            { return 0, nil }
func (m *mockClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) { return nil, nil }
func (m *mockClient) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMember, error) { return nil, nil }
func (m *mockClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) { return nil, nil }
func (m *mockClient) ZCard(ctx context.Context, key string) (int64, error)                    { return 0, nil }
func (m *mockClient) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) { return 0, nil }
func (m *mockClient) Publish(ctx context.Context, channel string, message interface{}) (int64, error) { return 0, nil }
func (m *mockClient) Subscribe(ctx context.Context, channels ...string) PubSub               { return &mockPubSub{} }
func (m *mockClient) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) { return "0-0", nil }
func (m *mockClient) XGroupCreateMkStream(ctx context.Context, stream, group, start string) error { return nil }
func (m *mockClient) XReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]StreamMessage, error) { return nil, nil }
func (m *mockClient) XAck(ctx context.Context, stream, group string, ids ...string) (int64, error) { return 0, nil }
func (m *mockClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) { return int64(1), nil }
func (m *mockClient) Info(ctx context.Context, section ...string) (string, error)             { return "", nil }
func (m *mockClient) PoolStats() *PoolStats                                                   { return &PoolStats{} }

type mockPubSub struct{ ch chan *Message }

func (p *mockPubSub) Close() error              { return nil }
func (p *mockPubSub) Channel() <-chan *Message { return make(chan *Message) }

func TestNewServiceWithClient(t *testing.T) {
	mock := newMockClient()
	cfg := DefaultConfig()
	svc, err := NewServiceWithClient(mock, cfg)
	if err != nil {
		t.Fatalf("NewServiceWithClient() error = %v", err)
	}
	ctx := context.Background()
	if err := svc.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
	if !svc.IsHealthy(ctx) {
		t.Error("expected IsHealthy() = true")
	}
}

func TestServiceSetGet(t *testing.T) {
	mock := newMockClient()
	cfg := DefaultConfig()
	svc, _ := NewServiceWithClient(mock, cfg)
	ctx := context.Background()

	_ = svc.SetString(ctx, "key", "value", time.Minute)
	val, err := svc.GetString(ctx, "key")
	if err != nil {
		t.Errorf("GetString() error = %v", err)
	}
	if val != "value" {
		t.Errorf("expected value, got %s", val)
	}
}

func TestServiceDelete(t *testing.T) {
	mock := newMockClient()
	cfg := DefaultConfig()
	svc, _ := NewServiceWithClient(mock, cfg)
	ctx := context.Background()

	_ = svc.SetString(ctx, "key", "value", time.Minute)
	deleted, err := svc.Delete(ctx, "key")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}

func TestServiceSetNX(t *testing.T) {
	mock := newMockClient()
	cfg := DefaultConfig()
	svc, _ := NewServiceWithClient(mock, cfg)
	ctx := context.Background()

	ok, _ := svc.SetNX(ctx, "key", "val1", time.Minute)
	if !ok {
		t.Error("expected first SetNX to succeed")
	}
	ok, _ = svc.SetNX(ctx, "key", "val2", time.Minute)
	if ok {
		t.Error("expected second SetNX to fail")
	}
}

func TestServiceGetOrSetString(t *testing.T) {
	mock := newMockClient()
	cfg := DefaultConfig()
	svc, _ := NewServiceWithClient(mock, cfg)
	ctx := context.Background()
	loaderCalled := 0

	loader := func() (string, error) {
		loaderCalled++
		return "loaded", nil
	}

	val, _ := svc.GetOrSetString(ctx, "key", time.Minute, loader)
	if val != "loaded" || loaderCalled != 1 {
		t.Errorf("expected loaded and loader called once")
	}

	val, _ = svc.GetOrSetString(ctx, "key", time.Minute, loader)
	if loaderCalled != 1 {
		t.Errorf("expected loader not called again")
	}
}

func TestDistributedLock(t *testing.T) {
	mock := newMockClient()
	cfg := DefaultConfig()
	svc, _ := NewServiceWithClient(mock, cfg)
	ctx := context.Background()

	lock, err := svc.AcquireLock(ctx, "resource", 30*time.Second)
	if err != nil {
		t.Errorf("AcquireLock() error = %v", err)
	}
	if lock == nil || !lock.IsValid() {
		t.Fatal("expected valid lock")
	}

	if err := svc.ReleaseLock(ctx, lock); err != nil {
		t.Errorf("ReleaseLock() error = %v", err)
	}
}

func TestDistributedLockIsValid(t *testing.T) {
	var lock *DistributedLock
	if lock.IsValid() {
		t.Error("nil lock should be invalid")
	}

	lock = &DistributedLock{TTL: time.Millisecond, AcquiredAt: time.Now().Add(-time.Second)}
	if lock.IsValid() {
		t.Error("expired lock should be invalid")
	}

	lock = &DistributedLock{TTL: 30 * time.Second, AcquiredAt: time.Now()}
	if !lock.IsValid() {
		t.Error("fresh lock should be valid")
	}
}

func TestRateLimit(t *testing.T) {
	mock := newMockClient()
	cfg := DefaultConfig()
	svc, _ := NewServiceWithClient(mock, cfg)
	ctx := context.Background()

	allowed, count, err := svc.RateLimit(ctx, "endpoint", 10, time.Minute)
	if err != nil {
		t.Errorf("RateLimit() error = %v", err)
	}
	if !allowed || count != 1 {
		t.Errorf("expected allowed=true, count=1")
	}
}

func TestServiceStats(t *testing.T) {
	mock := newMockClient()
	cfg := DefaultConfig()
	svc, _ := NewServiceWithClient(mock, cfg)
	ctx := context.Background()

	_ = svc.SetString(ctx, "k", "v", time.Minute)
	_, _ = svc.GetString(ctx, "k")
	_, _ = svc.GetString(ctx, "missing")

	stats := svc.Stats()
	if stats.Operations < 3 {
		t.Errorf("expected at least 3 operations")
	}
}

func TestServiceHealthCheck(t *testing.T) {
	mock := newMockClient()
	cfg := DefaultConfig()
	svc, _ := NewServiceWithClient(mock, cfg)
	ctx := context.Background()

	health := svc.HealthCheck(ctx)
	if !health.Healthy {
		t.Error("expected healthy = true")
	}
}

func TestNewServiceWithNilClient(t *testing.T) {
	cfg := DefaultConfig()
	_, err := NewServiceWithClient(nil, cfg)
	if err == nil {
		t.Error("expected error for nil client")
	}
}
