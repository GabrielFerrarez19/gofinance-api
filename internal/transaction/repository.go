package transaction

import (
	"context"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool, q *sqlc.Querier) *Repository {
	return &Repository{
		pool: pool,
		q:    sqlc.New(pool),
	}
}

func (r *Repository) Create(ctx context.Context, args sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
	return r.q.CreateTransaction(ctx, args)
}

func (r *Repository) GetById(ctx context.Context, id pgtype.UUID) (sqlc.Transaction, error) {
	return r.q.GetTransactionByID(ctx, id)
}

func (r *Repository) ListByAccount(ctx context.Context, accID pgtype.UUID) ([]sqlc.Transaction, error) {
	return r.q.ListTransactionsByAccount(ctx, accID)
}

func (r *Repository) ListByUser(ctx context.Context, userID pgtype.UUID) ([]sqlc.Transaction, error) {
	return r.q.ListTransactionsByUser(ctx, userID)
}

func (r *Repository) ListByPeriod(ctx context.Context, userID pgtype.UUID, from, to pgtype.Timestamptz) ([]sqlc.Transaction, error) {
	params := sqlc.ListTransactionsByPeriodParams{
		UserID: userID,
		Date:   from,
		Date_2: to,
	}

	return r.q.ListTransactionsByPeriod(ctx, params)
}

func (r *Repository) Updated(ctx context.Context, arg sqlc.UpdateTransactionParams) (sqlc.Transaction, error) {
	return r.q.UpdateTransaction(ctx, arg)
}

func (r *Repository) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	return r.q.SoftDeleteTransaction(ctx, id)
}
