package account

import (
	"context"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, userID pgtype.UUID, req models.CreateAccountRequest) (models.AccountResponse, error) {
	acc, err := s.repo.Create(ctx, sqlc.CreateAccountParams{
		UserID:      userID,
		Name:        req.Name,
		Type:        string(req.Type),
		Balance:     req.Balance,
		Currency:    req.Currency,
		Description: req.Description,
	})
	if err != nil {
		return models.AccountResponse{}, err
	}
	return toAccountResponse(acc), nil
}

func (s *Service) GetByID(ctx context.Context, id pgtype.UUID) (models.AccountResponse, error) {
	acc, err := s.repo.GetById(ctx, id)
	if err != nil {
		return models.AccountResponse{}, err
	}
	return toAccountResponse(acc), nil
}

func (s *Service) ListByUser(ctx context.Context, userID pgtype.UUID) ([]models.AccountResponse, error) {
	accs, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return []models.AccountResponse{}, err
	}
	out := make([]models.AccountResponse, 0, len(accs))

	for _, a := range accs {
		out = append(out, toAccountResponse(a))
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, id pgtype.UUID, req models.UpdateAccountRequest) (models.AccountResponse, error) {
	acc, err := s.repo.Update(ctx, sqlc.UpdateAccountParams{
		ID:          id,
		Name:        strPtrOrEmpty(req.Name),
		Type:        strPtrOrEmpty((*string)(req.Type)),
		Balance:     floatPtrToPgNumeric(req.Balance),
		Currency:    strPtrOrEmpty(req.Currency),
		Description: strPtrToPgText(req.Description),
		IsActive:    toPgBool(req.IsActive),
	})
	if err != nil {
		return models.AccountResponse{}, err
	}
	return toAccountResponse(acc), nil
}

func (s *Service) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func toPgBool(p *bool) pgtype.Bool {
	if p == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *p, Valid: true}
}

func toAccountResponse(a sqlc.Account) models.AccountResponse {
	return models.AccountResponse{
		ID:          uuidFromPg(a.ID),
		UserID:      uuidFromPg(a.UserID),
		Name:        a.Name,
		Type:        models.AccountType(a.Type),
		Balance:     floatFromNumeric(a.Balance),
		Currency:    a.Currency,
		Description: a.Description.String,
		IsActive:    a.IsActive.Bool,
		CreatedAt:   a.CreatedAt.Time,
		UpdatedAt:   a.UpdatedAt.Time,
	}
}

// Funções helper para conversão de tipos
func strPtrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func strPtrToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func floatPtrToPgNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{}
	}
	var num pgtype.Numeric
	num.Scan(*f)
	return num
}

func floatFromNumeric(n pgtype.Numeric) float64 {
	val, _ := n.Float64Value()
	return val.Float64
}

func uuidFromPg(pg pgtype.UUID) uuid.UUID {
	u, _ := uuid.FromBytes(pg.Bytes[:])
	return u
}
