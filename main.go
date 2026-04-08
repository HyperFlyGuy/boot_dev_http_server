package main

import (
	"chirpy/internal/database"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load(".env")
	db_url := os.Getenv("DB_URL")
	server_secret := os.Getenv("SECRET")
	polka_key := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", db_url)
	if err != nil {
		fmt.Printf("Error connecting to database: %v", err)
	}
	platform := os.Getenv("PLATFORM")
	cfg := apiConfig{
		dbQueries: database.New(db),
		platform:  platform,
		secret:    server_secret,
		polkakey:  polka_key,
	}
	//Router
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", ReadinessHandler)
	mux.HandleFunc("POST /api/chirps", cfg.CreateChirpHandler)
	mux.HandleFunc("GET /api/chirps", cfg.GetChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.GetChirpHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.DeleteChirpHandler)
	mux.HandleFunc("POST /api/users", cfg.CreateUserHandler)
	mux.HandleFunc("PUT /api/users", cfg.UpdateUserPasswordHandler)
	mux.HandleFunc("POST /api/revoke", cfg.RevokeTokenHandler)
	mux.HandleFunc("POST /api/refresh", cfg.RefreshTokenHandler)
	mux.HandleFunc("POST /api/login", cfg.LoginUserHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.RequestsCountHandler)
	mux.HandleFunc("POST /admin/reset", cfg.RequestsResetHandler)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.ChirpyRedUpgradeHandler)
	// Server Configuration
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Run the server
	server.ListenAndServe()
}
