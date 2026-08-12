package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/product-service/internal/repository"
	"github.com/ishimweBonheur/order-management/product-service/internal/service"

)

type ProductHandler struct {
	service *service.ProductService
}

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	product, err := h.service.CreateProduct(
		r.Context(),
		req.Name,
		req.Description,
		req.Price,
		req.Stock,
		req.Category,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) GetProducts(
	w http.ResponseWriter,
	r *http.Request,
) {
	filters := parseProductFilters(r)

	products, total, err := h.service.GetProducts(
		r.Context(),
		filters,
	)

	if err != nil {
		WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"An internal server error occurred",
		)
		return
	}

	totalPages := 0

	if filters.Limit > 0 {
		totalPages = (total + filters.Limit - 1) / filters.Limit
	}

	response := map[string]interface{}{
		"data": products,
		"pagination": map[string]interface{}{
			"page":        filters.Page,
			"limit":       filters.Limit,
			"total":       total,
			"total_pages": totalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to get product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)
	if err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_PRODUCT_ID",
			"Product ID is invalid",
		)
		return
	}

	var request UpdateProductRequest

	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST_BODY",
			"Request body is invalid",
		)
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			WriteError(
				w,
				http.StatusNotFound,
				"PRODUCT_NOT_FOUND",
				"Product was not found",
			)
			return
		}

		WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"An internal server error occurred",
		)
		return
	}

	product.Name = request.Name
	product.Description = request.Description
	product.Price = request.Price
	product.Stock = request.Stock
	product.Category = request.Category

	err = h.service.UpdateProduct(r.Context(), product)
	if err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"PRODUCT_UPDATE_FAILED",
			err.Error(),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")

	id, err := uuid.Parse(idString)
	if err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_PRODUCT_ID",
			"Product ID is invalid",
		)
		return
	}

	err = h.service.DeleteProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			WriteError(
				w,
				http.StatusNotFound,
				"PRODUCT_NOT_FOUND",
				"Product was not found",
			)
			return
		}

		WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"An internal server error occurred",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseProductFilters(r *http.Request) repository.ProductFilters {
	query := r.URL.Query()

	page := 1
	limit := 20

	if value := query.Get("page"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			page = parsed
		}
	}

	if value := query.Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			limit = parsed
		}
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	return repository.ProductFilters{
		Page:     page,
		Limit:    limit,
		Search:   query.Get("search"),
		Category: query.Get("category"),
		Sort:     query.Get("sort"),
		Order:    query.Get("order"),
	}
}