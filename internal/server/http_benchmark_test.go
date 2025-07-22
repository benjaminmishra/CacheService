package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"cache-service/internal/cache"
)

func BenchmarkHandleGet(b *testing.B) {
	metrics := createTestMetrics(b)
	cacheInstance, _ := cache.NewCache(context.Background(), cache.WithMetrics(metrics))
	cacheInstance.Set("foo", []byte("bar"))
	httpServer := newHttpServer(":0", cacheInstance)
	request := httptest.NewRequest(http.MethodGet, "/foo", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		httpServer.Handler.ServeHTTP(recorder, request)
	}
}

func BenchmarkHandleSet(b *testing.B) {
	metrics := createTestMetrics(b)
	cacheInstance, _ := cache.NewCache(context.Background(), cache.WithMetrics(metrics))
	httpServer := newHttpServer(":0", cacheInstance)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodPost, "/key", bytes.NewBufferString("value"))
		recorder := httptest.NewRecorder()
		httpServer.Handler.ServeHTTP(recorder, request)
	}
}
