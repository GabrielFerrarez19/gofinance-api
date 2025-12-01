package account

import (
	"context"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryInterface interface {
	Create(ctx context.Context, arg sqlc.CreateAccountParams) (sqlc.Account, error)
	GetById(ctx context.Context, id pgtype.UUID) (sqlc.Account, error)
	ListByUser(ctx context.Context, user_id pgtype.UUID) ([]sqlc.Account, error)
	Update(ctx context.Context, args sqlc.UpdateAccountParams) (sqlc.Account, error)
	SoftDelete(ctx context.Context, id pgtype.UUID) error
	UpdateBalance(ctx context.Context, id pgtype.UUID, newBalance pgtype.Numeric) (sqlc.Account, error)
}

type Repository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: sqlc.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg sqlc.CreateAccountParams) (sqlc.Account, error) {
	return r.q.CreateAccount(ctx, arg)
}

func (r *Repository) GetById(ctx context.Context, id pgtype.UUID) (sqlc.Account, error) {
	return r.q.GetAccountByID(ctx, id)
}

func (r *Repository) ListByUser(ctx context.Context, user_id pgtype.UUID) ([]sqlc.Account, error) {
	return r.q.ListAccountsByUser(ctx, user_id)
}

func (r *Repository) Update(ctx context.Context, args sqlc.UpdateAccountParams) (sqlc.Account, error) {
	return r.q.UpdateAccount(ctx, args)
}

func (r *Repository) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	return r.q.SoftDeleteAccount(ctx, id)
}

func (r *Repository) UpdateBalance(ctx context.Context, id pgtype.UUID, newBalance pgtype.Numeric) (sqlc.Account, error) {
	return r.q.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
		ID:      id,
		Balance: newBalance,
	})
}
