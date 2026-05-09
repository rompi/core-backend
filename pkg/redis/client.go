package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type clientWrapper struct {
	rdb    redis.UniversalClient
	config *Config
}

// NewClient creates a new Redis client.
func NewClient(cfg *Config) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	tlsConfig, err := cfg.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}
	var rdb redis.UniversalClient
	if cfg.ClusterEnabled {
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: cfg.ClusterNodes, Password: cfg.Password, Username: cfg.Username,
			PoolSize: cfg.PoolSize, MinIdleConns: cfg.MinIdleConns,
			DialTimeout: cfg.ConnectTimeout, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout,
			MaxRetries: cfg.MaxRetries, TLSConfig: tlsConfig,
		})
	} else if cfg.SentinelEnabled {
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName: cfg.SentinelMasterName, SentinelAddrs: cfg.SentinelAddresses,
			SentinelPassword: cfg.SentinelPassword, Password: cfg.Password, Username: cfg.Username, DB: cfg.DB,
			PoolSize: cfg.PoolSize, MinIdleConns: cfg.MinIdleConns,
			DialTimeout: cfg.ConnectTimeout, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout,
			MaxRetries: cfg.MaxRetries, TLSConfig: tlsConfig,
		})
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr: cfg.Address(), Password: cfg.Password, Username: cfg.Username, DB: cfg.DB,
			PoolSize: cfg.PoolSize, MinIdleConns: cfg.MinIdleConns,
			DialTimeout: cfg.ConnectTimeout, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout,
			MaxRetries: cfg.MaxRetries, TLSConfig: tlsConfig,
		})
	}
	return &clientWrapper{rdb: rdb, config: cfg}, nil
}

func (c *clientWrapper) Ping(ctx context.Context) error { return c.rdb.Ping(ctx).Err() }
func (c *clientWrapper) Close() error                   { return c.rdb.Close() }

func (c *clientWrapper) Get(ctx context.Context, key string) (string, error) {
	result, err := c.rdb.Get(ctx, c.config.WithKeyPrefix(key)).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return result, err
}

func (c *clientWrapper) Set(ctx context.Context, key string, value interface{}, ttl int64) error {
	exp := time.Duration(ttl) * time.Second
	if ttl <= 0 && c.config.DefaultTTL > 0 {
		exp = c.config.DefaultTTL
	}
	return c.rdb.Set(ctx, c.config.WithKeyPrefix(key), value, exp).Err()
}

func (c *clientWrapper) SetNX(ctx context.Context, key string, value interface{}, ttl int64) (bool, error) {
	return c.rdb.SetNX(ctx, c.config.WithKeyPrefix(key), value, time.Duration(ttl)*time.Second).Result()
}

func (c *clientWrapper) Del(ctx context.Context, keys ...string) (int64, error) {
	prefixed := make([]string, len(keys))
	for i, k := range keys {
		prefixed[i] = c.config.WithKeyPrefix(k)
	}
	return c.rdb.Del(ctx, prefixed...).Result()
}

func (c *clientWrapper) Exists(ctx context.Context, keys ...string) (int64, error) {
	prefixed := make([]string, len(keys))
	for i, k := range keys {
		prefixed[i] = c.config.WithKeyPrefix(k)
	}
	return c.rdb.Exists(ctx, prefixed...).Result()
}

func (c *clientWrapper) Expire(ctx context.Context, key string, ttl int64) (bool, error) {
	return c.rdb.Expire(ctx, c.config.WithKeyPrefix(key), time.Duration(ttl)*time.Second).Result()
}

func (c *clientWrapper) ExpireAt(ctx context.Context, key string, timestamp int64) (bool, error) {
	return c.rdb.ExpireAt(ctx, c.config.WithKeyPrefix(key), time.Unix(timestamp, 0)).Result()
}

func (c *clientWrapper) Persist(ctx context.Context, key string) (bool, error) {
	return c.rdb.Persist(ctx, c.config.WithKeyPrefix(key)).Result()
}

func (c *clientWrapper) TTL(ctx context.Context, key string) (int64, error) {
	result, err := c.rdb.TTL(ctx, c.config.WithKeyPrefix(key)).Result()
	return int64(result.Seconds()), err
}

func (c *clientWrapper) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.rdb.Keys(ctx, c.config.WithKeyPrefix(pattern)).Result()
}

func (c *clientWrapper) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return c.rdb.Scan(ctx, cursor, c.config.WithKeyPrefix(match), count).Result()
}

func (c *clientWrapper) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, c.config.WithKeyPrefix(key)).Result()
}

func (c *clientWrapper) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.IncrBy(ctx, c.config.WithKeyPrefix(key), value).Result()
}

func (c *clientWrapper) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) {
	return c.rdb.IncrByFloat(ctx, c.config.WithKeyPrefix(key), value).Result()
}

func (c *clientWrapper) Decr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Decr(ctx, c.config.WithKeyPrefix(key)).Result()
}

func (c *clientWrapper) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.DecrBy(ctx, c.config.WithKeyPrefix(key), value).Result()
}

func (c *clientWrapper) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.rdb.HSet(ctx, c.config.WithKeyPrefix(key), values...).Result()
}

func (c *clientWrapper) HGet(ctx context.Context, key, field string) (string, error) {
	result, err := c.rdb.HGet(ctx, c.config.WithKeyPrefix(key), field).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return result, err
}

func (c *clientWrapper) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, c.config.WithKeyPrefix(key)).Result()
}

func (c *clientWrapper) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return c.rdb.HDel(ctx, c.config.WithKeyPrefix(key), fields...).Result()
}

func (c *clientWrapper) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.rdb.HExists(ctx, c.config.WithKeyPrefix(key), field).Result()
}

func (c *clientWrapper) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return c.rdb.HIncrBy(ctx, c.config.WithKeyPrefix(key), field, incr).Result()
}

func (c *clientWrapper) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.rdb.LPush(ctx, c.config.WithKeyPrefix(key), values...).Result()
}

func (c *clientWrapper) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.rdb.RPush(ctx, c.config.WithKeyPrefix(key), values...).Result()
}

func (c *clientWrapper) LPop(ctx context.Context, key string) (string, error) {
	result, err := c.rdb.LPop(ctx, c.config.WithKeyPrefix(key)).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return result, err
}

func (c *clientWrapper) RPop(ctx context.Context, key string) (string, error) {
	result, err := c.rdb.RPop(ctx, c.config.WithKeyPrefix(key)).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return result, err
}

func (c *clientWrapper) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.LRange(ctx, c.config.WithKeyPrefix(key), start, stop).Result()
}

func (c *clientWrapper) LLen(ctx context.Context, key string) (int64, error) {
	return c.rdb.LLen(ctx, c.config.WithKeyPrefix(key)).Result()
}

func (c *clientWrapper) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.rdb.SAdd(ctx, c.config.WithKeyPrefix(key), members...).Result()
}

func (c *clientWrapper) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.rdb.SRem(ctx, c.config.WithKeyPrefix(key), members...).Result()
}

func (c *clientWrapper) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, c.config.WithKeyPrefix(key)).Result()
}

func (c *clientWrapper) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.rdb.SIsMember(ctx, c.config.WithKeyPrefix(key), member).Result()
}

func (c *clientWrapper) SCard(ctx context.Context, key string) (int64, error) {
	return c.rdb.SCard(ctx, c.config.WithKeyPrefix(key)).Result()
}

func (c *clientWrapper) ZAdd(ctx context.Context, key string, members ...ZMember) (int64, error) {
	redisMembers := make([]redis.Z, len(members))
	for i, m := range members {
		redisMembers[i] = redis.Z{Score: m.Score, Member: m.Member}
	}
	return c.rdb.ZAdd(ctx, c.config.WithKeyPrefix(key), redisMembers...).Result()
}

func (c *clientWrapper) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.rdb.ZRem(ctx, c.config.WithKeyPrefix(key), members...).Result()
}

func (c *clientWrapper) ZScore(ctx context.Context, key, member string) (float64, error) {
	result, err := c.rdb.ZScore(ctx, c.config.WithKeyPrefix(key), member).Result()
	if err == redis.Nil {
		return 0, ErrKeyNotFound
	}
	return result, err
}

func (c *clientWrapper) ZRank(ctx context.Context, key, member string) (int64, error) {
	result, err := c.rdb.ZRank(ctx, c.config.WithKeyPrefix(key), member).Result()
	if err == redis.Nil {
		return 0, ErrKeyNotFound
	}
	return result, err
}

func (c *clientWrapper) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRange(ctx, c.config.WithKeyPrefix(key), start, stop).Result()
}

func (c *clientWrapper) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMember, error) {
	result, err := c.rdb.ZRangeWithScores(ctx, c.config.WithKeyPrefix(key), start, stop).Result()
	if err != nil {
		return nil, err
	}
	members := make([]ZMember, len(result))
	for i, r := range result {
		members[i] = ZMember{Score: r.Score, Member: r.Member}
	}
	return members, nil
}

func (c *clientWrapper) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRevRange(ctx, c.config.WithKeyPrefix(key), start, stop).Result()
}

func (c *clientWrapper) ZCard(ctx context.Context, key string) (int64, error) {
	return c.rdb.ZCard(ctx, c.config.WithKeyPrefix(key)).Result()
}

func (c *clientWrapper) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) {
	return c.rdb.ZRemRangeByScore(ctx, c.config.WithKeyPrefix(key), min, max).Result()
}

func (c *clientWrapper) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return c.rdb.Publish(ctx, c.config.WithKeyPrefix(channel), message).Result()
}

func (c *clientWrapper) Subscribe(ctx context.Context, channels ...string) PubSub {
	prefixed := make([]string, len(channels))
	for i, ch := range channels {
		prefixed[i] = c.config.WithKeyPrefix(ch)
	}
	return &pubSubWrapper{ps: c.rdb.Subscribe(ctx, prefixed...), config: c.config}
}

func (c *clientWrapper) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	return c.rdb.XAdd(ctx, &redis.XAddArgs{Stream: c.config.WithKeyPrefix(stream), Values: values}).Result()
}

func (c *clientWrapper) XGroupCreateMkStream(ctx context.Context, stream, group, start string) error {
	return c.rdb.XGroupCreateMkStream(ctx, c.config.WithKeyPrefix(stream), group, start).Err()
}

func (c *clientWrapper) XReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]StreamMessage, error) {
	args := make([]string, len(streams)*2)
	for i, s := range streams {
		args[i] = c.config.WithKeyPrefix(s)
		args[i+len(streams)] = ">"
	}
	result, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: group, Consumer: consumer, Streams: args, Count: count, Block: block,
	}).Result()
	if err != nil {
		return nil, err
	}
	messages := make([]StreamMessage, len(result))
	for i, r := range result {
		entries := make([]StreamEntry, len(r.Messages))
		for j, m := range r.Messages {
			entries[j] = StreamEntry{ID: m.ID, Values: m.Values}
		}
		messages[i] = StreamMessage{Stream: r.Stream, Messages: entries}
	}
	return messages, nil
}

func (c *clientWrapper) XAck(ctx context.Context, stream, group string, ids ...string) (int64, error) {
	return c.rdb.XAck(ctx, c.config.WithKeyPrefix(stream), group, ids...).Result()
}

func (c *clientWrapper) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	prefixed := make([]string, len(keys))
	for i, k := range keys {
		prefixed[i] = c.config.WithKeyPrefix(k)
	}
	return c.rdb.Eval(ctx, script, prefixed, args...).Result()
}

func (c *clientWrapper) Info(ctx context.Context, section ...string) (string, error) {
	return c.rdb.Info(ctx, section...).Result()
}

func (c *clientWrapper) PoolStats() *PoolStats {
	stats := c.rdb.PoolStats()
	return &PoolStats{
		Hits: stats.Hits, Misses: stats.Misses, Timeouts: stats.Timeouts,
		TotalConns: stats.TotalConns, IdleConns: stats.IdleConns, StaleConns: stats.StaleConns,
	}
}

type pubSubWrapper struct {
	ps     *redis.PubSub
	config *Config
}

func (p *pubSubWrapper) Close() error { return p.ps.Close() }

func (p *pubSubWrapper) Channel() <-chan *Message {
	ch := p.ps.Channel()
	msgCh := make(chan *Message, cap(ch))
	go func() {
		for msg := range ch {
			msgCh <- &Message{Channel: msg.Channel, Pattern: msg.Pattern, Payload: msg.Payload}
		}
		close(msgCh)
	}()
	return msgCh
}
