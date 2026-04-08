package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashed_password, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashed_password, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	valid, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return valid, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	api_string := headers.Get("Authorization")
	if api_string == "" {
		return "", fmt.Errorf("Authorization Header does not exist")
	}
	res := strings.TrimPrefix(api_string, "ApiKey ")
	return res, nil
}
