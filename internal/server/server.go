package server

import (
	"cache-service/internal/cache"
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

// CacheServer provides an http server which can be used to interact with the cache
// can be extended to have multiple protocols such as gRPC RESP TCP etc
type CacheServer struct {
	Http *http.Server
}

// NewCacheServer constructs a server instance
func NewCacheServer(port int, cache *cache.Cache) (*CacheServer, error) {
	if cache == nil {
		return nil, fmt.Errorf("cache cannot be nil")
	}
	if port <= 0 {
		return nil, fmt.Errorf("port cannot be less than or equal to 0")
	}

	httpAddr := fmt.Sprintf(":%d", port)

	httpServer := newHttpServer(httpAddr, cache)
	return &CacheServer{Http: httpServer}, nil
}

// Start launches the HTTP and other servers asynchronously and returns an error channel
func (s *CacheServer) Start() <-chan ProtocolError {
	errChannel := make(chan ProtocolError, 1)

	slog.Info("Starting http server")

	// Start the HTTP server async
	go func() {
		err := s.Http.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			errChannel <- ProtocolError{
				Protocol: "http",
				Address:  s.Http.Addr,
				Err:      err,
			}
		}
	}()

	slog.Info("Listening to http requests on Address", "addr", s.Http.Addr)

	return errChannel
}

// Shutdown gracefully stops the HTTP server within the provided context
func (s *CacheServer) Shutdown(ctx context.Context) error {
	return s.Http.Shutdown(ctx)
}
