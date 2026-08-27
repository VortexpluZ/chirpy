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

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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
			log.Fatal(err)
			return
		}
	})
}

func (cfg *apiConfig) createChirp() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body   string    `json:"body"`
			UserId uuid.UUID `json:"user_id"`
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
		chirp, err := cfg.database.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   params.Body,
			UserID: params.UserId,
		})
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}
		respondWithJSON(w, 201, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	})
}

func (cfg *apiConfig) getChirps() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		chirps, err := cfg.database.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			log.Fatalln(err)
			return
		}

		_chirps := make([]Chirp, len(chirps))
		for i := 0; i < len(chirps); i++ {
			_chirps[i] = Chirp{
				ID:        chirps[i].ID,
				CreatedAt: chirps[i].CreatedAt,
				UpdatedAt: chirps[i].UpdatedAt,
				Body:      chirps[i].Body,
				UserID:    chirps[i].UserID,
			}
		}

		respondWithJSON(w, 200, _chirps)

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
	mux.Handle("POST /api/users", config.createUser())
	mux.Handle("POST /api/chirps", config.createChirp())
	mux.Handle("GET /api/chirps", config.getChirps())
	mux.Handle("GET /admin/metrics", config.getMetrics())
	mux.Handle("POST /admin/reset", config.reset())
	server := &http.Server{Handler: mux, Addr: ":8080"}
	server.ListenAndServe()
}
