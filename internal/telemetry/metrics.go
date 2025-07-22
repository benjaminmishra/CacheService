package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// NewMeterProvider creates a basic meter provider that exports metrics to stdout
// Can be modified or a new one can plugged in to export metrics to something else
func NewMeterProvider(ctx context.Context) (*sdkmetric.MeterProvider, func(), error) {
	exporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		log.Fatalf("failed to create stdout exporter: %v", err)
		return nil, nil, err
	}

	reader := sdkmetric.NewPeriodicReader(exporter)
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	shutdown := func() {
		err = provider.Shutdown(ctx)
		if err != nil {
			log.Fatalf("failed to shutdown meter provider: %v", err)
		}
	}

	return provider, shutdown, nil
}

type CacheMetrics struct {
	Hits       metric.Int64Counter
	Misses     metric.Int64Counter
	Sets       metric.Int64Counter
	Evictions  metric.Int64Counter
	ItemCount  metric.Int64Counter
	Latency    metric.Float64Histogram
	ErrorCount metric.Float64Counter
}

// NewCacheMetrics creates metric counters used by the cache.
func NewCacheMetrics(m metric.Meter) (*CacheMetrics, error) {
	hits, err := m.Int64Counter("cache_hits")
	if err != nil {
		return nil, err
	}

	misses, err := m.Int64Counter("cache_misses")
	if err != nil {
		return nil, err
	}

	sets, err := m.Int64Counter("cache_sets")
	if err != nil {
		return nil, err
	}

	evictions, err := m.Int64Counter("cache_evictions")
	if err != nil {
		return nil, err
	}

	itemCount, err := m.Int64Counter("cache_item_count")
	if err != nil {
		return nil, err
	}

	getLatency, err := m.Float64Histogram("cache_get_latency")
	if err != nil {
		return nil, err
	}

	errorCount, err := m.Float64Counter("cache_error_count")
	if err != nil {
		return nil, err
	}

	return &CacheMetrics{
		Hits:       hits,
		Misses:     misses,
		Sets:       sets,
		Evictions:  evictions,
		ItemCount:  itemCount,
		Latency:    getLatency,
		ErrorCount: errorCount,
	}, nil
}
