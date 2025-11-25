package account

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func (m *MockService) Create(ctx context.Context, userID pgtype.UUID, req models.CreateAccountRequest) (models.AccountResponse, error) {
	args := m.Called(ctx, userID, req)
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

func (m *MockService) Update(ctx context.Context, id pgtype.UUID, req models.UpdateAccountRequest) (models.AccountResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return models.AccountResponse{}, args.Error(1)
	}
	return args.Get(0).(models.AccountResponse), args.Error(1)
}

func (m *MockService) Delete(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestHancdler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    any
		mockResponse   models.AccountResponse
		modkError      error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: models.CreateAccountRequest{
				Name:        "Gabriel Cristiano Ferrarez",
				Type:        "checking",
				Balance:     1000.50,
				Currency:    "BRL",
				Description: "Conta pricipal",
			},
			mockResponse: models.AccountResponse{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Name:        "Gabriel Cristiano Ferrarez",
				Type:        "checking",
				Balance:     1000.50,
				Currency:    "BRL",
				Description: "Conta principal",
				IsActive:    true,
				CreatedAt:   time.Now(),
			},
			modkError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]any{
				"name": "Gabriel",
			},
			mockResponse:   models.AccountResponse{},
			modkError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: models.CreateAccountRequest{
				Name:        "Gabriel Cristiano Ferrarez",
				Type:        "checking",
				Balance:     1000.50,
				Currency:    "BRL",
				Description: "Conta pricipal",
			},
			mockResponse:   models.AccountResponse{},
			modkError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewHandler(mockService)

			userID := uuid.New()
			router := gin.New()
			router.POST("/accounts", func(c *gin.Context) {
				c.Set("user_id", userID)
				handler.Create(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusInternalServerError {
				expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
				mockService.On("Create", mock.Anything, expectedUserID, mock.Anything).Return(tt.mockResponse, tt.modkError)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				var response models.AccountResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
			}
			if tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusInternalServerError {
				mockService.AssertExpectations(t)
			}

		})
	}
}
