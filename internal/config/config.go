package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"cache-service/internal/evictors"
)

type Config struct {
	Port           int
	CacheTTL       time.Duration
	MaxCacheSize   int64
	MaxKeys        int
	EvictorFactory func() evictors.Evictor
}

// LoadConfig loads the configuration from environment variables with defaults.
// can be extended to load from files or other sources.
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()
	var err error

	if err = loadEnvVar(&cfg.Port, "PORT", strconv.Atoi); err != nil {
		return nil, err
	}

	if err = loadEnvVar(&cfg.CacheTTL, "CACHE_TTL", time.ParseDuration); err != nil {
		return nil, err
	}

	if err = loadEnvVar(&cfg.MaxCacheSize, "MAX_CACHE_SIZE", func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }); err != nil {
		return nil, err
	}

	if err = loadEnvVar(&cfg.MaxKeys, "MAX_KEYS", strconv.Atoi); err != nil {
		return nil, err
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation error: %w", err)
	}

	return cfg, nil
}

func DefaultConfig() *Config {
	return &Config{
		Port:           8080,
		CacheTTL:       30 * time.Minute,
		MaxCacheSize:   1024 * 1024 * 1024,
		MaxKeys:        2_000_000,
		EvictorFactory: func() evictors.Evictor { return evictors.NewLRUEvictor() },
	}
}

func loadEnvVar[T any](targetField *T, key string, parser func(string) (T, error)) error {
	if value := os.Getenv(key); value != "" {
		parsedValue, err := parser(value)
		if err != nil {
			return fmt.Errorf("invalid value for %s: %w", key, err)
		}

		*targetField = parsedValue
	}

	return nil
}

func validateConfig(cfg *Config) error {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", cfg.Port)
	}
	if cfg.CacheTTL <= 0 {
		return fmt.Errorf("CACHE_TTL must be a positive duration, got %s", cfg.CacheTTL)
	}
	if cfg.MaxCacheSize <= 0 {
		return fmt.Errorf("MAX_CACHE_SIZE must be a positive integer, got %d", cfg.MaxCacheSize)
	}
	if cfg.MaxKeys <= 0 {
		return fmt.Errorf("MAX_KEYS must be a positive integer, got %d", cfg.MaxKeys)
	}

	return nil
}
