package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/VortexpluZ/chirpy/internal/auth"
	"github.com/VortexpluZ/chirpy/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Bad Request")
		log.Println(err)
		return
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		log.Println(err)
		return
	}
	user, err := cfg.database.CreateUser(r.Context(),
		database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashed_password,
		})

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		log.Println(err)
		return
	}

	respondWithJSON(w, 201, User{
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		ID:        user.ID,
		Email:     user.Email,
	})

}

func setTokenExpirationTime(expiresIn int) int {
	if expiresIn <= 3600 && expiresIn > 0 {
		return expiresIn
	}
	return 3600
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresIn int    `json:"expires_in_seconds,omitempty"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Bad Request")
		log.Println(err)
		return
	}

	user, err := cfg.database.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		log.Println(err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		log.Println(err)
		return
	}

	if !match {
		respondWithError(w, 401, "Unauthorized")
		log.Println("invalid password attempt for user")
		return
	}
	expTime, _ := time.ParseDuration(fmt.Sprintf("%vs", setTokenExpirationTime(params.ExpiresIn)))

	log.Println(expTime)
	token, err := auth.MakeJWT(user.ID, cfg.secret, expTime)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		log.Println(err)
		return
	}

	respondWithJSON(w, 200, User{
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		ID:        user.ID,
		Email:     user.Email,
		Token:     token,
	})

}
