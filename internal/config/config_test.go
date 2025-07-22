package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != 8080 {
		t.Fatalf("unexpected port")
	}
	if cfg.CacheTTL != 30*time.Minute {
		t.Fatalf("unexpected ttl")
	}
	if cfg.MaxCacheSize == 0 {
		t.Fatalf("max cache size not set")
	}
}

func TestLoadConfig(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("CACHE_TTL", "1h")
	os.Setenv("MAX_CACHE_SIZE", "2048")
	t.Cleanup(func() { os.Clearenv() })

	cfg, err := LoadConfig()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9090 {
		t.Fatalf("expected port 9090")
	}
	if cfg.CacheTTL != time.Hour {
		t.Fatalf("expected ttl 1h")
	}
	if cfg.MaxCacheSize != 2048 {
		t.Fatalf("expected size 2048")
	}
}

func TestValidateConfig(t *testing.T) {
	cfg := &Config{Port: 0, CacheTTL: -1, MaxCacheSize: -1}

	if err := validateConfig(cfg); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestLoadConfigDefaultMaxKeys(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("CACHE_TTL", "")
	t.Setenv("MAX_CACHE_SIZE", "")
	t.Setenv("MAX_KEYS", "")

	cfg, err := LoadConfig()

	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.MaxKeys != 2_000_000 {
		t.Errorf("expected default MaxKeys 2000000, got %d", cfg.MaxKeys)
	}
}

func TestLoadConfigMaxKeysFromEnv(t *testing.T) {
	t.Setenv("MAX_KEYS", "12345")
	t.Setenv("PORT", "")
	t.Setenv("CACHE_TTL", "")
	t.Setenv("MAX_CACHE_SIZE", "")

	cfg, err := LoadConfig()

	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.MaxKeys != 12345 {
		t.Errorf("expected MaxKeys 12345, got %d", cfg.MaxKeys)
	}
}

func TestLoadConfigInvalidEnv(t *testing.T) {
	t.Setenv("PORT", "notnum")

	if _, err := LoadConfig(); err == nil {
		t.Fatalf("expected error for invalid PORT")
	}
}
