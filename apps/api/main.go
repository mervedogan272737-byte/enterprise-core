package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"enterprise-core/api/internal/cache"
	"enterprise-core/api/internal/config"
	"enterprise-core/api/internal/database"
	"enterprise-core/api/internal/health"
	apihttp "enterprise-core/api/internal/http"
)

func main() {
	cfg := config.Load()
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

	router := apihttp.NewRouter(healthHandler)

	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Enterprise API running on http://localhost:%s", cfg.APIPort)
	log.Printf("Health check: http://localhost:%s/health", cfg.APIPort)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
