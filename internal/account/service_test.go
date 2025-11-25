package account

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

type MokeRepository struct {
	mock.Mock
}

func (m *MokeRepository) Create(ctx context.Context, arg sqlc.CreateAccountParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return sqlc.Account{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Account), args.Error(1)
}

func (m *MokeRepository) GetById(ctx context.Context, id pgtype.UUID) (sqlc.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return sqlc.Account{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Account), args.Error(1)
}

func (m *MokeRepository) ListByUser(ctx context.Context, userID pgtype.UUID) ([]sqlc.Account, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return []sqlc.Account{}, args.Error(1)
	}
	return args.Get(0).([]sqlc.Account), args.Error(1)
}

func (m *MokeRepository) Update(ctx context.Context, arg sqlc.UpdateAccountParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return sqlc.Account{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Account), args.Error(1)
}

func (m *MokeRepository) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MokeRepository) UpdateBalance(ctx context.Context, id pgtype.UUID, newBalance pgtype.Numeric) (sqlc.Account, error) {
	args := m.Called(ctx, id, newBalance)
	if args.Get(0) == nil {
		return sqlc.Account{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Account), args.Error(1)
}

func createTestAccount() sqlc.Account {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	var balance pgtype.Numeric
	balance.Scan("1000.50")

	return sqlc.Account{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		UserID:      pgtype.UUID{Bytes: userID, Valid: true},
		Name:        "Gabriel Cristiano Ferrarez",
		Type:        "checking",
		Balance:     balance,
		Currency:    "BRL",
		Description: pgtype.Text{String: "Conta pricipal", Valid: true},
		IsActive:    pgtype.Bool{Bool: true, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		DeletedAt:   pgtype.Timestamptz{},
	}
}

func TestService_CreateAccount(t *testing.T) {
	tests := []struct {
		name        string
		req         models.CreateAccountRequest
		mockAccount sqlc.Account
		mockError   error
		wantErr     bool
		description string
	}{
		{
			name: "success",
			req: models.CreateAccountRequest{
				Name:        "Gabriel Cristiano Ferrarez",
				Type:        "checking",
				Balance:     1000.50,
				Currency:    "BRL",
				Description: "Conta pricipal",
			},
			mockAccount: createTestAccount(),
			mockError:   nil,
			wantErr:     false,
			description: "must create account successfully",
		}, {
			name: "repository error",
			req: models.CreateAccountRequest{
				Name:        "Gabriel Cristiano Ferrarez",
				Type:        "checking",
				Balance:     1000.50,
				Currency:    "BRL",
				Description: "Conta pricipal",
			},
			mockAccount: sqlc.Account{},
			mockError:   errors.New("database error"),
			wantErr:     true,
			description: "should return an error when repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MokeRepository)
			service := NewService(mockRepo)

			userID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
			mockRepo.On("Create", mock.Anything, mock.Anything).Return(tt.mockAccount, tt.mockError)

			result, err := service.Create(context.Background(), userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.Type, result.Type)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		id          pgtype.UUID
		mockAccount sqlc.Account
		mockError   error
		wantErr     bool
	}{
		{
			name:        "success",
			id:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockAccount: createTestAccount(),
			mockError:   nil,
			wantErr:     false,
		},
		{
			name:        "account not found",
			id:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockAccount: createTestAccount(),
			mockError:   errors.New("account not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mokeRepo := new(MokeRepository)
			service := NewService(mokeRepo)
			mokeRepo.On("GetById", mock.Anything, tt.id).Return(tt.mockAccount, tt.mockError)

			result, err := service.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.mockAccount.Name, result.Name)
			}
			mokeRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListByUser(t *testing.T) {
	tests := []struct {
		name         string
		userID       pgtype.UUID
		mockAccounts []sqlc.Account
		mockError    error
		wantErr      bool
		wantLength   int
	}{
		{
			name:         "success",
			userID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockAccounts: []sqlc.Account{createTestAccount()},
			mockError:    nil,
			wantErr:      false,
			wantLength:   1,
		}, {
			name:         "user not found",
			userID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockAccounts: []sqlc.Account{},
			mockError:    errors.New("user not found"),
			wantErr:      true,
			wantLength:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mokeRepo := new(MokeRepository)
			service := NewService(mokeRepo)

			mokeRepo.On("ListByUser", mock.Anything, tt.userID).Return(tt.mockAccounts, tt.mockError)

			result, err := service.ListByUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLength)
			}
			mokeRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	accountID := uuid.New()
	id := pgtype.UUID{Bytes: accountID, Valid: true}
	ty := models.AccountType("checking")

	tests := []struct {
		name        string
		id          pgtype.UUID
		req         models.UpdateAccountRequest
		mockAccount sqlc.Account
		mockError   error
		wantErr     bool
	}{
		{
			name: "success",
			id:   id,
			req: models.UpdateAccountRequest{
				Name:        stringPtr("Gabriel Cristiano Updated"),
				Type:        &ty,
				Balance:     floatPtr(1100.50),
				Currency:    stringPtr("BRL"),
				Description: stringPtr("Conta pricipal"),
			},
			mockAccount: createTestAccount(),
			mockError:   nil,
			wantErr:     false,
		},
		{
			name: "repository error",
			id:   id,
			req: models.UpdateAccountRequest{
				Name:        stringPtr("Gabriel Cristiano Updated"),
				Type:        &ty,
				Balance:     floatPtr(1100.50),
				Currency:    stringPtr("BRL"),
				Description: stringPtr("Conta pricipal"),
			},
			mockAccount: sqlc.Account{},
			mockError:   errors.New("database error"),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MokeRepository)
			service := NewService(mockRepo)

			mockRepo.On("Update", mock.Anything, mock.Anything).Return(tt.mockAccount, tt.mockError)

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

// Helper function
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
