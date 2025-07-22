package cache

import (
	"context"
	"fmt"
	"time"

	"cache-service/internal/evictors"
	"cache-service/internal/telemetry"

	"github.com/google/uuid"
	"stathat.com/c/consistent"
)

type shardManager struct {
	shardMap   map[string]*cacheShard
	ring       *consistent.Consistent
	shardCount int
}

func newShardManager(
	ctx context.Context,
	shardCount int,
	ttl time.Duration,
	maxSizePerShard int64,
	maxKeysPerShard int,
	evictorFactory func() evictors.Evictor,
	metrics *telemetry.CacheMetrics,
) (*shardManager, error) {

	shardMap := make(map[string]*cacheShard, shardCount)
	shardNodes := make([]string, 0, shardCount)
	ring := consistent.New()

	for range shardCount {
		shardID := uuid.New().String()
		shardNodes = append(shardNodes, shardID)

		// Create a new instance of evictor per shard
		// Reuse the same metrics instance as we want to aggregate metrics across shards
		shard, err := newShard(ctx, shardID, ttl, maxSizePerShard, maxKeysPerShard, evictorFactory(), metrics)

		if err != nil {
			return nil, fmt.Errorf("failed to create shard: %w", err)
		}

		shardMap[shardID] = shard
	}

	// Sets the shards on an consistent hasing ring ensure even distribution
	// TODO: We can still may have unevent distribution, but this is good enough for now
	ring.Set(shardNodes)

	return &shardManager{
		shardMap:   shardMap,
		ring:       ring,
		shardCount: shardCount,
	}, nil
}

func (sm *shardManager) GetShard(shardKey string) *cacheShard {
	shardID, err := sm.ring.Get(shardKey)

	if err != nil {
		// TODO: Implement robust fallback as this may result in hot spots
		// For now, we return the first shard as a fallback
		for _, shard := range sm.shardMap {
			return shard
		}
	}

	return sm.shardMap[shardID]
}
