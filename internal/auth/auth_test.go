package auth

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestMakeJWT(t *testing.T) {

	godotenv.Load()
	tokenSecret := os.Getenv("TOKEN_STRING")

	id, err := uuid.Parse("bbbff1ab-2214-4f9a-a0a6-1789526c61ad")
	if err != nil {
		t.Fatalf("error parsing id: %s", err)
	}

	token, err := MakeJWT(id, tokenSecret, time.Duration(10))
	if err != nil {
		t.Fatalf("error making jwt: %s", err)
	}

	t.Log(token)
}

func TestValidateJWT(t *testing.T) {

	godotenv.Load()
	tokenSecret := os.Getenv("TOKEN_STRING")

	id, err := uuid.Parse("bbbff1ab-2214-4f9a-a0a6-1789526c61ad")
	if err != nil {
		t.Fatalf("error parsing id: %s", err)
	}

	duration, _ := time.ParseDuration("15s")

	token, err := MakeJWT(id, tokenSecret, duration)
	if err != nil {
		t.Fatalf("error making jwt: %s", err)
	}

	returnedID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("error validating jwt: %s", err)
	}

	t.Log(returnedID)
}

// Unit test for GetBearerToken function
func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		expectErr bool
	}{
		{
			name: "Valid token",
			headers: http.Header{
				"Authorization": {"Bearer some-valid-token"},
			},
			wantToken: "some-valid-token",
			expectErr: false,
		},
		{
			name:      "No Authorization header",
			headers:   http.Header{},
			wantToken: "",
			expectErr: true,
		},
		{
			name: "Empty Authorization header",
			headers: http.Header{
				"Authorization": {""},
			},
			wantToken: "",
			expectErr: true,
		},
		{
			name: "Invalid Authorization format",
			headers: http.Header{
				"Authorization": {"InvalidTokenFormat"},
			},
			wantToken: "",
			expectErr: true,
		},
		{
			name: "Empty token after Bearer",
			headers: http.Header{
				"Authorization": {"Bearer "},
			},
			wantToken: "",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)

			if (err != nil) != tt.expectErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.expectErr)
				return
			}

			if token != tt.wantToken {
				t.Errorf("GetBearerToken() = %v, want %v", token, tt.wantToken)
			}
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		expectErr bool
	}{
		{
			name: "Valid token",
			headers: http.Header{
				"Authorization": {"ApiKey some-valid-token"},
			},
			wantToken: "some-valid-token",
			expectErr: false,
		},
		{
			name:      "No Authorization header",
			headers:   http.Header{},
			wantToken: "",
			expectErr: true,
		},
		{
			name: "Empty Authorization header",
			headers: http.Header{
				"Authorization": {""},
			},
			wantToken: "",
			expectErr: true,
		},
		{
			name: "Invalid Authorization format",
			headers: http.Header{
				"Authorization": {"InvalidTokenFormat"},
			},
			wantToken: "",
			expectErr: true,
		},
		{
			name: "Empty token after Bearer",
			headers: http.Header{
				"Authorization": {"ApiKey "},
			},
			wantToken: "",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetAPIKey(tt.headers)

			if (err != nil) != tt.expectErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.expectErr)
				return
			}

			if token != tt.wantToken {
				t.Errorf("GetBearerToken() = %v, want %v", token, tt.wantToken)
			}
		})
	}
}
