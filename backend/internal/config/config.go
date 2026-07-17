package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	MeilisearchURL string
	MeilisearchKey string
	JWTSecret      string
	S3Endpoint     string
	S3Bucket       string
	S3AccessKey    string
	S3SecretKey    string
	S3Region       string
	CORSOrigins    []string
	FrontendURL    string
	Env            string
	// RunWorkers embeds asynq workers in the API process (default true).
	// Set false when running dedicated plexus-worker replicas.
	RunWorkers bool
	// AllowRegistration controls open self-signup (default: on in development/test, off otherwise).
	AllowRegistration bool
	// MetricsAuthToken protects metrics/debug endpoints and /health/detailed (required outside development).
	MetricsAuthToken string
	// EncryptionKey encrypts webhook secrets at rest (falls back to JWT_SECRET).
	EncryptionKey string
	// MetricsEnabled starts the dedicated metrics/pprof listener (default true).
	MetricsEnabled bool
	// MetricsListenAddress is the bind address for /metrics and /debug/pprof.
	MetricsListenAddress string
	// LogLevel is slog level: debug, info, warn, error.
	LogLevel string
}

var insecureJWTSecrets = map[string]struct{}{
	"":                                    {},
	"change-me-in-production":             {},
	"dev-jwt-secret-change-in-production": {},
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                 getEnv("PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://plexus:plexus@localhost:5432/plexus?sslmode=disable"),
		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379"),
		MeilisearchURL:       getEnv("MEILISEARCH_URL", "http://localhost:7700"),
		MeilisearchKey:       getEnv("MEILISEARCH_KEY", ""),
		JWTSecret:            getEnv("JWT_SECRET", "change-me-in-production"),
		S3Endpoint:           getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3Bucket:             getEnv("S3_BUCKET", "plexus"),
		S3AccessKey:          getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:          getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Region:             getEnv("S3_REGION", "us-east-1"),
		CORSOrigins:          strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173,http://127.0.0.1:5173,app://plexus"), ","),
		FrontendURL:          getEnv("FRONTEND_URL", "http://localhost:3000"),
		Env:                  getEnv("ENV", "development"),
		RunWorkers:           getEnv("RUN_WORKERS", "true") != "false",
		MetricsAuthToken:     getEnv("METRICS_AUTH_TOKEN", ""),
		EncryptionKey:        getEnv("ENCRYPTION_KEY", ""),
		MetricsEnabled:       getEnv("METRICS_ENABLED", "true") != "false",
		MetricsListenAddress: getEnv("METRICS_LISTEN_ADDRESS", ":9090"),
		LogLevel:             strings.ToLower(getEnv("LOG_LEVEL", "info")),
	}
	cfg.AllowRegistration = parseAllowRegistration(cfg.Env)
	if cfg.EncryptionKey == "" {
		cfg.EncryptionKey = cfg.JWTSecret
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func parseAllowRegistration(env string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ALLOW_REGISTRATION"))) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return env == "development" || env == "test"
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development" || c.Env == "test"
}

// Validate enforces production hardening.
func (c *Config) Validate() error {
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error (got %q)", c.LogLevel)
	}
	if c.IsDevelopment() {
		return nil
	}
	if _, bad := insecureJWTSecrets[c.JWTSecret]; bad || len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be set to a strong value (≥32 chars) when ENV=%s", c.Env)
	}
	if c.MetricsAuthToken == "" {
		return fmt.Errorf("METRICS_AUTH_TOKEN must be set when ENV=%s", c.Env)
	}
	if _, bad := insecureJWTSecrets[c.EncryptionKey]; bad || len(c.EncryptionKey) < 32 {
		return fmt.Errorf("ENCRYPTION_KEY (or JWT_SECRET) must be a strong value (≥32 chars) when ENV=%s", c.Env)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
