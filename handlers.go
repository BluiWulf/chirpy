package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func statusHandler(resW http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/api/healthz" {
		http.NotFound(resW, req)
		return
	}
	resW.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resW.WriteHeader(http.StatusOK)
	resW.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(resW http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/admin/metrics" {
		http.NotFound(resW, req)
		return
	}
	count := cfg.fileserverHits.Load()
	resW.Header().Set("Content-Type", "text/html")
	resW.WriteHeader(http.StatusOK)
	resW.Write([]byte(fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", count)))
}

func (cfg *apiConfig) resetHandler(resW http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/admin/reset" {
		http.NotFound(resW, req)
		return
	}
	if cfg.platform != "dev" {
		respondWithError(resW, http.StatusForbidden, "Unauthorized Request")
		return
	}
	resW.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)

	err := cfg.dbQueries.ResetUsers(req.Context())
	if err != nil {
		fmt.Printf("Error reseting users database: %s", err)
		respondWithError(resW, http.StatusInternalServerError, "Unable to reset users database")
		return
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resW http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(resW, req)
	})
}

func (cfg *apiConfig) chirpHandler(resW http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/api/chirps" {
		http.NotFound(resW, req)
		return
	}

	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error decoding JSON: %s", err)
		respondWithError(resW, http.StatusBadRequest, "Invalid request")
		return
	}
	if len(params.Body) > maxLen {
		respondWithError(resW, http.StatusBadRequest, "Chirp is too long")
		return
	}
	dbParams := database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserID,
	}
	chirp, err := cfg.dbQueries.CreateChirp(req.Context(), dbParams)
	if err != nil {
		fmt.Printf("Error adding new chirp to database: %s", err)
		respondWithError(resW, http.StatusInternalServerError, "Unable to add new chirp to database")
		return
	}

	resp := Chirp{
		ID:        chirp.ID.UUID,
		CreatedAt: chirp.CreatedAt.Time,
		UpdatedAt: chirp.UpdatedAt.Time,
		Body:      profanityCheck(chirp.Body),
		UserID:    chirp.UserID,
	}
	respondWithJson(resW, http.StatusCreated, resp)
}

func (cfg *apiConfig) usersHandler(resW http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/api/users" {
		http.NotFound(resW, req)
		return
	}

	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error decoding JSON: %s", err)
		respondWithError(resW, http.StatusBadRequest, "Invalid request")
		return
	}

	user, err := cfg.dbQueries.CreateUser(req.Context(), params.Email)
	if err != nil {
		fmt.Printf("Error adding new user to database: %s", err)
		respondWithError(resW, http.StatusInternalServerError, "Unable to add new user to database")
		return
	}
	resp := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email,
	}
	respondWithJson(resW, http.StatusCreated, resp)
}
