package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	authhandler "enterprise-core/api/internal/auth/handler"
	authrepository "enterprise-core/api/internal/auth/repository"
	authservice "enterprise-core/api/internal/auth/service"
	authsession "enterprise-core/api/internal/auth/session"
	"enterprise-core/api/internal/auth/token"
	"enterprise-core/api/internal/cache"
	"enterprise-core/api/internal/config"
	"enterprise-core/api/internal/database"
	"enterprise-core/api/internal/health"
	apihttp "enterprise-core/api/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	redisClient, err := cache.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	healthHandler := health.Handler{
		DB:    db,
		Redis: redisClient,
	}

	userRepository := authrepository.NewRepository(db)

	accessTokenTTL := 24 * time.Hour
	refreshTokenTTL := 30 * 24 * time.Hour

	if ttlMinutes := os.Getenv("JWT_ACCESS_TOKEN_TTL_MINUTES"); ttlMinutes != "" {
		if minutes, err := strconv.Atoi(ttlMinutes); err == nil && minutes > 0 {
			accessTokenTTL = time.Duration(minutes) * time.Minute
		}
	}

	if ttlDays := os.Getenv("JWT_REFRESH_TOKEN_TTL_DAYS"); ttlDays != "" {
		if days, err := strconv.Atoi(ttlDays); err == nil && days > 0 {
			refreshTokenTTL = time.Duration(days) * 24 * time.Hour
		}
	}

	tokenManager := token.NewManager(
		cfg.JWTSecret,
		accessTokenTTL,
	)

	sessionManager := authsession.NewManager(
		redisClient,
		refreshTokenTTL,
	)

	authService := authservice.NewService(
		userRepository,
		tokenManager,
		sessionManager,
	)

	authHandler := authhandler.NewAuthHandler(
		authService,
	)

	router := apihttp.NewRouter(
		healthHandler,
		authHandler,
		tokenManager,
	)

	server := &http.Server{
		Addr:              "0.0.0.0:" + cfg.APIPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf(
		"Enterprise API starting on 0.0.0.0:%s",
		cfg.APIPort,
	)

	log.Printf(
		"Health check endpoint: http://localhost:%s/health",
		cfg.APIPort,
	)

	log.Printf(
		"API login endpoint: http://localhost:%s/auth/login",
		cfg.APIPort,
	)

	log.Printf(
		"API refresh endpoint: http://localhost:%s/auth/refresh",
		cfg.APIPort,
	)

	log.Printf(
		"API logout endpoint: http://localhost:%s/auth/logout",
		cfg.APIPort,
	)

	log.Printf(
		"API me endpoint: http://localhost:%s/auth/me",
		cfg.APIPort,
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
