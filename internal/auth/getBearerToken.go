package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no authorization")
	}

	tokenBearer := strings.Split(authHeader, " ")[0]
	if tokenBearer != "Bearer" {
		return "", fmt.Errorf("invalid format")
	}

	tokenString := strings.Split(authHeader, " ")[1]

	return tokenString, nil
}
