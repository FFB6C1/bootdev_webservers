package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/FFB6C1/bootdev_webservers/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	database       *database.Queries
}

func main() {
	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Could not open database:", err)
	}
	dbQueries := database.New(db)
	apiConfig := apiConfig{
		fileServerHits: atomic.Int32{},
		database:       dbQueries,
	}
	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiConfig.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiConfig.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
