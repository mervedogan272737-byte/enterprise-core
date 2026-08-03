package customer

import (
	"context"
	"errors"
	"strings"

	"enterprise-core/api/internal/customer/repository"
)

var (
	ErrInvalidCustomer = errors.New("invalid customer")
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	organizationID,
	name,
	email,
	phone,
	company,
	notes string,
) (*repository.Customer, error) {
	organizationID = strings.TrimSpace(organizationID)
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)
	company = strings.TrimSpace(company)
	notes = strings.TrimSpace(notes)

	if organizationID == "" || name == "" {
		return nil, ErrInvalidCustomer
	}

	return s.repo.Create(
		ctx,
		organizationID,
		name,
		email,
		phone,
		company,
		notes,
	)
}

func (s *Service) List(
	ctx context.Context,
	organizationID string,
) ([]repository.Customer, error) {
	organizationID = strings.TrimSpace(organizationID)

	if organizationID == "" {
		return nil, ErrInvalidCustomer
	}

	return s.repo.List(ctx, organizationID)
}

func (s *Service) Get(
	ctx context.Context,
	organizationID,
	customerID string,
) (*repository.Customer, error) {
	organizationID = strings.TrimSpace(organizationID)
	customerID = strings.TrimSpace(customerID)

	if organizationID == "" || customerID == "" {
		return nil, ErrInvalidCustomer
	}

	return s.repo.Get(
		ctx,
		organizationID,
		customerID,
	)
}

func (s *Service) Update(
	ctx context.Context,
	organizationID,
	customerID,
	name,
	email,
	phone,
	company,
	notes,
	status string,
) (*repository.Customer, error) {
	organizationID = strings.TrimSpace(organizationID)
	customerID = strings.TrimSpace(customerID)
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)
	company = strings.TrimSpace(company)
	notes = strings.TrimSpace(notes)
	status = strings.TrimSpace(strings.ToLower(status))

	if organizationID == "" ||
		customerID == "" ||
		name == "" {
		return nil, ErrInvalidCustomer
	}

	if status == "" {
		status = "active"
	}

	if status != "active" && status != "inactive" {
		return nil, ErrInvalidCustomer
	}

	return s.repo.Update(
		ctx,
		organizationID,
		customerID,
		name,
		email,
		phone,
		company,
		notes,
		status,
	)
}

func (s *Service) Delete(
	ctx context.Context,
	organizationID,
	customerID string,
) error {
	organizationID = strings.TrimSpace(organizationID)
	customerID = strings.TrimSpace(customerID)

	if organizationID == "" || customerID == "" {
		return ErrInvalidCustomer
	}

	return s.repo.Delete(
		ctx,
		organizationID,
		customerID,
	)
}
