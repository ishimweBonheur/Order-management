package handler

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/auth-service/internal/platform/api"
	sharedauth "github.com/ishimweBonheur/order-management/auth-service/internal/platform/auth"
	"github.com/ishimweBonheur/order-management/auth-service/internal/repository"
	"github.com/ishimweBonheur/order-management/auth-service/internal/service"
	"net/http"
)

type Handler struct{ service *service.AuthService }

func New(s *service.AuthService) *Handler { return &Handler{service: s} }

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !api.DecodeJSON(w, r, &req) {
		return
	}
	u, err := h.service.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		code := "REGISTRATION_FAILED"
		message := "Registration failed"
		if errors.Is(err, repository.ErrEmailExists) {
			code = "EMAIL_EXISTS"
			message = "Email is already registered"
		}
		api.Error(w, http.StatusBadRequest, code, message)
		return
	}
	api.JSON(w, http.StatusCreated, u)
}
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !api.DecodeJSON(w, r, &req) {
		return
	}
	token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		api.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email or password is incorrect")
		return
	}
	api.JSON(w, http.StatusOK, map[string]string{"access_token": token, "token_type": "Bearer"})
}
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	idText, ok := sharedauth.UserID(r.Context())
	id, err := uuid.Parse(idText)
	if !ok || err != nil {
		api.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required")
		return
	}
	u, err := h.service.Me(r.Context(), id)
	if err != nil {
		api.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "User was not found")
		return
	}
	api.JSON(w, http.StatusOK, u)
}
func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.Error(w, http.StatusBadRequest, "INVALID_USER_ID", "User ID is invalid")
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if !api.DecodeJSON(w, r, &req) {
		return
	}
	u, err := h.service.AssignRole(r.Context(), id, req.Role)
	if errors.Is(err, service.ErrInvalidInput) {
		api.Error(w, http.StatusBadRequest, "INVALID_ROLE", "Role must be customer or admin")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		api.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "User was not found")
		return
	}
	if errors.Is(err, repository.ErrAdminExists) {
		api.Error(w, http.StatusConflict, "ADMIN_EXISTS", "Only one admin is allowed")
		return
	}
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An internal error occurred")
		return
	}
	api.JSON(w, http.StatusOK, u)
}
func Health(w http.ResponseWriter, _ *http.Request) {
	api.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
