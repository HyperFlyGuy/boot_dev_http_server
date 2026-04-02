package main

import (
	"net/http"
)

func main() {
	cfg := apiConfig{}
	//Router
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", ReadinessHandler)
	mux.HandleFunc("POST /api/validate_chirp", ValidateChirpHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.RequestsCountHandler)
	mux.HandleFunc("POST /admin/reset", cfg.RequestsResetHandler)
	// Server Configuration
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Run the server
	server.ListenAndServe()
}
