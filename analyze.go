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

	// We are intentionally searching for the very last URL we inserted, 
	// using a column that has NO index.
	query := `EXPLAIN ANALYZE SELECT * FROM urls WHERE original_url = 'https://sakarya.edu.tr/page/999999';`

	fmt.Println("🔍 Analyzing Query Execution Plan...")
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("\n--- 📄 PostgreSQL Receipt ---")
	for rows.Next() {
		var planLine string
		if err := rows.Scan(&planLine); err != nil {
			log.Fatal(err)
		}
		fmt.Println(planLine)
	}
}
