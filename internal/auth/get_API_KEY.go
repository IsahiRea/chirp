package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no authorization")
	}

	tokenBearer := strings.Split(authHeader, " ")[0]
	if tokenBearer != "ApiKey" {
		return "", fmt.Errorf("invalid format")
	}

	key := strings.Split(authHeader, " ")[1]

	return key, nil
}
