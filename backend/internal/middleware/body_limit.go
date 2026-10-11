package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

const MaxRequestBytes = 1 << 20

// Read the complete bounded body before dispatch. This also rejects oversized
// trailing data that a single JSON Decode would otherwise leave unread.
func LimitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.Body == http.NoBody {
			next.ServeHTTP(w, r)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxRequestBytes))
		r.Body.Close()
		if err != nil {
			var limit *http.MaxBytesError
			status, message := http.StatusBadRequest, `{"error":"unable to read request body"}`
			if errors.As(err, &limit) {
				status, message = http.StatusRequestEntityTooLarge, `{"error":"request body exceeds 1 MiB"}`
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(message))
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}
