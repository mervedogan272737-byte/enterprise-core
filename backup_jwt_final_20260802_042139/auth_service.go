package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	Issuer   = "enterprise-api"
	Audience = "enterprise-api"
)

type Service struct {
	JWTSecret string
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func NewService(secret string) *Service {
	return &Service{
		JWTSecret: secret,
	}
}

func (s *Service) HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (s *Service) CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	) == nil
}

func (s *Service) GenerateToken(
	userID string,
	email string,
	role string,
	tokenType string,
	duration time.Duration,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  strings.ToLower(strings.TrimSpace(email)),
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{Audience},
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}

	newToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return newToken.SignedString(
		[]byte(s.JWTSecret),
	)
}

func (s *Service) ValidateToken(
	ctx context.Context,
	tokenString string,
) (*Claims, error) {
	_ = ctx

	tokenString = strings.TrimSpace(tokenString)

	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimSpace(
			strings.TrimPrefix(tokenString, "Bearer "),
		)
	}

	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	claims := &Claims{}

	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(parsedToken *jwt.Token) (interface{}, error) {
			if parsedToken.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(s.JWTSecret), nil
		},
		jwt.WithIssuer(Issuer),
		jwt.WithAudience(Audience),
	)

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Type != "access" && claims.Type != "refresh" {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}
