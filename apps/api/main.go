package main

import (
	"context"
	"log"
	"net/http"
	"time"

	authhandler "enterprise-core/api/internal/auth/handler"
	authrepository "enterprise-core/api/internal/auth/repository"
	authservice "enterprise-core/api/internal/auth/service"
	"enterprise-core/api/internal/cache"
	"enterprise-core/api/internal/config"
	"enterprise-core/api/internal/database"
	"enterprise-core/api/internal/health"
	apihttp "enterprise-core/api/internal/http"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// PostgreSQL bağlantısı
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	// Redis bağlantısı
	redisClient, err := cache.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	// Health handler
	healthHandler := health.Handler{
		DB:    db,
		Redis: redisClient,
	}

	// User repository
	userRepository := authrepository.NewRepository(db)

	// Auth service
	authService := authservice.NewService(userRepository)

	// Auth handler
	authHandler := authhandler.NewAuthHandler(authService)

	// HTTP Router
	router := apihttp.NewRouter(
		healthHandler,
		authHandler,
	)

	// HTTP Server
	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf(
		"Enterprise API running on http://localhost:%s",
		cfg.APIPort,
	)

	log.Printf(
		"Health check: http://localhost:%s/health",
		cfg.APIPort,
	)

	// Server başlat
	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
