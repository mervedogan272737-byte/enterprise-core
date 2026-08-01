package token

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAccessToken(t *testing.T) {
	manager := NewManager("test-secret-key", time.Hour)

	tokenString, err := manager.GenerateAccessToken(
		"user-123",
		"user@example.com",
		"Test User",
		"user",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected token string")
	}

	parts := strings.Split(tokenString, ".")

	if len(parts) != 3 {
		t.Fatalf("expected JWT with 3 parts, got %d", len(parts))
	}
}

func TestValidateAccessToken_ValidToken(t *testing.T) {
	manager := NewManager("test-secret-key", time.Hour)

	tokenString, err := manager.GenerateAccessToken(
		"user-123",
		"user@example.com",
		"Test User",
		"user",
	)

	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := manager.ValidateAccessToken(tokenString)

	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Fatalf("expected user ID %q, got %q", "user-123", claims.UserID)
	}

	if claims.Email != "user@example.com" {
		t.Fatalf("expected email %q, got %q", "user@example.com", claims.Email)
	}

	if claims.FullName != "Test User" {
		t.Fatalf("expected full name %q, got %q", "Test User", claims.FullName)
	}

	if claims.Role != "user" {
		t.Fatalf("expected role %q, got %q", "user", claims.Role)
	}

	if claims.Subject != "user-123" {
		t.Fatalf("expected subject %q, got %q", "user-123", claims.Subject)
	}

	if claims.Issuer != Issuer {
		t.Fatalf("expected issuer %q, got %q", Issuer, claims.Issuer)
	}

	if len(claims.Audience) != 1 || claims.Audience[0] != Audience {
		t.Fatalf("expected audience %q, got %v", Audience, claims.Audience)
	}

	if claims.IssuedAt == nil {
		t.Fatal("expected issued-at claim")
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expected expiration claim")
	}

	if !claims.ExpiresAt.After(claims.IssuedAt.Time) {
		t.Fatal("expected expiration time to be after issued-at time")
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	manager := NewManager("test-secret-key", time.Hour)

	_, err := manager.ValidateAccessToken("this-is-not-a-valid-jwt")

	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAccessToken_ExpiredToken(t *testing.T) {
	manager := NewManager("test-secret-key", -time.Second)

	tokenString, err := manager.GenerateAccessToken(
		"user-123",
		"user@example.com",
		"Test User",
		"user",
	)

	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	_, err = manager.ValidateAccessToken(tokenString)

	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for expired token, got %v", err)
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	manager := NewManager("correct-secret", time.Hour)

	tokenString, err := manager.GenerateAccessToken(
		"user-123",
		"user@example.com",
		"Test User",
		"user",
	)

	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	wrongManager := NewManager("wrong-secret", time.Hour)

	_, err = wrongManager.ValidateAccessToken(tokenString)

	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for wrong secret, got %v", err)
	}
}

func TestValidateAccessToken_RejectsDifferentSigningMethod(t *testing.T) {
	manager := NewManager("test-secret-key", time.Hour)

	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS384,
		Claims{
			UserID:   "user-123",
			Email:    "user@example.com",
			FullName: "Test User",
			Role:     "user",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    Issuer,
				Audience:  jwt.ClaimStrings{Audience},
				Subject:   "user-123",
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		},
	)

	tokenString, err := jwtToken.SignedString([]byte("test-secret-key"))

	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = manager.ValidateAccessToken(tokenString)

	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for different signing method, got %v", err)
	}
}

func TestValidateAccessToken_RejectsWrongIssuer(t *testing.T) {
	manager := NewManager("test-secret-key", time.Hour)

	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			UserID:   "user-123",
			Email:    "user@example.com",
			FullName: "Test User",
			Role:     "user",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "wrong-issuer",
				Audience:  jwt.ClaimStrings{Audience},
				Subject:   "user-123",
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		},
	)

	tokenString, err := jwtToken.SignedString([]byte("test-secret-key"))

	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = manager.ValidateAccessToken(tokenString)

	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for wrong issuer, got %v", err)
	}
}

func TestValidateAccessToken_RejectsWrongAudience(t *testing.T) {
	manager := NewManager("test-secret-key", time.Hour)

	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			UserID:   "user-123",
			Email:    "user@example.com",
			FullName: "Test User",
			Role:     "user",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    Issuer,
				Audience:  jwt.ClaimStrings{"wrong-audience"},
				Subject:   "user-123",
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		},
	)

	tokenString, err := jwtToken.SignedString([]byte("test-secret-key"))

	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = manager.ValidateAccessToken(tokenString)

	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for wrong audience, got %v", err)
	}
}
