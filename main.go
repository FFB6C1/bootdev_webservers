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
	db             *database.Queries
	platform       string
	secret         string
}

func main() {
	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Could not open database:", err)
	}
	dbQueries := database.New(db)
	apiConfig := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
	}
	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiConfig.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiConfig.resetHandler)
	mux.HandleFunc("POST /api/chirps", apiConfig.postNewChirpHandler)
	mux.HandleFunc("GET /api/chirps", apiConfig.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.getChirpByIdHandler)
	mux.HandleFunc("POST /api/users", apiConfig.newUserHandler)
	mux.HandleFunc("POST /api/login", apiConfig.loginHandler)
	mux.HandleFunc("POST /api/refresh", apiConfig.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apiConfig.revokeHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
