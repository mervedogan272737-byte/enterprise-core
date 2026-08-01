package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	UserID   string `json:"UserID"`
	Email    string `json:"Email"`
	FullName string `json:"FullName"`
	Role     string `json:"Role"`
	jwt.RegisteredClaims
}

type Manager struct {
	SecretKey      string
	AccessTokenTTL time.Duration
}

func NewManager(secretKey string, accessTokenTTL time.Duration) *Manager {
	return &Manager{
		SecretKey:      secretKey,
		AccessTokenTTL: accessTokenTTL,
	}
}

func (m *Manager) GenerateAccessToken(
	userID string,
	email string,
	fullName string,
	role string,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:   userID,
		Email:    email,
		FullName: fullName,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.AccessTokenTTL)),
		},
	}

	newToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return newToken.SignedString(
		[]byte(m.SecretKey),
	)
}

func (m *Manager) ValidateAccessToken(
	tokenString string,
) (*Claims, error) {
	claims := &Claims{}

	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(parsedToken *jwt.Token) (interface{}, error) {
			if parsedToken.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return []byte(m.SecretKey), nil
		},
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
