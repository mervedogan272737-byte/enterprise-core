package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	APIPort     string
	DatabaseURL string
	RedisURL    string
	CORSOrigins string
	JWTSecret   string
}

func Load() (Config, error) {
	_ = godotenv.Load("../../.env")

	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}

	return Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		APIPort: getEnv("API_PORT", "8080"),
		DatabaseURL: getEnv(
			"DATABASE_URL",
			"postgres://enterprise_user:enterprise_password@localhost:5432/enterprise_db?sslmode=disable",
		),
		RedisURL: getEnv(
			"REDIS_URL",
			"redis://localhost:6379",
		),
		CORSOrigins: getEnv(
			"CORS_ORIGINS",
			"http://localhost:3000",
		),
		JWTSecret: jwtSecret,
	}, nil
}

func getEnv(
	key string,
	fallback string,
) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
