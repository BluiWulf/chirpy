package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	resW.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resW http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(resW, req)
	})
}

func validateHandler(resW http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/api/validate_chirp" {
		http.NotFound(resW, req)
		return
	}

	type parameters struct {
		Body string `json:"body"`
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
	resp := returnVals{CleanedBody: profanityCheck(params.Body)}
	respondWithJson(resW, http.StatusOK, resp)
}
