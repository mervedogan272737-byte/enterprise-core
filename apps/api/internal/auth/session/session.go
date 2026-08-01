package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrSessionNotFound     = errors.New("session not found")
)

const (
	refreshTokenBytes = 32
	sessionKeyPrefix  = "auth:session:"
)

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Manager struct {
	Redis           *redis.Client
	RefreshTokenTTL time.Duration
}

func NewManager(
	redisClient *redis.Client,
	refreshTokenTTL time.Duration,
) *Manager {
	return &Manager{
		Redis:           redisClient,
		RefreshTokenTTL: refreshTokenTTL,
	}
}

func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, refreshTokenBytes)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}

func sessionKey(tokenHash string) string {
	return sessionKeyPrefix + tokenHash
}

func (m *Manager) CreateSession(
	ctx context.Context,
	userID string,
	email string,
	fullName string,
	role string,
) (string, *Session, error) {
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return "", nil, err
	}

	now := time.Now().UTC()

	session := &Session{
		ID:        HashRefreshToken(refreshToken),
		UserID:    userID,
		Email:     email,
		FullName:  fullName,
		Role:      role,
		CreatedAt: now,
		ExpiresAt: now.Add(m.RefreshTokenTTL),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return "", nil, fmt.Errorf("marshal session: %w", err)
	}

	key := sessionKey(session.ID)

	if err := m.Redis.Set(
		ctx,
		key,
		data,
		m.RefreshTokenTTL,
	).Err(); err != nil {
		return "", nil, fmt.Errorf("store session: %w", err)
	}

	return refreshToken, session, nil
}

func (m *Manager) GetSession(
	ctx context.Context,
	refreshToken string,
) (*Session, error) {
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	tokenHash := HashRefreshToken(refreshToken)

	data, err := m.Redis.Get(
		ctx,
		sessionKey(tokenHash),
	).Result()

	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	var session Session

	if err := json.Unmarshal(
		[]byte(data),
		&session,
	); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		_ = m.RevokeSession(
			ctx,
			refreshToken,
		)

		return nil, ErrInvalidRefreshToken
	}

	return &session, nil
}

func (m *Manager) RevokeSession(
	ctx context.Context,
	refreshToken string,
) error {
	if refreshToken == "" {
		return ErrInvalidRefreshToken
	}

	tokenHash := HashRefreshToken(refreshToken)

	if err := m.Redis.Del(
		ctx,
		sessionKey(tokenHash),
	).Err(); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}

func (m *Manager) RotateSession(
	ctx context.Context,
	oldRefreshToken string,
) (string, *Session, error) {
	oldSession, err := m.GetSession(
		ctx,
		oldRefreshToken,
	)
	if err != nil {
		return "", nil, err
	}

	if err := m.RevokeSession(
		ctx,
		oldRefreshToken,
	); err != nil {
		return "", nil, err
	}

	return m.CreateSession(
		ctx,
		oldSession.UserID,
		oldSession.Email,
		oldSession.FullName,
		oldSession.Role,
	)
}
