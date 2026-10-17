package api

import (
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestCORSRequiresConfiguredOrigin(t *testing.T) {
	for _, tc := range []struct {
		origins      []string
		origin, want string
	}{
		{nil, "https://untrusted.example", ""},
		{[]string{"https://app.example"}, "https://app.example", "https://app.example"},
		{[]string{"https://app.example"}, "https://untrusted.example", ""},
	} {
		router := NewRouter(nil, zap.NewNop(), tc.origins, nil, nil, nil, nil, nil, nil)
		r := httptest.NewRequest("GET", "/healthz", nil)
		r.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != tc.want {
			t.Fatalf("origins=%v request=%q: allowed %q", tc.origins, tc.origin, got)
		}
	}
}

func TestWorkflowDeleteRequiresAuthentication(t *testing.T) {
	router := NewRouter(nil, zap.NewNop(), nil, nil, nil, nil, nil, nil, nil)
	request := httptest.NewRequest("DELETE", "/workflows/11111111-1111-4111-8111-111111111111", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 401 {
		t.Fatalf("unauthenticated delete returned %d", response.Code)
	}
}
