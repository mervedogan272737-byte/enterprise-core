package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

	if !strings.Contains(
		rec.Body.String(),
		"merve.user@example.com",
	) {
		t.Fatalf(
			"expected response to contain user email, got %s",
			rec.Body.String(),
		)
	}
}

func TestAuthHandler_AdminMeWithClaims(t *testing.T) {
	handler := &AuthHandler{}

	req := httptest.NewRequest(
		http.MethodGet,
		"/auth/admin/me",
		nil,
	)

	claims := &token.Claims{
		UserID:   "admin-user-id",
		Email:    "admin@example.com",
		FullName: "Admin User",
		Role:     "admin",
	}

	ctx := context.WithValue(
		req.Context(),
		middleware.ClaimsContextKeyForTest(),
		claims,
	)

	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.AdminMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var response meResponse

	if err := json.NewDecoder(
		rec.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.ID != "admin-user-id" {
		t.Fatalf(
			"expected ID %q, got %q",
			"admin-user-id",
			response.ID,
		)
	}

	if response.Email != "admin@example.com" {
		t.Fatalf(
			"expected email %q, got %q",
			"admin@example.com",
			response.Email,
		)
	}

	if response.FullName != "Admin User" {
		t.Fatalf(
			"expected full name %q, got %q",
			"Admin User",
			response.FullName,
		)
	}

	if response.Role != "admin" {
		t.Fatalf(
			"expected role %q, got %q",
			"admin",
			response.Role,
		)
	}
}
