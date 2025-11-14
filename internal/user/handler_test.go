package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateUser(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return models.UserResponse{}, args.Error(1)
	}
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *MockService) GetUserByID(ctx context.Context, id pgtype.UUID) (models.UserResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return models.UserResponse{}, args.Error(1)
	}
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *MockService) GetUserByEmail(ctx context.Context, email string) (models.UserResponse, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return models.UserResponse{}, args.Error(1)
	}
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *MockService) ValidatorPassword(ctx context.Context, email, password string) (models.UserResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return models.UserResponse{}, args.Error(1)
	}
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *MockService) UpdateUser(ctx context.Context, id pgtype.UUID, req models.UpdateUserRequest) (models.UserResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return models.UserResponse{}, args.Error(1)
	}
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *MockService) DeletedUser(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) ListUsers(ctx context.Context) ([]models.UserResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return []models.UserResponse{}, args.Error(1)
	}
	return args.Get(0).([]models.UserResponse), args.Error(1)
}

func TestHandler_CreatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   models.UserResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: models.CreateUserRequest{
				FullName: "João Silva",
				Email:    "joao@test.com",
				Password: "123456",
			},
			mockResponse: models.UserResponse{
				ID:       uuid.New(),
				FullName: "João Silva",
				Email:    "joao@test.com",
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"full_name": "João",
				// email missing
			},
			mockResponse:   models.UserResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: models.CreateUserRequest{
				FullName: "João Silva",
				Email:    "joao@test.com",
				Password: "123456",
			},
			mockResponse:   models.UserResponse{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockService)
			handler := NewHandler(mockService)

			router := gin.New()
			router.POST("/users", handler.CreatedUser)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusCreated {
				mockService.On("CreateUser", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)
			}

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				var response models.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	validID := userID.String()

	tests := []struct {
		name           string
		userID         string
		mockResponse   models.UserResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:   "success",
			userID: validID,
			mockResponse: models.UserResponse{
				ID:       userID,
				FullName: "João Silva",
				Email:    "joao@test.com",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid UUID",
			userID:         "invalid-uuid",
			mockResponse:   models.UserResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "user not found",
			userID:         validID,
			mockResponse:   models.UserResponse{},
			mockError:      errors.New("user not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockService)
			handler := NewHandler(mockService)

			router := gin.New()
			router.GET("/users/:id", handler.GetUser)

			req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusOK {
				id := pgtype.UUID{Bytes: userID, Valid: true}
				mockService.On("GetUserByID", mock.Anything, id).Return(tt.mockResponse, tt.mockError)
			}

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response models.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Email, response.Email)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_ListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockResponse   []models.UserResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			mockResponse: []models.UserResponse{
				{ID: uuid.New(), FullName: "João Silva", Email: "joao@test.com"},
				{ID: uuid.New(), FullName: "Maria Silva", Email: "maria@test.com"},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			mockResponse:   nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockService)
			handler := NewHandler(mockService)

			router := gin.New()
			router.GET("/users", handler.ListUsers)

			req := httptest.NewRequest("GET", "/users", nil)
			w := httptest.NewRecorder()

			mockService.On("ListUsers", mock.Anything).Return(tt.mockResponse, tt.mockError)

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response []models.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response, len(tt.mockResponse))
			}
			mockService.AssertExpectations(t)
		})
	}
}
