package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"enterprise-core/api/internal/auth/repository"
	"enterprise-core/api/internal/auth/token"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenGeneration    = errors.New("token generation failed")
)

type RegisterRequest struct {
	Email    string
	Password string
	FullName string
}

type LoginRequest struct {
	Email    string
	Password string
}

type Service struct {
	Users        *repository.Repository
	TokenManager *token.Manager
}

func NewService(
	users *repository.Repository,
	tokenManager *token.Manager,
) *Service {
	return &Service{
		Users:        users,
		TokenManager: tokenManager,
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

func (s *Service) Login(
	ctx context.Context,
	req LoginRequest,
) (*repository.User, string, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	if email == "" || !strings.Contains(email, "@") {
		return nil, "", ErrInvalidCredentials
	}

	if req.Password == "" {
		return nil, "", ErrInvalidCredentials
	}

	user, err := s.Users.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	accessToken, err := s.TokenManager.GenerateAccessToken(
		user.ID,
		user.Email,
		user.FullName,
		user.Role,
	)
	if err != nil {
		return nil, "", fmt.Errorf(
			"%w: %v",
			ErrTokenGeneration,
			err,
		)
	}

	return user, accessToken, nil
}
