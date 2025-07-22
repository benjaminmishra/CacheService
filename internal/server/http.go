package server

import (
	"embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"

	"cache-service/internal/cache"
)

// TODO: Move the docs to a separate package instead of embedding them here.

//go:embed docs/*
var docsFS embed.FS

var docsSubFS = func() fs.FS {
	sub, err := fs.Sub(docsFS, "docs")
	if err != nil {
		return docsFS
	}
	return sub
}()

func newHttpServer(addr string, cache *cache.Cache) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		key := r.PathValue("key")
		handleGet(cache, w, r, key)
	})

	mux.HandleFunc("POST /api/v1/cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		key := r.PathValue("key")

		handleSet(cache, w, r, key)
	})

	mux.HandleFunc("/health", handleHealth)
	mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.FS(docsSubFS))))
	mux.Handle("/openapi.yaml", http.FileServer(http.FS(docsSubFS)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/docs/swagger.html", http.StatusFound)
			return
		}
	})

	mux.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func handleSet(store *cache.Cache, w http.ResponseWriter, r *http.Request, key string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	if err := store.Set(key, body); err != nil {
		switch {
		case errors.Is(err, cache.ErrCacheFull):
			respondWithError(w, "cache is full", http.StatusInsufficientStorage)
		case errors.Is(err, cache.ErrInvalidValue):
			respondWithError(w, "invalid value", http.StatusBadRequest)
		case errors.Is(err, cache.ErrValueTooLarge):
			respondWithError(w, "value too large", http.StatusRequestEntityTooLarge)
		default:
			respondWithError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleGet(store *cache.Cache, w http.ResponseWriter, _ *http.Request, key string) {
	val, err := store.Get(key)

	if err != nil {
		switch {
		case errors.Is(err, cache.ErrNotFound):
			respondWithError(w, "key not found", http.StatusNotFound)
		case errors.Is(err, cache.ErrExpired):
			respondWithError(w, "key expired", http.StatusNotFound)
		default:
			respondWithError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(val)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func respondWithError(w http.ResponseWriter, message string, code int) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
