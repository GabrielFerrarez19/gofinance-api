package report

import (
	"context"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
	q *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository{
	return &Repository{
		pool: pool,	
		q: sqlc.New(pool),
	}
}

func (r *Repository) Create(ctx context.Context, arg sqlc.CreateReportParams)(sqlc.Report, error){
	return r.q.CreateReport(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID)(sqlc.Report, error){
	return r.q.GetReportByID(ctx,id)
}

func (r *Repository) ListByUser(ctx context.Context, user_id pgtype.UUID)([]sqlc.Report, error){
	return r.q.ListReportsByUser(ctx,user_id)
}

func (r *Repository) Update(ctx context.Context, arg sqlc.UpdateReportParams)(sqlc.Report, error){
	return r.q.UpdateReport(ctx,arg)
}

func (r *Repository) Delete(ctx context.Context, id pgtype.UUID)(error){
	return r.q.DeleteReport(ctx,id)
}