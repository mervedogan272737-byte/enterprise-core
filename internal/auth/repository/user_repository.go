package repository

import (
"context"
"errors"
"fmt"
"strings"

"github.com/jackc/pgx/v5"
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
email = strings.ToLower(strings.TrimSpace(email))
fullName = strings.TrimSpace(fullName)

user := &User{}

err := r.DB.QueryRow(
ctx,
`
INSERT INTO users (
email,
password_hash,
full_name
)
VALUES ($1, $2, $3)
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
return nil, fmt.Errorf("create user: %w", err)
}

return user, nil
}

func (r *Repository) FindByEmail(
ctx context.Context,
email string,
) (*User, error) {
email = strings.ToLower(strings.TrimSpace(email))

user := &User{}

err := r.DB.QueryRow(
ctx,
`
SELECT
id,
email,
password_hash,
full_name,
role,
is_active
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

if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}

if err != nil {
return nil, fmt.Errorf("find user by email: %w", err)
}

return user, nil
}

func (r *Repository) FindByID(
ctx context.Context,
id string,
) (*User, error) {
id = strings.TrimSpace(id)

user := &User{}

err := r.DB.QueryRow(
ctx,
`
SELECT
id,
email,
password_hash,
full_name,
role,
is_active
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

if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}

if err != nil {
return nil, fmt.Errorf("find user by id: %w", err)
}

return user, nil
}

func (r *Repository) ListUsers(
ctx context.Context,
) ([]*User, error) {
rows, err := r.DB.Query(
ctx,
`
SELECT
id,
email,
full_name,
role,
is_active
FROM users
ORDER BY email ASC
`,
)
if err != nil {
return nil, fmt.Errorf("list users: %w", err)
}
defer rows.Close()

users := make([]*User, 0)

for rows.Next() {
user := &User{}

if err := rows.Scan(
&user.ID,
&user.Email,
&user.FullName,
&user.Role,
&user.IsActive,
); err != nil {
return nil, fmt.Errorf("scan user: %w", err)
}

users = append(users, user)
}

if err := rows.Err(); err != nil {
return nil, fmt.Errorf("iterate users: %w", err)
}

return users, nil
}

func (r *Repository) SetActive(
ctx context.Context,
id string,
active bool,
) error {
id = strings.TrimSpace(id)

result, err := r.DB.Exec(
ctx,
`
UPDATE users
SET is_active = $1
WHERE id = $2
`,
active,
id,
)

if err != nil {
return fmt.Errorf("set user active: %w", err)
}

if result.RowsAffected() == 0 {
return errors.New("user not found")
}

return nil
}

func (r *Repository) DeleteUser(
ctx context.Context,
id string,
) error {
id = strings.TrimSpace(id)

result, err := r.DB.Exec(
ctx,
`
DELETE FROM users
WHERE id = $1
`,
id,
)

if err != nil {
return fmt.Errorf("delete user: %w", err)
}

if result.RowsAffected() == 0 {
return errors.New("user not found")
}

return nil
}
