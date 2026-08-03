package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"enterprise-core/api/internal/admin"
	auth "enterprise-core/api/internal/auth"
	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/auth/repository"
	"enterprise-core/api/internal/auth/session"
	"enterprise-core/api/internal/cache"
	"enterprise-core/api/internal/config"
	"enterprise-core/api/internal/database"
	"enterprise-core/api/internal/health"
	enterprisehttp "enterprise-core/api/internal/http"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration load failed: %v", err)
	}

	db, err := database.Connect(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(
		ctx,
		db,
		"./migrations",
	); err != nil {
		log.Fatalf(
			"database migrations failed: %v",
			err,
		)
	}

	redisClient, err := cache.Connect(
		ctx,
		cfg.RedisURL,
	)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	userRepository := repository.NewRepository(db)

	refreshTTL := 24 * time.Hour

	sessionManager := session.NewManager(
		redisClient,
		refreshTTL,
	)

	authService := auth.NewService(
		cfg.JWTSecret,
	)

	authHandler := auth.NewHandler(
		userRepository,
		authService,
		sessionManager,
	)

	adminHandler := admin.NewHandler(
		userRepository,
	)

	healthHandler := health.Handler{
		DB:    db,
		Redis: redisClient,
	}

	r := enterprisehttp.NewRouter(
		healthHandler,
		authHandler,
		adminHandler,
		authmiddleware.TokenValidator(authService),
	)

	server := &http.Server{
		Addr:         ":" + cfg.APIPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf(
		"Enterprise API starting on 0.0.0.0:%s",
		cfg.APIPort,
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf(
			"server failed: %v",
			err,
		)
	}
}
