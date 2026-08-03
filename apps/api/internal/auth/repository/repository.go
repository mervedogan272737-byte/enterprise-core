package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExist = errors.New("email already exists")
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
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	user := &User{}

	err := r.db.QueryRow(
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

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	id string,
) (*User, error) {
	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`
SELECT id, email, password_hash, full_name, role, is_active
FROM users
WHERE id = $1
`,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) CreateUser(
	ctx context.Context,
	email,
	passwordHash,
	fullName string,
) (*User, error) {
	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`
INSERT INTO users (
email,
password_hash,
full_name,
role,
is_active
)
VALUES ($1, $2, $3, 'user', TRUE)
RETURNING
id,
email,
password_hash,
full_name,
role,
is_active
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
		return nil, err
	}

	return user, nil
}

func (r *Repository) CreateUserWithOrganization(
	ctx context.Context,
	email,
	passwordHash,
	fullName,
	organizationName,
	organizationSlug string,
) (*User, string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, "", err
	}

	defer tx.Rollback(ctx)

	user := &User{}

	err = tx.QueryRow(
		ctx,
		`
INSERT INTO users (
email,
password_hash,
full_name,
role,
is_active
)
VALUES ($1, $2, $3, 'user', TRUE)
RETURNING
id,
email,
password_hash,
full_name,
role,
is_active
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
		return nil, "", err
	}

	var organizationID string

	err = tx.QueryRow(
		ctx,
		`
INSERT INTO organizations (
name,
slug,
is_active
)
VALUES ($1, $2, TRUE)
RETURNING id
`,
		organizationName,
		organizationSlug,
	).Scan(&organizationID)

	if err != nil {
		return nil, "", err
	}

	_, err = tx.Exec(
		ctx,
		`
INSERT INTO organization_members (
organization_id,
user_id,
role
)
VALUES ($1, $2, 'owner')
`,
		organizationID,
		user.ID,
	)

	if err != nil {
		return nil, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", err
	}

	return user, organizationID, nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT id, email, COALESCE(full_name, ''), role, is_active, created_at
 FROM users
 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User

		if err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.FullName,
			&u.Role,
			&u.IsActive,
		); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) SetActive(
	ctx context.Context,
	userID string,
	active bool,
) error {
	commandTag, err := r.db.Exec(
		ctx,
		`UPDATE users
 SET is_active = $1,
     updated_at = NOW()
 WHERE id = $2`,
		active,
		userID,
	)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repository) DeleteUser(
	ctx context.Context,
	userID string,
) error {
	commandTag, err := r.db.Exec(
		ctx,
		`DELETE FROM users
 WHERE id = $1`,
		userID,
	)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repository) CreateAdminIfNotExists(
	ctx context.Context,
	email,
	passwordHash,
	fullName string,
) (*User, error) {
	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`
INSERT INTO users (
email,
password_hash,
full_name,
role,
is_active
)
VALUES ($1, $2, $3, 'admin', TRUE)
ON CONFLICT (email) DO NOTHING
RETURNING
id,
email,
password_hash,
full_name,
role,
is_active
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

	if errors.Is(err, pgx.ErrNoRows) {
		return r.FindByEmail(ctx, email)
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}
