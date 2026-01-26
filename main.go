package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func muxHandler(resW http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/healthz" {
		http.NotFound(resW, req)
		return
	}
	resW.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resW.WriteHeader(http.StatusOK)
	resW.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(resW http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/metrics" {
		http.NotFound(resW, req)
		return
	}
	count := cfg.fileserverHits.Load()
	resW.WriteHeader(http.StatusOK)
	resW.Write([]byte(fmt.Sprintf("Hits: %v", count)))
}

func (cfg *apiConfig) resetHandler(resW http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/reset" {
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

func main() {
	var apiCfg apiConfig

	mux := http.NewServeMux()
	server := http.Server{}

	server.Handler = mux
	server.Addr = ":8080"

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("/healthz", muxHandler)
	mux.HandleFunc("/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("/reset", apiCfg.resetHandler)

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("error starting server: %v", err)
		return
	}
}
