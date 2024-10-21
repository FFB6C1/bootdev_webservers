package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(writer, request)
	})
	return handler
}

func main() {
	apiConfig := apiConfig{
		fileServerHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /api/metrics", apiConfig.metricsHandler)
	mux.HandleFunc("POST /api/reset", apiConfig.resetHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}

func readinessHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(writer http.ResponseWriter, _ *http.Request) {
	hitsString := fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())
	writer.Write([]byte(hitsString))
}

func (cfg *apiConfig) resetHandler(writer http.ResponseWriter, _ *http.Request) {
	cfg.fileServerHits.Store(0)
	writer.Write([]byte("file server hits reset."))
}
