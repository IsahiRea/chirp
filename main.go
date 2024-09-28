package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	cfg.fileserverHits.Add(1)

	return next
}

func (cfg *apiConfig) hits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)

	output := "Hits: " + strconv.Itoa(int(cfg.fileserverHits.Load()))

	result := []byte(output)
	w.Write(result)
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)

	cfg.fileserverHits.Store(0)
}

func readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)

	result := []byte("ok")
	w.Write(result)
}

func main() {
	mux := http.NewServeMux()

	apiCfg := apiConfig{}

	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("/app/assets/logo", http.StripPrefix("/app/", http.FileServer(http.Dir("logo.png"))))
	mux.HandleFunc("/healthz", readiness)
	mux.HandleFunc("/metrics", apiCfg.hits)
	mux.HandleFunc("/reset", apiCfg.reset)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
