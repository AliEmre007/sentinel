//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	fmt.Println("🏗️ Building B-Tree Index on original_url... (This takes a few seconds for 1M rows)")
	
	// Create the B-Tree Index
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);`)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}
	fmt.Println("✅ B-Tree Index built successfully!")

	// Re-run the autopsy
	query := `EXPLAIN ANALYZE SELECT * FROM urls WHERE original_url = 'https://sakarya.edu.tr/page/999999';`

	fmt.Println("\n🔍 Re-running Query Execution Plan...")
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("\n--- 📄 NEW PostgreSQL Receipt ---")
	for rows.Next() {
		var planLine string
		if err := rows.Scan(&planLine); err != nil {
			log.Fatal(err)
		}
		fmt.Println(planLine)
	}
}
