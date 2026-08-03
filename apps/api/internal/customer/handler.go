package customer

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

type customerRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Company string `json:"company"`
	Notes   string `json:"notes"`
	Status  string `json:"status"`
}

func (h *Handler) currentUserID(r *http.Request) (string, bool) {
	claims := authmiddleware.GetClaims(r)

	if claims == nil || strings.TrimSpace(claims.UserID) == "" {
		return "", false
	}

	return strings.TrimSpace(claims.UserID), true
}

func (h *Handler) authorizeOrganization(
	r *http.Request,
) (string, bool) {
	userID, ok := h.currentUserID(r)
	if !ok {
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
		userID,
	)

	if err != nil || !isMember {
		return "", false
	}

	return organizationID, true
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := h.authorizeOrganization(r)
	if !ok {
		http.Error(
			w,
			"forbidden",
			http.StatusForbidden,
		)
		return
	}

	var req customerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	customer, err := h.service.Create(
		r.Context(),
		organizationID,
		req.Name,
		req.Email,
		req.Phone,
		req.Company,
		req.Notes,
	)

	if err != nil {
		if errors.Is(err, ErrInvalidCustomer) {
			http.Error(
				w,
				"invalid customer",
				http.StatusBadRequest,
			)
			return
		}

		http.Error(
			w,
			"failed to create customer",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(customer)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := h.authorizeOrganization(r)
	if !ok {
		http.Error(
			w,
			"forbidden",
			http.StatusForbidden,
		)
		return
	}

	customers, err := h.service.List(
		r.Context(),
		organizationID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to list customers",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(customers)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := h.authorizeOrganization(r)
	if !ok {
		http.Error(
			w,
			"forbidden",
			http.StatusForbidden,
		)
		return
	}

	customerID := strings.TrimSpace(
		chi.URLParam(r, "customerID"),
	)

	if customerID == "" {
		http.Error(
			w,
			"invalid customer id",
			http.StatusBadRequest,
		)
		return
	}

	customer, err := h.service.Get(
		r.Context(),
		organizationID,
		customerID,
	)

	if err != nil {
		http.Error(
			w,
			"customer not found",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(customer)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := h.authorizeOrganization(r)
	if !ok {
		http.Error(
			w,
			"forbidden",
			http.StatusForbidden,
		)
		return
	}

	customerID := strings.TrimSpace(
		chi.URLParam(r, "customerID"),
	)

	if customerID == "" {
		http.Error(
			w,
			"invalid customer id",
			http.StatusBadRequest,
		)
		return
	}

	var req customerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	customer, err := h.service.Update(
		r.Context(),
		organizationID,
		customerID,
		req.Name,
		req.Email,
		req.Phone,
		req.Company,
		req.Notes,
		req.Status,
	)

	if err != nil {
		if errors.Is(err, ErrInvalidCustomer) {
			http.Error(
				w,
				"invalid customer",
				http.StatusBadRequest,
			)
			return
		}

		http.Error(
			w,
			"failed to update customer",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(customer)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := h.authorizeOrganization(r)
	if !ok {
		http.Error(
			w,
			"forbidden",
			http.StatusForbidden,
		)
		return
	}

	customerID := strings.TrimSpace(
		chi.URLParam(r, "customerID"),
	)

	if customerID == "" {
		http.Error(
			w,
			"invalid customer id",
			http.StatusBadRequest,
		)
		return
	}

	if err := h.service.Delete(
		r.Context(),
		organizationID,
		customerID,
	); err != nil {
		http.Error(
			w,
			"failed to delete customer",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
