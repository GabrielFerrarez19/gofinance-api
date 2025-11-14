package user

import (
	"context"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryInterface interface {
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
	GetUserByEmail(ctx context.Context, email string) (sqlc.User, error)
	UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error)
	DeletedUser(ctx context.Context, id pgtype.UUID) error
	ListUsers(ctx context.Context) ([]sqlc.User, error)
}

type Repository struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *Repository) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	return r.q.CreateUser(ctx, arg)
}

func (r *Repository) GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *Repository) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	return r.q.UpdateUser(ctx, arg)
}

func (r *Repository) DeletedUser(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteUser(ctx, id)
}

func (r *Repository) ListUsers(ctx context.Context) ([]sqlc.User, error) {
	return r.q.ListUsers(ctx)
}
