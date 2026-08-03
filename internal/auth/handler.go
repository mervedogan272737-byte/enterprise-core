package auth

import (
"encoding/json"
"net/http"
"strings"
"time"

authmiddleware "enterprise-core/api/internal/auth/middleware"
"enterprise-core/api/internal/auth/repository"
"enterprise-core/api/internal/auth/session"
"enterprise-core/api/internal/response"
)

type Handler struct {
Users    *repository.Repository
Auth     *Service
Sessions *session.Store
}

type registerRequest struct {
Email    string `json:"email"`
Password string `json:"password"`
FullName string `json:"full_name"`
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
sessions *session.Store,
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

if req.Email == "" || req.Password == "" {
response.Error(w, http.StatusBadRequest, "email and password are required")
return
}

existing, err := h.Users.FindByEmail(r.Context(), req.Email)
if err != nil {
response.Error(w, http.StatusInternalServerError, "database error")
return
}

if existing != nil {
response.Error(w, http.StatusConflict, "email already registered")
return
}

passwordHash, err := h.Auth.HashPassword(req.Password)
if err != nil {
response.Error(w, http.StatusBadRequest, err.Error())
return
}

u, err := h.Users.CreateUser(
r.Context(),
req.Email,
passwordHash,
req.FullName,
)
if err != nil {
response.Error(w, http.StatusInternalServerError, "failed to create user")
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

u, err := h.Users.FindByEmail(r.Context(), req.Email)
if err != nil || u == nil {
response.Error(w, http.StatusUnauthorized, "invalid email or password")
return
}

if !u.IsActive {
response.Error(w, http.StatusForbidden, "user account is inactive")
return
}

if !h.Auth.CheckPassword(u.PasswordHash, req.Password) {
response.Error(w, http.StatusUnauthorized, "invalid email or password")
return
}

accessTTL := 24 * time.Hour
refreshTTL := 30 * 24 * time.Hour

accessToken, err := h.Auth.GenerateToken(
u.ID,
u.Email,
u.Role,
"access",
accessTTL,
)
if err != nil {
response.Error(w, http.StatusInternalServerError, "failed to generate access token")
return
}

refreshToken, err := h.Auth.GenerateToken(
u.ID,
u.Email,
u.Role,
"refresh",
refreshTTL,
)
if err != nil {
response.Error(w, http.StatusInternalServerError, "failed to generate refresh token")
return
}

if err := h.Sessions.SaveRefreshToken(
r.Context(),
u.ID,
refreshToken,
refreshTTL,
); err != nil {
response.Error(w, http.StatusInternalServerError, "failed to create session")
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
response.Error(w, http.StatusBadRequest, "refresh token is required")
return
}

sessionData, err := h.Sessions.GetRefreshToken(
r.Context(),
req.RefreshToken,
)
if err != nil {
response.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
return
}

claims, err := h.Auth.ValidateToken(
r.Context(),
req.RefreshToken,
)
if err != nil || claims.Type != "refresh" {
response.Error(w, http.StatusUnauthorized, "invalid refresh token")
return
}

if claims.UserID != sessionData.UserID {
response.Error(w, http.StatusUnauthorized, "invalid session")
return
}

u, err := h.Users.FindByID(
r.Context(),
claims.UserID,
)
if err != nil || u == nil || !u.IsActive {
response.Error(w, http.StatusUnauthorized, "user unavailable")
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
response.Error(w, http.StatusInternalServerError, "failed to generate access token")
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

if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
token := strings.TrimSpace(req.RefreshToken)

if token != "" {
_ = h.Sessions.DeleteRefreshToken(
r.Context(),
token,
)
}
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
response.Error(w, http.StatusUnauthorized, "unauthorized")
return
}

u, err := h.Users.FindByID(
r.Context(),
claims.UserID,
)
if err != nil || u == nil {
response.Error(w, http.StatusUnauthorized, "user not found")
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
