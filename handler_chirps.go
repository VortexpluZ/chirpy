package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/VortexpluZ/chirpy/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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
			log.Println(err)
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

func (cfg *apiConfig) getChirp() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		chirpID, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			respondWithError(w, 400, "Bad Request")
			log.Println(err)
			return
		}

		chirp, err := cfg.database.GetChirp(r.Context(), chirpID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(w, http.StatusNotFound, "Chirp not found")
				return
			}

			respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirp")
			log.Printf("Error fetching chirp %s: %v", chirpID, err)
			return
		}

		_chirp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

		respondWithJSON(w, 200, _chirp)

	})
}
