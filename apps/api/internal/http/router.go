package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	authhandler "enterprise-core/api/internal/auth/handler"
	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/auth/token"
	"enterprise-core/api/internal/health"
)

func NewRouter(
	healthHandler health.Handler,
	authHandler *authhandler.AuthHandler,
	tokenManager *token.Manager,
) http.Handler {
	r := chi.NewRouter()

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
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
			"X-CSRF-Token",
		},
		ExposedHeaders: []string{
			"Link",
		},
		AllowCredentials: true,
		MaxAge: 300,
	}))

	// Health
	r.Get(
		"/health",
		healthHandler.All,
	)

	// Auth
	r.Route(
		"/auth",
		func(r chi.Router) {
			// Register
			r.Post(
				"/register",
				authHandler.Register,
			)

			// Login
			r.Post(
				"/login",
				authHandler.Login,
			)

			// Refresh token
			r.Post(
				"/refresh",
				authHandler.Refresh,
			)

			// Logout
			r.Post(
				"/logout",
				authHandler.Logout,
			)

			// Current authenticated user
			r.With(
				authmiddleware.JWTAuth(tokenManager),
			).Get(
				"/me",
				authHandler.Me,
			)

			// Admin only
			r.With(
				authmiddleware.JWTAuth(tokenManager),
				authmiddleware.RequireRole("admin"),
			).Get(
				"/admin/me",
				authHandler.AdminMe,
			)
		},
	)

	return r
}