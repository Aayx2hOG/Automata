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
