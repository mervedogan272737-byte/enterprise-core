package service

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"enterprise-core/api/internal/auth/token"
)

func TestRegister_InvalidEmail(t *testing.T) {
	s := &Service{}

	_, err := s.Register(
		context.Background(),
		RegisterRequest{
			Email:    "invalid-email",
			Password: "Test123456!",
			FullName: "Test User",
		},
	)

	if err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestRegister_InvalidPassword(t *testing.T) {
	s := &Service{}

	_, err := s.Register(
		context.Background(),
		RegisterRequest{
			Email:    "test@example.com",
			Password: "123",
			FullName: "Test User",
		},
	)

	if err != ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLogin_InvalidEmail(t *testing.T) {
	s := &Service{}

	_, _, err := s.Login(
		context.Background(),
		LoginRequest{
			Email:    "invalid-email",
			Password: "Test123456!",
		},
	)

	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_EmptyPassword(t *testing.T) {
	s := &Service{}

	_, _, err := s.Login(
		context.Background(),
		LoginRequest{
			Email:    "test@example.com",
			Password: "",
		},
	)

	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestPasswordHashVerification(t *testing.T) {
	password := "Test123456!"

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	err = bcrypt.CompareHashAndPassword(
		hash,
		[]byte(password),
	)

	if err != nil {
		t.Fatalf("expected password hash to match: %v", err)
	}

	err = bcrypt.CompareHashAndPassword(
		hash,
		[]byte("WrongPassword123!"),
	)

	if err == nil {
		t.Fatal("expected wrong password to fail")
	}
}

func TestTokenManagerIntegration(t *testing.T) {
	manager := token.NewManager(
		"test-secret-key",
		time.Hour,
	)

	tokenString, err := manager.GenerateAccessToken(
		"user-123",
		"user@example.com",
		"Test User",
		"user",
	)

	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected access token")
	}

	claims, err := manager.ValidateAccessToken(
		tokenString,
	)

	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Fatalf(
			"expected user ID %q, got %q",
			"user-123",
			claims.UserID,
		)
	}

	if claims.Email != "user@example.com" {
		t.Fatalf(
			"expected email %q, got %q",
			"user@example.com",
			claims.Email,
		)
	}

	if claims.FullName != "Test User" {
		t.Fatalf(
			"expected full name %q, got %q",
			"Test User",
			claims.FullName,
		)
	}

	if claims.Role != "user" {
		t.Fatalf(
			"expected role %q, got %q",
			"user",
			claims.Role,
		)
	}
}
