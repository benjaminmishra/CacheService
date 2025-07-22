package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cache-service/internal/cache"
	"cache-service/internal/config"
	"cache-service/internal/evictors"
	"cache-service/internal/server"
	"cache-service/internal/telemetry"
)

func main() {

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(logHandler))

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration, shutting down", "error", err)
		os.Exit(1)
	}

	cacheCtx, cacheCancel := context.WithCancel(context.Background())
	defer cacheCancel()

	meterProvider, shutdownMeterProvider, err := telemetry.NewMeterProvider(cacheCtx)
	if err != nil {
		slog.Error("failed to create meter provider:", "err", err)
	} else {
		defer shutdownMeterProvider()
	}

	// Get a meter for the cache
	meter := meterProvider.Meter("cache-service/cache")
	cacheMetrics, err := telemetry.NewCacheMetrics(meter)
	if err != nil {
		slog.Error("failed to create cache metrics:", "err", err)
	}

	// Create a cache, this also creates the shards of the cache
	cache, err := cache.NewCache(
		cacheCtx,
		cache.WithMaxSize(cfg.MaxCacheSize),
		cache.WithMaxKeys(cfg.MaxKeys),
		cache.WithTTL(cfg.CacheTTL),
		cache.WithShardCount(512),
		cache.WithMetrics(cacheMetrics),
		cache.WithEvictorFactory(func() evictors.Evictor { return evictors.NewLRUEvictor() }))

	if err != nil {
		slog.Error("failed to create cache:", "err", err)
		os.Exit(1)
	}

	// Create a cache server for communicating with the ourside world
	cacheServer, err := server.NewCacheServer(cfg.Port, cache)
	if err != nil {
		slog.Error("failed to create cache server:", "err", err)
		os.Exit(1)
	}

	errChannel := cacheServer.Start()

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case protocolErr := <-errChannel:
		slog.Error("cache server error:", "protocol", protocolErr.Protocol, "error", protocolErr.Err)
		performGracefulShutdown(cacheServer, cacheCancel)
		os.Exit(1)

	case sig := <-shutdownChannel:
		slog.Info("received signal", "sig", sig, "message", "shutting down cache server")
		performGracefulShutdown(cacheServer, cacheCancel)
		slog.Info("cache server shut down gracefully")
	}
}

func performGracefulShutdown(cacheServer *server.CacheServer, cacheCancel context.CancelFunc) {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	cacheCancel()

	if err := cacheServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("error during server shutdown:", "err", err)
	}
}
