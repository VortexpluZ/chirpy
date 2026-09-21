package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/VortexpluZ/chirpy/internal/auth"
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

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	chirp, err := cfg.database.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})

}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {

	author_id := r.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error

	if author_id != "" {
		user_id, err := uuid.Parse(author_id)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Bad Request")
			log.Println(err)
			return
		}
		chirps, err = cfg.database.GetChirpsByAuthor(r.Context(), user_id)
	} else {
		chirps, err = cfg.database.GetChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
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

	sort_way := r.URL.Query().Get("sort")
	if sort_way != "" {
		switch sort_way {
		case "asc":
			slices.SortFunc(_chirps, func(a Chirp, b Chirp) int {
				if a.CreatedAt.Equal(b.CreatedAt) {
					return 0
				} else if a.CreatedAt.Before(b.CreatedAt) {
					return -1
				} else {
					return 1
				}
			})
		case "desc":
			slices.SortFunc(_chirps, func(a Chirp, b Chirp) int {
				if a.CreatedAt.Equal(b.CreatedAt) {
					return 0
				} else if a.CreatedAt.Before(b.CreatedAt) {
					return 1
				} else {
					return -1
				}
			})
		}
	}

	respondWithJSON(w, http.StatusOK, _chirps)

}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request")
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

	respondWithJSON(w, http.StatusOK, _chirp)

}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		log.Println(err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		log.Println(err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request")
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

	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "Cannot delete other user's chirp")
		return
	}

	_, err = cfg.database.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			log.Println(err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Could not delete chirp")
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
