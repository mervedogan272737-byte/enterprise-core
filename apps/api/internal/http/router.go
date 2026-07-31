package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"enterprise-core/api/internal/auth/handler"
	"enterprise-core/api/internal/health"
)

func NewRouter(
	healthHandler health.Handler,
	authHandler *handler.AuthHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", healthHandler.All)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
	})

	return r
}
