package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"enterprise-core/api/internal/admin"
	"enterprise-core/api/internal/auth"
	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/auth/repository"
	"enterprise-core/api/internal/auth/session"
	"enterprise-core/api/internal/cache"
	"enterprise-core/api/internal/config"
	"enterprise-core/api/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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

	redisClient, err := cache.Connect(
		ctx,
		cfg.RedisURL,
	)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	userRepository := repository.NewRepository(
		db,
	)

	sessionStore := session.NewManager(
		redisClient,
		7*24*time.Hour,
	)

	authService := auth.NewService(
		cfg.JWTSecret,
	)

	authHandler := auth.NewHandler(
		userRepository,
		authService,
		sessionStore,
	)

	adminHandler := admin.NewHandler(
		userRepository,
	)

	r := chi.NewRouter()

	allowedOrigins := strings.Split(
		cfg.CORSOrigins,
		",",
	)

	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(
			allowedOrigins[i],
		)
	}

	r.Use(
		cors.Handler(
			cors.Options{
				AllowedOrigins: allowedOrigins,

				AllowedMethods: []string{
					"GET",
					"POST",
					"PUT",
					"PATCH",
					"DELETE",
					"OPTIONS",
				},

				AllowedHeaders: []string{
					"Accept",
					"Authorization",
					"Content-Type",
				},

				AllowCredentials: true,

				MaxAge: 300,
			},
		),
	)

	r.Get(
		"/health",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(
				http.StatusOK,
			)

			_, _ = w.Write(
				[]byte(
					`{"status":"ok","service":"enterprise-api","database":"ok","redis":"ok"}`,
				),
			)
		},
	)

	r.Route(
		"/auth",
		func(r chi.Router) {
			r.Post(
				"/register",
				authHandler.Register,
			)

			r.Post(
				"/login",
				authHandler.Login,
			)

			r.Post(
				"/refresh",
				authHandler.Refresh,
			)

			r.Post(
				"/logout",
				authHandler.Logout,
			)

			r.With(
				authmiddleware.JWTAuth(
					authService,
				),
			).Get(
				"/me",
				authHandler.Me,
			)
		},
	)

	r.Route(
		"/admin",
		func(r chi.Router) {
			r.Use(
				authmiddleware.JWTAuth(
					authService,
				),
			)

			r.Use(
				authmiddleware.RequireRole(
					"admin",
				),
			)

			r.Get(
				"/me",
				adminHandler.Me,
			)
		},
	)

	server := &http.Server{
		Addr: ":" + cfg.APIPort,

		Handler: r,

		ReadTimeout: 15 * time.Second,

		WriteTimeout: 15 * time.Second,

		IdleTimeout: 60 * time.Second,
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
