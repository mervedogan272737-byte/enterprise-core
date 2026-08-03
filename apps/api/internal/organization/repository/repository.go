package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type Organization struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	IsActive bool   `json:"is_active"`
}

type Member struct {
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	Email          string `json:"email"`
	FullName       string `json:"full_name"`
	Role           string `json:"role"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, name, slug string) (*Organization, error) {
	org := &Organization{}

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO organizations (name, slug)
 VALUES ($1, $2)
 RETURNING id, name, slug, is_active`,
		name,
		slug,
	).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return org, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Organization, error) {
	org := &Organization{}

	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, slug, is_active
 FROM organizations
 WHERE id = $1`,
		id,
	).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return org, nil
}

func (r *Repository) ListByUserID(ctx context.Context, userID string) ([]Organization, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT o.id, o.name, o.slug, o.is_active
 FROM organizations o
 INNER JOIN organization_members om
     ON om.organization_id = o.id
 WHERE om.user_id = $1
   AND o.is_active = TRUE
 ORDER BY o.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var organizations []Organization

	for rows.Next() {
		var org Organization

		if err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Slug,
			&org.IsActive,
		); err != nil {
			return nil, err
		}

		organizations = append(organizations, org)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return organizations, nil
}

func (r *Repository) UserExists(
	ctx context.Context,
	userID string,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`SELECT EXISTS(
SELECT 1
FROM users
WHERE id = $1
  AND is_active = TRUE
)`,
		userID,
	).Scan(&exists)

	return exists, err
}

func (r *Repository) AddMember(
	ctx context.Context,
	organizationID,
	userID,
	role string,
) error {
	exists, err := r.UserExists(
		ctx,
		userID,
	)
	if err != nil {
		return err
	}

	if !exists {
		return ErrUserNotFound
	}

	_, err = r.db.Exec(
		ctx,
		`INSERT INTO organization_members
 (organization_id, user_id, role)
 VALUES ($1, $2, $3)
 ON CONFLICT (organization_id, user_id)
 DO UPDATE SET role = EXCLUDED.role`,
		organizationID,
		userID,
		role,
	)

	return err
}

func (r *Repository) IsMember(
	ctx context.Context,
	organizationID,
	userID string,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`SELECT EXISTS(
SELECT 1
FROM organization_members
WHERE organization_id = $1
  AND user_id = $2
)`,
		organizationID,
		userID,
	).Scan(&exists)

	return exists, err
}

func (r *Repository) GetMemberRole(
	ctx context.Context,
	organizationID,
	userID string,
) (string, error) {
	var role string

	err := r.db.QueryRow(
		ctx,
		`SELECT role
 FROM organization_members
 WHERE organization_id = $1
   AND user_id = $2`,
		organizationID,
		userID,
	).Scan(&role)

	if err == pgx.ErrNoRows {
		return "", nil
	}

	return role, err
}

func (r *Repository) ListMembers(
	ctx context.Context,
	organizationID string,
) ([]Member, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
om.organization_id,
om.user_id,
u.email,
COALESCE(u.full_name, ''),
om.role
 FROM organization_members om
 INNER JOIN users u
     ON u.id = om.user_id
 WHERE om.organization_id = $1
 ORDER BY om.created_at ASC`,
		organizationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member

	for rows.Next() {
		var member Member

		if err := rows.Scan(
			&member.OrganizationID,
			&member.UserID,
			&member.Email,
			&member.FullName,
			&member.Role,
		); err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *Repository) RemoveMember(
	ctx context.Context,
	organizationID,
	userID string,
) error {
	_, err := r.db.Exec(
		ctx,
		`DELETE FROM organization_members
 WHERE organization_id = $1
   AND user_id = $2`,
		organizationID,
		userID,
	)

	return err
}
