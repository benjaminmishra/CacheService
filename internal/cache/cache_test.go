package cache

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"cache-service/internal/evictors"
)

const (
	setErrStr string = "set error : %v"
)

type testEvictor struct{}

func (testEvictor) OnSet(string)         {}
func (testEvictor) OnGet(string)         {}
func (testEvictor) OnDelete(string)      {}
func (testEvictor) Evict(n int) []string { return []string{"a"} }

func TestCacheSetGet(t *testing.T) {
	metrics := createTestMetrics(t)
	cacheInstance, err := NewCache(context.Background(), WithMetrics(metrics))

	if err != nil {
		t.Fatalf("new cache error: %v", err)
	}
	if err := cacheInstance.Set("a", []byte("b")); err != nil {
		t.Fatalf(setErrStr, err)
	}

	value, err := cacheInstance.Get("a")
	if err != nil || string(value) != "b" {
		t.Fatalf("unexpected get %s %v", value, err)
	}
}

func TestCacheExpiration(t *testing.T) {
	metrics := createTestMetrics(t)
	cacheInstance, _ := NewCache(context.Background(), WithTTL(10*time.Millisecond), WithMetrics(metrics))

	if err := cacheInstance.Set("a", []byte("b")); err != nil {
		t.Fatalf(setErrStr, err)
	}

	time.Sleep(20 * time.Millisecond)
	if _, err := cacheInstance.Get("a"); err != ErrExpired {
		t.Fatalf("expected expired error")
	}
}

func TestCacheGetNotFound(t *testing.T) {
	metrics := createTestMetrics(t)
	cacheInstance, _ := NewCache(context.Background(), WithMetrics(metrics))

	if _, err := cacheInstance.Get("missing"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCacheSetInvalidValue(t *testing.T) {
	metrics := createTestMetrics(t)
	cacheInstance, _ := NewCache(context.Background(), WithShardCount(1), WithMetrics(metrics))

	if err := cacheInstance.Set("k", []byte("")); err != ErrInvalidValue {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestCacheValueTooLarge(t *testing.T) {
	metrics := createTestMetrics(t)
	cacheInstance, _ := NewCache(context.Background(), WithShardCount(2), WithMaxSize(2), WithMetrics(metrics))
	if err := cacheInstance.Set("k", []byte("too")); err != ErrValueTooLarge {
		t.Fatalf("expected ErrValueTooLarge, got %v", err)
	}
}

func TestCacheSetMaxKeysExceeded(t *testing.T) {
	metrics := createTestMetrics(t)

	cacheInstance, _ := NewCache(
		context.Background(),
		WithShardCount(1),
		WithMaxKeys(1),
		WithEvictorFactory(func() evictors.Evictor { return testEvictor{} }),
		WithMetrics(metrics))

	if err := cacheInstance.Set("a", []byte("1")); err != nil {
		t.Fatalf(setErrStr, err)
	}
	if err := cacheInstance.Set("b", []byte("2")); err != ErrTooManyKeys {
		t.Fatalf("expected ErrTooManyKeys, got %v", err)
	}
}

func TestCacheConcurrency(t *testing.T) {
	metrics := createTestMetrics(t)
	cache, err := NewCache(context.Background(), WithMetrics(metrics))
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 1000        // number of concurrent workers
	numOpsPerGouroutine := 10000 // operations per worker

	numKeys := 100

	for range numGoroutines {
		wg.Add(1)

		// Each goroutine performs a mix of get and set operations
		// here we keep the ratio of reads to writes at 80% reads and 20% writes
		go func() {
			defer wg.Done()

			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for j := range numOpsPerGouroutine {
				key := fmt.Sprintf("key-%d", r.Intn(numKeys))
				value := fmt.Appendf(nil, "value-%d", j)

				operationNum := r.Intn(100)
				if operationNum < 80 {
					_, _ = cache.Get(key)
				} else {
					_ = cache.Set(key, value)
				}
			}
		}()
	}

	wg.Wait()
}
