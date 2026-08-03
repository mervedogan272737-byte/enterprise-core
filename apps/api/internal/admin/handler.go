package admin

import (
	"encoding/json"
	"net/http"

	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/auth/repository"
	"enterprise-core/api/internal/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Users *repository.Repository
}

func NewHandler(users *repository.Repository) *Handler {
	return &Handler{
		Users: users,
	}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := authmiddleware.GetClaims(r)

	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"admin":   true,
			"user_id": claims.UserID,
			"email":   claims.Email,
			"role":    claims.Role,
		},
	)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Users.ListUsers(r.Context())
	if err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to list users",
		)
		return
	}

	response.JSON(w, http.StatusOK, users)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.Users.FindByID(
		r.Context(),
		id,
	)

	if err != nil || user == nil {
		response.Error(
			w,
			http.StatusNotFound,
			"user not found",
		)
		return
	}

	response.JSON(w, http.StatusOK, user)
}

func (h *Handler) SetUserActive(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")

	var req struct {
		Active bool `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	if err := h.Users.SetActive(
		r.Context(),
		id,
		req.Active,
	); err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to update user",
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"status":  "ok",
			"user_id": id,
			"active":  req.Active,
		},
	)
}

func (h *Handler) DeleteUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")

	if err := h.Users.DeleteUser(
		r.Context(),
		id,
	); err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to delete user",
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"status":  "ok",
			"user_id": id,
		},
	)
}
