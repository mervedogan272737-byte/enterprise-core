package organization

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/organization/repository"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createOrganizationRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type addMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (h *Handler) currentUserID(r *http.Request) (string, bool) {
	claims := authmiddleware.GetClaims(r)

	if claims == nil || strings.TrimSpace(claims.UserID) == "" {
		return "", false
	}

	return strings.TrimSpace(claims.UserID), true
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createOrganizationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	org, err := h.service.Create(
		r.Context(),
		req.Name,
		req.Slug,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.AddMember(
		r.Context(),
		org.ID,
		userID,
		"owner",
	); err != nil {
		http.Error(
			w,
			"failed to add organization owner",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(org)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	organizations, err := h.service.List(
		r.Context(),
		userID,
	)
	if err != nil {
		http.Error(
			w,
			"failed to list organizations",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(organizations)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	if id == "" {
		http.Error(
			w,
			"invalid organization id",
			http.StatusBadRequest,
		)
		return
	}

	isMember, err := h.service.IsMember(
		r.Context(),
		id,
		userID,
	)
	if err != nil {
		http.Error(
			w,
			"failed to check organization membership",
			http.StatusInternalServerError,
		)
		return
	}

	if !isMember {
		http.Error(
			w,
			"forbidden",
			http.StatusForbidden,
		)
		return
	}

	org, err := h.service.Get(
		r.Context(),
		id,
	)
	if err != nil {
		http.Error(
			w,
			"organization not found",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(org)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	organizationID := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	if organizationID == "" {
		http.Error(
			w,
			"invalid organization id",
			http.StatusBadRequest,
		)
		return
	}

	isMember, err := h.service.IsMember(
		r.Context(),
		organizationID,
		userID,
	)
	if err != nil {
		http.Error(
			w,
			"failed to check organization membership",
			http.StatusInternalServerError,
		)
		return
	}

	if !isMember {
		http.Error(
			w,
			"forbidden",
			http.StatusForbidden,
		)
		return
	}

	members, err := h.service.ListMembers(
		r.Context(),
		organizationID,
	)
	if err != nil {
		http.Error(
			w,
			"failed to list organization members",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(members)
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := h.currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	organizationID := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	if organizationID == "" {
		http.Error(
			w,
			"invalid organization id",
			http.StatusBadRequest,
		)
		return
	}

	requesterRole, err := h.service.GetMemberRole(
		r.Context(),
		organizationID,
		requesterID,
	)
	if err != nil {
		http.Error(
			w,
			"failed to check requester role",
			http.StatusInternalServerError,
		)
		return
	}

	if requesterRole != "owner" && requesterRole != "admin" {
		http.Error(
			w,
			"forbidden",
			http.StatusForbidden,
		)
		return
	}

	var req addMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	targetUserID := strings.TrimSpace(req.UserID)
	role := strings.TrimSpace(
		strings.ToLower(req.Role),
	)

	if targetUserID == "" {
		http.Error(
			w,
			"invalid user id",
			http.StatusBadRequest,
		)
		return
	}

	if role == "" {
		role = "member"
	}

	if role != "member" && role != "admin" {
		http.Error(
			w,
			"invalid member role",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.service.AddMember(
		r.Context(),
		organizationID,
		targetUserID,
		role,
	); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(
				w,
				"user not found",
				http.StatusNotFound,
			)
			return
		}

		http.Error(
			w,
			"failed to add organization member",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := h.currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	organizationID := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	targetUserID := strings.TrimSpace(
		chi.URLParam(r, "userID"),
	)

	if organizationID == "" || targetUserID == "" {
		http.Error(
			w,
			"invalid member information",
			http.StatusBadRequest,
		)
		return
	}

	err := h.service.RemoveMember(
		r.Context(),
		organizationID,
		requesterID,
		targetUserID,
	)

	if err != nil {
		if errors.Is(err, ErrForbidden) {
			http.Error(
				w,
				"forbidden",
				http.StatusForbidden,
			)
			return
		}

		if errors.Is(err, ErrInvalidMember) {
			http.Error(
				w,
				"invalid member",
				http.StatusBadRequest,
			)
			return
		}

		http.Error(
			w,
			"failed to remove organization member",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
