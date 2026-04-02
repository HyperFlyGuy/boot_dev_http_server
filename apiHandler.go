package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

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
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits.Load())
}

func ValidateChirpHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Body string `json:"body"`
	}
	type CleanedResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}
	decoder := json.NewDecoder(req.Body)
	params := reqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request parameters")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	field := cleanResponse(params.Body)
	res := CleanedResponse{
		CleanedBody: field,
	}
	respondWithJSON(w, 200, res)
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
