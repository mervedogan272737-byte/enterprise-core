package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/auth/token"
)

func TestAuthHandler_MeRequiresClaims(t *testing.T) {
	handler := &AuthHandler{}

	req := httptest.NewRequest(
		http.MethodGet,
		"/auth/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestAuthHandler_AdminMeRequiresClaims(t *testing.T) {
	handler := &AuthHandler{}

	req := httptest.NewRequest(
		http.MethodGet,
		"/auth/admin/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.AdminMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestAuthHandler_MeWithClaims(t *testing.T) {
	handler := &AuthHandler{}

	req := httptest.NewRequest(
		http.MethodGet,
		"/auth/me",
		nil,
	)

	claims := &token.Claims{
		UserID:   "test-user-id",
		Email:    "merve.user@example.com",
		FullName: "Merve User",
		Role:     "user",
	}

	ctx := context.WithValue(
		req.Context(),
		middleware.ClaimsContextKeyForTest(),
		claims,
	)

	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
