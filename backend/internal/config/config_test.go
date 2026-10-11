package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvironmentPrecedenceAndBindings(t *testing.T) {
	file := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(file, []byte("JWT_SECRET=file-value-that-must-not-override-env\nHTTP_ALLOWED_ORIGINS=https://file.example\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "environment-secret-with-at-least-32-characters")
	t.Setenv("HTTP_ALLOWED_ORIGINS", "https://app.example")
	t.Setenv("HTTP_TRUSTED_PROXIES", "127.0.0.1/32,::1/128")
	t.Setenv("EXECUTION_MAX_PENDING_RUNS", "50")
	t.Setenv("EXECUTION_MAX_PENDING_PER_USER", "5")
	t.Setenv("EXECUTION_RUNS_PER_MINUTE", "10")
	cfg, err := load(file)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Auth.JWTSecret != "environment-secret-with-at-least-32-characters" || cfg.HTTP.AllowedOrigins != "https://app.example" {
		t.Fatal("environment was overridden or not bound")
	}
	if cfg.Execution.MaxPendingRuns != 50 || cfg.Execution.MaxPendingPerUser != 5 || cfg.Execution.RunsPerMinute != 10 {
		t.Fatalf("admission bindings: %+v", cfg.Execution)
	}
	proxies, err := cfg.HTTP.TrustedProxyPrefixes()
	if err != nil || len(proxies) != 2 {
		t.Fatalf("proxy configuration: %v %v", proxies, err)
	}
}

func TestInvalidLimitsAndProxies(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "test-secret-with-at-least-32-characters")
	for _, value := range []string{"not-a-cidr", "0.0.0.0/0", "::/0"} {
		t.Setenv("HTTP_TRUSTED_PROXIES", value)
		if _, err := load(filepath.Join(t.TempDir(), "missing.env")); err == nil {
			t.Fatalf("accepted trusted proxy %q", value)
		}
	}
	t.Setenv("HTTP_TRUSTED_PROXIES", "")
	t.Setenv("EXECUTION_MAX_PENDING_RUNS", "0")
	if _, err := load(filepath.Join(t.TempDir(), "missing.env")); err == nil {
		t.Fatal("accepted zero admission limit")
	}
}

func TestMissingLocalFileStillLoadsRepositoryEnvironment(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, ".env")
	if err := os.WriteFile(file, []byte("HTTP_PORT=8099\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "test-secret-with-at-least-32-characters")
	t.Setenv("HTTP_PORT", "")
	if err := os.Unsetenv("HTTP_PORT"); err != nil {
		t.Fatal(err)
	}
	cfg, err := load(filepath.Join(dir, "missing.env"), file)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTP.Port != 8099 {
		t.Fatalf("repository environment not loaded: port=%d", cfg.HTTP.Port)
	}
}
