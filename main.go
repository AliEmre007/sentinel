package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

// db holds the global database connection pool.
var db *sql.DB

// ---------------------------------------------------------
// DATA MODELS
// ---------------------------------------------------------
type ShortenRequest struct {
	OriginalURL string `json:"original_url"`
}

type ShortenResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
}

// ---------------------------------------------------------
// CORE LOGIC FUNCTIONS
// ---------------------------------------------------------

// generateShortCode creates a random alphanumeric string.
func generateShortCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))] 
	}
	return string(b)
}

// handleShorten processes POST requests to create new short links.
func handleShorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OriginalURL == "" {
		http.Error(w, "Invalid request payload. 'original_url' is required.", http.StatusBadRequest)
		return
	}

	shortCode := generateShortCode(6)

	// BEST PRACTICE: Parameterized query to prevent SQL injection.
	insertSQL := `INSERT INTO urls (short_code, original_url) VALUES ($1, $2)`
	_, err := db.Exec(insertSQL, shortCode, req.OriginalURL)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{
		ShortCode:   shortCode,
		OriginalURL: req.OriginalURL,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleRedirect processes GET requests and forwards the user to the original URL.
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	// BEST PRACTICE: Modern Go 1.22 routing safely extracts the variable from the URL.
	shortCode := r.PathValue("code")
	if shortCode == "" {
		http.Error(w, "Short code is missing", http.StatusBadRequest)
		return
	}

	var originalURL string
	querySQL := `SELECT original_url FROM urls WHERE short_code = $1`
	
	err := db.QueryRow(querySQL, shortCode).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			// If the code doesn't exist in the DB, return our custom 404.
			http.Error(w, "404 - URL not found in Sentinel", http.StatusNotFound)
			return
		}
		log.Printf("Database error during redirect lookup: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// BEST PRACTICE: Perform the physical HTTP 302 Redirect
	http.Redirect(w, r, originalURL, http.StatusFound)
}

// handleHealth provides a simple status check for the API.
func handleHealth(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Sentinel API is Live Reloading! 🚀")
}

// ---------------------------------------------------------
// INITIALIZATION & ROUTING
// ---------------------------------------------------------

// initDB initializes the PostgreSQL connection and creates the table if missing.
func initDB() {
	var err error
	
	// BEST PRACTICE: Read secrets from the environment, never from the code.
	connStr := os.Getenv("DATABASE_URL")
	
	// BEST PRACTICE: Fail-Fast. If the DevOps engineer forgot to set the .env file, 
	// crash immediately with a clear error message.
	if connStr == "" {
		log.Fatal("CRITICAL: DATABASE_URL environment variable is not set!")
	}
	
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to allocate DB pool: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Database is unreachable: ", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_code VARCHAR(10) UNIQUE NOT NULL,
		original_url TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Failed to create 'urls' table: ", err)
	}
	fmt.Println("Database connected & 'urls' table verified! 🟢")
}

func main() {
	initDB()
	defer db.Close()

	// BEST PRACTICE: Exactly one handler per route pattern
	http.HandleFunc("GET /health", handleHealth)
	http.HandleFunc("POST /shorten", handleShorten)
	http.HandleFunc("GET /{code}", handleRedirect)

	fmt.Println("Sentinel Server starting on port 8080...")
	
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
