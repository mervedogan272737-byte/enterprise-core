package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-core/api/internal/auth/token"
)

func TestRequireRole_AllowsAdmin(t *testing.T) {
	handler := RequireRole("admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	claims := &token.Claims{
		Role: "admin",
	}

	ctx := context.WithValue(
		req.Context(),
		claimsContextKey,
		claims,
	)

	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRequireRole_RejectsUser(t *testing.T) {
	handler := RequireRole("admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	claims := &token.Claims{
		Role: "user",
	}

	ctx := context.WithValue(
		req.Context(),
		claimsContextKey,
		claims,
	)

	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestRequireRole_RejectsMissingClaims(t *testing.T) {
	handler := RequireRole("admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}
