package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var db *sql.DB
var rdb *redis.Client
var ctx = context.Background()

// ---------------------------------------------------------
// 🛡️ THE LUA SCRIPT (Token Bucket Algorithm)
// ---------------------------------------------------------
const tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local bucket = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(bucket[1])
local last_refill = tonumber(bucket[2])

if not tokens then
    tokens = capacity
    last_refill = now
else
    local time_passed = now - last_refill
    local new_tokens = math.floor(time_passed * refill_rate)
    tokens = math.min(capacity, tokens + new_tokens)
    if new_tokens > 0 then
        last_refill = now
    end
end

if tokens >= 1 then
    redis.call("HMSET", key, "tokens", tokens - 1, "last_refill", last_refill)
    redis.call("EXPIRE", key, 60)
    return 1
else
    return 0
end
`

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
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully!")
}

func generateShortCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// ---------------------------------------------------------
// BUSINESS LOGIC
// ---------------------------------------------------------
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sentinel API is Live with Redis and Rate Limiting! 🚀\n"))
}

type ShortenRequest struct {
	OriginalURL string `json:"original_url"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
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

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]

	cachedURL, err := rdb.Get(ctx, shortCode).Result()
	if err == nil {
		log.Printf("⚡ CACHE HIT for %s", shortCode)
		http.Redirect(w, r, cachedURL, http.StatusFound)
		return
	}

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

	err = rdb.Set(ctx, shortCode, originalURL, 24*time.Hour).Err()
	if err != nil {
		log.Printf("Warning: Failed to cache %s: %v", shortCode, err)
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// ---------------------------------------------------------
// 🛡️ THE MIDDLEWARE (The Bouncer)
// ---------------------------------------------------------
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		
		bucketKey := "rate_limit:" + ip
		capacity := 5      // Maximum burst of 5 requests
		refillRate := 1    // Refill 1 token per second
		now := time.Now().Unix()

		result, err := rdb.Eval(ctx, tokenBucketScript, []string{bucketKey}, capacity, refillRate, now).Result()
		if err != nil {
			log.Printf("Rate limiter error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if result.(int64) == 0 {
			log.Printf("🛑 BLOCKED: %s is making too many requests!", ip)
			http.Error(w, "429 Too Many Requests - Slow down!", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func main() {
	initDB()
	initRedis()

	http.HandleFunc("/health", handleHealth)
	
	// We cast handleShorten to an http.HandlerFunc so the Go compiler accepts it into the middleware
	http.HandleFunc("/shorten", rateLimitMiddleware(http.HandlerFunc(handleShorten)))
	
	http.HandleFunc("/", handleRedirect)

	log.Println("Sentinel Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
