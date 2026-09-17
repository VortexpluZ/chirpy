package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/VortexpluZ/chirpy/internal/auth"
	"github.com/VortexpluZ/chirpy/internal/database"
	_ "github.com/lib/pq"
)

type Token struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {

	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	dbToken, err := cfg.database.GetRefreshToken(r.Context(), refresh_token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Println(err)
		return
	}

	if dbToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Println("Refresh token revoked")
		return
	}

	if time.Now().After(dbToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Println("Refresh token expired")
		return
	}

	JwtExpTime, _ := time.ParseDuration(fmt.Sprintf("%vs", 3600))
	token, err := auth.MakeJWT(dbToken.UserID, cfg.secret, JwtExpTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		log.Println(err)
		return
	}

	respondWithJSON(w, http.StatusOK, Token{
		Token: token,
	})
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {

	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	dbToken, err := cfg.database.GetRefreshToken(r.Context(), refresh_token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Println(err)
		return
	}

	if dbToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Println("Refresh token revoked")
		return
	}

	if time.Now().After(dbToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Println("Refresh token expired")
		return
	}

	cfg.database.UpdateRefreshToken(r.Context(),
		database.UpdateRefreshTokenParams{
			Token:     refresh_token,
			RevokedAt: sql.NullTime{Time: time.Now(), Valid: true}})

	w.WriteHeader(http.StatusNoContent)
}
