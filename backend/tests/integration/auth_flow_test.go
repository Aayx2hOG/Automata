package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type authResponse struct {
	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func doJSON(t *testing.T, router http.Handler, method, path string, body interface{}, token string) (*httptest.ResponseRecorder, []byte) {
	t.Helper()

	var reqBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&reqBody).Encode(body); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec, rec.Body.Bytes()
}

func TestAuthFlow_FullLifecycle(t *testing.T) {
	env := setupTestEnv(t)

	email := fmt.Sprintf("integration-test-%d@example.com", time.Now().UnixNano())
	password := "correct-horse-battery-staple"

	rec, body := doJSON(t, env.router, http.MethodPost, "/auth/register", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d, body: %s", rec.Code, body)
	}
	var registerResp authResponse
	if err := json.Unmarshal(body, &registerResp); err != nil {
		t.Fatalf("register: decode response: %v", err)
	}
	if registerResp.AccessToken == "" || registerResp.RefreshToken == "" {
		t.Fatal("register: expected tokens in response, got none")
	}

	rec, _ = doJSON(t, env.router, http.MethodPost, "/auth/register", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate register: expected 409, got %d", rec.Code)
	}

	rec, body = doJSON(t, env.router, http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d, body: %s", rec.Code, body)
	}
	var loginResp authResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		t.Fatalf("login: decode response: %v", err)
	}

	rec, _ = doJSON(t, env.router, http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": "wrong-password",
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password login: expected 401, got %d", rec.Code)
	}

	rec, _ = doJSON(t, env.router, http.MethodGet, "/users/me", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("/users/me without token: expected 401, got %d", rec.Code)
	}

	rec, body = doJSON(t, env.router, http.MethodGet, "/users/me", nil, loginResp.AccessToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("/users/me with token: expected 200, got %d, body: %s", rec.Code, body)
	}
	var meResp struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &meResp); err != nil {
		t.Fatalf("/users/me: decode response: %v", err)
	}
	if meResp.Email != email {
		t.Fatalf("/users/me: expected email %q, got %q", email, meResp.Email)
	}

	rec, body = doJSON(t, env.router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": loginResp.RefreshToken,
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh: expected 200, got %d, body: %s", rec.Code, body)
	}
	var refreshResp authResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		t.Fatalf("refresh: decode response: %v", err)
	}
	if refreshResp.RefreshToken == loginResp.RefreshToken {
		t.Fatal("refresh: expected a new refresh token, got the same one back")
	}

	rec, _ = doJSON(t, env.router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": loginResp.RefreshToken,
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("reused refresh token: expected 401, got %d", rec.Code)
	}

	rec, _ = doJSON(t, env.router, http.MethodPost, "/auth/logout", map[string]string{
		"refresh_token": refreshResp.RefreshToken,
	}, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout: expected 204, got %d", rec.Code)
	}

	rec, _ = doJSON(t, env.router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": refreshResp.RefreshToken,
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout: expected 401, got %d", rec.Code)
	}
}

func TestAuthFlow_ValidationRejectsBadInput(t *testing.T) {
	env := setupTestEnv(t)

	cases := []struct {
		name string
		body map[string]string
	}{
		{"invalid email", map[string]string{"email": "not-an-email", "password": "password123"}},
		{"short password", map[string]string{"email": "valid@example.com", "password": "short"}},
		{"missing email", map[string]string{"password": "password123"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec, body := doJSON(t, env.router, http.MethodPost, "/auth/register", tc.body, "")
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d, body: %s", rec.Code, body)
			}
		})
	}
}
