package redis

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type service struct {
	client     Client
	config     *Config
	serializer Serializer
	mu         sync.RWMutex
	operations int64
	cacheHits  int64
	cacheMisses int64
	errors     int64
	totalLatency int64
}

// NewService creates a new Redis service.
func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	client, err := NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	serializer, err := NewSerializer(cfg.Serializer)
	if err != nil {
		return nil, fmt.Errorf("failed to create serializer: %w", err)
	}
	return &service{client: client, config: cfg, serializer: serializer}, nil
}

// NewServiceWithClient creates a service with a custom client.
func NewServiceWithClient(client Client, cfg *Config) (Service, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}
	serializer, err := NewSerializer(cfg.Serializer)
	if err != nil {
		return nil, fmt.Errorf("failed to create serializer: %w", err)
	}
	return &service{client: client, config: cfg, serializer: serializer}, nil
}

func (s *service) Client() Client { return s.client }
func (s *service) Ping(ctx context.Context) error {
	start := time.Now()
	err := s.client.Ping(ctx)
	s.recordOperation(time.Since(start), err)
	return err
}
func (s *service) Close() error { return s.client.Close() }
func (s *service) IsHealthy(ctx context.Context) bool { return s.client.Ping(ctx) == nil }
func (s *service) HealthCheck(ctx context.Context) *HealthStatus {
	status := &HealthStatus{CheckedAt: time.Now()}
	start := time.Now()
	err := s.client.Ping(ctx)
	status.Latency = time.Since(start)
	if err != nil {
		status.Healthy = false
		status.Error = err.Error()
	} else {
		status.Healthy = true
		status.PoolStats = s.client.PoolStats()
	}
	return status
}

func (s *service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	data, err := s.serializer.Serialize(value)
	if err != nil {
		s.recordOperation(time.Since(start), err)
		return NewErrorWithCause(CodeSerializationFailed, "failed to serialize", err).WithKey(key)
	}
	err = s.client.Set(ctx, key, data, int64(ttl.Seconds()))
	s.recordOperation(time.Since(start), err)
	return err
}

func (s *service) Get(ctx context.Context, key string, dest interface{}) error {
	start := time.Now()
	data, err := s.client.Get(ctx, key)
	if err != nil {
		if IsKeyNotFound(err) {
			atomic.AddInt64(&s.cacheMisses, 1)
		}
		s.recordOperation(time.Since(start), err)
		return err
	}
	atomic.AddInt64(&s.cacheHits, 1)
	err = s.serializer.Deserialize([]byte(data), dest)
	s.recordOperation(time.Since(start), err)
	return err
}

func (s *service) GetString(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result, err := s.client.Get(ctx, key)
	if IsKeyNotFound(err) {
		atomic.AddInt64(&s.cacheMisses, 1)
	} else if err == nil {
		atomic.AddInt64(&s.cacheHits, 1)
	}
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	start := time.Now()
	err := s.client.Set(ctx, key, value, int64(ttl.Seconds()))
	s.recordOperation(time.Since(start), err)
	return err
}

func (s *service) Delete(ctx context.Context, keys ...string) (int64, error) {
	start := time.Now()
	result, err := s.client.Del(ctx, keys...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) Exists(ctx context.Context, keys ...string) (int64, error) {
	start := time.Now()
	result, err := s.client.Exists(ctx, keys...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, loader func() (interface{}, error)) error {
	start := time.Now()
	data, err := s.client.Get(ctx, key)
	if err == nil {
		atomic.AddInt64(&s.cacheHits, 1)
		return s.serializer.Deserialize([]byte(data), dest)
	}
	if !IsKeyNotFound(err) {
		s.recordOperation(time.Since(start), err)
		return err
	}
	atomic.AddInt64(&s.cacheMisses, 1)
	value, err := loader()
	if err != nil {
		s.recordOperation(time.Since(start), err)
		return err
	}
	serialized, err := s.serializer.Serialize(value)
	if err != nil {
		s.recordOperation(time.Since(start), err)
		return err
	}
	_ = s.client.Set(ctx, key, serialized, int64(ttl.Seconds()))
	err = s.serializer.Deserialize(serialized, dest)
	s.recordOperation(time.Since(start), err)
	return err
}

func (s *service) GetOrSetString(ctx context.Context, key string, ttl time.Duration, loader func() (string, error)) (string, error) {
	start := time.Now()
	data, err := s.client.Get(ctx, key)
	if err == nil {
		atomic.AddInt64(&s.cacheHits, 1)
		s.recordOperation(time.Since(start), nil)
		return data, nil
	}
	if !IsKeyNotFound(err) {
		s.recordOperation(time.Since(start), err)
		return "", err
	}
	atomic.AddInt64(&s.cacheMisses, 1)
	value, err := loader()
	if err != nil {
		s.recordOperation(time.Since(start), err)
		return "", err
	}
	_ = s.client.Set(ctx, key, value, int64(ttl.Seconds()))
	s.recordOperation(time.Since(start), nil)
	return value, nil
}

func (s *service) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	start := time.Now()
	data, err := s.serializer.Serialize(value)
	if err != nil {
		s.recordOperation(time.Since(start), err)
		return false, err
	}
	result, err := s.client.SetNX(ctx, key, data, int64(ttl.Seconds()))
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	start := time.Now()
	result, err := s.client.Expire(ctx, key, int64(ttl.Seconds()))
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error) {
	start := time.Now()
	result, err := s.client.ExpireAt(ctx, key, tm.Unix())
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) Persist(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	result, err := s.client.Persist(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) TTL(ctx context.Context, key string) (time.Duration, error) {
	start := time.Now()
	result, err := s.client.TTL(ctx, key)
	s.recordOperation(time.Since(start), err)
	return time.Duration(result) * time.Second, err
}

func (s *service) Incr(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	result, err := s.client.Incr(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	start := time.Now()
	result, err := s.client.IncrBy(ctx, key, value)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) {
	start := time.Now()
	result, err := s.client.IncrByFloat(ctx, key, value)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) Decr(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	result, err := s.client.Decr(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	start := time.Now()
	result, err := s.client.DecrBy(ctx, key, value)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	start := time.Now()
	serialized := make([]interface{}, 0, len(values)*2)
	for field, value := range values {
		data, err := s.serializer.Serialize(value)
		if err != nil {
			s.recordOperation(time.Since(start), err)
			return err
		}
		serialized = append(serialized, field, string(data))
	}
	_, err := s.client.HSet(ctx, key, serialized...)
	s.recordOperation(time.Since(start), err)
	return err
}

func (s *service) HGet(ctx context.Context, key, field string, dest interface{}) error {
	start := time.Now()
	data, err := s.client.HGet(ctx, key, field)
	if err != nil {
		s.recordOperation(time.Since(start), err)
		return err
	}
	err = s.serializer.Deserialize([]byte(data), dest)
	s.recordOperation(time.Since(start), err)
	return err
}

func (s *service) HGetString(ctx context.Context, key, field string) (string, error) {
	start := time.Now()
	result, err := s.client.HGet(ctx, key, field)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	start := time.Now()
	result, err := s.client.HGetAll(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	start := time.Now()
	result, err := s.client.HDel(ctx, key, fields...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) HExists(ctx context.Context, key, field string) (bool, error) {
	start := time.Now()
	result, err := s.client.HExists(ctx, key, field)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	start := time.Now()
	result, err := s.client.HIncrBy(ctx, key, field, incr)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	start := time.Now()
	serialized := make([]interface{}, len(values))
	for i, v := range values {
		data, err := s.serializer.Serialize(v)
		if err != nil {
			s.recordOperation(time.Since(start), err)
			return 0, err
		}
		serialized[i] = string(data)
	}
	result, err := s.client.LPush(ctx, key, serialized...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	start := time.Now()
	serialized := make([]interface{}, len(values))
	for i, v := range values {
		data, err := s.serializer.Serialize(v)
		if err != nil {
			s.recordOperation(time.Since(start), err)
			return 0, err
		}
		serialized[i] = string(data)
	}
	result, err := s.client.RPush(ctx, key, serialized...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) LPopString(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result, err := s.client.LPop(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) RPopString(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result, err := s.client.RPop(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	startTime := time.Now()
	result, err := s.client.LRange(ctx, key, start, stop)
	s.recordOperation(time.Since(startTime), err)
	return result, err
}

func (s *service) LLen(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	result, err := s.client.LLen(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	start := time.Now()
	result, err := s.client.SAdd(ctx, key, members...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	start := time.Now()
	result, err := s.client.SRem(ctx, key, members...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) SMembers(ctx context.Context, key string) ([]string, error) {
	start := time.Now()
	result, err := s.client.SMembers(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	start := time.Now()
	result, err := s.client.SIsMember(ctx, key, member)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) SCard(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	result, err := s.client.SCard(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) ZAdd(ctx context.Context, key string, members ...ZMember) (int64, error) {
	start := time.Now()
	result, err := s.client.ZAdd(ctx, key, members...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	start := time.Now()
	result, err := s.client.ZRem(ctx, key, members...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) ZScore(ctx context.Context, key, member string) (float64, error) {
	start := time.Now()
	result, err := s.client.ZScore(ctx, key, member)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) ZRank(ctx context.Context, key, member string) (int64, error) {
	start := time.Now()
	result, err := s.client.ZRank(ctx, key, member)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	startTime := time.Now()
	result, err := s.client.ZRange(ctx, key, start, stop)
	s.recordOperation(time.Since(startTime), err)
	return result, err
}

func (s *service) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMember, error) {
	startTime := time.Now()
	result, err := s.client.ZRangeWithScores(ctx, key, start, stop)
	s.recordOperation(time.Since(startTime), err)
	return result, err
}

func (s *service) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	startTime := time.Now()
	result, err := s.client.ZRevRange(ctx, key, start, stop)
	s.recordOperation(time.Since(startTime), err)
	return result, err
}

func (s *service) ZCard(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	result, err := s.client.ZCard(ctx, key)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	start := time.Now()
	data, err := s.serializer.Serialize(message)
	if err != nil {
		s.recordOperation(time.Since(start), err)
		return 0, err
	}
	result, err := s.client.Publish(ctx, channel, string(data))
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) Subscribe(ctx context.Context, handler MessageHandler, channels ...string) (Subscription, error) {
	ps := s.client.Subscribe(ctx, channels...)
	sub := &subscription{ps: ps, handler: handler, done: make(chan struct{})}
	go sub.listen()
	return sub, nil
}

func (s *service) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*DistributedLock, error) {
	return s.AcquireLockWithRetry(ctx, key, ttl, 0, 0)
}

func (s *service) AcquireLockWithRetry(ctx context.Context, key string, ttl time.Duration, retryCount int, retryDelay time.Duration) (*DistributedLock, error) {
	start := time.Now()
	lockKey := fmt.Sprintf("lock:%s", key)
	lockValue := uuid.New().String()
	for i := 0; i <= retryCount; i++ {
		acquired, err := s.client.SetNX(ctx, lockKey, lockValue, int64(ttl.Seconds()))
		if err != nil {
			s.recordOperation(time.Since(start), err)
			return nil, err
		}
		if acquired {
			s.recordOperation(time.Since(start), nil)
			return &DistributedLock{Key: lockKey, Value: lockValue, TTL: ttl, AcquiredAt: time.Now()}, nil
		}
		if i < retryCount {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
			}
		}
	}
	s.recordOperation(time.Since(start), ErrLockFailed)
	return nil, NewErrorWithKey(CodeLockFailed, "failed to acquire lock", key)
}

func (s *service) ReleaseLock(ctx context.Context, lock *DistributedLock) error {
	if lock == nil {
		return ErrLockNotHeld
	}
	start := time.Now()
	script := `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
	result, err := s.client.Eval(ctx, script, []string{lock.Key}, lock.Value)
	s.recordOperation(time.Since(start), err)
	if err != nil {
		return err
	}
	if result.(int64) == 0 {
		return ErrLockNotHeld
	}
	return nil
}

func (s *service) ExtendLock(ctx context.Context, lock *DistributedLock, ttl time.Duration) error {
	if lock == nil {
		return ErrLockNotHeld
	}
	start := time.Now()
	script := `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`
	result, err := s.client.Eval(ctx, script, []string{lock.Key}, lock.Value, ttl.Milliseconds())
	s.recordOperation(time.Since(start), err)
	if err != nil {
		return err
	}
	if result.(int64) == 0 {
		return ErrLockNotHeld
	}
	lock.TTL = ttl
	return nil
}

func (s *service) RateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error) {
	start := time.Now()
	rateLimitKey := fmt.Sprintf("ratelimit:%s", key)
	count, err := s.client.Incr(ctx, rateLimitKey)
	if err != nil {
		s.recordOperation(time.Since(start), err)
		return false, 0, err
	}
	if count == 1 {
		_, _ = s.client.Expire(ctx, rateLimitKey, int64(window.Seconds()))
	}
	s.recordOperation(time.Since(start), nil)
	return count <= limit, count, nil
}

func (s *service) StreamAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	start := time.Now()
	result, err := s.client.XAdd(ctx, stream, values)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) StreamCreateGroup(ctx context.Context, stream, group, startID string) error {
	start := time.Now()
	err := s.client.XGroupCreateMkStream(ctx, stream, group, startID)
	s.recordOperation(time.Since(start), err)
	return err
}

func (s *service) StreamReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]StreamMessage, error) {
	start := time.Now()
	result, err := s.client.XReadGroup(ctx, group, consumer, streams, count, block)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) StreamAck(ctx context.Context, stream, group string, ids ...string) (int64, error) {
	start := time.Now()
	result, err := s.client.XAck(ctx, stream, group, ids...)
	s.recordOperation(time.Since(start), err)
	return result, err
}

func (s *service) PoolStats() *PoolStats { return s.client.PoolStats() }

func (s *service) Stats() *ServiceStats {
	ops := atomic.LoadInt64(&s.operations)
	var avgLatency time.Duration
	if ops > 0 {
		avgLatency = time.Duration(atomic.LoadInt64(&s.totalLatency) / ops)
	}
	return &ServiceStats{
		Operations:     ops,
		CacheHits:      atomic.LoadInt64(&s.cacheHits),
		CacheMisses:    atomic.LoadInt64(&s.cacheMisses),
		Errors:         atomic.LoadInt64(&s.errors),
		AverageLatency: avgLatency,
		PoolStats:      s.client.PoolStats(),
	}
}

func (s *service) recordOperation(latency time.Duration, err error) {
	atomic.AddInt64(&s.operations, 1)
	atomic.AddInt64(&s.totalLatency, int64(latency))
	if err != nil {
		atomic.AddInt64(&s.errors, 1)
	}
}

type subscription struct {
	ps      PubSub
	handler MessageHandler
	done    chan struct{}
	once    sync.Once
}

func (s *subscription) listen() {
	ch := s.ps.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if s.handler != nil {
				s.handler(msg)
			}
		case <-s.done:
			return
		}
	}
}

func (s *subscription) Close() error {
	s.once.Do(func() { close(s.done) })
	return s.ps.Close()
}

func (s *subscription) Channel() <-chan *Message { return s.ps.Channel() }
