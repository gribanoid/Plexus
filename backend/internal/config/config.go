package config

import (
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
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://plexus:plexus@localhost:5432/plexus?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		MeilisearchURL:   getEnv("MEILISEARCH_URL", "http://localhost:7700"),
		MeilisearchKey:   getEnv("MEILISEARCH_KEY", ""),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
		S3Endpoint:       getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3Bucket:         getEnv("S3_BUCKET", "plexus"),
		S3AccessKey:      getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:      getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Region:         getEnv("S3_REGION", "us-east-1"),
		CORSOrigins:      strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000,app://plexus"), ","),
		FrontendURL:      getEnv("FRONTEND_URL", "http://localhost:3000"),
		Env:              getEnv("ENV", "development"),
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
