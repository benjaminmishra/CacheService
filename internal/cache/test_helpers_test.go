package cache

import (
	"testing"

	"cache-service/internal/evictors"
	"cache-service/internal/telemetry"
	"go.opentelemetry.io/otel/metric/noop"
)

func newLRUEvictorForTest() evictors.Evictor { return evictors.NewLRUEvictor() }

func createTestMetrics(tb testing.TB) *telemetry.CacheMetrics {
	tb.Helper()
	metrics, err := telemetry.NewCacheMetrics(noop.NewMeterProvider().Meter("test"))
	if err != nil {
		tb.Fatalf("failed to create metrics: %v", err)
	}
	return metrics
}
