package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"enterprise-core/api/internal/auth/repository"
)

var (
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type RegisterRequest struct {
	Email    string
	Password string
	FullName string
}

type Service struct {
	Users *repository.Repository
}

func NewService(users *repository.Repository) *Service {
	return &Service{
		Users: users,
	}
}

func (s *Service) Register(
	ctx context.Context,
	req RegisterRequest,
) (*repository.User, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	fullName := strings.TrimSpace(req.FullName)

	if email == "" || !strings.Contains(email, "@") {
		return nil, ErrInvalidEmail
	}

	if len(req.Password) < 8 {
		return nil, ErrInvalidPassword
	}

	existingUser, err := s.Users.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.Users.CreateUser(
		ctx,
		email,
		string(passwordHash),
		fullName,
	)
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}

	return user, nil
}
