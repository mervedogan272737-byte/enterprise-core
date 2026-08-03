package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/auth/repository"
	"enterprise-core/api/internal/auth/session"
	"enterprise-core/api/internal/response"

	"time"
)

type Handler struct {
	Users    *repository.Repository
	Auth     *Service
	Sessions *session.Manager
}

type registerRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	FullName         string `json:"full_name"`
	OrganizationName string `json:"organization_name"`
	OrganizationSlug string `json:"organization_slug"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func NewHandler(
	users *repository.Repository,
	authService *Service,
	sessions *session.Manager,
) *Handler {
	return &Handler{
		Users:    users,
		Auth:     authService,
		Sessions: sessions,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)
	req.OrganizationName = strings.TrimSpace(req.OrganizationName)
	req.OrganizationSlug = strings.ToLower(
		strings.TrimSpace(req.OrganizationSlug),
	)

	if req.Email == "" ||
		req.Password == "" ||
		req.FullName == "" ||
		req.OrganizationName == "" ||
		req.OrganizationSlug == "" {
		response.Error(
			w,
			http.StatusBadRequest,
			"email, password, full_name, organization_name and organization_slug are required",
		)
		return
	}

	if len(req.Password) < 8 {
		response.Error(
			w,
			http.StatusBadRequest,
			"password must be at least 8 characters",
		)
		return
	}

	existing, err := h.Users.FindByEmail(
		r.Context(),
		req.Email,
	)

	if err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			"database error",
		)
		return
	}

	if existing != nil {
		response.Error(
			w,
			http.StatusConflict,
			"email already registered",
		)
		return
	}

	passwordHash, err := h.Auth.HashPassword(
		req.Password,
	)

	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	u, organizationID, err :=
		h.Users.CreateUserWithOrganization(
			r.Context(),
			req.Email,
			passwordHash,
			req.FullName,
			req.OrganizationName,
			req.OrganizationSlug,
		)

	if err != nil {
		response.Error(
			w,
			http.StatusConflict,
			"failed to create account or organization",
		)
		return
	}

	response.JSON(
		w,
		http.StatusCreated,
		map[string]interface{}{
			"user": map[string]interface{}{
				"id":        u.ID,
				"email":     u.Email,
				"full_name": u.FullName,
				"role":      u.Role,
				"is_active": u.IsActive,
			},
			"organization": map[string]interface{}{
				"id":   organizationID,
				"name": req.OrganizationName,
				"slug": req.OrganizationSlug,
				"role": "owner",
			},
		},
	)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	u, err := h.Users.FindByEmail(
		r.Context(),
		req.Email,
	)

	if err != nil || u == nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"invalid email or password",
		)
		return
	}

	if !u.IsActive {
		response.Error(
			w,
			http.StatusForbidden,
			"user account is inactive",
		)
		return
	}

	if !h.Auth.CheckPassword(
		u.PasswordHash,
		req.Password,
	) {
		response.Error(
			w,
			http.StatusUnauthorized,
			"invalid email or password",
		)
		return
	}

	accessTTL := 24 * time.Hour

	accessToken, err := h.Auth.GenerateToken(
		u.ID,
		u.Email,
		u.Role,
		"access",
		accessTTL,
	)

	if err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to generate access token",
		)
		return
	}

	refreshToken, _, err :=
		h.Sessions.CreateSession(
			r.Context(),
			u.ID,
			u.Email,
			u.FullName,
			u.Role,
		)

	if err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to create session",
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
			"expires_in":    int(accessTTL.Seconds()),
			"user": map[string]interface{}{
				"id":        u.ID,
				"email":     u.Email,
				"full_name": u.FullName,
				"role":      u.Role,
				"is_active": u.IsActive,
			},
		},
	)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)

	if req.RefreshToken == "" {
		response.Error(
			w,
			http.StatusBadRequest,
			"refresh token is required",
		)
		return
	}

	sessionData, err := h.Sessions.GetSession(
		r.Context(),
		req.RefreshToken,
	)

	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"invalid or expired refresh token",
		)
		return
	}

	u, err := h.Users.FindByID(
		r.Context(),
		sessionData.UserID,
	)

	if err != nil || u == nil || !u.IsActive {
		response.Error(
			w,
			http.StatusUnauthorized,
			"user unavailable",
		)
		return
	}

	accessTTL := 24 * time.Hour

	accessToken, err := h.Auth.GenerateToken(
		u.ID,
		u.Email,
		u.Role,
		"access",
		accessTTL,
	)

	if err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			"failed to generate access token",
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   int(accessTTL.Seconds()),
		},
	)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err == nil &&
		strings.TrimSpace(req.RefreshToken) != "" {

		_ = h.Sessions.RevokeSession(
			r.Context(),
			strings.TrimSpace(req.RefreshToken),
		)
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]string{
			"message": "logged out successfully",
		},
	)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := authmiddleware.GetClaims(r)

	if claims == nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	u, err := h.Users.FindByID(
		r.Context(),
		claims.UserID,
	)

	if err != nil || u == nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"user not found",
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"user": map[string]interface{}{
				"id":        u.ID,
				"email":     u.Email,
				"full_name": u.FullName,
				"role":      u.Role,
				"is_active": u.IsActive,
			},
		},
	)
}

var _ = errors.New
