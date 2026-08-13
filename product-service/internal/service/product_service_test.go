package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/product-service/internal/cache"
	"github.com/ishimweBonheur/order-management/product-service/internal/model"
	"github.com/ishimweBonheur/order-management/product-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

func newTestService(t *testing.T) (*ProductService, repository.ProductRepository) {
	t.Helper()

	repo := repository.NewMemoryProductRepository()

	srv, err := newServiceWithCache(repo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	return srv, repo
}

func newServiceWithCache(repo repository.ProductRepository) (*ProductService, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	c := cache.NewProductCache(client, 5*time.Minute)

	return NewProductService(repo, c, nil), nil
}

func TestCreateProduct(t *testing.T) {
	srv, repo := newTestService(t)

	product, err := srv.CreateProduct(
		context.Background(),
		"Water",
		"Mineral water",
		1.50,
		100,
		"drinks",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if product.Name != "Water" {
		t.Errorf("expected name Water, got %s", product.Name)
	}

	if product.Price != 1.50 {
		t.Errorf("expected price 1.50, got %f", product.Price)
	}

	if product.Stock != 100 {
		t.Errorf("expected stock 100, got %d", product.Stock)
	}

	if product.Category != "drinks" {
		t.Errorf("expected category drinks, got %s", product.Category)
	}

	if product.ID == uuid.Nil {
		t.Error("expected product to have an ID")
	}

	_, err = repo.GetByID(context.Background(), product.ID)
	if err != nil {
		t.Errorf("expected product to be persisted, got %v", err)
	}
}

func TestCreateProductValidation(t *testing.T) {
	srv, _ := newTestService(t)

	tests := []struct {
		name        string
		productName string
		price       float64
		stock       int
	}{
		{
			name:        "empty name",
			productName: "   ",
			price:       10,
			stock:       5,
		},
		{
			name:        "negative price",
			productName: "Water",
			price:       -1,
			stock:       5,
		},
		{
			name:        "negative stock",
			productName: "Water",
			price:       10,
			stock:       -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := srv.CreateProduct(
				context.Background(),
				tt.productName,
				"description",
				tt.price,
				tt.stock,
				"drinks",
			)

			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestGetProduct(t *testing.T) {
	srv, _ := newTestService(t)

	product, err := srv.CreateProduct(
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

	got, err := srv.GetProduct(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ID != product.ID {
		t.Errorf("expected ID %s, got %s", product.ID, got.ID)
	}

	if got.Name != product.Name {
		t.Errorf("expected name %s, got %s", product.Name, got.Name)
	}
}

func TestGetProductNotFound(t *testing.T) {
	srv, _ := newTestService(t)

	_, err := srv.GetProduct(
		context.Background(),
		uuid.New(),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetProducts(t *testing.T) {
	srv, _ := newTestService(t)

	for i := 0; i < 5; i++ {
		_, err := srv.CreateProduct(
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

	products, total, err := srv.GetProducts(
		context.Background(),
		repository.ProductFilters{
			Page:  1,
			Limit: 2,
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}

	if len(products) != 2 {
		t.Errorf("expected 2 products, got %d", len(products))
	}
}

func TestUpdateProduct(t *testing.T) {
	srv, _ := newTestService(t)

	product, err := srv.CreateProduct(
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

	product.Name = "Sparkling Water"
	product.Price = 2.00

	err = srv.UpdateProduct(context.Background(), product)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := srv.GetProduct(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("failed to get product: %v", err)
	}

	if got.Name != "Sparkling Water" {
		t.Errorf("expected name Sparkling Water, got %s", got.Name)
	}

	if got.Price != 2.00 {
		t.Errorf("expected price 2.00, got %f", got.Price)
	}
}

func TestUpdateProductValidation(t *testing.T) {
	srv, _ := newTestService(t)

	product, err := srv.CreateProduct(
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

	tests := []struct {
		name        string
		productName string
		price       float64
		stock       int
	}{
		{
			name:        "empty name",
			productName: "   ",
			price:       10,
			stock:       5,
		},
		{
			name:        "negative price",
			productName: "Water",
			price:       -1,
			stock:       5,
		},
		{
			name:        "negative stock",
			productName: "Water",
			price:       10,
			stock:       -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product.Name = tt.productName
			product.Price = tt.price
			product.Stock = tt.stock

			err := srv.UpdateProduct(context.Background(), product)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	srv, _ := newTestService(t)

	product, err := srv.CreateProduct(
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

	err = srv.DeleteProduct(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = srv.GetProduct(context.Background(), product.ID)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestCacheHit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	c := cache.NewProductCache(client, 5*time.Minute)

	product := &model.Product{
		ID:        uuid.New(),
		Name:      "Water",
		Price:     1.50,
		Stock:     100,
		Category:  "drinks",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := c.Set(context.Background(), product); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	got, err := c.Get(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("expected cache hit, got %v", err)
	}

	if got.ID != product.ID {
		t.Errorf("expected ID %s, got %s", product.ID, got.ID)
	}

	if got.Name != product.Name {
		t.Errorf("expected name %s, got %s", product.Name, got.Name)
	}
}

func TestCacheMiss(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	c := cache.NewProductCache(client, 5*time.Minute)

	_, err = c.Get(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected cache miss, got hit")
	}
}

func TestCacheTTL(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	c := cache.NewProductCache(client, 1*time.Second)

	product := &model.Product{
		ID:        uuid.New(),
		Name:      "Water",
		Price:     1.50,
		Stock:     100,
		Category:  "drinks",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := c.Set(context.Background(), product); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	mr.FastForward(2 * time.Second)

	_, err = c.Get(context.Background(), product.ID)
	if err == nil {
		t.Fatal("expected cache miss after TTL expiry, got hit")
	}
}

func TestCacheInvalidation(t *testing.T) {
	srv, _ := newTestService(t)

	product, err := srv.CreateProduct(
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

	// First get populates the cache
	if _, err := srv.GetProduct(context.Background(), product.ID); err != nil {
		t.Fatalf("failed to get product: %v", err)
	}

	// Update the product
	product.Name = "Updated Water"
	if err := srv.UpdateProduct(context.Background(), product); err != nil {
		t.Fatalf("failed to update product: %v", err)
	}

	// Get should reflect the update (cache was updated)
	got, err := srv.GetProduct(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("failed to get product: %v", err)
	}

	if got.Name != "Updated Water" {
		t.Errorf("expected updated name, got %s", got.Name)
	}
}
