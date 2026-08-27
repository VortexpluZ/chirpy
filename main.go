package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/VortexpluZ/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database       *database.Queries
	platform       string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getMetrics() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`<html>
		<body>
		  <h1>Welcome, Chirpy Admin</h1>
		  <p>Chirpy has been visited %d times!</p>
		</body>
	  </html>`, cfg.fileserverHits.Load())))
	})
}

func (cfg *apiConfig) resetMetrics() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)
	})
}

func (cfg *apiConfig) reset() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.platform != "dev" {
			respondWithError(w, 403, "Forbidden")
			return
		}
		err := cfg.database.TruncateUsers(r.Context())
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnVals struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, returnVals{
		Error: msg,
	})
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

func validateChirp() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		type parameters struct {
			Body string `json:"body"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}
		if len(params.Body) > 140 {
			respondWithError(w, 400, "Chirp is too long")
			return
		}
		type returnVals struct {
			CleanedBody string `json:"cleaned_body"`
		}
		respondWithJSON(w, 200, returnVals{
			CleanedBody: profaneRewrite(params.Body),
		})

	})
}

func healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
}

func (cfg *apiConfig) createUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email string `json:"email"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		user, err := cfg.database.CreateUser(r.Context(), params.Email)
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		respondWithJSON(w, 201, User{
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			ID:        user.ID,
			Email:     user.Email,
		})
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
	mux.Handle("POST /api/validate_chirp", validateChirp())
	mux.Handle("POST /api/users", config.createUser())
	mux.Handle("GET /admin/metrics", config.getMetrics())
	mux.Handle("POST /admin/reset", config.reset())
	server := &http.Server{Handler: mux, Addr: ":8080"}
	server.ListenAndServe()
}
