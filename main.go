package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/VortexpluZ/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database       *database.Queries
	platform       string
}

func profaneRewrite(text string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(text, " ")
	for i, word := range words {
		lowered := strings.ToLower(word)
		if _, exists := badWords[lowered]; exists {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error when stablishing connection to DB")
		return
	}
	dbQueries := database.New(db)
	mux := http.NewServeMux()
	config := apiConfig{database: dbQueries, platform: platform}
	mux.Handle("/app/", http.StripPrefix("/app", config.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("GET /api/healthz", healthz())
	mux.Handle("POST /api/users", config.createUser())
	mux.Handle("POST /api/login", config.login())
	mux.Handle("POST /api/chirps", config.createChirp())
	mux.Handle("GET /api/chirps", config.getChirps())
	mux.Handle("GET /api/chirps/{chirpID}", config.getChirp())
	mux.Handle("GET /admin/metrics", config.getMetrics())
	mux.Handle("POST /admin/reset", config.reset())
	server := &http.Server{Handler: mux, Addr: ":8080"}
	server.ListenAndServe()
}
