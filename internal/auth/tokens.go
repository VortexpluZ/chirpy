package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Missing Authorization header")
	}
	token := strings.Split(authHeader, "Bearer ")
	if len(token) != 2 {
		return "", errors.New("Invalid Authorization header")
	}
	return token[1], nil
}
