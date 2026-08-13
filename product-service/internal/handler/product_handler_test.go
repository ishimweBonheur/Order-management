package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/product-service/internal/cache"
	"github.com/ishimweBonheur/order-management/product-service/internal/repository"
	"github.com/ishimweBonheur/order-management/product-service/internal/service"
	"github.com/redis/go-redis/v9"
)

func newTestHandler(t *testing.T) *ProductHandler {
	t.Helper()

	repo := repository.NewMemoryProductRepository()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	c := cache.NewProductCache(client, 5*time.Minute)

	srv := service.NewProductService(repo, c, nil)

	return NewProductHandler(srv)
}

func newTestRouter(t *testing.T, h *ProductHandler) *chi.Mux {
	t.Helper()

	router := chi.NewRouter()

	router.Route("/products", func(r chi.Router) {
		r.Post("/", h.CreateProduct)
		r.Get("/", h.GetProducts)
		r.Get("/{id}", h.GetProduct)
		r.Put("/{id}", h.UpdateProduct)
		r.Delete("/{id}", h.DeleteProduct)
	})

	return router
}

func TestCreateProductHandler(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	body := map[string]interface{}{
		"name":        "Water",
		"description": "Mineral water",
		"price":       1.50,
		"stock":       100,
		"category":    "drinks",
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		"/products",
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var product map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &product); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if product["name"] != "Water" {
		t.Errorf("expected name Water, got %v", product["name"])
	}

	if product["price"] != 1.50 {
		t.Errorf("expected price 1.50, got %v", product["price"])
	}
}

func TestCreateProductHandlerInvalidBody(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	req := httptest.NewRequest(
		http.MethodPost,
		"/products",
		bytes.NewReader([]byte("{invalid")),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateProductHandlerValidationError(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	body := map[string]interface{}{
		"name":        "",
		"description": "Mineral water",
		"price":       1.50,
		"stock":       100,
		"category":    "drinks",
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		"/products",
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetProductHandler(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	// Create a product first
	product, err := h.service.CreateProduct(
		context.Background(),
		"Water",
		"Mineral water",
		1.50,
		100,
		"drinks",
	)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/"+product.ID.String(),
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got["id"] != product.ID.String() {
		t.Errorf("expected id %s, got %v", product.ID.String(), got["id"])
	}
}

func TestGetProductHandlerNotFound(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/"+uuid.New().String(),
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetProductHandlerInvalidID(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/invalid-id",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetProductsHandler(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	for i := 0; i < 5; i++ {
		_, err := h.service.CreateProduct(
			context.Background(),
			"Product "+string(rune('A'+i)),
			"description",
			float64(i+1),
			10,
			"drinks",
		)
		if err != nil {
			t.Fatalf("failed to create product: %v", err)
		}
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/products?page=1&limit=2",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		Data       []map[string]interface{} `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			Limit      int `json:"limit"`
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
		} `json:"pagination"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Data) != 2 {
		t.Errorf("expected 2 products, got %d", len(response.Data))
	}

	if response.Pagination.Total != 5 {
		t.Errorf("expected total 5, got %d", response.Pagination.Total)
	}

	if response.Pagination.TotalPages != 3 {
		t.Errorf("expected total_pages 3, got %d", response.Pagination.TotalPages)
	}
}

func TestGetProductsHandlerFilters(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	_, err := h.service.CreateProduct(
		context.Background(),
		"Water",
		"Mineral water",
		1.50,
		100,
		"drinks",
	)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	_, err = h.service.CreateProduct(
		context.Background(),
		"Chips",
		"Potato chips",
		2.50,
		50,
		"snacks",
	)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/products?category=drinks",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Data) != 1 {
		t.Errorf("expected 1 product, got %d", len(response.Data))
	}

	if response.Data[0]["category"] != "drinks" {
		t.Errorf("expected category drinks, got %v", response.Data[0]["category"])
	}
}

func TestUpdateProductHandler(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	product, err := h.service.CreateProduct(
		context.Background(),
		"Water",
		"Mineral water",
		1.50,
		100,
		"drinks",
	)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	body := map[string]interface{}{
		"name":        "Sparkling Water",
		"description": "Sparkling mineral water",
		"price":       2.00,
		"stock":       80,
		"category":    "drinks",
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPut,
		"/products/"+product.ID.String(),
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got["name"] != "Sparkling Water" {
		t.Errorf("expected name Sparkling Water, got %v", got["name"])
	}
}

func TestUpdateProductHandlerNotFound(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	body := map[string]interface{}{
		"name":        "Water",
		"description": "Mineral water",
		"price":       1.50,
		"stock":       100,
		"category":    "drinks",
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPut,
		"/products/"+uuid.New().String(),
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestDeleteProductHandler(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	product, err := h.service.CreateProduct(
		context.Background(),
		"Water",
		"Mineral water",
		1.50,
		100,
		"drinks",
	)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodDelete,
		"/products/"+product.ID.String(),
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestDeleteProductHandlerNotFound(t *testing.T) {
	h := newTestHandler(t)
	router := newTestRouter(t, h)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/products/"+uuid.New().String(),
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestParseProductFilters(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/products?page=2&limit=50&search=water&category=drinks&sort=price&order=asc",
		nil,
	)

	filters := parseProductFilters(req)

	if filters.Page != 2 {
		t.Errorf("expected page 2, got %d", filters.Page)
	}

	if filters.Limit != 50 {
		t.Errorf("expected limit 50, got %d", filters.Limit)
	}

	if filters.Search != "water" {
		t.Errorf("expected search water, got %s", filters.Search)
	}

	if filters.Category != "drinks" {
		t.Errorf("expected category drinks, got %s", filters.Category)
	}

	if filters.Sort != "price" {
		t.Errorf("expected sort price, got %s", filters.Sort)
	}

	if filters.Order != "asc" {
		t.Errorf("expected order asc, got %s", filters.Order)
	}
}

func TestParseProductFiltersDefaults(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/products",
		nil,
	)

	filters := parseProductFilters(req)

	if filters.Page != 1 {
		t.Errorf("expected default page 1, got %d", filters.Page)
	}

	if filters.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", filters.Limit)
	}
}

func TestParseProductFiltersInvalidValues(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/products?page=0&limit=0",
		nil,
	)

	filters := parseProductFilters(req)

	if filters.Page != 1 {
		t.Errorf("expected page reset to 1, got %d", filters.Page)
	}

	if filters.Limit != 20 {
		t.Errorf("expected limit reset to 20, got %d", filters.Limit)
	}
}

func TestParseProductFiltersMaxLimit(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/products?limit=1000",
		nil,
	)

	filters := parseProductFilters(req)

	if filters.Limit != 100 {
		t.Errorf("expected limit capped at 100, got %d", filters.Limit)
	}
}
