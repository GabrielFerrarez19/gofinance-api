package user

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
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return sqlc.User{}, args.Error(1)
	}
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return sqlc.User{}, args.Error(1)
	}
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return sqlc.User{}, args.Error(1)
	}
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return sqlc.User{}, args.Error(1)
	}
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockRepository) DeletedUser(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListUsers(ctx context.Context) ([]sqlc.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.User), args.Error(1)
}

func createTestUser() sqlc.User {
	userID := uuid.New()
	now := time.Now()
	return sqlc.User{
		ID:           pgtype.UUID{Bytes: userID, Valid: true},
		FullName:     "João Silva",
		Email:        "joao@test.com",
		PasswordHash: "$2a$10$TgJ3Bv8a6c5fef2zQ9vG1uUeIm7H6zN2qGQkY7m3X4Cqk6bAq3l7a",
		IsActive:     pgtype.Bool{Bool: true, Valid: true},
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	}
}

func TestService_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		req         models.CreateUserRequest
		mockUser    sqlc.User
		mockError   error
		wantErr     bool
		description string
	}{
		{
			name: "success",
			req: models.CreateUserRequest{
				FullName: "João Silva",
				Email:    "joao@test.com",
				Password: "123456",
			},
			mockUser:    createTestUser(),
			mockError:   nil,
			wantErr:     false,
			description: "must create user successfully",
		},
		{
			name: "repository error",
			req: models.CreateUserRequest{
				FullName: "João Silva",
				Email:    "joao@test.com",
				Password: "123456",
			},
			mockUser:    sqlc.User{},
			mockError:   errors.New("database error"),
			wantErr:     true,
			description: "should return an error when repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewService(mockRepo)

			mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(tt.mockUser, tt.mockError)

			result, err := service.CreateUser(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.req.FullName, result.FullName)
				assert.Equal(t, tt.req.Email, result.Email)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetUserByID(t *testing.T) {
	tests := []struct {
		name      string
		id        pgtype.UUID
		mockUser  sqlc.User
		mockError error
		wantErr   bool
	}{
		{
			name:      "success",
			id:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockUser:  createTestUser(),
			mockError: nil,
			wantErr:   false,
		},
		{
			name:      "user not found",
			id:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockUser:  sqlc.User{},
			mockError: errors.New("user not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mokeRepo := new(MockRepository)
			service := NewService(mokeRepo)

			mokeRepo.On("GetUserByID", mock.Anything, tt.id).Return(tt.mockUser, tt.mockError)

			result, err := service.GetUserByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.mockUser.Email, result.Email)
			}
			mokeRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		mockUser  sqlc.User
		mockError error
		wantErr   bool
	}{
		{
			name:      "success",
			email:     "joao@test.com",
			mockUser:  createTestUser(),
			mockError: nil,
			wantErr:   false,
		},
		{
			name:      "user not found",
			email:     "notfound@test.com",
			mockUser:  sqlc.User{},
			mockError: errors.New("user not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewService(mockRepo)

			mockRepo.On("GetUserByEmail", mock.Anything, tt.email).Return(tt.mockUser, tt.mockError)

			result, err := service.GetUserByEmail(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.email, result.Email)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ValidatorPassword(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		mockUser    sqlc.User
		mockError   error
		wantErr     bool
		description string
	}{
		{
			name:        "success - valid password",
			email:       "joao@test.com",
			password:    "123456",
			mockUser:    createTestUser(),
			mockError:   nil,
			wantErr:     false,
			description: "must validate correct password",
		},
		{
			name:        "invalid password",
			email:       "joao@test.com",
			password:    "wrongpassword",
			mockUser:    createTestUser(),
			mockError:   nil,
			wantErr:     true,
			description: "should return an error for incorrect password",
		},
		{
			name:        "user not found",
			email:       "notfound@test.com",
			password:    "123456",
			mockUser:    sqlc.User{},
			mockError:   errors.New("user not found"),
			wantErr:     true,
			description: "should return an error when user does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewService(mockRepo)

			mockRepo.On("GetUserByEmail", mock.Anything, tt.email).Return(tt.mockUser, tt.mockError)

			result, err := service.ValidatorPassword(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.email, result.Email)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_UpdateUser(t *testing.T) {
	userID := uuid.New()
	id := pgtype.UUID{Bytes: userID, Valid: true}

	tests := []struct {
		name      string
		id        pgtype.UUID
		req       models.UpdateUserRequest
		mockUser  sqlc.User
		mockError error
		wantErr   bool
	}{
		{
			name: "success",
			id:   id,
			req: models.UpdateUserRequest{
				FullName: *stringPtr("João Silva Updated"),
				Email:    *stringPtr("joao.updated@test.com"),
			},
			mockUser:  createTestUser(),
			mockError: nil,
			wantErr:   false,
		},
		{
			name: "repository error",
			id:   id,
			req: models.UpdateUserRequest{
				FullName: *stringPtr("João Silva Updated"),
			},
			mockUser:  sqlc.User{},
			mockError: errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewService(mockRepo)

			mockRepo.On("UpdateUser", mock.Anything, mock.Anything).Return(tt.mockUser, tt.mockError)

			result, err := service.UpdateUser(context.Background(), tt.id, tt.req)

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

func TestService_ListUsers(t *testing.T) {
	tests := []struct {
		name       string
		mockUsers  []sqlc.User
		mockError  error
		wantErr    bool
		wantLength int
	}{
		{
			name:       "success - empty list",
			mockUsers:  []sqlc.User{},
			mockError:  nil,
			wantErr:    false,
			wantLength: 0,
		},
		{
			name: "success - with users",
			mockUsers: []sqlc.User{
				createTestUser(),
				createTestUser(),
			},
			mockError:  nil,
			wantErr:    false,
			wantLength: 2,
		},
		{
			name:       "repository error",
			mockUsers:  nil,
			mockError:  errors.New("database error"),
			wantErr:    true,
			wantLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewService(mockRepo)

			mockRepo.On("ListUsers", mock.Anything).Return(tt.mockUsers, tt.mockError)

			result, err := service.ListUsers(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Len(t, result, tt.wantLength)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
