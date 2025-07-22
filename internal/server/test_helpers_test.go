package server

import (
	"testing"

	"cache-service/internal/telemetry"
	"go.opentelemetry.io/otel/metric/noop"
)

func createTestMetrics(tb testing.TB) *telemetry.CacheMetrics {
	tb.Helper()
	metrics, err := telemetry.NewCacheMetrics(noop.NewMeterProvider().Meter("test"))
	if err != nil {
		tb.Fatalf("failed to create metrics: %v", err)
	}
	return metrics
}
