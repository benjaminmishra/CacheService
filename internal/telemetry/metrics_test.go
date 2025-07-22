package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/metric/noop"
)

func TestNewMeterProvider(t *testing.T) {
	provider, shutdown, err := NewMeterProvider(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if provider == nil || shutdown == nil {
		t.Fatalf("provider or shutdown was nil")
	}
	shutdown()
}

func TestNewCacheMetrics(t *testing.T) {
	metrics, err := NewCacheMetrics(noop.NewMeterProvider().Meter("test"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics == nil || metrics.Hits == nil || metrics.Misses == nil || metrics.Sets == nil || metrics.Evictions == nil || metrics.ItemCount == nil || metrics.Latency == nil || metrics.ErrorCount == nil {
		t.Fatalf("metrics not properly initialized")
	}
}
