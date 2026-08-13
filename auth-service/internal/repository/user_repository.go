package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ishimweBonheur/order-management/auth-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound    = errors.New("user not found")
	ErrEmailExists = errors.New("email already exists")
)

type UserRepository interface {
	Create(context.Context, *model.User) error
	ByEmail(context.Context, string) (*model.User, error)
	ByID(context.Context, uuid.UUID) (*model.User, error)
}
type Postgres struct{ db *pgxpool.Pool }

func NewPostgres(db *pgxpool.Pool) *Postgres { return &Postgres{db: db} }

const userColumns = `id,name,email,password_hash,role,created_at,updated_at`

func (r *Postgres) Create(ctx context.Context, u *model.User) error {
	_, err := r.db.Exec(ctx, `INSERT INTO users (`+userColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7)`, u.ID, u.Name, u.Email, u.PasswordHash, u.Role, u.CreatedAt, u.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrEmailExists
	}
	return err
}
func scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}
func (r *Postgres) ByEmail(ctx context.Context, email string) (*model.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email=$1`, email))
}
func (r *Postgres) ByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id=$1`, id))
}
