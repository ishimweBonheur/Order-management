package handler

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/order-service/internal/model"
	"github.com/ishimweBonheur/order-management/order-service/internal/platform/api"
	sharedauth "github.com/ishimweBonheur/order-management/order-service/internal/platform/auth"
	"github.com/ishimweBonheur/order-management/order-service/internal/repository"
	"github.com/ishimweBonheur/order-management/order-service/internal/service"
	"net/http"
	"strconv"
)

type Handler struct{ s *service.Service }

func New(s *service.Service) *Handler { return &Handler{s: s} }
func identity(r *http.Request) (uuid.UUID, string, bool) {
	text, ok := sharedauth.UserID(r.Context())
	if !ok {
		return uuid.Nil, "", false
	}
	id, err := uuid.Parse(text)
	role, _ := sharedauth.Role(r.Context())
	return id, role, err == nil
}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		} `json:"items"`
	}
	if !api.DecodeJSON(w, r, &req) {
		return
	}
	items := make([]model.CreateItem, 0, len(req.Items))
	for index, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			api.Error(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", fmt.Sprintf("items[%d].product_id is not a valid UUID", index))
			return
		}
		items = append(items, model.CreateItem{ProductID: productID, Quantity: item.Quantity})
	}
	uid, _, ok := identity(r)
	if !ok {
		api.Error(w, 401, "UNAUTHORIZED", "Authentication is required")
		return
	}
	o, err := h.s.Create(r.Context(), uid, items)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrder) || errors.Is(err, repository.ErrInsufficientStock) {
			api.Error(w, http.StatusBadRequest, "ORDER_CREATION_FAILED", "Order items or available stock are invalid")
		} else {
			api.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An internal error occurred")
		}
		return
	}
	api.JSON(w, 201, o)
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := identity(r)
	if !ok {
		api.Error(w, 401, "UNAUTHORIZED", "Authentication is required")
		return
	}
	var filter *uuid.UUID
	if role != "admin" {
		filter = &uid
	}
	page, limit := 1, 20
	if value, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && value > 0 {
		page = value
	}
	if value, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && value > 0 {
		limit = value
	}
	if limit > 100 {
		limit = 100
	}
	orders, total, err := h.s.List(r.Context(), filter, page, limit)
	if err != nil {
		api.Error(w, 500, "INTERNAL_SERVER_ERROR", "An internal error occurred")
		return
	}
	api.JSON(w, 200, map[string]any{"data": orders, "pagination": map[string]int{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	uid, role, ok := identity(r)
	if err != nil || !ok {
		api.Error(w, 400, "INVALID_ORDER_ID", "Order ID is invalid")
		return
	}
	o, err := h.s.Get(r.Context(), id, uid, role == "admin")
	if errors.Is(err, repository.ErrNotFound) {
		api.Error(w, 404, "ORDER_NOT_FOUND", "Order was not found")
		return
	}
	if err != nil {
		api.Error(w, 500, "INTERNAL_SERVER_ERROR", "An internal error occurred")
		return
	}
	api.JSON(w, 200, o)
}
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	var req struct {
		Status string `json:"status"`
	}
	if err != nil || !api.DecodeJSON(w, r, &req) {
		return
	}
	o, err := h.s.Status(r.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrder) {
			api.Error(w, http.StatusBadRequest, "ORDER_STATUS_UPDATE_FAILED", "Order status is invalid")
		} else if errors.Is(err, repository.ErrNotFound) {
			api.Error(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Order was not found")
		} else {
			api.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An internal error occurred")
		}
		return
	}
	api.JSON(w, 200, o)
}
func Health(w http.ResponseWriter, _ *http.Request) {
	api.JSON(w, 200, map[string]string{"status": "ok"})
}
