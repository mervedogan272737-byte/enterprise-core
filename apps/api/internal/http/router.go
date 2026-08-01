package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

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

			r.With(
				authmiddleware.JWTAuth(tokenManager),
			).Get(
				"/me",
				authHandler.Me,
			)
		},
	)

	return r
}