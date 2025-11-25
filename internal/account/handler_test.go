package account

import (
	"context"

	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Create(ctx context.Context, req models.CreateAccountRequest) (models.AccountResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return models.AccountResponse{}, args.Error(1)
	}
	return args.Get(0).(models.AccountResponse), args.Error(1)
}

func (m *MockService) GetByID(ctx context.Context, id pgtype.UUID) (models.AccountResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return models.AccountResponse{}, args.Error(1)
	}
	return args.Get(0).(models.AccountResponse), args.Error(1)
}

func (m *MockService) ListByUser(ctx context.Context, userID pgtype.UUID) ([]models.AccountResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return []models.AccountResponse{}, args.Error(1)
	}
	return args.Get(0).([]models.AccountResponse), args.Error(1)
}

func (m *MockService) Update(ctx context.Context, req models.UpdateAccountRequest) (models.AccountResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return models.AccountResponse{}, args.Error(1)
	}
	return args.Get(0).(models.AccountResponse), args.Error(1)
}

func (m *MockService) Delete(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
