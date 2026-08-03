package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
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
	loadEnvFile(".env")

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

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)

		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")

		if key == "" {
			continue
		}

		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
