package transaction

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

func (m *MockService) Create(ctx context.Context, userID pgtype.UUID, req models.CreateTransactionRequest) (models.TransactionResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return models.TransactionResponse{}, args.Error(1)
	}
	return args.Get(0).(models.TransactionResponse), args.Error(1)
}

func (m *MockService) GetByID(ctx context.Context, id pgtype.UUID) (models.TransactionResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return models.TransactionResponse{}, args.Error(1)
	}
	return args.Get(0).(models.TransactionResponse), args.Error(1)
}

func (m *MockService) ListByAccount(ctx context.Context, accountID pgtype.UUID) ([]models.TransactionResponse, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return []models.TransactionResponse{}, args.Error(1)
	}
	return args.Get(0).([]models.TransactionResponse), args.Error(1)
}

func (m *MockService) ListByUser(ctx context.Context, userID pgtype.UUID) ([]models.TransactionResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return []models.TransactionResponse{}, args.Error(1)
	}
	return args.Get(0).([]models.TransactionResponse), args.Error(1)
}

func (m *MockService) Update(ctx context.Context, id pgtype.UUID, req models.UpdateTransactionRequest) (models.TransactionResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return models.TransactionResponse{}, args.Error(1)
	}
	return args.Get(0).(models.TransactionResponse), args.Error(1)
}

func (m *MockService) Delete(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type TestHandler struct {
	service *MockService
}

func NewTestHandler(service *MockService) *TestHandler {
	return &TestHandler{
		service: service,
	}
}

func (h *TestHandler) Create(c *gin.Context) {
	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	raw, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := pgtype.UUID{Bytes: raw.(uuid.UUID), Valid: true}
	out, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *TestHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	idUUID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id "})
		return
	}
	id := pgtype.UUID{Bytes: idUUID, Valid: true}

	out, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *TestHandler) ListByAccount(c *gin.Context) {
	accStr := c.Param("account_id")
	accUUID, err := uuid.Parse(accStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	accountID := pgtype.UUID{Bytes: accUUID, Valid: true}

	out, err := h.service.ListByAccount(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *TestHandler) ListByUser(c *gin.Context) {
	raw, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := pgtype.UUID{Bytes: raw.(uuid.UUID), Valid: true}

	out, err := h.service.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list"})
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *TestHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	idUUID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	id := pgtype.UUID{Bytes: idUUID, Valid: true}

	var req models.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *TestHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	idUUID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	id := pgtype.UUID{Bytes: idUUID, Valid: true}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	accountID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		setUserID      bool
		mockResponse   models.TransactionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: models.CreateTransactionRequest{
				AccountID:   accountID,
				Type:        models.TransactionIncome,
				Amount:      1000.50,
				Description: "Salário",
				Date:        time.Now(),
			},
			setUserID: true,
			mockResponse: models.TransactionResponse{
				ID:          uuid.New(),
				UserID:      userID,
				AccountID:   accountID,
				Type:        models.TransactionIncome,
				Amount:      1000.50,
				Description: "Salário",
				Status:      models.StatusCompleted,
				Date:        time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"account_id": accountID.String(),
			},
			setUserID:      true,
			mockResponse:   models.TransactionResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthorized - no user_id",
			requestBody: models.CreateTransactionRequest{
				AccountID:   accountID,
				Type:        models.TransactionIncome,
				Amount:      1000.50,
				Description: "Salário",
				Date:        time.Now(),
			},
			setUserID:      false,
			mockResponse:   models.TransactionResponse{},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "service error",
			requestBody: models.CreateTransactionRequest{
				AccountID:   accountID,
				Type:        models.TransactionIncome,
				Amount:      1000.50,
				Description: "Salário",
				Date:        time.Now(),
			},
			setUserID:      true,
			mockResponse:   models.TransactionResponse{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.POST("/transactions", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.Create(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.setUserID && (tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusInternalServerError) {
				expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
				mockService.On("Create", mock.Anything, expectedUserID, mock.Anything).Return(tt.mockResponse, tt.mockError)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				var response models.TransactionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
			}
			if tt.setUserID && (tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusInternalServerError) {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	txID := uuid.New()

	tests := []struct {
		name           string
		id             string
		mockResponse   models.TransactionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			id:   txID.String(),
			mockResponse: models.TransactionResponse{
				ID:          txID,
				UserID:      uuid.New(),
				AccountID:   uuid.New(),
				Type:        models.TransactionIncome,
				Amount:      1000.50,
				Description: "Salário",
				Status:      models.StatusCompleted,
				Date:        time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "invalid-uuid",
			mockResponse:   models.TransactionResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			id:             txID.String(),
			mockResponse:   models.TransactionResponse{},
			mockError:      errors.New("not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.GET("/transactions/:id", handler.GetByID)

			req := httptest.NewRequest("GET", "/transactions/"+tt.id, nil)
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
				var response models.TransactionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, response.ID)
			}
			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusNotFound {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_ListByAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	accountID := uuid.New()

	tests := []struct {
		name           string
		accountID      string
		mockResponse   []models.TransactionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:      "success",
			accountID: accountID.String(),
			mockResponse: []models.TransactionResponse{
				{
					ID:          uuid.New(),
					AccountID:   accountID,
					Type:        models.TransactionIncome,
					Amount:      1000.50,
					Description: "Salário",
					Status:      models.StatusCompleted,
					Date:        time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid account id",
			accountID:      "invalid-uuid",
			mockResponse:   []models.TransactionResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			accountID:      accountID.String(),
			mockResponse:   []models.TransactionResponse{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.GET("/transactions/account/:account_id", handler.ListByAccount)

			req := httptest.NewRequest("GET", "/transactions/account/"+tt.accountID, nil)
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusInternalServerError {
				parsedID, err := uuid.Parse(tt.accountID)
				if err == nil {
					id := pgtype.UUID{Bytes: parsedID, Valid: true}
					mockService.On("ListByAccount", mock.Anything, id).Return(tt.mockResponse, tt.mockError)
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response []models.TransactionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response, len(tt.mockResponse))
			}
			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusInternalServerError {
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
		mockResponse   []models.TransactionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:      "success",
			setUserID: true,
			mockResponse: []models.TransactionResponse{
				{
					ID:          uuid.New(),
					UserID:      userID,
					AccountID:   uuid.New(),
					Type:        models.TransactionIncome,
					Amount:      1000.50,
					Description: "Salário",
					Status:      models.StatusCompleted,
					Date:        time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user_id",
			setUserID:      false,
			mockResponse:   []models.TransactionResponse{},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			setUserID:      true,
			mockResponse:   []models.TransactionResponse{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.GET("/transactions", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.ListByUser(c)
			})

			req := httptest.NewRequest("GET", "/transactions", nil)
			w := httptest.NewRecorder()

			if tt.setUserID {
				expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
				mockService.On("ListByUser", mock.Anything, expectedUserID).Return(tt.mockResponse, tt.mockError)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response []models.TransactionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response, len(tt.mockResponse))
			}
			if tt.setUserID {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	txID := uuid.New()
	validID := txID.String()

	tests := []struct {
		name           string
		txID           string
		requestBody    interface{}
		mockResponse   models.TransactionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success - update amount",
			txID: validID,
			requestBody: models.UpdateTransactionRequest{
				Amount: floatPtr(2000.75),
			},
			mockResponse: models.TransactionResponse{
				ID:          txID,
				Amount:      2000.75,
				Description: "Salário atualizado",
				Status:      models.StatusCompleted,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid UUID",
			txID:           "invalid-uuid",
			requestBody:    models.UpdateTransactionRequest{Amount: floatPtr(1000.0)},
			mockResponse:   models.TransactionResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request body",
			txID:           validID,
			requestBody:    map[string]interface{}{"amount": "invalid"},
			mockResponse:   models.TransactionResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "not found",
			txID: validID,
			requestBody: models.UpdateTransactionRequest{
				Amount: floatPtr(2000.75),
			},
			mockResponse:   models.TransactionResponse{},
			mockError:      errors.New("not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.PUT("/transactions/:id", handler.Update)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/transactions/"+tt.txID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusNotFound {
				parsedID, err := uuid.Parse(tt.txID)
				if err == nil {
					id := pgtype.UUID{Bytes: parsedID, Valid: true}
					mockService.On("Update", mock.Anything, id, mock.Anything).Return(tt.mockResponse, tt.mockError)
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response models.TransactionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, response.ID)
			}
			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusNotFound {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	txID := uuid.New()
	validID := txID.String()

	tests := []struct {
		name           string
		txID           string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			txID:           validID,
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid UUID",
			txID:           "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			txID:           validID,
			mockError:      errors.New("not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.DELETE("/transactions/:id", handler.Delete)

			req := httptest.NewRequest("DELETE", "/transactions/"+tt.txID, nil)
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusNoContent || (tt.expectedStatus == http.StatusNotFound && tt.txID == validID) {
				parsedID, err := uuid.Parse(tt.txID)
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

func floatPtr(f float64) *float64 {
	return &f
}





