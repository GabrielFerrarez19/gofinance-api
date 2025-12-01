package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, args sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
	argsCalled := m.Called(ctx, args)
	if argsCalled.Get(0) == nil {
		return sqlc.Transaction{}, argsCalled.Error(1)
	}
	return argsCalled.Get(0).(sqlc.Transaction), argsCalled.Error(1)
}

func (m *MockRepository) GetById(ctx context.Context, id pgtype.UUID) (sqlc.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return sqlc.Transaction{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Transaction), args.Error(1)
}

func (m *MockRepository) ListByAccount(ctx context.Context, accID pgtype.UUID) ([]sqlc.Transaction, error) {
	args := m.Called(ctx, accID)
	if args.Get(0) == nil {
		return []sqlc.Transaction{}, args.Error(1)
	}
	return args.Get(0).([]sqlc.Transaction), args.Error(1)
}

func (m *MockRepository) ListByUser(ctx context.Context, userID pgtype.UUID) ([]sqlc.Transaction, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return []sqlc.Transaction{}, args.Error(1)
	}
	return args.Get(0).([]sqlc.Transaction), args.Error(1)
}

func (m *MockRepository) ListByPeriod(ctx context.Context, userID pgtype.UUID, from, to pgtype.Timestamptz) ([]sqlc.Transaction, error) {
	args := m.Called(ctx, userID, from, to)
	if args.Get(0) == nil {
		return []sqlc.Transaction{}, args.Error(1)
	}
	return args.Get(0).([]sqlc.Transaction), args.Error(1)
}

func (m *MockRepository) Updated(ctx context.Context, arg sqlc.UpdateTransactionParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return sqlc.Transaction{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Transaction), args.Error(1)
}

func (m *MockRepository) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func createTestTransaction() sqlc.Transaction {
	id := uuid.New()
	userID := uuid.New()
	accountID := uuid.New()
	now := time.Now()

	var amount pgtype.Numeric
	amount.Scan("1000.50")

	return sqlc.Transaction{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		UserID:      pgtype.UUID{Bytes: userID, Valid: true},
		AccountID:   pgtype.UUID{Bytes: accountID, Valid: true},
		CategoryID:  pgtype.UUID{},
		Type:        "income",
		Amount:      amount,
		Description: "Salário",
		Status:      "completed",
		Date:        pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		DeletedAt:   pgtype.Timestamptz{},
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name        string
		req         models.CreateTransactionRequest
		mockTx      sqlc.Transaction
		mockError   error
		wantErr     bool
		description string
	}{
		{
			name: "success",
			req: models.CreateTransactionRequest{
				AccountID:   uuid.New(),
				Type:        models.TransactionIncome,
				Amount:      1000.50,
				Description: "Salário",
				Date:        time.Now(),
			},
			mockTx:      createTestTransaction(),
			mockError:   nil,
			wantErr:     false,
			description: "must create transaction successfully",
		},
		{
			name: "repository error",
			req: models.CreateTransactionRequest{
				AccountID:   uuid.New(),
				Type:        models.TransactionIncome,
				Amount:      1000.50,
				Description: "Salário",
				Date:        time.Now(),
			},
			mockTx:      sqlc.Transaction{},
			mockError:   errors.New("database error"),
			wantErr:     true,
			description: "should return an error when repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := &Service{repo: mockRepo}

			userID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
			mockRepo.On("Create", mock.Anything, mock.Anything).Return(tt.mockTx, tt.mockError)

			result, err := service.Create(context.Background(), userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.req.Description, result.Description)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        pgtype.UUID
		mockTx    sqlc.Transaction
		mockError error
		wantErr   bool
	}{
		{
			name:      "success",
			id:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockTx:    createTestTransaction(),
			mockError: nil,
			wantErr:   false,
		},
		{
			name:      "transaction not found",
			id:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockTx:    sqlc.Transaction{},
			mockError: errors.New("transaction not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := &Service{repo: mockRepo}

			mockRepo.On("GetById", mock.Anything, tt.id).Return(tt.mockTx, tt.mockError)

			result, err := service.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListByAccount(t *testing.T) {
	tests := []struct {
		name        string
		accountID   pgtype.UUID
		mockTxs     []sqlc.Transaction
		mockError   error
		wantErr     bool
		wantLength  int
		description string
	}{
		{
			name:        "success",
			accountID:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockTxs:     []sqlc.Transaction{createTestTransaction()},
			mockError:   nil,
			wantErr:     false,
			wantLength:  1,
			description: "must list transactions by account successfully",
		},
		{
			name:        "repository error",
			accountID:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockTxs:     []sqlc.Transaction{},
			mockError:   errors.New("database error"),
			wantErr:     true,
			wantLength:  0,
			description: "should return an error when repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := &Service{repo: mockRepo}

			mockRepo.On("ListByAccount", mock.Anything, tt.accountID).Return(tt.mockTxs, tt.mockError)

			result, err := service.ListByAccount(context.Background(), tt.accountID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLength)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListByUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      pgtype.UUID
		mockTxs     []sqlc.Transaction
		mockError   error
		wantErr     bool
		wantLength  int
		description string
	}{
		{
			name:        "success",
			userID:      pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockTxs:     []sqlc.Transaction{createTestTransaction()},
			mockError:   nil,
			wantErr:     false,
			wantLength:  1,
			description: "must list transactions by user successfully",
		},
		{
			name:        "repository error",
			userID:      pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockTxs:     []sqlc.Transaction{},
			mockError:   errors.New("database error"),
			wantErr:     true,
			wantLength:  0,
			description: "should return an error when repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := &Service{repo: mockRepo}

			mockRepo.On("ListByUser", mock.Anything, tt.userID).Return(tt.mockTxs, tt.mockError)

			result, err := service.ListByUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLength)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	txID := uuid.New()
	id := pgtype.UUID{Bytes: txID, Valid: true}

	tests := []struct {
		name        string
		id          pgtype.UUID
		req         models.UpdateTransactionRequest
		mockTx      sqlc.Transaction
		mockError   error
		wantErr     bool
		description string
	}{
		{
			name: "success",
			id:   id,
			req: models.UpdateTransactionRequest{
				Amount: floatPtr(2000.75),
			},
			mockTx:      createTestTransaction(),
			mockError:   nil,
			wantErr:     false,
			description: "must update transaction successfully",
		},
		{
			name: "repository error",
			id:   id,
			req: models.UpdateTransactionRequest{
				Amount: floatPtr(2000.75),
			},
			mockTx:      sqlc.Transaction{},
			mockError:   errors.New("database error"),
			wantErr:     true,
			description: "should return an error when repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := &Service{repo: mockRepo}

			mockRepo.On("Updated", mock.Anything, mock.Anything).Return(tt.mockTx, tt.mockError)

			result, err := service.Update(context.Background(), tt.id, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		id          pgtype.UUID
		mockError   error
		wantErr     bool
		description string
	}{
		{
			name:        "success",
			id:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockError:   nil,
			wantErr:     false,
			description: "must delete transaction successfully",
		},
		{
			name:        "repository error",
			id:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockError:   errors.New("database error"),
			wantErr:     true,
			description: "should return an error when repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := &Service{repo: mockRepo}

			mockRepo.On("SoftDelete", mock.Anything, tt.id).Return(tt.mockError)

			err := service.Delete(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
