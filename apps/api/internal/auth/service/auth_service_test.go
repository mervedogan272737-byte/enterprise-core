package service

import (
	"context"
	"testing"
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
