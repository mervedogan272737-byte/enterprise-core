package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"enterprise-core/api/internal/auth/token"

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

func (s *Service) CheckPassword(
	hash string,
	password string,
) bool {
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

	claims := token.Claims{
		UserID: strings.TrimSpace(userID),
		Email:  strings.ToLower(strings.TrimSpace(email)),
		Role:   strings.TrimSpace(role),
		Type:   strings.TrimSpace(tokenType),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: Issuer,
			Audience: jwt.ClaimStrings{
				Audience,
			},
			Subject: userID,
			IssuedAt: jwt.NewNumericDate(
				now,
			),
			NotBefore: jwt.NewNumericDate(
				now,
			),
			ExpiresAt: jwt.NewNumericDate(
				now.Add(duration),
			),
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
) (*token.Claims, error) {
	_ = ctx

	tokenString = strings.TrimSpace(
		tokenString,
	)

	if strings.HasPrefix(
		tokenString,
		"Bearer ",
	) {
		tokenString = strings.TrimSpace(
			strings.TrimPrefix(
				tokenString,
				"Bearer ",
			),
		)
	}

	if tokenString == "" {
		return nil, errors.New(
			"token is required",
		)
	}

	claims := &token.Claims{}

	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(
			parsedToken *jwt.Token,
		) (interface{}, error) {
			if parsedToken.Method != jwt.SigningMethodHS256 {
				return nil, errors.New(
					"unexpected signing method",
				)
			}

			return []byte(
				s.JWTSecret,
			), nil
		},
		jwt.WithIssuer(
			Issuer,
		),
		jwt.WithAudience(
			Audience,
		),
	)

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, errors.New(
			"invalid token",
		)
	}

	if claims.Type != "access" &&
		claims.Type != "refresh" {
		return nil, errors.New(
			"invalid token type",
		)
	}

	return claims, nil
}
