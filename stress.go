package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	const numRequests = 1000
	const targetURL = "http://localhost:8080/shorten"
	payload := []byte(`{"original_url": "https://sakarya.edu.tr"}`)

	// The Conductor
	var wg sync.WaitGroup

	// Thread-safe counters (Best Practice)
	var successCount int32
	var rateLimitedCount int32
	var errorCount int32

	fmt.Printf("🚀 Launching %d concurrent Goroutines against %s...\n", numRequests, targetURL)
	
	// Start the stopwatch
	startTime := time.Now()

	// Spin up 1,000 Goroutines instantly
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		
		go func() {
			defer wg.Done() // Tell the WaitGroup this routine is finished when it exits
			
			req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(payload))
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				return
			}
			defer resp.Body.Close()

			// Tally the results atomically
			if resp.StatusCode == 200 {
				atomic.AddInt32(&successCount, 1)
			} else if resp.StatusCode == 429 {
				atomic.AddInt32(&rateLimitedCount, 1)
			} else {
				atomic.AddInt32(&errorCount, 1)
			}
		}()
	}

	// Block the main thread until all 1,000 Goroutines are done
	wg.Wait()
	
	// Stop the stopwatch
	duration := time.Since(startTime)

	// Print the Autopsy
	fmt.Println("\n--- 📊 Load Test Results ---")
	fmt.Printf("Total Time:         %v\n", duration)
	fmt.Printf("✅ Success (200):   %d\n", successCount)
	fmt.Printf("🛑 Blocked (429):   %d\n", rateLimitedCount)
	fmt.Printf("❌ Errors (500+):   %d\n", errorCount)
}
