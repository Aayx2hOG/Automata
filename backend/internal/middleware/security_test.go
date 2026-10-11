package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

func TestUntrustedForwardingHeadersCannotBypassLimit(t *testing.T) {
	t.Setenv("TEST_DISABLE_RATE_LIMIT", "0")
	limiter := NewRateLimiter(0.001, 1)
	h := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for i := 0; i < 3; i++ {
		r := httptest.NewRequest("POST", "/auth/login", nil)
		r.RemoteAddr = "198.51.100.10:1234"
		r.Header.Set("X-Forwarded-For", fmt.Sprintf("192.0.2.%d", i+1))
		r.Header.Set("X-Real-IP", fmt.Sprintf("192.0.2.%d", i+1))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		want := http.StatusTooManyRequests
		if i == 0 {
			want = http.StatusNoContent
		}
		if w.Code != want {
			t.Fatalf("request %d: %d, want %d", i, w.Code, want)
		}
	}
}

func TestTrustedProxyChain(t *testing.T) {
	proxies := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/24")}
	for _, tc := range []struct{ header, want string }{
		{"192.0.2.1, 198.51.100.20, 10.0.0.2", "198.51.100.20"},
		{"198.51.100.20", "198.51.100.20"},
		{"invalid", "10.0.0.1"},
	} {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "10.0.0.1:1234"
		r.Header.Set("X-Forwarded-For", tc.header)
		if got := clientIP(r, proxies); got != tc.want {
			t.Fatalf("%q: got %q, want %q", tc.header, got, tc.want)
		}
	}
}

func TestBodyLimitIncludesTrailingData(t *testing.T) {
	for _, body := range []string{`{}`, `{}` + strings.Repeat(" ", MaxRequestBytes)} {
		called := false
		h := LimitRequestBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(http.StatusNoContent) }))
		r := httptest.NewRequest("POST", "/workflows", strings.NewReader(body))
		r.ContentLength = -1 // Also cover chunked requests with no declared length.
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if len(body) > MaxRequestBytes {
			if called || w.Code != http.StatusRequestEntityTooLarge {
				t.Fatalf("oversized body: called=%v status=%d", called, w.Code)
			}
		} else if !called || w.Code != http.StatusNoContent {
			t.Fatal("valid body rejected")
		}
	}
}

func TestProductionCannotDisableRateLimit(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("TEST_DISABLE_RATE_LIMIT", "1")
	if NewRateLimiter(1, 1).rps != 1 {
		t.Fatal("production rate limiting disabled")
	}
}
