package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/product-service/internal/model"
)

var ErrProductNotFound = errors.New("product not found")

type MemoryProductRepository struct {
	mu       sync.RWMutex
	products map[uuid.UUID]model.Product
}

func NewMemoryProductRepository() *MemoryProductRepository {
	return &MemoryProductRepository{
		products: make(map[uuid.UUID]model.Product),
	}
}

func (r *MemoryProductRepository) Create(ctx context.Context, product *model.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[product.ID] = *product

	return nil
}

func (r *MemoryProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, exists := r.products[id]
	if !exists {
		return nil, ErrProductNotFound
	}

	return &product, nil
}

func (r *MemoryProductRepository) GetAll(ctx context.Context) ([]model.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]model.Product, 0, len(r.products))

	for _, product := range r.products {
		products = append(products, product)
	}

	return products, nil
}

func (r *MemoryProductRepository) Update(ctx context.Context, product *model.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[product.ID]; !exists {
		return ErrProductNotFound
	}

	r.products[product.ID] = *product

	return nil
}

func (r *MemoryProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[id]; !exists {
		return ErrProductNotFound
	}

	delete(r.products, id)

	return nil
}
