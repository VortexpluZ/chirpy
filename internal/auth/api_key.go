package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Missing Authorization header")
	}
	token := strings.Split(authHeader, "ApiKey ")
	if len(token) != 2 {
		return "", errors.New("Invalid ApiKey header")
	}
	return token[1], nil
}
