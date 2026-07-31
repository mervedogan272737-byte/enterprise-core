package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	APIPort     string
	DatabaseURL string
	RedisURL    string
	CORSOrigins string
}

func Load() Config {
	_ = godotenv.Load("../../.env")

	return Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		APIPort:     getEnv("API_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://enterprise_user:enterprise_password@localhost:5432/enterprise_db?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
