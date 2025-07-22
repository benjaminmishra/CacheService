package cache

import (
	"context"
	"fmt"
	"time"

	"cache-service/internal/evictors"
	"cache-service/internal/telemetry"
)

const (
	DefaultShardCount = 256
	DefaultMaxSize    = 1024 * 1024 * 1024 // 1 GB
	DefaultMaxKeys    = 2_000_000          // 2 million keys
	DefaultTTL        = 30 * time.Minute
)

// Cache is a sharded in-memory cache.
type Cache struct {
	shardManager   *shardManager
	ttl            time.Duration
	maxSize        int64
	maxKeys        int
	shardCount     int
	evictorFactory func() evictors.Evictor
	metrics        *telemetry.CacheMetrics
}

// NewCache constructs a Cache instance using the provided options
func NewCache(ctx context.Context, opts ...CacheOption) (*Cache, error) {
	c := &Cache{
		ttl:            DefaultTTL,
		maxSize:        DefaultMaxSize,
		maxKeys:        DefaultMaxKeys,
		shardCount:     DefaultShardCount,
		evictorFactory: func() evictors.Evictor { return evictors.NewLRUEvictor() },
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	maxSizePerShard := c.maxSize / int64(c.shardCount)
	maxKeysPerShard := c.maxKeys / c.shardCount

	shardManagerInstance, err := newShardManager(ctx, c.shardCount, c.ttl, maxSizePerShard, maxKeysPerShard, c.evictorFactory, c.metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create shard manager: %w", err)
	}

	c.shardManager = shardManagerInstance
	return c, nil
}

func (c *Cache) Get(key string) ([]byte, error) {
	shard := c.shardManager.GetShard(key)
	val, err := shard.get(key)
	return val, err
}

func (c *Cache) Set(key string, value []byte) error {
	shard := c.shardManager.GetShard(key)
	if err := shard.set(key, value); err != nil {
		return err
	}
	return nil
}
