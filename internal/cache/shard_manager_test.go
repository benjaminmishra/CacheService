package cache

import (
	"context"
	"testing"
	"time"

	"cache-service/internal/evictors"

	"stathat.com/c/consistent"
)

func TestShardManager(t *testing.T) {
	metrics := createTestMetrics(t)
	shardManager, err := newShardManager(context.Background(), 8, time.Minute, 128, 10, NewLRUEvictor, metrics)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(shardManager.shardMap) != 8 {
		t.Fatalf("expected 8 shards")
	}

	shard1 := shardManager.GetShard("a")
	shard2 := shardManager.GetShard("a")
	if shard1 != shard2 {
		t.Fatalf("expected consistent shard")
	}
}

func NewLRUEvictor() evictors.Evictor { return evictors.NewLRUEvictor() }

func TestShardManagerFallback(t *testing.T) {
	metrics := createTestMetrics(t)
	shardManager, err := newShardManager(context.Background(), 1, time.Minute, 128, 10, NewLRUEvictor, metrics)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	shardManager.ring = consistent.New()
	shard := shardManager.GetShard("key")
	if shard == nil {
		t.Fatalf("expected fallback shard")
	}
}
