package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateShortCode(t *testing.T) {
	code1 := generateShortCode(6)
	code2 := generateShortCode(6)
	if len(code1) != 6 || code1 == code2 {
		t.Errorf("Short code logic failed")
	}
}

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handleHealth(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}
}
