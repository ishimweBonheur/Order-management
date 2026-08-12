package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ishimweBonheur/order-management/product-service/internal/model"

)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	GetAll(ctx context.Context, filters ProductFilters) ([]model.Product, int, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresProductRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProductRepository(
	db *pgxpool.Pool,
) *PostgresProductRepository {
	return &PostgresProductRepository{
		db: db,
	}
}

func (r *PostgresProductRepository) Create(
	ctx context.Context,
	product *model.Product,
) error {
	query := `
		INSERT INTO products (
			id,
			name,
			description,
			price,
			stock,
			category,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.Category,
		product.CreatedAt,
		product.UpdatedAt,
	)

	return err
}

func (r *PostgresProductRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.Product, error) {
	query := `
		SELECT
			id,
			name,
			description,
			price,
			stock,
			category,
			created_at,
			updated_at
		FROM products
		WHERE id = $1
	`

	var product model.Product

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.Category,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}

		return nil, err
	}

	return &product, nil
}

func (r *PostgresProductRepository) GetAll(
	ctx context.Context,
	filters ProductFilters,
) ([]model.Product, int, error) {

	offset := (filters.Page - 1) * filters.Limit

	args := make([]any, 0)

	where := "WHERE 1=1"

	if filters.Search != "" {
		args = append(args, "%"+filters.Search+"%")

		where += `
			AND (
				name ILIKE $` + strconv.Itoa(len(args)) + `
				OR description ILIKE $` + strconv.Itoa(len(args)) + `
			)
		`
	}

	if filters.Category != "" {
		args = append(args, filters.Category)

		where += `
			AND category = $` + strconv.Itoa(len(args)) + `
		`
	}

	countQuery := `
		SELECT COUNT(*)
		FROM products
		` + where

	var total int

	err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	sortColumn := "created_at"

	switch filters.Sort {
	case "name":
		sortColumn = "name"
	case "price":
		sortColumn = "price"
	case "stock":
		sortColumn = "stock"
	case "created_at":
		sortColumn = "created_at"
	}

	order := "DESC"

	if strings.EqualFold(filters.Order, "asc") {
		order = "ASC"
	}

	args = append(args, filters.Limit)
	limitPosition := len(args)

	args = append(args, offset)
	offsetPosition := len(args)

	query := `
		SELECT
			id,
			name,
			description,
			price,
			stock,
			category,
			created_at,
			updated_at
		FROM products
		` + where + `
		ORDER BY ` + sortColumn + ` ` + order + `
		LIMIT $` + strconv.Itoa(limitPosition) + `
		OFFSET $` + strconv.Itoa(offsetPosition)

	rows, err := r.db.Query(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	products := make([]model.Product, 0)

	for rows.Next() {
		var product model.Product

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.Category,
			&product.CreatedAt,
			&product.UpdatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *PostgresProductRepository) Update(
	ctx context.Context,
	product *model.Product,
) error {
	query := `
		UPDATE products
		SET
			name = $1,
			description = $2,
			price = $3,
			stock = $4,
			category = $5,
			updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.Exec(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.Category,
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *PostgresProductRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	query := `
		DELETE FROM products
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		id,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrProductNotFound
	}

	return nil
}
