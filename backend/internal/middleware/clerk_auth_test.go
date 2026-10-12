package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
)

type testClerkAuth struct {
	user  *models.User
	err   error
	calls int
}

func (a *testClerkAuth) Authenticate(context.Context, string) (*models.User, error) {
	a.calls++
	return a.user, a.err
}

func TestClerkAuthentication(t *testing.T) {
	id := uuid.New()
	for _, tc := range []struct {
		name, header  string
		user          *models.User
		err           error
		status, calls int
	}{
		{"missing", "", nil, nil, 401, 0},
		{"wrong scheme", "Basic token", nil, nil, 401, 0},
		{"extra fields", "Bearer token extra", nil, nil, 401, 0},
		{"invalid token", "Bearer bad", nil, errors.New("invalid"), 401, 1},
		{"disabled user", "Bearer token", &models.User{ID: id}, nil, 401, 1},
		{"valid session", "Bearer token", &models.User{ID: id, Role: models.RoleUser, IsActive: true}, nil, 204, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auth := &testClerkAuth{user: tc.user, err: tc.err}
			handler := RequireClerkAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got, ok := UserIDFromContext(r.Context())
				if !ok || got != id {
					t.Fatal("incorrect user identity")
				}
				role, ok := RoleFromContext(r.Context())
				if !ok || role != models.RoleUser {
					t.Fatal("incorrect role")
				}
				w.WriteHeader(204)
			}))
			req := httptest.NewRequest("GET", "/users/me", nil)
			req.Header.Set("Authorization", tc.header)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != tc.status || auth.calls != tc.calls {
				t.Fatalf("status=%d calls=%d", res.Code, auth.calls)
			}
		})
	}
}
