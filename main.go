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

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))

}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	dbQueries := database.New(db)
	mux := http.NewServeMux()
	config := apiConfig{database: dbQueries, platform: platform}
	mux.Handle("/app/", http.StripPrefix("/app", config.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/users", config.createUser)
	mux.HandleFunc("POST /api/login", config.login)
	mux.HandleFunc("POST /api/chirps", config.createChirp)
	mux.HandleFunc("GET /api/chirps", config.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", config.getChirp)
	mux.HandleFunc("GET /admin/metrics", config.getMetrics)
	mux.HandleFunc("POST /admin/reset", config.reset)
	server := &http.Server{Handler: mux, Addr: ":8080"}
	log.Printf("serving on port :8080")
	log.Fatal(server.ListenAndServe())
}
