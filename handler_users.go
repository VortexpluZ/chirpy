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

type UserIdentity struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	IsRed     bool      `json:"is_chirpy_red"`
}

type User struct {
	UserIdentity
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
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
		respondWithError(w, http.StatusBadRequest, "Bad Request")
		log.Println(err)
		return
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		log.Println(err)
		return
	}
	user, err := cfg.database.CreateUser(r.Context(),
		database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashed_password,
		})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		log.Println(err)
		return
	}

	respondWithJSON(w, http.StatusCreated, UserIdentity{
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		ID:        user.ID,
		Email:     user.Email,
		IsRed:     user.IsChirpyRed,
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request")
		log.Println(err)
		return
	}

	user, err := cfg.database.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Println(err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Println(err)
		return
	}

	if !match {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Println("invalid password attempt for user")
		return
	}
	JwtExpTime, _ := time.ParseDuration(fmt.Sprintf("%vs", 3600))
	tokenExpTime, _ := time.ParseDuration(fmt.Sprintf("%vh", 1440))

	token, err := auth.MakeJWT(user.ID, cfg.secret, JwtExpTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		log.Println(err)
		return
	}

	refresh_token := auth.MakeRefreshToken()
	cfg.database.CreateRefreshToken(r.Context(),
		database.CreateRefreshTokenParams{
			Token:     refresh_token,
			ExpiresAt: time.Now().Add(tokenExpTime),
			UserID:    user.ID,
		})

	respondWithJSON(w, http.StatusOK, User{
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		ID:           user.ID,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refresh_token,
		IsRed:        user.IsChirpyRed,
	})

}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
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

	type parameters struct {
		NewEmail    string `json:"email"`
		NewPassword string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request")
		log.Println(err)
		return
	}

	hashed_password, err := auth.HashPassword(params.NewPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		log.Println(err)
		return
	}

	updatedUser, _ := cfg.database.UpdateUser(r.Context(),
		database.UpdateUserParams{
			ID:             userId,
			Email:          params.NewEmail,
			HashedPassword: hashed_password})

	respondWithJSON(w, http.StatusOK, UserIdentity{
		ID:        updatedUser.ID,
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		IsRed:     updatedUser.IsChirpyRed,
	})
}
