package http

import (
	"net/http"

	authservice "enterprise-core/api/internal/auth"
	authhandler "enterprise-core/api/internal/auth/handler"
	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/health"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRouter(
	healthHandler health.Handler,
	authHandler *authhandler.AuthHandler,
	adminHandler *authhandler.AdminHandler,
	authService *authservice.Service,
) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3001",
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
		MaxAge:           300,
	}))

	r.Get(
		"/health",
		healthHandler.All,
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
				authmiddleware.JWTAuth(authService),
			).Get(
				"/me",
				authHandler.Me,
			)

			r.With(
				authmiddleware.JWTAuth(authService),
				authmiddleware.RequireRole("admin"),
			).Get(
				"/admin/me",
				authHandler.AdminMe,
			)
		},
	)

	r.Route(
		"/admin",
		func(r chi.Router) {
			r.Use(
				authmiddleware.JWTAuth(authService),
				authmiddleware.RequireRole("admin"),
			)

			r.Get(
				"/users",
				adminHandler.ListUsers,
			)

			r.Get(
				"/users/{id}",
				adminHandler.GetUser,
			)

			r.Patch(
				"/users/{id}/active",
				adminHandler.SetUserActive,
			)

			r.Delete(
				"/users/{id}",
				adminHandler.DeleteUser,
			)
		},
	)

	return r
}
