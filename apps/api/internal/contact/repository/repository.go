package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Contact struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	CustomerID     string `json:"customer_id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Position       string `json:"position"`
	Notes          string `json:"notes"`
	Status         string `json:"status"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	organizationID,
	customerID,
	firstName,
	lastName,
	email,
	phone,
	position,
	notes string,
) (*Contact, error) {
	contact := &Contact{}

	err := r.db.QueryRow(
		ctx,
		`
        INSERT INTO contacts (
            organization_id,
            customer_id,
            first_name,
            last_name,
            email,
            phone,
            position,
            notes
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        RETURNING
            id,
            organization_id,
            customer_id,
            first_name,
            last_name,
            COALESCE(email,''),
            COALESCE(phone,''),
            COALESCE(position,''),
            COALESCE(notes,''),
            status
        `,
		organizationID,
		customerID,
		firstName,
		lastName,
		email,
		phone,
		position,
		notes,
	).Scan(
		&contact.ID,
		&contact.OrganizationID,
		&contact.CustomerID,
		&contact.FirstName,
		&contact.LastName,
		&contact.Email,
		&contact.Phone,
		&contact.Position,
		&contact.Notes,
		&contact.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("create contact: %w", err)
	}

	return contact, nil
}

func (r *Repository) List(
	ctx context.Context,
	organizationID,
	customerID string,
) ([]Contact, error) {
	rows, err := r.db.Query(
		ctx,
		`
        SELECT
            id,
            organization_id,
            customer_id,
            first_name,
            last_name,
            COALESCE(email,''),
            COALESCE(phone,''),
            COALESCE(position,''),
            COALESCE(notes,''),
            status
        FROM contacts
        WHERE organization_id = $1
          AND customer_id = $2
        ORDER BY created_at DESC
        `,
		organizationID,
		customerID,
	)

	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}

	defer rows.Close()

	contacts := make([]Contact, 0)

	for rows.Next() {
		var contact Contact

		if err := rows.Scan(
			&contact.ID,
			&contact.OrganizationID,
			&contact.CustomerID,
			&contact.FirstName,
			&contact.LastName,
			&contact.Email,
			&contact.Phone,
			&contact.Position,
			&contact.Notes,
			&contact.Status,
		); err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}

		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list contacts rows: %w", err)
	}

	return contacts, nil
}

func (r *Repository) Get(
	ctx context.Context,
	organizationID,
	customerID,
	contactID string,
) (*Contact, error) {
	contact := &Contact{}

	err := r.db.QueryRow(
		ctx,
		`
        SELECT
            id,
            organization_id,
            customer_id,
            first_name,
            last_name,
            COALESCE(email,''),
            COALESCE(phone,''),
            COALESCE(position,''),
            COALESCE(notes,''),
            status
        FROM contacts
        WHERE organization_id = $1
          AND customer_id = $2
          AND id = $3
        `,
		organizationID,
		customerID,
		contactID,
	).Scan(
		&contact.ID,
		&contact.OrganizationID,
		&contact.CustomerID,
		&contact.FirstName,
		&contact.LastName,
		&contact.Email,
		&contact.Phone,
		&contact.Position,
		&contact.Notes,
		&contact.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("get contact: %w", err)
	}

	return contact, nil
}

func (r *Repository) Update(
	ctx context.Context,
	organizationID,
	customerID,
	contactID,
	firstName,
	lastName,
	email,
	phone,
	position,
	notes,
	status string,
) (*Contact, error) {
	contact := &Contact{}

	err := r.db.QueryRow(
		ctx,
		`
        UPDATE contacts
        SET
            first_name = $4,
            last_name = $5,
            email = $6,
            phone = $7,
            position = $8,
            notes = $9,
            status = $10,
            updated_at = NOW()
        WHERE organization_id = $1
          AND customer_id = $2
          AND id = $3
        RETURNING
            id,
            organization_id,
            customer_id,
            first_name,
            last_name,
            COALESCE(email,''),
            COALESCE(phone,''),
            COALESCE(position,''),
            COALESCE(notes,''),
            status
        `,
		organizationID,
		customerID,
		contactID,
		firstName,
		lastName,
		email,
		phone,
		position,
		notes,
		status,
	).Scan(
		&contact.ID,
		&contact.OrganizationID,
		&contact.CustomerID,
		&contact.FirstName,
		&contact.LastName,
		&contact.Email,
		&contact.Phone,
		&contact.Position,
		&contact.Notes,
		&contact.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("update contact: %w", err)
	}

	return contact, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	organizationID,
	customerID,
	contactID string,
) error {
	result, err := r.db.Exec(
		ctx,
		`
        DELETE FROM contacts
        WHERE organization_id = $1
          AND customer_id = $2
          AND id = $3
        `,
		organizationID,
		customerID,
		contactID,
	)

	if err != nil {
		return fmt.Errorf("delete contact: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("contact not found")
	}

	return nil
}
