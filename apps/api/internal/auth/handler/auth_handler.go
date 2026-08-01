package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/auth/service"
)

type AuthHandler struct {
	Service *service.Service
}

func NewAuthHandler(authService *service.Service) *AuthHandler {
	return &AuthHandler{
		Service: authService,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type registerResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	Role        string `json:"role"`
	AccessToken string `json:"access_token"`
}

type meResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

func (h *AuthHandler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	user, err := h.Service.Register(
		r.Context(),
		service.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
			FullName: req.FullName,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			http.Error(
				w,
				"invalid email",
				http.StatusBadRequest,
			)

		case errors.Is(err, service.ErrInvalidPassword):
			http.Error(
				w,
				"password must be at least 8 characters",
				http.StatusBadRequest,
			)

		case errors.Is(err, service.ErrUserAlreadyExists):
			http.Error(
				w,
				"user already exists",
				http.StatusConflict,
			)

		default:
			http.Error(
				w,
				"registration failed",
				http.StatusInternalServerError,
			)
		}

		return
	}

	response := registerResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	user, accessToken, err := h.Service.Login(
		r.Context(),
		service.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			http.Error(
				w,
				"invalid email or password",
				http.StatusUnauthorized,
			)

		case errors.Is(err, service.ErrTokenGeneration):
			http.Error(
				w,
				"token generation failed",
				http.StatusInternalServerError,
			)

		default:
			http.Error(
				w,
				"login failed",
				http.StatusInternalServerError,
			)
		}

		return
	}

	response := loginResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		Role:        user.Role,
		AccessToken: accessToken,
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, ok := middleware.ClaimsFromContext(
		r.Context(),
	)

	if !ok {
		http.Error(
			w,
			"unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	response := meResponse{
		ID:       claims.UserID,
		Email:    claims.Email,
		Role:     claims.Role,
		FullName: claims.FullName,
	}

	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response)
}
