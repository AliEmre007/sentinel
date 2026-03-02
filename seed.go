//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	fmt.Println("🚀 Starting massive data injection using PostgreSQL COPY protocol...")
	startTime := time.Now()

	txn, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("urls", "short_code", "original_url"))
	if err != nil {
		log.Fatal(err)
	}

	const numRows = 1000000

	for i := 0; i < numRows; i++ {
		uniqueCode := fmt.Sprintf("BLK%06d", i)
		
		_, err := stmt.Exec(uniqueCode, fmt.Sprintf("https://sakarya.edu.tr/page/%d", i))
		if err != nil {
			log.Fatalf("Buffer error at row %d: %v", i, err)
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = txn.Commit()
	if err != nil {
		log.Fatal(err)
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Successfully injected %d rows in %v!\n", numRows, duration)
}
