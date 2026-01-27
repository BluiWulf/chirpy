package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

const maxLen = 140
const rootPath = "."
const port = "8080"

var profanity = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	var apiCfg apiConfig

	mux := http.NewServeMux()
	server := http.Server{}
	server.Handler = mux
	server.Addr = ":" + port

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(rootPath))))
	mux.Handle("/app/", fsHandler)

	// API Endpoints
	mux.HandleFunc("GET /api/healthz", statusHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateHandler)

	// Admin Endpoints
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("error starting server: %v", err)
		return
	}
}
