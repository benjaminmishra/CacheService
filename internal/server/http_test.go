package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"cache-service/internal/cache"
)

func TestHandleSetGet(t *testing.T) {
	ctx := context.Background()
	metrics := createTestMetrics(t)
	cacheInstance, _ := cache.NewCache(ctx, cache.WithMetrics(metrics))
	httpServer := newHttpServer(":0", cacheInstance)

	// Get key api test
	responseRecorder := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/cache/foo", bytes.NewBufferString("bar"))
	httpServer.Handler.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected code: %d", responseRecorder.Code)
	}

	// Set key api test
	responseRecorder = httptest.NewRecorder()

	request = httptest.NewRequest(http.MethodGet, "/api/v1/cache/foo", nil)
	httpServer.Handler.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected code: %d", responseRecorder.Code)
	}
	if responseRecorder.Body.String() != "bar" {
		t.Fatalf("unexpected body: %s", responseRecorder.Body.String())
	}
}

func TestHealthEndpointOK(t *testing.T) {
	srv := newHttpServer(":0", nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	expected := `{"status":"ok"}`
	if rr.Body.String() != expected {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestHealthEndpointMethodNotAllowed(t *testing.T) {
	srv := newHttpServer(":0", nil)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rr.Code)
	}
}

func TestHandleSetMissingValue(t *testing.T) {
	metrics := createTestMetrics(t)
	c, _ := cache.NewCache(context.Background(), cache.WithMetrics(metrics))
	srv := newHttpServer(":0", c)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cache/a", nil)

	srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleGetNotFound(t *testing.T) {
	metrics := createTestMetrics(t)
	c, _ := cache.NewCache(context.Background(), cache.WithMetrics(metrics))
	srv := newHttpServer(":0", c)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cache/missing", nil)

	srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleSetCacheFull(t *testing.T) {
	metrics := createTestMetrics(t)
	c, _ := cache.NewCache(context.Background(), cache.WithShardCount(1), cache.WithMaxSize(2), cache.WithMetrics(metrics))
	srv := newHttpServer(":0", c)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cache/foo", bytes.NewBufferString("aaa"))

	srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
}

func TestDocsEndpoints(t *testing.T) {
	srv := newHttpServer(":0", nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs/swagger.html", nil)
	srv.Handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	srv.Handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
