package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/product-service/internal/model"
	"github.com/ishimweBonheur/order-management/product-service/internal/repository"

)

type ProductService struct {
	repository repository.ProductRepository
}

type createProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
}

func NewProductService(repository repository.ProductRepository) *ProductService {
	return &ProductService{
		repository: repository,
	}
}

func (s *ProductService) CreateProduct(
	ctx context.Context,
	name string,
	description string,
	price float64,
	stock int,
	category string,
) (*model.Product, error) {

	name = strings.TrimSpace(name)
	category = strings.TrimSpace(category)

	if name == "" {
		return nil, errors.New("product name is required")
	}

	if price < 0 {
		return nil, errors.New("product price cannot be negative")
	}

	if stock < 0 {
		return nil, errors.New("product stock cannot be negative")
	}

	now := time.Now()

	product := &model.Product{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Category:    category,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.repository.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetProduct(
	ctx context.Context,
	id uuid.UUID,
) (*model.Product, error) {

	return s.repository.GetByID(ctx, id)
}

func (s *ProductService) GetProducts(
	ctx context.Context, filters repository.ProductFilters,
) ([]model.Product, int, error) {

	return s.repository.GetAll(ctx, filters)
}

func (s *ProductService) UpdateProduct(
	ctx context.Context,
	product *model.Product,
) error {

	product.Name = strings.TrimSpace(product.Name)
	product.Category = strings.TrimSpace(product.Category)

	if product.Name == "" {
		return errors.New("product name is required")
	}

	if product.Price < 0 {
		return errors.New("product price cannot be negative")
	}

	if product.Stock < 0 {
		return errors.New("product stock cannot be negative")
	}

	product.UpdatedAt = time.Now()

	return s.repository.Update(ctx, product)
} 

func (s *ProductService) DeleteProduct(
	ctx context.Context,
	id uuid.UUID,
) error {

	return s.repository.Delete(ctx, id)
}
