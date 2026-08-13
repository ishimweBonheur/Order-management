package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/order-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var (
	ErrNotFound          = errors.New("order not found")
	ErrInsufficientStock = errors.New("insufficient product stock")
)

type Repository interface {
	Create(context.Context, uuid.UUID, []model.CreateItem) (*model.Order, error)
	List(context.Context, *uuid.UUID) ([]model.Order, error)
	ByID(context.Context, uuid.UUID) (*model.Order, error)
	UpdateStatus(context.Context, uuid.UUID, string) (*model.Order, error)
}
type Postgres struct{ db *pgxpool.Pool }

func NewPostgres(db *pgxpool.Pool) *Postgres { return &Postgres{db: db} }
func (r *Postgres) Create(ctx context.Context, userID uuid.UUID, input []model.CreateItem) (order *model.Order, err error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	now := time.Now().UTC()
	order = &model.Order{ID: uuid.New(), UserID: userID, Status: model.StatusPending, Items: make([]model.Item, 0, len(input)), CreatedAt: now, UpdatedAt: now}
	for _, requested := range input {
		var price float64
		var stock int
		err = tx.QueryRow(ctx, `SELECT price,stock FROM products WHERE id=$1 FOR UPDATE`, requested.ProductID).Scan(&price, &stock)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: product %s", ErrInsufficientStock, requested.ProductID)
		}
		if err != nil {
			return nil, err
		}
		if requested.Quantity < 1 || stock < requested.Quantity {
			return nil, fmt.Errorf("%w: product %s", ErrInsufficientStock, requested.ProductID)
		}
		item := model.Item{ID: uuid.New(), OrderID: order.ID, ProductID: requested.ProductID, Quantity: requested.Quantity, UnitPrice: price}
		order.Items = append(order.Items, item)
		order.TotalAmount += price * float64(requested.Quantity)
	}
	_, err = tx.Exec(ctx, `INSERT INTO orders(id,user_id,status,total_amount,created_at,updated_at)VALUES($1,$2,$3,$4,$5,$6)`, order.ID, order.UserID, order.Status, order.TotalAmount, now, now)
	if err != nil {
		return nil, err
	}
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `INSERT INTO order_items(id,order_id,product_id,quantity,unit_price)VALUES($1,$2,$3,$4,$5)`, item.ID, item.OrderID, item.ProductID, item.Quantity, item.UnitPrice)
		if err != nil {
			return nil, err
		}
		tag, e := tx.Exec(ctx, `UPDATE products SET stock=stock-$1,updated_at=$2 WHERE id=$3 AND stock >= $1`, item.Quantity, now, item.ProductID)
		if e != nil {
			return nil, e
		}
		if tag.RowsAffected() != 1 {
			return nil, ErrInsufficientStock
		}
	}
	err = tx.Commit(ctx)
	return order, err
}
func (r *Postgres) List(ctx context.Context, userID *uuid.UUID) ([]model.Order, error) {
	q := `SELECT id,user_id,status,total_amount,created_at,updated_at FROM orders`
	args := []any{}
	if userID != nil {
		q += ` WHERE user_id=$1`
		args = append(args, *userID)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Order{}
	for rows.Next() {
		var o model.Order
		if err = rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
func (r *Postgres) ByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var o model.Order
	err := r.db.QueryRow(ctx, `SELECT id,user_id,status,total_amount,created_at,updated_at FROM orders WHERE id=$1`, id).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, `SELECT id,order_id,product_id,quantity,unit_price FROM order_items WHERE order_id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var i model.Item
		if err = rows.Scan(&i.ID, &i.OrderID, &i.ProductID, &i.Quantity, &i.UnitPrice); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, i)
	}
	return &o, rows.Err()
}
func (r *Postgres) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*model.Order, error) {
	tag, err := r.db.Exec(ctx, `UPDATE orders SET status=$1,updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return r.ByID(ctx, id)
}
