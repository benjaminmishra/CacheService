package cache

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
)

func BenchmarkCacheReadWrite(b *testing.B) {
	metrics := createTestMetrics(b)
	cache, _ := NewCache(context.Background(), WithMetrics(metrics))
	cache.Set("k", []byte("v"))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i%5 == 0 {
			cache.Set("k", []byte("v"))
		} else {
			cache.Get("k")
		}
	}
}

func BenchmarkCacheSet(b *testing.B) {
	metrics := createTestMetrics(b)
	cache, _ := NewCache(context.Background(), WithShardCount(4), WithMetrics(metrics))
	value := []byte("value")

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		cache.Set(fmt.Sprintf("k%d", i), value)
	}
}

func BenchmarkCacheGetMiss(b *testing.B) {
	metrics := createTestMetrics(b)
	cache, _ := NewCache(context.Background(), WithShardCount(4), WithMetrics(metrics))

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		cache.Get(fmt.Sprintf("missing%d", i))
	}
}

func BenchmarkCacheParallelReads(b *testing.B) {
	metrics := createTestMetrics(b)
	cache, _ := NewCache(context.Background(), WithShardCount(4), WithMetrics(metrics))
	cache.Set("k", []byte("v"))

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("k")
		}
	})
}

func BenchmarkCacheParallelWrites(b *testing.B) {
	metrics := createTestMetrics(b)
	cacheInstance, _ := NewCache(context.Background(), WithShardCount(4), WithMetrics(metrics))
	value := []byte("value")
	var counter uint64

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := fmt.Sprintf("k%d", atomic.AddUint64(&counter, 1))
			cacheInstance.Set(key, value)
		}
	})
}
