package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Customer struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Company        string `json:"company"`
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
	name,
	email,
	phone,
	company,
	notes string,
) (*Customer, error) {
	customer := &Customer{}

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO customers
(organization_id, name, email, phone, company, notes)
 VALUES ($1, $2, $3, $4, $5, $6)
 RETURNING id, organization_id, name,
           COALESCE(email, ''),
           COALESCE(phone, ''),
           COALESCE(company, ''),
           COALESCE(notes, ''),
           status`,
		organizationID,
		name,
		email,
		phone,
		company,
		notes,
	).Scan(
		&customer.ID,
		&customer.OrganizationID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.Company,
		&customer.Notes,
		&customer.Status,
	)

	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (r *Repository) List(
	ctx context.Context,
	organizationID string,
) ([]Customer, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
id,
organization_id,
name,
COALESCE(email, ''),
COALESCE(phone, ''),
COALESCE(company, ''),
COALESCE(notes, ''),
status
 FROM customers
 WHERE organization_id = $1
 ORDER BY created_at DESC`,
		organizationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := make([]Customer, 0)

	for rows.Next() {
		var customer Customer

		if err := rows.Scan(
			&customer.ID,
			&customer.OrganizationID,
			&customer.Name,
			&customer.Email,
			&customer.Phone,
			&customer.Company,
			&customer.Notes,
			&customer.Status,
		); err != nil {
			return nil, err
		}

		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return customers, nil
}

func (r *Repository) Get(
	ctx context.Context,
	organizationID,
	customerID string,
) (*Customer, error) {
	customer := &Customer{}

	err := r.db.QueryRow(
		ctx,
		`SELECT
id,
organization_id,
name,
COALESCE(email, ''),
COALESCE(phone, ''),
COALESCE(company, ''),
COALESCE(notes, ''),
status
 FROM customers
 WHERE organization_id = $1
   AND id = $2`,
		organizationID,
		customerID,
	).Scan(
		&customer.ID,
		&customer.OrganizationID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.Company,
		&customer.Notes,
		&customer.Status,
	)

	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (r *Repository) Update(
	ctx context.Context,
	organizationID,
	customerID,
	name,
	email,
	phone,
	company,
	notes,
	status string,
) (*Customer, error) {
	customer := &Customer{}

	err := r.db.QueryRow(
		ctx,
		`UPDATE customers
 SET name = $3,
     email = $4,
     phone = $5,
     company = $6,
     notes = $7,
     status = $8,
     updated_at = NOW()
 WHERE organization_id = $1
   AND id = $2
 RETURNING id, organization_id, name,
           COALESCE(email, ''),
           COALESCE(phone, ''),
           COALESCE(company, ''),
           COALESCE(notes, ''),
           status`,
		organizationID,
		customerID,
		name,
		email,
		phone,
		company,
		notes,
		status,
	).Scan(
		&customer.ID,
		&customer.OrganizationID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.Company,
		&customer.Notes,
		&customer.Status,
	)

	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	organizationID,
	customerID string,
) error {
	_, err := r.db.Exec(
		ctx,
		`DELETE FROM customers
 WHERE organization_id = $1
   AND id = $2`,
		organizationID,
		customerID,
	)

	return err
}
