package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/product-service/internal/model"
	"github.com/redis/go-redis/v9"
)

func newTestCache(t *testing.T) (*ProductCache, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	c := NewProductCache(client, 5*time.Minute)

	return c, mr
}

func newTestProduct() *model.Product {
	now := time.Now()

	return &model.Product{
		ID:          uuid.New(),
		Name:        "Water",
		Description: "Mineral water",
		Price:       1.50,
		Stock:       100,
		Category:    "drinks",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestProductCacheSetAndGet(t *testing.T) {
	c, _ := newTestCache(t)

	product := newTestProduct()

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

	if got.Description != product.Description {
		t.Errorf("expected description %s, got %s", product.Description, got.Description)
	}

	if got.Price != product.Price {
		t.Errorf("expected price %f, got %f", product.Price, got.Price)
	}

	if got.Stock != product.Stock {
		t.Errorf("expected stock %d, got %d", product.Stock, got.Stock)
	}

	if got.Category != product.Category {
		t.Errorf("expected category %s, got %s", product.Category, got.Category)
	}
}

func TestProductCacheGetMissing(t *testing.T) {
	c, _ := newTestCache(t)

	_, err := c.Get(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestProductCacheDelete(t *testing.T) {
	c, _ := newTestCache(t)

	product := newTestProduct()

	if err := c.Set(context.Background(), product); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	if err := c.Delete(context.Background(), product.ID); err != nil {
		t.Fatalf("failed to delete cache: %v", err)
	}

	_, err := c.Get(context.Background(), product.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestProductCacheTTLExpiry(t *testing.T) {
	c, mr := newTestCache(t)

	product := newTestProduct()

	if err := c.Set(context.Background(), product); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	mr.FastForward(6 * time.Minute)

	_, err := c.Get(context.Background(), product.ID)
	if err == nil {
		t.Fatal("expected cache miss after TTL expiry, got hit")
	}
}

func TestProductCacheOverwrite(t *testing.T) {
	c, _ := newTestCache(t)

	product := newTestProduct()

	if err := c.Set(context.Background(), product); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	product.Name = "Sparkling Water"
	product.Price = 2.00

	if err := c.Set(context.Background(), product); err != nil {
		t.Fatalf("failed to update cache: %v", err)
	}

	got, err := c.Get(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("expected cache hit, got %v", err)
	}

	if got.Name != "Sparkling Water" {
		t.Errorf("expected name Sparkling Water, got %s", got.Name)
	}

	if got.Price != 2.00 {
		t.Errorf("expected price 2.00, got %f", got.Price)
	}
}

func TestProductCacheKeyFormat(t *testing.T) {
	c, mr := newTestCache(t)

	product := newTestProduct()

	if err := c.Set(context.Background(), product); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	expectedKey := "product:" + product.ID.String()

	if !mr.Exists(expectedKey) {
		t.Errorf("expected key %s to exist in Redis", expectedKey)
	}
}
