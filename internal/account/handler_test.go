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

func TestHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		id             string
		mockResponse   models.AccountResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			id:   uuid.New().String(),
			mockResponse: models.AccountResponse{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				Name:        "Gabriel Cristiano Ferrarez",
				Type:        "checking",
				Balance:     1000.50,
				Currency:    "BRL",
				Description: "conta principal",
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			id:             uuid.New().String(),
			mockResponse:   models.AccountResponse{},
			mockError:      errors.New("not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			id:             "invalid-uuid",
			mockResponse:   models.AccountResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewHandler(mockService)

			router := gin.New()
			router.GET("/account/:id", handler.GetByID)

			req := httptest.NewRequest("GET", "/account/"+tt.id, nil)
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusNotFound {
				parsedID, err := uuid.Parse(tt.id)
				if err == nil {
					id := pgtype.UUID{Bytes: parsedID, Valid: true}
					mockService.On("GetByID", mock.Anything, id).Return(tt.mockResponse, tt.mockError)
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response models.AccountResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Name, response.Name)
				assert.Equal(t, tt.mockResponse.ID, response.ID)
			}
			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusNotFound {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_ListByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	tests := []struct {
		name           string
		setUserID      bool
		mockResponse   []models.AccountResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:      "success",
			setUserID: true,
			mockResponse: []models.AccountResponse{
				{
					ID:          uuid.New(),
					UserID:      userID,
					Name:        "Gabriel Cristiano Ferrarez",
					Type:        "checking",
					Balance:     1000.50,
					Currency:    "BRL",
					Description: "conta principal",
					IsActive:    true,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          uuid.New(),
					UserID:      userID,
					Name:        "Conta Poupança",
					Type:        "savings",
					Balance:     5000.00,
					Currency:    "BRL",
					Description: "conta secundaria",
					IsActive:    true,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user_id in context",
			setUserID:      false,
			mockResponse:   []models.AccountResponse{},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			setUserID:      true,
			mockResponse:   []models.AccountResponse{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewHandler(mockService)

			router := gin.New()
			router.GET("/accounts", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.ListByUser(c)
			})

			req := httptest.NewRequest("GET", "/accounts", nil)
			w := httptest.NewRecorder()

			if tt.setUserID {
				expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
				mockService.On("ListByUser", mock.Anything, expectedUserID).Return(tt.mockResponse, tt.mockError)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response []models.AccountResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response, len(tt.mockResponse))
				if len(tt.mockResponse) > 0 {
					assert.Equal(t, tt.mockResponse[0].Name, response[0].Name)
				}
			}
			if tt.setUserID {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	accountID := uuid.New()
	validID := accountID.String()

	tests := []struct {
		name           string
		accountID      string
		requestBody    any
		mockResponse   models.AccountResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:      "success - update name",
			accountID: validID,
			requestBody: models.UpdateAccountRequest{
				Name: stringPtr("Rafael test updated"),
			},
			mockResponse: models.AccountResponse{
				ID:          accountID,
				UserID:      uuid.New(),
				Name:        "Rafael test updated",
				Type:        "checking",
				Balance:     1000.50,
				Currency:    "BRL",
				Description: "conta principal",
				IsActive:    true,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:      "success - update balance",
			accountID: validID,
			requestBody: models.UpdateAccountRequest{
				Balance: floatPtr(2000.75),
			},
			mockResponse: models.AccountResponse{
				ID:          accountID,
				UserID:      uuid.New(),
				Name:        "Rafael test",
				Type:        "checking",
				Balance:     2000.75,
				Currency:    "BRL",
				Description: "conta principal",
				IsActive:    true,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid UUID",
			accountID:      "invalid-uuid",
			requestBody:    models.UpdateAccountRequest{Name: stringPtr("Test")},
			mockResponse:   models.AccountResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request body",
			accountID:      validID,
			requestBody:    map[string]any{"name": 123},
			mockResponse:   models.AccountResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "account not found",
			accountID: validID,
			requestBody: models.UpdateAccountRequest{
				Name: stringPtr("Rafael test updated"),
			},
			mockResponse:   models.AccountResponse{},
			mockError:      errors.New("account not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewHandler(mockService)

			router := gin.New()
			router.PUT("/accounts/:id", handler.Update)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/accounts/"+tt.accountID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusNotFound {
				parsedID, err := uuid.Parse(tt.accountID)
				if err == nil {
					id := pgtype.UUID{Bytes: parsedID, Valid: true}
					mockService.On("Update", mock.Anything, id, mock.Anything).Return(tt.mockResponse, tt.mockError)
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response models.AccountResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, response.ID)
				assert.Equal(t, tt.mockResponse.Name, response.Name)
			}
			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusNotFound {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	accountID := uuid.New()
	validID := accountID.String()

	tests := []struct {
		name           string
		accountID      string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			accountID:      validID,
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid UUID",
			accountID:      "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "account not found",
			accountID:      validID,
			mockError:      errors.New("account not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "service error",
			accountID:      validID,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewHandler(mockService)

			router := gin.New()
			router.DELETE("/accounts/:id", handler.Delete)

			req := httptest.NewRequest("DELETE", "/accounts/"+tt.accountID, nil)
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusNoContent || (tt.expectedStatus == http.StatusNotFound && tt.accountID == validID) {
				parsedID, err := uuid.Parse(tt.accountID)
				if err == nil {
					id := pgtype.UUID{Bytes: parsedID, Valid: true}
					mockService.On("Delete", mock.Anything, id).Return(tt.mockError)
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusNoContent {
				assert.Empty(t, w.Body.String())
			}
			if tt.expectedStatus == http.StatusNoContent || tt.expectedStatus == http.StatusNotFound {
				mockService.AssertExpectations(t)
			}
		})
	}
}
