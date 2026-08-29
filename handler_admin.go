package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func (cfg *apiConfig) getMetrics(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`<html>
		<body>
		  <h1>Welcome, Chirpy Admin</h1>
		  <p>Chirpy has been visited %d times!</p>
		</body>
	  </html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {

	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {

	if cfg.platform != "dev" {
		respondWithError(w, 403, "Forbidden")
		return
	}
	err := cfg.database.TruncateUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		log.Println(err)
		return
	}

}
