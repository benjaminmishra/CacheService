package cache

import (
	"testing"
	"time"

	"cache-service/internal/evictors"
	"cache-service/internal/telemetry"
)

const (
	unexpectedErrStr string = "unexpected error: %v"
)

func TestWithTTL(t *testing.T) {
	cacheInstance := &Cache{}
	if err := WithTTL(5 * time.Second)(cacheInstance); err != nil {
		t.Fatalf(unexpectedErrStr, err)
	}
	if cacheInstance.ttl != 5*time.Second {
		t.Fatalf("expected ttl to be set")
	}
	if err := WithTTL(0)(cacheInstance); err == nil {
		t.Fatalf("expected error for ttl <= 0")
	}
}

func TestWithMaxSize(t *testing.T) {
	cacheInstance := &Cache{}
	if err := WithMaxSize(1024)(cacheInstance); err != nil {
		t.Fatalf(unexpectedErrStr, err)
	}
	if cacheInstance.maxSize != 1024 {
		t.Fatalf("expected maxSize to be set")
	}
	if err := WithMaxSize(0)(cacheInstance); err == nil {
		t.Fatalf("expected error for maxSize <= 0")
	}
}

func TestWithMaxKeys(t *testing.T) {
	cacheInstance := &Cache{}

	if err := WithMaxKeys(10)(cacheInstance); err != nil {
		t.Fatalf(unexpectedErrStr, err)
	}

	if cacheInstance.maxKeys != 10 {
		t.Fatalf("expected maxKeys to be set")
	}

	if err := WithMaxKeys(0)(cacheInstance); err == nil {
		t.Fatalf("expected error for maxKeys <= 0")
	}
}

func TestWithShardCount(t *testing.T) {
	cacheInstance := &Cache{}
	if err := WithShardCount(4)(cacheInstance); err != nil {
		t.Fatalf(unexpectedErrStr, err)
	}
	if cacheInstance.shardCount != 4 {
		t.Fatalf("expected shardCount to be set")
	}
	if err := WithShardCount(0)(cacheInstance); err == nil {
		t.Fatalf("expected error for shardCount <= 0")
	}
}

func TestWithEvictorFactory(t *testing.T) {
	cacheInstance := &Cache{}
	factory := func() evictors.Evictor { return evictors.NewLRUEvictor() }
	if err := WithEvictorFactory(factory)(cacheInstance); err != nil {
		t.Fatalf(unexpectedErrStr, err)
	}
	if cacheInstance.evictorFactory == nil {
		t.Fatalf("expected evictorFactory to be set")
	}
	if err := WithEvictorFactory(nil)(cacheInstance); err == nil {
		t.Fatalf("expected error for nil factory")
	}
}

func TestWithMetrics(t *testing.T) {
	cacheInstance := &Cache{}
	m := &telemetry.CacheMetrics{}
	if err := WithMetrics(m)(cacheInstance); err != nil {
		t.Fatalf(unexpectedErrStr, err)
	}
	if cacheInstance.metrics != m {
		t.Fatalf("expected metrics to be set")
	}
	if err := WithMetrics(nil)(cacheInstance); err == nil {
		t.Fatalf("expected error for nil metrics")
	}
}
