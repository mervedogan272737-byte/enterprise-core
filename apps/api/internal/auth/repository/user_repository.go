package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	FullName     string `json:"full_name"`
	Role         string `json:"role"`
	IsActive     bool   `json:"is_active"`
}

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		DB: db,
	}
}

func (r *Repository) CreateUser(
	ctx context.Context,
	email string,
	passwordHash string,
	fullName string,
) (*User, error) {
	user := &User{}

	err := r.DB.QueryRow(
		ctx,
		`
INSERT INTO users (email, password_hash, full_name)
VALUES ($1, $2, $3)
RETURNING id, email, password_hash, full_name, role, is_active
`,
		email,
		passwordHash,
		fullName,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *Repository) FindByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	user := &User{}

	err := r.DB.QueryRow(
		ctx,
		`
SELECT id, email, password_hash, full_name, role, is_active
FROM users
WHERE email = $1
`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}
