package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-core/api/internal/auth/handler"
	"enterprise-core/api/internal/auth/token"
	"enterprise-core/api/internal/health"
)

func TestNewRouter_HealthRouteRegistered(t *testing.T) {
	router := NewRouter(
		health.Handler{},
		&handler.AuthHandler{},
		&token.Manager{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/unknown-route",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}
