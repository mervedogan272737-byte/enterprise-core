package organization

import (
	"context"
	"errors"
	"strings"

	"enterprise-core/api/internal/organization/repository"
)

var (
	ErrInvalidOrganization  = errors.New("invalid organization")
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrForbidden            = errors.New("forbidden")
	ErrInvalidMember        = errors.New("invalid member")
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	name,
	slug string,
) (*repository.Organization, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(strings.ToLower(slug))

	if name == "" || slug == "" {
		return nil, ErrInvalidOrganization
	}

	return s.repo.Create(ctx, name, slug)
}

func (s *Service) Get(
	ctx context.Context,
	id string,
) (*repository.Organization, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return nil, ErrInvalidOrganization
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(
	ctx context.Context,
	userID string,
) ([]repository.Organization, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) AddMember(
	ctx context.Context,
	organizationID,
	userID,
	role string,
) error {
	organizationID = strings.TrimSpace(organizationID)
	userID = strings.TrimSpace(userID)
	role = strings.TrimSpace(strings.ToLower(role))

	if organizationID == "" || userID == "" {
		return ErrInvalidMember
	}

	if role == "" {
		role = "member"
	}

	return s.repo.AddMember(
		ctx,
		organizationID,
		userID,
		role,
	)
}

func (s *Service) IsMember(
	ctx context.Context,
	organizationID,
	userID string,
) (bool, error) {
	return s.repo.IsMember(
		ctx,
		organizationID,
		userID,
	)
}

func (s *Service) GetMemberRole(
	ctx context.Context,
	organizationID,
	userID string,
) (string, error) {
	return s.repo.GetMemberRole(
		ctx,
		organizationID,
		userID,
	)
}

func (s *Service) ListMembers(
	ctx context.Context,
	organizationID string,
) ([]repository.Member, error) {
	organizationID = strings.TrimSpace(organizationID)

	if organizationID == "" {
		return nil, ErrInvalidOrganization
	}

	return s.repo.ListMembers(
		ctx,
		organizationID,
	)
}

func (s *Service) RemoveMember(
	ctx context.Context,
	organizationID,
	requesterID,
	targetUserID string,
) error {
	organizationID = strings.TrimSpace(organizationID)
	requesterID = strings.TrimSpace(requesterID)
	targetUserID = strings.TrimSpace(targetUserID)

	if organizationID == "" ||
		requesterID == "" ||
		targetUserID == "" {
		return ErrInvalidMember
	}

	requesterRole, err := s.repo.GetMemberRole(
		ctx,
		organizationID,
		requesterID,
	)
	if err != nil {
		return err
	}

	if requesterRole != "owner" &&
		requesterRole != "admin" {
		return ErrForbidden
	}

	targetRole, err := s.repo.GetMemberRole(
		ctx,
		organizationID,
		targetUserID,
	)
	if err != nil {
		return err
	}

	if targetRole == "" {
		return ErrInvalidMember
	}

	if targetRole == "owner" {
		return ErrForbidden
	}

	return s.repo.RemoveMember(
		ctx,
		organizationID,
		targetUserID,
	)
}
