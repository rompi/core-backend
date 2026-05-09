package redis

import "time"

// ZMember represents a member of a sorted set.
type ZMember struct {
	Score  float64
	Member interface{}
}

// ZRangeBy represents options for sorted set range operations.
type ZRangeBy struct {
	Min    string
	Max    string
	Offset int64
	Count  int64
}

// GeoLocation represents a geographic location.
type GeoLocation struct {
	Name      string
	Longitude float64
	Latitude  float64
	Distance  float64
	GeoHash   int64
}

// GeoRadiusQuery represents options for geo radius queries.
type GeoRadiusQuery struct {
	Radius      float64
	Unit        string
	WithCoord   bool
	WithDist    bool
	WithGeoHash bool
	Count       int
	Sort        string
}

// StreamEntry represents an entry in a Redis stream.
type StreamEntry struct {
	ID     string
	Values map[string]interface{}
}

// StreamMessage represents a message from a stream read operation.
type StreamMessage struct {
	Stream   string
	Messages []StreamEntry
}

// Message represents a pub/sub message.
type Message struct {
	Channel      string
	Pattern      string
	Payload      string
	PayloadSlice []string
}

// PoolStats contains connection pool statistics.
type PoolStats struct {
	Hits       uint32
	Misses     uint32
	Timeouts   uint32
	TotalConns uint32
	IdleConns  uint32
	StaleConns uint32
}

// HealthStatus represents Redis health status.
type HealthStatus struct {
	Healthy   bool
	Latency   time.Duration
	Error     string
	PoolStats *PoolStats
	CheckedAt time.Time
}

// ServiceStats contains service statistics.
type ServiceStats struct {
	Operations     int64
	CacheHits      int64
	CacheMisses    int64
	Errors         int64
	AverageLatency time.Duration
	PoolStats      *PoolStats
}

// DistributedLock represents a distributed lock.
type DistributedLock struct {
	Key        string
	Value      string
	TTL        time.Duration
	AcquiredAt time.Time
}

// IsValid checks if the lock is still valid.
func (l *DistributedLock) IsValid() bool {
	if l == nil {
		return false
	}
	return time.Since(l.AcquiredAt) < l.TTL
}

// SlowLogEntry represents an entry in the slow log.
type SlowLogEntry struct {
	ID         int64
	Time       time.Time
	Duration   time.Duration
	Args       []string
	ClientAddr string
	ClientName string
}
