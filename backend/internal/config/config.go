package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env      string         `mapstructure:"env"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Worker   WorkerConfig   `mapstructure:"worker"`
}

type WorkerConfig struct {
	Concurrency   int `mapstructure:"concurrency"`
	QueueCapacity int `mapstructure:"queue_capacity"`
}

type HTTPConfig struct {
	Port            int           `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	AllowedOrigins  string        `mapstructure:"allowed_origins"`
}

type DatabaseConfig struct {
	URL      string `mapstructure:"url"`
	MaxConns int32  `mapstructure:"max_conns"`
}

type AuthConfig struct {
	JWTSecret       string        `mapstructure:"jwt_secret"`
	CredentialsKey  string        `mapstructure:"credentials_key"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
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
	// Load the shared repository environment, with backend-local compatibility.
	_ = godotenv.Overload("../.env", ".env")

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("http.port", 8080)
	v.SetDefault("http.shutdown_timeout", "10s")
	v.SetDefault("database.max_conns", 10)
	v.SetDefault("worker.concurrency", 4)
	v.SetDefault("worker.queue_capacity", 64)
	v.SetDefault("auth.access_token_ttl", "15m")
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
	if err := v.BindEnv("auth.credentials_key", "CREDENTIALS_ENCRYPTION_KEY"); err != nil {
		return nil, fmt.Errorf("bind CREDENTIALS_ENCRYPTION_KEY: %w", err)
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
	if cfg.Worker.Concurrency < 1 || cfg.Worker.QueueCapacity < 1 {
		return nil, fmt.Errorf("worker concurrency and queue capacity must be positive")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if len(cfg.Auth.CredentialsKey) != 32 {
		return nil, fmt.Errorf("CREDENTIALS_ENCRYPTION_KEY must be exactly 32 bytes (AES-256)")
	}

	return &cfg, nil
}
