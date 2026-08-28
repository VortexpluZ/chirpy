package main

import (
	"encoding/json"
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
}

func (cfg *apiConfig) createUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}

func (cfg *apiConfig) login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			log.Println(err)
			return
		}

		respondWithJSON(w, 200, User{
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			ID:        user.ID,
			Email:     user.Email,
		})
	})
}
