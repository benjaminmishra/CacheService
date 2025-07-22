package cache

import (
	"cache-service/internal/evictors"
	"cache-service/internal/telemetry"
	"fmt"
	"time"
)

type CacheOption func(*Cache) error

func WithTTL(ttl time.Duration) CacheOption {
	return func(c *Cache) error {
		if ttl <= 0 {
			return fmt.Errorf("ttl must be positive, got %v", ttl)
		}
		c.ttl = ttl
		return nil
	}
}

func WithMaxSize(maxSize int64) CacheOption {
	return func(c *Cache) error {
		if maxSize <= 0 {
			return fmt.Errorf("maxSize must be positive, got %d", maxSize)
		}
		c.maxSize = maxSize
		return nil
	}
}

func WithMaxKeys(maxKeys int) CacheOption {
	return func(c *Cache) error {
		if maxKeys <= 0 {
			return fmt.Errorf("maxKeys must be positive, got %d", maxKeys)
		}
		c.maxKeys = maxKeys
		return nil
	}
}

func WithShardCount(shardCount int) CacheOption {
	return func(c *Cache) error {
		if shardCount <= 0 {
			return fmt.Errorf("shardCount must be positive, got %d", shardCount)
		}
		c.shardCount = shardCount
		return nil
	}
}

func WithEvictorFactory(factory func() evictors.Evictor) CacheOption {
	return func(c *Cache) error {
		if factory == nil {
			return fmt.Errorf("evictor factory must not be nil")
		}
		c.evictorFactory = factory
		return nil
	}
}

func WithMetrics(metrics *telemetry.CacheMetrics) CacheOption {
	return func(c *Cache) error {
		if metrics == nil {
			return fmt.Errorf("metrics must not be nil")
		}
		c.metrics = metrics
		return nil
	}
}
