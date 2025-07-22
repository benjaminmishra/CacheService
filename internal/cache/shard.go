package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cache-service/internal/evictors"
	"cache-service/internal/telemetry"
)

type cacheShard struct {
	id          string
	items       map[string]*cacheItem
	mu          sync.RWMutex
	ttl         time.Duration
	maxSize     int64
	maxKeys     int
	currentSize int64
	evictor     evictors.Evictor
	metrics     *telemetry.CacheMetrics // for tracking metrics at shard level
	ctx         context.Context
}

type cacheItem struct {
	Value     []byte
	ExpiresAt time.Time
	Size      int64
}

func (c *cacheItem) isExpired() bool {
	if c.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(c.ExpiresAt)
}

// newShard creates a new shard instace
func newShard(ctx context.Context, shardId string, ttl time.Duration, maxSize int64, maxKeys int, evictorInstance evictors.Evictor, metrics *telemetry.CacheMetrics) (*cacheShard, error) {

	if ttl <= 0 {
		return nil, fmt.Errorf("ttl must be positive, got %v", ttl)
	}

	if maxSize <= 0 {
		return nil, fmt.Errorf("maxSize must be positive, got %d", maxSize)
	}

	if maxKeys <= 0 {
		return nil, fmt.Errorf("maxKeys must be positive, got %d", maxKeys)
	}

	if evictorInstance == nil {
		return nil, fmt.Errorf("evictor instance must not be nil")
	}

	if metrics == nil {
		return nil, fmt.Errorf("metrics must not be nil")
	}

	c := &cacheShard{
		id:      shardId,
		items:   make(map[string]*cacheItem),
		ttl:     ttl,
		maxSize: maxSize,
		maxKeys: maxKeys,
		evictor: evictorInstance,
		metrics: metrics,
		ctx:     ctx,
	}

	return c, nil
}

func (c *cacheShard) set(key string, value []byte) error {
	incomingItemSize := int64(len(value))
	if incomingItemSize == 0 {
		return ErrInvalidValue
	}

	// Check if the item exceeds the maximum size of the entire shard
	if c.maxSize > 0 && incomingItemSize > c.maxSize {
		c.metrics.ErrorCount.Add(c.ctx, 1)
		return ErrValueTooLarge
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	_, exists := c.items[key]
	if !exists && (c.maxKeys > 0 && len(c.items)+1 > c.maxKeys) {
		c.metrics.ErrorCount.Add(c.ctx, 1)
		return ErrTooManyKeys
	}

	extraSpaceNeeded := incomingItemSize
	if exists {
		extraSpaceNeeded -= c.items[key].Size
	}

	// Try to make space if the incoming item needs more space
	for c.maxSize > 0 && c.currentSize+extraSpaceNeeded > c.maxSize {
		if !c.makeSpaceLocked(extraSpaceNeeded) {
			c.metrics.ErrorCount.Add(c.ctx, 1)
			return ErrCacheFull
		}
	}

	c.setLocked(key, value, incomingItemSize)
	return nil
}

func (c *cacheShard) setLocked(key string, value []byte, itemSize int64) {
	if oldItem, exists := c.items[key]; exists {
		c.currentSize -= oldItem.Size
	} else {
		c.metrics.ItemCount.Add(c.ctx, 1)
	}

	item := &cacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
		Size:      itemSize,
	}

	c.items[key] = item
	c.currentSize += itemSize
	c.metrics.Sets.Add(c.ctx, 1)

	c.evictor.OnSet(key)
}

func (c *cacheShard) get(key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]

	if !exists {
		c.metrics.Misses.Add(c.ctx, 1)
		return nil, ErrNotFound
	}

	// Cleanup expired items
	if item.isExpired() {
		c.removeKeyLocked(key)
		c.metrics.Misses.Add(c.ctx, 1)
		return nil, ErrExpired
	}

	c.evictor.OnGet(key)
	c.metrics.Hits.Add(c.ctx, 1)

	return item.Value, nil
}

func (c *cacheShard) makeSpaceLocked(neededSpace int64) bool {
	c.cleanupExpiredLocked()

	if c.currentSize <= c.maxSize-neededSpace {
		return true
	}

	// Estimate how many items to evict:
	// For now just try to figure out the average size of each item value
	// and calculate how many items we need to evict.
	averageItemSize := c.currentSize / (int64(len(c.items)) + 1)
	if averageItemSize == 0 {
		averageItemSize = 1
	}

	spaceToFree := c.currentSize - (c.maxSize - neededSpace)
	countToEvict := int(spaceToFree/averageItemSize) + 1 // +1 just to be safe

	keysToEvict := c.evictor.Evict(countToEvict)
	if len(keysToEvict) == 0 {
		return false
	}

	// Remove the evicted keys from the cache.
	for _, key := range keysToEvict {
		c.removeKeyLocked(key)
	}

	return c.currentSize <= c.maxSize-neededSpace
}

func (c *cacheShard) cleanupExpiredLocked() {
	var keysToDelete []*string = make([]*string, 0, len(c.items))

	for key, item := range c.items {
		if item.isExpired() {
			keysToDelete = append(keysToDelete, &key)
		}
	}

	for _, keyToDelete := range keysToDelete {
		c.removeKeyLocked(*keyToDelete)
	}
}

func (c *cacheShard) removeKeyLocked(key string) {
	item, exists := c.items[key]

	if !exists {
		return
	}

	c.currentSize -= item.Size
	delete(c.items, key)

	c.metrics.ItemCount.Add(c.ctx, -1)
}
