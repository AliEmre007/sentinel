package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Unit Test: Tests the pure business logic
func TestGenerateShortCode(t *testing.T) {
	expectedLength := 99
	code1 := generateShortCode(expectedLength)
	code2 := generateShortCode(expectedLength)

	if len(code1) != expectedLength || len(code2) != expectedLength {
		t.Errorf("Failed length check. Got %d and %d", len(code1), len(code2))
	}
	if code1 == code2 {
		t.Errorf("Collision detected! Both codes were: %s", code1)
	}
}

// Integration Test: Tests the HTTP Request/Response cycle
func TestHealthEndpoint(t *testing.T) {
	// 1. Create a simulated HTTP GET request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	// 2. Create a ResponseRecorder to act as our "fake web browser" to capture the output
	rr := httptest.NewRecorder()

	// 3. Call the handler directly, passing in our fake browser and fake request
	handleHealth(rr, req)

	// 4. Assert the HTTP Status Code is exactly 200 (OK)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// 5. Assert the API returned the correct body text
	expected := "Sentinel API is Live Reloading! 🚀"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
