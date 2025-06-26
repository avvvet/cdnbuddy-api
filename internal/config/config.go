package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Environment string
	LogLevel    string
	DatabaseURL string
	NATSUrl     string

	// CDN Provider credentials
	CacheFlyToken    string
	CloudflareToken  string
	CloudflareZoneID string

	// JWT
	JWTSecret string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8081"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/cdnbuddy?sslmode=disable"),
		NATSUrl:     getEnv("NATS_URL", "nats://localhost:4222"),

		CacheFlyToken:    getEnv("CACHEFLY_TOKEN", ""),
		CloudflareToken:  getEnv("CLOUDFLARE_TOKEN", ""),
		CloudflareZoneID: getEnv("CLOUDFLARE_ZONE_ID", ""),

		JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
