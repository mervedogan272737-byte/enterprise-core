package handler

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
			http.Error(w, "invalid email", http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidPassword):
			http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		case errors.Is(err, service.ErrUserAlreadyExists):
			http.Error(w, "user already exists", http.StatusConflict)
		default:
			http.Error(w, "registration failed", http.StatusInternalServerError)
		}
		return
	}

	response := registerResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(response)
}
