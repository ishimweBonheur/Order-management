package repository

import (
	"context"
	"errors"
	"sort"
	"strings"
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

func (r *MemoryProductRepository) GetAll(
	ctx context.Context,
	filters ProductFilters,
) ([]model.Product, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]model.Product, 0, len(r.products))

	for _, product := range r.products {
		products = append(products, product)
	}

	if filters.Search != "" {
		search := strings.ToLower(filters.Search)

		filtered := make([]model.Product, 0, len(products))

		for _, product := range products {
			if strings.Contains(strings.ToLower(product.Name), search) ||
				strings.Contains(strings.ToLower(product.Description), search) {
				filtered = append(filtered, product)
			}
		}

		products = filtered
	}

	if filters.Category != "" {
		filtered := make([]model.Product, 0, len(products))

		for _, product := range products {
			if product.Category == filters.Category {
				filtered = append(filtered, product)
			}
		}

		products = filtered
	}

	switch filters.Sort {
	case "name":
		sort.Slice(products, func(i, j int) bool {
			return products[i].Name < products[j].Name
		})
	case "price":
		sort.Slice(products, func(i, j int) bool {
			return products[i].Price < products[j].Price
		})
	case "stock":
		sort.Slice(products, func(i, j int) bool {
			return products[i].Stock < products[j].Stock
		})
	default:
		sort.Slice(products, func(i, j int) bool {
			return products[i].CreatedAt.After(products[j].CreatedAt)
		})
	}

	if strings.EqualFold(filters.Order, "asc") {
		// Reverse for ascending order
		for i, j := 0, len(products)-1; i < j; i, j = i+1, j-1 {
			products[i], products[j] = products[j], products[i]
		}
	}

	total := len(products)

	offset := (filters.Page - 1) * filters.Limit

	if offset > total {
		offset = total
	}

	end := offset + filters.Limit

	if end > total {
		end = total
	}

	if offset < 0 {
		offset = 0
	}

	return products[offset:end], total, nil
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
