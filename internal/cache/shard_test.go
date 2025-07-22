package cache

import (
	"context"
	"testing"
	"time"
)

func TestShardSetGet(t *testing.T) {
	metrics := createTestMetrics(t)
	shard, err := newShard(context.Background(), "s1", time.Second, 1024, 10, newLRUEvictorForTest(), metrics)
	if err != nil {
		t.Fatalf("newShard error: %v", err)
	}
	if err := shard.set("a", []byte("b")); err != nil {
		t.Fatalf("set error: %v", err)
	}

	value, err := shard.get("a")
	if err != nil || string(value) != "b" {
		t.Fatalf("unexpected get result %s %v", value, err)
	}
}

func TestShardExpiration(t *testing.T) {
	metrics := createTestMetrics(t)
	shard, _ := newShard(context.Background(), "s1", 10*time.Millisecond, 1024, 10, newLRUEvictorForTest(), metrics)

	if err := shard.set("a", []byte("b")); err != nil {
		t.Fatalf("set error: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	if _, err := shard.get("a"); err != ErrExpired {
		t.Fatalf("expected expired error, got %v", err)
	}
}

func TestMakeSpaceLocked(t *testing.T) {
	metrics := createTestMetrics(t)
	shard, _ := newShard(context.Background(), "s1", time.Minute, 5, 2, newLRUEvictorForTest(), metrics)

	shard.set("a", []byte("aa"))
	shard.set("b", []byte("bb"))

	shard.mu.Lock()
	ok := shard.makeSpaceLocked(2)
	shard.mu.Unlock()

	if !ok {
		t.Fatalf("expected to free space")
	}
}

func TestShardGetNotFound(t *testing.T) {
	metrics := createTestMetrics(t)
	shard, _ := newShard(context.Background(), "s1", time.Minute, 10, 10, newLRUEvictorForTest(), metrics)

	if _, err := shard.get("missing"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound")
	}
}

func TestShardSetInvalidValue(t *testing.T) {
	metrics := createTestMetrics(t)
	shard, _ := newShard(context.Background(), "s1", time.Minute, 10, 10, newLRUEvictorForTest(), metrics)

	if err := shard.set("k", []byte("")); err != ErrInvalidValue {
		t.Fatalf("expected ErrInvalidValue")
	}
}

func TestShardSetValueTooLarge(t *testing.T) {
	metrics := createTestMetrics(t)
	shard, _ := newShard(context.Background(), "s1", time.Minute, 2, 10, newLRUEvictorForTest(), metrics)

	if err := shard.set("k", []byte("too")); err != ErrValueTooLarge {
		t.Fatalf("expected ErrValueTooLarge")
	}
}

func TestShardSetOversized(t *testing.T) {
	metrics := createTestMetrics(t)
	shard, _ := newShard(context.Background(), "s1", time.Minute, 2, 10, newLRUEvictorForTest(), metrics)

	if err := shard.set("k", []byte("123")); err != ErrValueTooLarge {
		t.Fatalf("expected ErrValueTooLarge")
	}
}
