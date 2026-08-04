package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aayx2hOG/automata/internal/api"
	"go.uber.org/zap"
)

func TestHealthEndpoint_WithoutDB(t *testing.T) {
	logger := zap.NewNop()
	router := api.NewRouter(nil, logger, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("expected CORS header to echo allowed origin, got %q", got)
	}
}

func TestHealthEndpoint_RejectsDisallowedOrigin(t *testing.T) {
	logger := zap.NewNop()
	router := api.NewRouter(nil, logger, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodOptions, "/healthz", nil)
	req.Header.Set("Origin", "http://shit.shit.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "http://shit.shit.com" {
		t.Error("disallowed origin should not be echoed back in CORS header")
	}
}
