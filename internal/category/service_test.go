package category

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

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) Create(ctx context.Context, arg sqlc.CreateCategoryParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return sqlc.Category{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Category), args.Error(1)
}

func (m *MockRepository) GetByID(ctx context.Context, id pgtype.UUID) (sqlc.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return sqlc.Category{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Category), args.Error(1)
}

func (m *MockRepository) GetByUserID(ctx context.Context, UserID pgtype.UUID) ([]sqlc.Category, error) {
	args := m.Called(ctx, UserID)
	if args.Get(0) == nil {
		return []sqlc.Category{}, args.Error(1)
	}
	return args.Get(0).([]sqlc.Category), args.Error(1)
}

func (m *MockRepository) GetByUserIDAndName(ctx context.Context, userID pgtype.UUID, name string) (sqlc.Category, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return sqlc.Category{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Category), args.Error(1)
}

func (m *MockRepository) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) UpdateCategory(ctx context.Context, arg sqlc.UpdateCategoryParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return sqlc.Category{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Category), args.Error(1)
}

func createTestCategory() sqlc.Category {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	return sqlc.Category{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		UserID:      pgtype.UUID{Bytes: userID, Valid: true},
		Name:        "Alimentação",
		Description: pgtype.Text{String: "Gastos com comida", Valid: true},
		Color:       "#FF5733",
		Icon:        pgtype.Text{String: "food", Valid: true},
		IsActive:    pgtype.Bool{Bool: true, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		DeletedAt:   pgtype.Timestamptz{},
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name           string
		req            models.CreateCategoryRequest
		userID         pgtype.UUID
		mockCategory   sqlc.Category
		mockGetError   error
		mockCreateError error
		wantErr        bool
		description    string
	}{
		{
			name: "success",
			req: models.CreateCategoryRequest{
				Name:        "Alimentação",
				Description: "Gastos com comida",
				Color:       "#FF5733",
				Icon:        "food",
			},
			userID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockCategory:   createTestCategory(),
			mockGetError:   errors.New("not found"),
			mockCreateError: nil,
			wantErr:        false,
			description:    "must create category successfully",
		},
		{
			name: "category already exists",
			req: models.CreateCategoryRequest{
				Name:        "Alimentação",
				Description: "Gastos com comida",
				Color:       "#FF5733",
				Icon:        "food",
			},
			userID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockCategory:   createTestCategory(),
			mockGetError:   nil,
			mockCreateError: nil,
			wantErr:        true,
			description:    "should return an error when category already exists",
		},
		{
			name: "repository error",
			req: models.CreateCategoryRequest{
				Name:        "Alimentação",
				Description: "Gastos com comida",
				Color:       "#FF5733",
				Icon:        "food",
			},
			userID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockCategory:   sqlc.Category{},
			mockGetError:   errors.New("not found"),
			mockCreateError: errors.New("database error"),
			wantErr:        true,
			description:    "should return an error when repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			service := &Service{repo: mockRepo}

			mockRepo.On("GetByUserIDAndName", mock.Anything, tt.userID, tt.req.Name).Return(tt.mockCategory, tt.mockGetError)
			if tt.mockGetError != nil {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(tt.mockCategory, tt.mockCreateError)
			}

			result, err := service.Create(context.Background(), tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.name == "category already exists" {
					assert.Contains(t, err.Error(), "category already exists")
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.req.Name, result.Name)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		id          pgtype.UUID
		mockCategory sqlc.Category
		mockError   error
		wantErr     bool
	}{
		{
			name:        "success",
			id:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockCategory: createTestCategory(),
			mockError:   nil,
			wantErr:     false,
		},
		{
			name:        "category not found",
			id:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockCategory: sqlc.Category{},
			mockError:   errors.New("category not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			service := &Service{repo: mockRepo}

			mockRepo.On("GetByID", mock.Anything, tt.id).Return(tt.mockCategory, tt.mockError)

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

func TestService_GetByUserID(t *testing.T) {
	tests := []struct {
		name         string
		userID       pgtype.UUID
		mockCategories []sqlc.Category
		mockError    error
		wantErr      bool
		wantLength   int
	}{
		{
			name:         "success",
			userID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockCategories: []sqlc.Category{createTestCategory()},
			mockError:    nil,
			wantErr:      false,
			wantLength:   1,
		},
		{
			name:         "repository error",
			userID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockCategories: []sqlc.Category{},
			mockError:    errors.New("database error"),
			wantErr:      true,
			wantLength:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			service := &Service{repo: mockRepo}

			mockRepo.On("GetByUserID", mock.Anything, tt.userID).Return(tt.mockCategories, tt.mockError)

			result, err := service.GetByUserID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLength)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	categoryID := uuid.New()
	userID := uuid.New()
	id := pgtype.UUID{Bytes: categoryID, Valid: true}
	uid := pgtype.UUID{Bytes: userID, Valid: true}

	tests := []struct {
		name           string
		id             pgtype.UUID
		userID         pgtype.UUID
		req            models.UpdateCategoryRequest
		mockCategory   sqlc.Category
		mockGetError   error
		mockUpdateError sqlc.Category
		wantErr        bool
		description    string
	}{
		{
			name:   "category not found",
			id:     id,
			userID: uid,
			req: models.UpdateCategoryRequest{
				Name: stringPtr("Alimentação Atualizada"),
			},
			mockCategory: sqlc.Category{},
			mockGetError: errors.New("not found"),
			wantErr:      true,
			description:  "should return an error when category not found",
		},
		{
			name:   "category does not belong to user",
			id:     id,
			userID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
			req: models.UpdateCategoryRequest{
				Name: stringPtr("Alimentação Atualizada"),
			},
			mockCategory:   createTestCategory(),
			mockGetError:   nil,
			mockUpdateError: sqlc.Category{},
			wantErr:        true,
			description:    "should return an error when category does not belong to user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			service := &Service{repo: mockRepo}

			mockRepo.On("GetByID", mock.Anything, tt.id).Return(tt.mockCategory, tt.mockGetError)

			result, err := service.Update(context.Background(), tt.id, tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.name == "category not found" {
					assert.Contains(t, err.Error(), "category not found")
				} else if tt.name == "category does not belong to user" {
					assert.Contains(t, err.Error(), "category does not belong to user")
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	id := pgtype.UUID{Bytes: categoryID, Valid: true}
	uid := pgtype.UUID{Bytes: userID, Valid: true}
	
	testCategory := createTestCategory()
	testCategory.UserID = uid
	testCategory.ID = id

	tests := []struct {
		name           string
		id             pgtype.UUID
		userID         pgtype.UUID
		mockCategory   sqlc.Category
		mockGetError   error
		mockDeleteError error
		wantErr        bool
		description    string
	}{
		{
			name:           "success",
			id:             id,
			userID:         uid,
			mockCategory:   testCategory,
			mockGetError:   nil,
			mockDeleteError: nil,
			wantErr:        false,
			description:    "must delete category successfully",
		},
		{
			name:           "category not found",
			id:             id,
			userID:         uid,
			mockCategory:   sqlc.Category{},
			mockGetError:   errors.New("not found"),
			mockDeleteError: nil,
			wantErr:        true,
			description:    "should return an error when category not found",
		},
		{
			name:           "category does not belong to user",
			id:             id,
			userID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockCategory:   createTestCategory(),
			mockGetError:   nil,
			mockDeleteError: nil,
			wantErr:        true,
			description:    "should return an error when category does not belong to user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			service := &Service{repo: mockRepo}

			mockRepo.On("GetByID", mock.Anything, tt.id).Return(tt.mockCategory, tt.mockGetError)
			if tt.mockGetError == nil && tt.mockCategory.UserID.Bytes == tt.userID.Bytes {
				mockRepo.On("SoftDelete", mock.Anything, tt.id).Return(tt.mockDeleteError)
			}

			err := service.Delete(context.Background(), tt.id, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}


