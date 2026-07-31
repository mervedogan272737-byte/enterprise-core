package http

import (
	"net/http"

	"enterprise-core/api/internal/health"
	"github.com/go-chi/chi/v5"
)

func NewRouter(healthHandler health.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", healthHandler.All)

	return r
}
