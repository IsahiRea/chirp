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

	"github.com/IsahiRea/chirp/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {

	type userVal struct {
		Email string `json:"email"`
	}

	type resVal struct {
		Id      uuid.UUID `json:"id"`
		Created time.Time `json:"created_at"`
		Updated time.Time `json:"updated_at"`
		Email   string    `json:"email"`
	}

	userReq := userVal{}
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), userReq.Email)
	if err != nil {
		log.Printf("Error finding user: %s", err)
		w.WriteHeader(500)
		return
	}

	resp := resVal{
		Id:      user.ID,
		Created: user.CreatedAt,
		Updated: user.UpdatedAt,
		Email:   user.Email,
	}

	data, err := json.Marshal(resp)
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

	type errors struct {
		ErrorMsg string `json:"error"`
	}

	type sendBack struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Clean     string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
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

	responseData := sendBack{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Clean:     chirp.Body,
		UserID:    chirp.UserID,
	}

	data, err := json.Marshal(&responseData)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
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

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to the database: %s", err)
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		dbQueries: dbQueries,
		platform:  platform,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("/app/assets/logo", http.StripPrefix("/app/", http.FileServer(http.Dir("logo.png"))))

	mux.HandleFunc("GET /api/healthz", readiness)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsers)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
