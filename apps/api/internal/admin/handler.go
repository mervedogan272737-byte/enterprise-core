package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/auth/repository"
	"enterprise-core/api/internal/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var (
	ErrCannotDeleteSelf     = errors.New("cannot delete current admin user")
	ErrCannotDeactivateSelf = errors.New("cannot deactivate current admin user")
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
	claims, ok := authmiddleware.ClaimsFromContext(r.Context())
	if !ok || claims == nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
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

	response.JSON(
		w,
		http.StatusOK,
		users,
	)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUserID(w, r)
	if !ok {
		return
	}

	user, err := h.Users.FindByID(
		r.Context(),
		id.String(),
	)
	if err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to retrieve user",
		)
		return
	}

	if user == nil {
		response.Error(
			w,
			http.StatusNotFound,
			"user not found",
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		user,
	)
}

func (h *Handler) SetUserActive(
	w http.ResponseWriter,
	r *http.Request,
) {
	targetID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	claims, ok := authmiddleware.ClaimsFromContext(r.Context())
	if !ok || claims == nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	currentUserID, err := uuid.Parse(
		strings.TrimSpace(claims.UserID),
	)
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"invalid authenticated user",
		)
		return
	}

	var req struct {
		Active *bool `json:"active"`
	}

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&req); err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	if req.Active == nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"active field is required",
		)
		return
	}

	if currentUserID == targetID && !*req.Active {
		response.Error(
			w,
			http.StatusConflict,
			"cannot deactivate your own admin account",
		)
		return
	}

	if err := h.Users.SetActive(
		r.Context(),
		targetID.String(),
		*req.Active,
	); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			response.Error(
				w,
				http.StatusNotFound,
				"user not found",
			)
			return
		}

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
			"user_id": targetID.String(),
			"active":  *req.Active,
		},
	)
}

func (h *Handler) DeleteUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	targetID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	claims, ok := authmiddleware.ClaimsFromContext(r.Context())
	if !ok || claims == nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	currentUserID, err := uuid.Parse(
		strings.TrimSpace(claims.UserID),
	)
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"invalid authenticated user",
		)
		return
	}

	if currentUserID == targetID {
		response.Error(
			w,
			http.StatusConflict,
			"cannot delete your own admin account",
		)
		return
	}

	if err := h.Users.DeleteUser(
		r.Context(),
		targetID.String(),
	); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			response.Error(
				w,
				http.StatusNotFound,
				"user not found",
			)
			return
		}

		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to delete user",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseUserID(
	w http.ResponseWriter,
	r *http.Request,
) (uuid.UUID, bool) {
	rawID := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	if rawID == "" {
		response.Error(
			w,
			http.StatusBadRequest,
			"user id is required",
		)
		return uuid.Nil, false
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid user id",
		)
		return uuid.Nil, false
	}

	return id, true
}
