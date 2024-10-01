package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/IsahiRea/chirp/internal/auth"
	"github.com/IsahiRea/chirp/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	tokenSecret    string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {

	type recieve struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	userReq := recieve{}
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	hashedPassword, err := auth.HashPassword(userReq.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(500)
		return
	}

	requestDataSend := database.CreateUserParams{
		Email:          userReq.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), requestDataSend)
	if err != nil {
		log.Printf("Error finding user: %s", err)
		w.WriteHeader(500)
		return
	}

	sendBack := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
	}

	data, err := json.Marshal(sendBack)
	if err != nil {
		log.Printf("Error during marshal: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error obtaining token: %s", err)
		w.WriteHeader(500)
		return
	}

	id, err := auth.ValidateJWT(tokenString, cfg.tokenSecret)
	if err != nil {
		log.Printf("Error validating token: %s", err)
		w.WriteHeader(401)
		return
	}

	type errors struct {
		ErrorMsg string `json:"error"`
	}

	type recieve struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	requestData := recieve{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {

		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if requestData.UserID != id {
		log.Printf("Error Unauthorized: %s", err)
		w.WriteHeader(401)
		return
	}

	if len(requestData.Body) > 140 {

		errorsResp := errors{ErrorMsg: "Chirp is too long"}
		data, err := json.Marshal(&errorsResp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(data)
		return
	}

	// Profane
	bannedWords := []string{"kerfuffle", "sharbert", "fornax", "Kerfuffle", "Sharbert", "Fornax"}

	for _, word := range bannedWords {
		requestData.Body = strings.ReplaceAll(requestData.Body, word, "****")
	}

	requestDataSend := database.CreateChirpParams{
		Body:   requestData.Body,
		UserID: requestData.UserID,
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), requestDataSend)
	if err != nil {
		log.Printf("Error finding chirp: %s", err)
		w.WriteHeader(500)
		return
	}

	data, err := json.Marshal(&chirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)

}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {

	allChirps, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error obtaining chirps: %s", err)
		w.WriteHeader(500)
		return
	}

	data, err := json.Marshal(allChirps)
	if err != nil {
		log.Printf("Error Marshalling all chirps: %s", err)
		w.WriteHeader(500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)

}

func (cfg *apiConfig) handlerGetChirpID(w http.ResponseWriter, r *http.Request) {
	uuidString := r.PathValue("chirpID")

	id, err := uuid.Parse(uuidString)
	if err != nil {
		log.Println("Invalid resource")
		w.WriteHeader(404)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), id)
	if err != nil {
		log.Printf("Error finding chirp by ID: %s", err)
		w.WriteHeader(500)
		return
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error Marshalling chirp by ID: %s", err)
		w.WriteHeader(500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)

	template := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`, int(cfg.fileserverHits.Load()))

	output := []byte(template)
	w.Write(output)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {

	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}

	if err := cfg.dbQueries.DeleteUsers(r.Context()); err != nil {
		log.Printf("Error Deleting users: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)

	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {

	type recieve struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	userReq := recieve{}
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.dbQueries.GetHashPassByEmail(r.Context(), userReq.Email)

	if err != nil {
		log.Printf("Error finding user: %s", err)
		w.WriteHeader(500)
		return
	}

	if err := auth.CheckPasswordHash(userReq.Password, user.HashedPassword); err != nil {
		log.Printf("Error email or password: %s", err)
		w.WriteHeader(401)
		return
	}

	// Create Tokens

	timeDurationJWT, err := time.ParseDuration("1h")
	if err != nil {
		log.Printf("Error parsing time duration: %s", err)
		w.WriteHeader(500)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, timeDurationJWT)
	if err != nil {
		log.Printf("Error creating JWT: %s", err)
		w.WriteHeader(500)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating refresh token: %s", err)
		w.WriteHeader(500)
		return
	}

	tokenReq := database.CreateRefeshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}

	if err := cfg.dbQueries.CreateRefeshToken(r.Context(), tokenReq); err != nil {
		log.Printf("Error saving refresh token: %s", err)
		w.WriteHeader(500)
		return
	}

	sendBack := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
		RToken    string    `json:"refresh_token"`
	}{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
		token,
		refreshToken,
	}

	data, err := json.Marshal(sendBack)
	if err != nil {
		log.Printf("Error during marshal: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error obtaining refresh token: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.dbQueries.GetUserFromRToken(r.Context(), token)
	if err != nil {
		log.Printf("Error saving refresh token: %s", err)
		w.WriteHeader(401)
		return
	}

	if time.Now().After(user.ExpiresAt) || user.RevokedAt.Valid {
		log.Printf("Error token expired: %s", err)
		w.WriteHeader(401)
		return
	}

	timeDurationJWT, err := time.ParseDuration("1h")
	if err != nil {
		log.Printf("Error parsing time duration: %s", err)
		w.WriteHeader(500)
		return
	}

	newAccessToken, err := auth.MakeJWT(user.UserID, cfg.tokenSecret, timeDurationJWT)
	if err != nil {
		log.Printf("Error creaing JWT: %s", err)
		w.WriteHeader(500)
		return
	}

	sendBack := struct {
		Token string `json:"token"`
	}{
		newAccessToken,
	}

	data, err := json.Marshal(sendBack)
	if err != nil {
		log.Printf("Error during marshal: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error obtaining refresh token: %s", err)
		w.WriteHeader(500)
		return
	}

	if err := cfg.dbQueries.RevokeRefreshToken(r.Context(), token); err != nil {
		log.Printf("Error revoking refresh token: %s", err)
		w.WriteHeader(401)
		return
	}

	w.WriteHeader(204)
}

func readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)

	result := []byte("ok")
	w.Write(result)
}

func main() {

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("TOKEN_STRING")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to the database: %s", err)
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		dbQueries:   dbQueries,
		platform:    platform,
		tokenSecret: tokenSecret,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("/app/assets/logo", http.StripPrefix("/app/", http.FileServer(http.Dir("logo.png"))))

	mux.HandleFunc("GET /api/healthz", readiness)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsers)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpID)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
