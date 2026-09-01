package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aayx2hOG/automata/internal/middleware"
)

func TestRateLimiter_EnforcesLimit(t *testing.T) {
	// 1 rps, burst of 2
	limiter := middleware.NewRateLimiter(1, 2)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 2 requests should succeed (burst of 2)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for initial burst request %d, got %d", i+1, rr.Code)
		}
	}

	// Next request exceeds burst and should return 429
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got %d", rr.Code)
	}
}
