package category

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

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
		q:    sqlc.New(pool),
	}
}

func (r *Repository) Create(ctx context.Context, arg sqlc.CreateCategoryParams) (sqlc.Category, error) {
	return r.q.CreateCategory(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (sqlc.Category, error) {
	return r.q.GetCategoryByID(ctx, id)
}

func (r *Repository) GetByUserID(ctx context.Context, UserID pgtype.UUID) ([]sqlc.Category, error) {
	return r.q.GetCategoriesByUserID(ctx, UserID)
}

func (r *Repository) Update(ctx context.Context, arg sqlc.UpdateCategoryParams) (sqlc.Category, error) {
	return r.q.UpdateCategory(ctx, arg)
}

func (r *Repository) UpdateCategory(ctx context.Context, arg sqlc.UpdateCategoryParams) (sqlc.Category, error) {
	return r.q.UpdateCategory(ctx, arg)
}

func (r *Repository) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeletedCategory(ctx, id)
}

func (r *Repository) GetByUserIDAndName(ctx context.Context, userID pgtype.UUID, name string) (sqlc.Category, error) {
	return r.q.GetCategoryByUserIDAndName(ctx, sqlc.GetCategoryByUserIDAndNameParams{
		UserID: userID,
		Name:   name,
	})
}
