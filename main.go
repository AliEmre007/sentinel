package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var db *sql.DB
var rdb *redis.Client
var ctx = context.Background() // Required by Redis for managing connection timeouts

type ShortenRequest struct {
	OriginalURL string `json:"original_url"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
}

// ---------------------------------------------------------
// INFRASTRUCTURE INITIALIZATION
// ---------------------------------------------------------

func initDB() {
	connStr := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_code VARCHAR(10) UNIQUE NOT NULL,
		original_url TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	log.Println("Connected to PostgreSQL successfully!")
}

func initRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	
	rdb = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	
	// Test the connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully!")
}

// ---------------------------------------------------------
// BUSINESS LOGIC
// ---------------------------------------------------------

func generateShortCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sentinel API is Live with Redis! 🚀\n"))
}

func handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	shortCode := generateShortCode(6)

	_, err := db.Exec("INSERT INTO urls (short_code, original_url) VALUES ($1, $2)", shortCode, req.OriginalURL)
	if err != nil {
		http.Error(w, "Failed to save to database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ShortenResponse{ShortCode: shortCode})
}

// THE NEW CACHE-ASIDE REDIRECT HANDLER
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]

	// 1. Check Redis First (Memory is fast)
	cachedURL, err := rdb.Get(ctx, shortCode).Result()
	if err == nil {
		log.Printf("⚡ CACHE HIT for %s", shortCode)
		http.Redirect(w, r, cachedURL, http.StatusFound)
		return
	}

	// 2. Cache Miss! Query PostgreSQL (Disk is slow)
	log.Printf("🐢 CACHE MISS for %s. Querying PostgreSQL...", shortCode)
	var originalURL string
	err = db.QueryRow("SELECT original_url FROM urls WHERE short_code = $1", shortCode).Scan(&originalURL)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "URL not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// 3. Save to Redis for the next person (Expires in 24 hours)
	err = rdb.Set(ctx, shortCode, originalURL, 24*time.Hour).Err()
	if err != nil {
		log.Printf("Warning: Failed to cache %s: %v", shortCode, err)
	}

	// 4. Redirect the user
	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	initDB()
	initRedis() // Boot up the cache

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/shorten", handleShorten)
	http.HandleFunc("/", handleRedirect)

	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
