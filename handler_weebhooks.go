package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/VortexpluZ/chirpy/internal/auth"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PolkaHook struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) upgradeUserHook(w http.ResponseWriter, r *http.Request) {

	apiToken, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if apiToken != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	decoder := json.NewDecoder(r.Body)
	polka := PolkaHook{}
	err = decoder.Decode(&polka)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if polka.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Invalid user action")
		return
	}

	_, err = cfg.database.UpdateUserToRed(r.Context(), polka.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
