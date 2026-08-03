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
	"enterprise-core/api/internal/auth/token"
	"enterprise-core/api/internal/cache"
	"enterprise-core/api/internal/config"
	"enterprise-core/api/internal/database"
	"enterprise-core/api/internal/session"
	"enterprise-core/api/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {
	ctx := context.Background()

	// ============================================================
	// CONFIG
	// ============================================================

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf(
			"configuration load failed: %v",
			err,
		)
	}

	// ============================================================
	// DATABASE
	// ============================================================

	db, err := database.Connect(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatalf(
			"database connection failed: %v",
			err,
		)
	}
	defer db.Close()

	// ============================================================
	// REDIS
	// ============================================================

	redisClient, err := cache.Connect(
		ctx,
		cfg.RedisURL,
	)
	if err != nil {
		log.Fatalf(
			"redis connection failed: %v",
			err,
		)
	}
	defer redisClient.Close()

	// ============================================================
	// JWT TOKEN MANAGER
	// ============================================================

	tokenManager := token.NewManager(
		cfg.JWTSecret,
		24*time.Hour,
	)

	// ============================================================
	// REPOSITORIES / SERVICES
	// ============================================================

	userRepository := user.NewRepository(
		db,
	)

	sessionStore := session.NewStore(
		redisClient,
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

	// ============================================================
	// ROUTER
	// ============================================================

	r := chi.NewRouter()

	// ============================================================
	// CORS
	// ============================================================

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

	// ============================================================
	// HEALTH
	// ============================================================

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

	// ============================================================
	// AUTH ROUTES
	// ============================================================

	r.Route(
		"/auth",
		func(r chi.Router) {

			// Public routes

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

			// Protected user endpoint

			r.With(
				authmiddleware.JWTAuth(
					tokenManager,
				),
			).Get(
				"/me",
				authHandler.Me,
			)
		},
	)

	// ============================================================
	// ADMIN ROUTES
	// ============================================================

	r.Route(
		"/admin",
		func(r chi.Router) {

			// JWT authentication

			r.Use(
				authmiddleware.JWTAuth(
					tokenManager,
				),
			)

			// Admin role authorization

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

	// ============================================================
	// HTTP SERVER
	// ============================================================

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
