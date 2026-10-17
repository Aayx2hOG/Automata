package config

import (
	"fmt"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	CredentialsEncryptionKey string          `mapstructure:"credentials_encryption_key"`
	Env                      string          `mapstructure:"env"`
	HTTP                     HTTPConfig      `mapstructure:"http"`
	Database                 DatabaseConfig  `mapstructure:"database"`
	Auth                     AuthConfig      `mapstructure:"auth"`
	Execution                ExecutionConfig `mapstructure:"execution"`
	Worker                   WorkerConfig    `mapstructure:"worker"`
}

type ExecutionConfig struct {
	MaxPendingRuns    int `mapstructure:"max_pending_runs"`
	MaxPendingPerUser int `mapstructure:"max_pending_per_user"`
	RunsPerMinute     int `mapstructure:"runs_per_minute"`
}

type WorkerConfig struct {
	Concurrency   int `mapstructure:"concurrency"`
	QueueCapacity int `mapstructure:"queue_capacity"`
}

type HTTPConfig struct {
	Port            int           `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	TrustedProxies  string        `mapstructure:"trusted_proxies"`
	AllowedOrigins  string        `mapstructure:"allowed_origins"`
}

type DatabaseConfig struct {
	URL      string `mapstructure:"url"`
	MaxConns int32  `mapstructure:"max_conns"`
}

type AuthConfig struct {
	RequireClerk           bool          `mapstructure:"require_clerk"`
	ClerkSecretKey         string        `mapstructure:"clerk_secret_key"`
	ClerkAuthorizedParties string        `mapstructure:"clerk_authorized_parties"`
	JWTSecret              string        `mapstructure:"jwt_secret"`
	AccessTokenTTL         time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL        time.Duration `mapstructure:"refresh_token_ttl"`
}

func (c HTTPConfig) AllowedOriginsList() []string {
	parts := strings.Split(c.AllowedOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func Load() (*Config, error) {
	return load(".env.local", ".env", "../.env")
}

func load(envFiles ...string) (*Config, error) {
	// Existing process variables always win. Local files are development defaults.
	for _, file := range envFiles {
		if err := godotenv.Load(file); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("load environment file %s: %w", file, err)
		}
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("credentials_encryption_key", "")

	v.SetDefault("env", "development")
	v.SetDefault("http.allowed_origins", "")
	v.SetDefault("http.trusted_proxies", "")
	v.SetDefault("execution.max_pending_runs", 10000)
	v.SetDefault("execution.max_pending_per_user", 100)
	v.SetDefault("execution.runs_per_minute", 60)
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.shutdown_timeout", "10s")
	v.SetDefault("database.max_conns", 10)
	v.SetDefault("worker.concurrency", 4)
	v.SetDefault("worker.queue_capacity", 64)
	v.SetDefault("auth.access_token_ttl", "15m")
	v.SetDefault("auth.require_clerk", false)
	v.SetDefault("auth.clerk_authorized_parties", "http://localhost:3000,http://127.0.0.1:3000")
	if err := v.BindEnv("auth.clerk_secret_key", "CLERK_SECRET_KEY"); err != nil {
		return nil, err
	}
	if err := v.BindEnv("auth.clerk_authorized_parties", "CLERK_AUTHORIZED_PARTIES"); err != nil {
		return nil, err
	}
	v.SetDefault("auth.refresh_token_ttl", "168h")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading the config file: %w", err)
		}
	}

	if err := v.BindEnv("database.url", "DATABASE_URL"); err != nil {
		return nil, fmt.Errorf("bind DATABASE_URL: %w", err)
	}
	if err := v.BindEnv("auth.jwt_secret", "JWT_SECRET"); err != nil {
		return nil, fmt.Errorf("bind JWT_SECRET: %w", err)
	}
	if err := v.BindEnv("env", "ENV"); err != nil {
		return nil, fmt.Errorf("bind ENV: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Auth.RequireClerk && strings.TrimSpace(cfg.Auth.ClerkSecretKey) == "" {
		return nil, fmt.Errorf("CLERK_SECRET_KEY is required when AUTH_REQUIRE_CLERK=true; use the same Clerk instance as the frontend")
	}
	if cfg.Worker.Concurrency < 1 || cfg.Worker.QueueCapacity < 1 {
		return nil, fmt.Errorf("worker concurrency and queue capacity must be positive")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	if cfg.HTTP.Port < 1 || cfg.HTTP.Port > 65535 || cfg.HTTP.ShutdownTimeout <= 0 {
		return nil, fmt.Errorf("HTTP port and shutdown timeout must be valid positive values")
	}
	if cfg.Database.MaxConns < 1 || cfg.Auth.AccessTokenTTL <= 0 || cfg.Auth.RefreshTokenTTL <= 0 {
		return nil, fmt.Errorf("database connections and token lifetimes must be positive")
	}
	if cfg.Execution.MaxPendingRuns < 1 || cfg.Execution.MaxPendingPerUser < 1 || cfg.Execution.RunsPerMinute < 1 {
		return nil, fmt.Errorf("execution admission limits must be positive")
	}
	if _, err := cfg.HTTP.TrustedProxyPrefixes(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c HTTPConfig) TrustedProxyPrefixes() ([]netip.Prefix, error) {
	var prefixes []netip.Prefix
	for _, entry := range strings.Split(c.TrustedProxies, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q", entry)
		}
		if prefix.Bits() == 0 {
			return nil, fmt.Errorf("trusted proxies must not include all addresses")
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}
