package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ishimweBonheur/order-management/product-service/internal/model"

)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	GetAll(ctx context.Context) ([]model.Product, error)
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
) ([]model.Product, error) {
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
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
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