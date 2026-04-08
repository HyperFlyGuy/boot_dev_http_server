package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) ChirpyRedUpgradeHandler(w http.ResponseWriter, req *http.Request) {
	type data_shape struct {
		UserID uuid.UUID `json:"user_id"`
	}
	type reqParams struct {
		Event string     `json:"event"`
		Data  data_shape `json:"data"`
	}
	decoder := json.NewDecoder(req.Body)
	params := reqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request parameters")
		return
	}
	//check for API Key
	key, err := auth.GetAPIKey(req.Header)
	if err != nil {
		respondWithError(w, 401, "Error fetching API Key")
		return
	}
	if key != cfg.polkakey {
		respondWithError(w, 401, "Invalid API Key")
		return
	}

	//check for valid event
	if params.Event == "user.upgraded" {
		err := cfg.dbQueries.UpgradeChirpyRed(context.Background(), params.Data.UserID)
		if err != nil {
			respondWithError(w, 404, "User not found")
			return
		}
		respondWithJSON(w, 204, "")
	}
	respondWithJSON(w, 204, "")
}

func (cfg *apiConfig) DeleteChirpHandler(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "Invalid chirp ID")
	}
	//Validate the access token
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, 401, "Error fetching bearer token")
		return
	}

	user_id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Error validating bearer token")
		return
	}
	// Check if chirp exists
	chirp, err := cfg.dbQueries.GetChirp(context.Background(), chirpID)
	if err != nil {
		respondWithError(w, 404, "Chirp was not found:")
		return
	}
	// Validate user info
	if chirp.UserID == user_id {
		err = cfg.dbQueries.DeleteChirp(context.Background(), chirpID)
		type res struct {
			Body string
		}
		respondWithJSON(w, 204, res{
			Body: "Chirp has been deleted",
		})
	} else {
		respondWithError(w, 403, "Unauthorized to delete chirp")
		return
	}
}

func (cfg *apiConfig) UpdateUserPasswordHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	//Validate the access token
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, 401, "Error fetching bearer token")
		return
	}

	user_id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Error validating bearer token")
		return
	}
	//Decode the request
	decoder := json.NewDecoder(req.Body)
	params := reqParams{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 401, "Error decoding request parameters")
		return
	}
	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		fmt.Println(err)
	}
	//update the user
	res, err := cfg.dbQueries.UpdateUser(context.Background(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashed_password,
		ID:             user_id,
	})
	if err != nil {
		fmt.Println(err)
	}
	user := User{
		ID:        res.ID,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
		Email:     res.Email,
	}
	respondWithJSON(w, 200, user)

}

func (cfg *apiConfig) RevokeTokenHandler(w http.ResponseWriter, req *http.Request) {
	refresh_token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	cfg.dbQueries.TokenRevokeUpdate(context.Background(), refresh_token)
	respondWithJSON(w, 204, "")
}

func (cfg *apiConfig) RefreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	refresh_token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	token_res, err := cfg.dbQueries.GetUserFromToken(context.Background(), refresh_token)
	if err != nil {
		respondWithError(w, 401, "Error fetching token")
	}
	if token_res.RevokedAt.Valid {
		respondWithError(w, 401, "Token has been revoked")
	}
	if token_res.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, "Token has expired")
	}
	access_token, err := auth.MakeJWT(token_res.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, 401, "Unable to create a new token")
		return
	}
	type Response struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, 200, Response{
		Token: access_token,
	})
}

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

func (cfg *apiConfig) LoginUserHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(req.Body)
	params := reqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request parameters")
		return
	}
	//Password Handling
	user_lookup, err := cfg.dbQueries.UserPasswordLookup(context.Background(), params.Email)
	if err != nil {
		fmt.Println(err)
		return
	}
	valid, err := auth.CheckPasswordHash(params.Password, user_lookup.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "Error comparing passwords")
		return
	}
	if valid == false {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	access_token, err := auth.MakeJWT(user_lookup.ID, cfg.secret, time.Duration(3600)*time.Second)
	if err != nil {
		respondWithError(w, 400, "Error creating token")
		return
	}
	refresh_token := auth.MakeRefreshToken()
	cfg.dbQueries.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:  refresh_token,
		UserID: user_lookup.ID,
	})

	type response struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}
	respondWithJSON(w, 200, response{
		ID:           user_lookup.ID,
		CreatedAt:    user_lookup.CreatedAt,
		UpdatedAt:    user_lookup.UpdatedAt,
		Email:        user_lookup.Email,
		IsChirpyRed:  user_lookup.IsChirpyRed,
		Token:        access_token,
		RefreshToken: refresh_token,
	})
}

func (cfg *apiConfig) CreateUserHandler(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(req.Body)
	params := reqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding request parameters")
		return
	}
	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		fmt.Println(err)
	}
	res, err := cfg.dbQueries.CreateUser(req.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashed_password,
	})
	if err != nil {
		fmt.Println(err)
	}
	user := User{
		ID:             res.ID,
		CreatedAt:      res.CreatedAt,
		UpdatedAt:      res.UpdatedAt,
		Email:          res.Email,
		HashedPassword: res.HashedPassword,
		IsChirpyRed:    res.IsChirpyRed,
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
		Body string `json:"body"`
	}
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, 401, "Error fetching bearer token")
		return
	}

	user_id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Error validating bearer token")
		return
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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
		UserID: user_id,
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
