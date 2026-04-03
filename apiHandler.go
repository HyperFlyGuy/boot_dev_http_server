package main

import (
	"chirpy/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (cfg *apiConfig) GetChirpHandler(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "Invalid chirp ID")
	}
	chirp, err := cfg.dbQueries.GetChirp(context.Background(), chirpID)
	if err != nil {
		respondWithError(w, 404, "Error retrieving chirp from database")
		return
	}
	res := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		User_ID:   chirp.UserID,
	}
	respondWithJSON(w, 200, res)
}
func (cfg *apiConfig) GetChirpsHandler(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(context.Background())
	if err != nil {
		respondWithError(w, 500, "Error retrieving chirps from database")
		return
	}
	var res []Chirp
	for _, chirp := range chirps {
		c := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			User_ID:   chirp.UserID,
		}
		res = append(res, c)
	}

	respondWithJSON(w, 200, res)
}

func (cfg *apiConfig) CreateUserHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(req.Body)
	params := reqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request parameters")
		return
	}
	res, err := cfg.dbQueries.CreateUser(req.Context(), params.Email)
	if err != nil {
		fmt.Println(err)
	}
	user := User{
		ID:        res.ID,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
		Email:     res.Email,
	}
	respondWithJSON(w, 201, user)

}

func ReadinessHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) RequestsCountHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %v times!</p></body></html>", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) RequestsResetHandler(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Forbidden action")
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	cfg.dbQueries.ResetDatabase(context.Background())
	fmt.Fprintf(w, "Database has been reset")
}

func (cfg *apiConfig) CreateChirpHandler(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request parameters")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	params.Body = cleanResponse(params.Body)
	chirp, err := cfg.dbQueries.CreateChirp(context.Background(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprint(err))
		return
	}
	res := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		User_ID:   chirp.UserID,
	}
	respondWithJSON(w, 201, res)
}

func cleanResponse(body string) string {
	split_body := strings.Fields(body)
	banned_words := []string{"kerfuffle", "sharbert", "fornax"}
	for _, banned_word := range banned_words {
		for i, word := range split_body {
			if strings.ToLower(word) == banned_word {
				split_body[i] = "****"
			}
		}
	}
	res := strings.Join(split_body, " ")
	return res
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "error: %v", msg)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Write(dat)
}
