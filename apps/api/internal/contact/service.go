package contact

import (
	"context"
	"errors"
	"strings"

	"enterprise-core/api/internal/contact/repository"
)

var ErrInvalidContact = errors.New("invalid contact")

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	organizationID,
	customerID,
	firstName,
	lastName,
	email,
	phone,
	position,
	notes string,
) (*repository.Contact, error) {
	organizationID = strings.TrimSpace(organizationID)
	customerID = strings.TrimSpace(customerID)
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	if organizationID == "" ||
		customerID == "" ||
		firstName == "" ||
		lastName == "" {
		return nil, ErrInvalidContact
	}

	return s.repo.Create(
		ctx,
		organizationID,
		customerID,
		firstName,
		lastName,
		strings.TrimSpace(email),
		strings.TrimSpace(phone),
		strings.TrimSpace(position),
		strings.TrimSpace(notes),
	)
}

func (s *Service) List(
	ctx context.Context,
	organizationID,
	customerID string,
) ([]repository.Contact, error) {
	if strings.TrimSpace(organizationID) == "" ||
		strings.TrimSpace(customerID) == "" {
		return nil, ErrInvalidContact
	}

	return s.repo.List(
		ctx,
		strings.TrimSpace(organizationID),
		strings.TrimSpace(customerID),
	)
}

func (s *Service) Get(
	ctx context.Context,
	organizationID,
	customerID,
	contactID string,
) (*repository.Contact, error) {
	if strings.TrimSpace(organizationID) == "" ||
		strings.TrimSpace(customerID) == "" ||
		strings.TrimSpace(contactID) == "" {
		return nil, ErrInvalidContact
	}

	return s.repo.Get(
		ctx,
		strings.TrimSpace(organizationID),
		strings.TrimSpace(customerID),
		strings.TrimSpace(contactID),
	)
}

func (s *Service) Update(
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
) (*repository.Contact, error) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	status = strings.ToLower(strings.TrimSpace(status))

	if status == "" {
		status = "active"
	}

	if status != "active" && status != "inactive" {
		return nil, ErrInvalidContact
	}

	if firstName == "" || lastName == "" {
		return nil, ErrInvalidContact
	}

	return s.repo.Update(
		ctx,
		strings.TrimSpace(organizationID),
		strings.TrimSpace(customerID),
		strings.TrimSpace(contactID),
		firstName,
		lastName,
		strings.TrimSpace(email),
		strings.TrimSpace(phone),
		strings.TrimSpace(position),
		strings.TrimSpace(notes),
		status,
	)
}

func (s *Service) Delete(
	ctx context.Context,
	organizationID,
	customerID,
	contactID string,
) error {
	return s.repo.Delete(
		ctx,
		strings.TrimSpace(organizationID),
		strings.TrimSpace(customerID),
		strings.TrimSpace(contactID),
	)
}
