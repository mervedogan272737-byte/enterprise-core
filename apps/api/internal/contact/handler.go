package contact

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/organization"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service             *Service
	organizationService *organization.Service
}

func NewHandler(
	service *Service,
	organizationService *organization.Service,
) *Handler {
	return &Handler{
		service:             service,
		organizationService: organizationService,
	}
}

type contactRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Position  string `json:"position"`
	Notes     string `json:"notes"`
	Status    string `json:"status"`
}

func (h *Handler) authorize(
	r *http.Request,
) (string, bool) {
	claims := authmiddleware.GetClaims(r)

	if claims == nil ||
		strings.TrimSpace(claims.UserID) == "" {
		return "", false
	}

	organizationID := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	if organizationID == "" {
		return "", false
	}

	isMember, err := h.organizationService.IsMember(
		r.Context(),
		organizationID,
		strings.TrimSpace(claims.UserID),
	)

	if err != nil || !isMember {
		return "", false
	}

	return organizationID, true
}

func (h *Handler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	organizationID, ok := h.authorize(r)

	if !ok {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	customerID := strings.TrimSpace(
		chi.URLParam(r, "customerID"),
	)

	var req contactRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	contact, err := h.service.Create(
		r.Context(),
		organizationID,
		customerID,
		req.FirstName,
		req.LastName,
		req.Email,
		req.Phone,
		req.Position,
		req.Notes,
	)

	if err != nil {
		if errors.Is(err, ErrInvalidContact) {
			http.Error(w, "invalid contact", http.StatusBadRequest)
			return
		}

		http.Error(
			w,
			"failed to create contact",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(contact)
}

func (h *Handler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	organizationID, ok := h.authorize(r)

	if !ok {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	contacts, err := h.service.List(
		r.Context(),
		organizationID,
		chi.URLParam(r, "customerID"),
	)

	if err != nil {
		http.Error(
			w,
			"failed to list contacts",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(contacts)
}

func (h *Handler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	organizationID, ok := h.authorize(r)

	if !ok {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	contact, err := h.service.Get(
		r.Context(),
		organizationID,
		chi.URLParam(r, "customerID"),
		chi.URLParam(r, "contactID"),
	)

	if err != nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(contact)
}

func (h *Handler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	organizationID, ok := h.authorize(r)

	if !ok {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req contactRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	contact, err := h.service.Update(
		r.Context(),
		organizationID,
		chi.URLParam(r, "customerID"),
		chi.URLParam(r, "contactID"),
		req.FirstName,
		req.LastName,
		req.Email,
		req.Phone,
		req.Position,
		req.Notes,
		req.Status,
	)

	if err != nil {
		if errors.Is(err, ErrInvalidContact) {
			http.Error(w, "invalid contact", http.StatusBadRequest)
			return
		}

		http.Error(
			w,
			"failed to update contact",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(contact)
}

func (h *Handler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	organizationID, ok := h.authorize(r)

	if !ok {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	err := h.service.Delete(
		r.Context(),
		organizationID,
		chi.URLParam(r, "customerID"),
		chi.URLParam(r, "contactID"),
	)

	if err != nil {
		http.Error(
			w,
			"failed to delete contact",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
