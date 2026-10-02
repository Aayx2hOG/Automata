package handlers

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

func TestDatabaseUnavailableResponse(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"connection refused", &net.OpError{Op: "dial", Net: "tcp", Err: fmt.Errorf("connection refused")}, true},
		{"connection lost", fmt.Errorf("commit: %w", io.EOF), true},
		{"timeout", context.DeadlineExceeded, true},
		{"shutdown", &pgconn.PgError{Code: "57P01"}, true},
		{"recovery", &pgconn.PgError{Code: "57P03"}, true},
		{"connections exhausted", &pgconn.PgError{Code: "53300"}, true},
		{"constraint violation", &pgconn.PgError{Code: "23505"}, false},
		{"SQL bug", &pgconn.PgError{Code: "42601"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			if got := respondDatabaseUnavailable(w, zap.NewNop(), tc.err); got != tc.want {
				t.Fatalf("handled = %v, want %v", got, tc.want)
			}
			if tc.want && (w.Code != 503 || w.Header().Get("Retry-After") != "5") {
				t.Fatalf("unexpected outage response: %d %v", w.Code, w.Header())
			}
		})
	}
}
