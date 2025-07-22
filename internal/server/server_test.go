package server

import (
	"context"
	"testing"
	"time"

	"cache-service/internal/cache"
)

func TestNewCacheServer(t *testing.T) {
	ctx := context.Background()
	metrics := createTestMetrics(t)
	cacheInstance, _ := cache.NewCache(ctx, cache.WithMetrics(metrics))

	if _, err := NewCacheServer(0, cacheInstance); err == nil {
		t.Fatalf("expected error for invalid port")
	}

	if _, err := NewCacheServer(8080, nil); err == nil {
		t.Fatalf("expected error for nil cache")
	}

	cacheServer, err := NewCacheServer(8080, cacheInstance)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go func() {
		cacheServer.Http.ListenAndServe()
	}()

	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := cacheServer.Shutdown(ctx2); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
}

func TestCacheServerStart(t *testing.T) {
	ctx := context.Background()
	metrics := createTestMetrics(t)
	cacheInstance, _ := cache.NewCache(ctx, cache.WithMetrics(metrics))

	_, err := NewCacheServer(0, cacheInstance)
	if err == nil {
		t.Fatalf("expected error for invalid port")
	}

	cacheServer, err := NewCacheServer(8081, cacheInstance)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCh := cacheServer.Start()

	time.Sleep(10 * time.Millisecond)

	ctx2, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := cacheServer.Shutdown(ctx2); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}

	select {
	case perr := <-errCh:
		t.Errorf("unexpected protocol error: %v", perr)
	default:
		// No error, as expected
	}
}
