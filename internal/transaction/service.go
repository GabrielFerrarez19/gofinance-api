package transaction

import (
	"context"
	"fmt"
	"time"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type RepositoryInterface interface {
	Create(ctx context.Context, args sqlc.CreateTransactionParams) (sqlc.Transaction, error)
	GetById(ctx context.Context, id pgtype.UUID) (sqlc.Transaction, error)
	ListByAccount(ctx context.Context, accID pgtype.UUID) ([]sqlc.Transaction, error)
	ListByUser(ctx context.Context, userID pgtype.UUID) ([]sqlc.Transaction, error)
	ListByPeriod(ctx context.Context, userID pgtype.UUID, from, to pgtype.Timestamptz) ([]sqlc.Transaction, error)
	Updated(ctx context.Context, arg sqlc.UpdateTransactionParams) (sqlc.Transaction, error)
	SoftDelete(ctx context.Context, id pgtype.UUID) error
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, userID pgtype.UUID, req models.CreateTransactionRequest) (models.TransactionResponse, error) {
	var amt pgtype.Numeric
	if err := amt.Scan(fmt.Sprintf("%.2f", req.Amount)); err != nil {
		return models.TransactionResponse{}, fmt.Errorf("invalid amount: %w", err)
	}

	stat := textFromString("")
	if req.Status != nil {
		stat = textFromString(string(*req.Status))
	}

	tx, err := s.repo.Create(ctx, sqlc.CreateTransactionParams{
		UserID:      userID,
		AccountID:   uuidToPg(req.AccountID),
		CategoryID:  uuidPtrToPg(req.CategoryID),
		Type:        string(req.Type),
		Amount:      amt,
		Description: req.Description,
		Column7:     stat,
		Date:        timeToPg(req.Date),
	})
	if err != nil {
		return models.TransactionResponse{}, err
	}
	return toTxResponse(tx), nil
}

func (s *Service) GetByID(ctx context.Context, id pgtype.UUID) (models.TransactionResponse, error) {
	tx, err := s.repo.GetById(ctx, id)
	if err != nil {
		return models.TransactionResponse{}, err
	}
	return toTxResponse(tx), nil
}

func (s *Service) ListByAccount(ctx context.Context, accountID pgtype.UUID) ([]models.TransactionResponse, error) {
	list, err := s.repo.ListByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	out := make([]models.TransactionResponse, 0, len(list))
	for _, t := range list {
		out = append(out, toTxResponse(t))
	}

	return out, nil
}

func (s *Service) ListByUser(ctx context.Context, userID pgtype.UUID) ([]models.TransactionResponse, error) {
	list, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]models.TransactionResponse, 0, len(list))
	for _, t := range list {
		out = append(out, toTxResponse(t))
	}
	return out, nil
}

func (s *Service) ListByPeriod(ctx context.Context, userID pgtype.UUID, from, to pgtype.Timestamptz) ([]models.TransactionResponse, error) {
	list, err := s.repo.ListByPeriod(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	out := make([]models.TransactionResponse, 0, len(list))
	for _, t := range list {
		out = append(out, toTxResponse(t))
	}

	return out, nil
}

func (s *Service) Update(ctx context.Context, id pgtype.UUID, req models.UpdateTransactionRequest) (models.TransactionResponse, error) {
	arg := sqlc.UpdateTransactionParams{
		ID:          id,
		AccountID:   uuidPtrToPg(req.AccountID),
		CategoryID:  uuidPtrToPg(req.CategoryID),
		Type:        textPtrToPg(req.Type),
		Amount:      floatPtrToPgNumeric(req.Amount),
		Description: strPtrToPgText(req.Description),
		Status:      strPtrToPgText((*string)(req.Status)),
		Date:        timePtrToPg(req.Date),
	}
	tx, err := s.repo.Updated(ctx, arg)
	if err != nil {
		return models.TransactionResponse{}, err
	}

	return toTxResponse(tx), nil
}

func (s *Service) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

/******** helpers ********/
func uuidToPg(u uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: u, Valid: true} }

func uuidPtrToPg(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

func floatPtrToPgNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{}
	}
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%.2f", *f))
	return n
}

func strPtrToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func textFromString(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func textPtrToPg(t *models.TransactionType) pgtype.Text {
	if t == nil {
		return pgtype.Text{}
	}
	v := string(*t)
	return pgtype.Text{String: v, Valid: true}
}
func timeToPg(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }
func timePtrToPg(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func toTxResponse(t sqlc.Transaction) models.TransactionResponse {
	amt, _ := t.Amount.Float64Value()
	var catPtr *uuid.UUID
	if t.CategoryID.Valid {
		v := uuid.UUID(t.CategoryID.Bytes)
		catPtr = &v
	}
	return models.TransactionResponse{
		ID:          t.ID.Bytes,
		UserID:      t.UserID.Bytes,
		AccountID:   t.AccountID.Bytes,
		CategoryID:  catPtr,
		Type:        models.TransactionType(t.Type),
		Amount:      amt.Float64,
		Description: t.Description,
		Status:      models.TransactionStatus(t.Status),
		Date:        t.Date.Time,
		CreatedAt:   t.CreatedAt.Time,
		UpdatedAt:   t.UpdatedAt.Time,
	}
}
