package category

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

func (m *MockService) Create(ctx context.Context, userID pgtype.UUID, req models.CreateCategoryRequest) (models.CategoryResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return models.CategoryResponse{}, args.Error(1)
	}
	return args.Get(0).(models.CategoryResponse), args.Error(1)
}

func (m *MockService) GetByID(ctx context.Context, id pgtype.UUID) (models.CategoryResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return models.CategoryResponse{}, args.Error(1)
	}
	return args.Get(0).(models.CategoryResponse), args.Error(1)
}

func (m *MockService) GetByUserID(ctx context.Context, userID pgtype.UUID) ([]models.CategoryResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return []models.CategoryResponse{}, args.Error(1)
	}
	return args.Get(0).([]models.CategoryResponse), args.Error(1)
}

func (m *MockService) Update(ctx context.Context, id pgtype.UUID, userID pgtype.UUID, req models.UpdateCategoryRequest) (models.CategoryResponse, error) {
	args := m.Called(ctx, id, userID, req)
	if args.Get(0) == nil {
		return models.CategoryResponse{}, args.Error(1)
	}
	return args.Get(0).(models.CategoryResponse), args.Error(1)
}

func (m *MockService) Delete(ctx context.Context, id pgtype.UUID, userID pgtype.UUID) error {
	args := m.Called(ctx, id, userID)
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
	var req models.CreateCategoryRequest
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	id := pgtype.UUID{Bytes: idUUID, Valid: true}
	out, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
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
	out, err := h.service.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	var req models.UpdateCategoryRequest
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
	id := pgtype.UUID{Bytes: idUUID, Valid: true}

	out, err := h.service.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	raw, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := pgtype.UUID{Bytes: raw.(uuid.UUID), Valid: true}
	id := pgtype.UUID{Bytes: idUUID, Valid: true}

	if err := h.service.Delete(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		setUserID      bool
		mockResponse   models.CategoryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: models.CreateCategoryRequest{
				Name:        "Alimentação",
				Description: "Gastos com comida",
				Color:       "#FF5733",
				Icon:        "food",
			},
			setUserID: true,
			mockResponse: models.CategoryResponse{
				ID:          uuid.New(),
				UserID:      userID,
				Name:        "Alimentação",
				Description: "Gastos com comida",
				Color:       "#FF5733",
				Icon:        "food",
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"name": "Alimentação",
			},
			setUserID:      true,
			mockResponse:   models.CategoryResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthorized - no user_id",
			requestBody: models.CreateCategoryRequest{
				Name:        "Alimentação",
				Description: "Gastos com comida",
				Color:       "#FF5733",
				Icon:        "food",
			},
			setUserID:      false,
			mockResponse:   models.CategoryResponse{},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "service error",
			requestBody: models.CreateCategoryRequest{
				Name:        "Alimentação",
				Description: "Gastos com comida",
				Color:       "#FF5733",
				Icon:        "food",
			},
			setUserID:      true,
			mockResponse:   models.CategoryResponse{},
			mockError:      errors.New("category already exists"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.POST("/categories", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.Create(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.setUserID && (tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusInternalServerError) {
				expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
				mockService.On("Create", mock.Anything, expectedUserID, mock.Anything).Return(tt.mockResponse, tt.mockError)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				var response models.CategoryResponse
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

	categoryID := uuid.New()

	tests := []struct {
		name           string
		id             string
		mockResponse   models.CategoryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			id:   categoryID.String(),
			mockResponse: models.CategoryResponse{
				ID:          categoryID,
				UserID:      uuid.New(),
				Name:        "Alimentação",
				Description: "Gastos com comida",
				Color:       "#FF5733",
				Icon:        "food",
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "invalid-uuid",
			mockResponse:   models.CategoryResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			id:             categoryID.String(),
			mockResponse:   models.CategoryResponse{},
			mockError:      errors.New("not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.GET("/categories/:id", handler.GetByID)

			req := httptest.NewRequest("GET", "/categories/"+tt.id, nil)
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
				var response models.CategoryResponse
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

func TestHandler_ListByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	tests := []struct {
		name           string
		setUserID      bool
		mockResponse   []models.CategoryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:      "success",
			setUserID: true,
			mockResponse: []models.CategoryResponse{
				{
					ID:          uuid.New(),
					UserID:      userID,
					Name:        "Alimentação",
					Description: "Gastos com comida",
					Color:       "#FF5733",
					Icon:        "food",
					IsActive:    true,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          uuid.New(),
					UserID:      userID,
					Name:        "Transporte",
					Description: "Gastos com transporte",
					Color:       "#33FF57",
					Icon:        "car",
					IsActive:    true,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user_id",
			setUserID:      false,
			mockResponse:   []models.CategoryResponse{},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			setUserID:      true,
			mockResponse:   []models.CategoryResponse{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.GET("/categories", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.ListByUser(c)
			})

			req := httptest.NewRequest("GET", "/categories", nil)
			w := httptest.NewRecorder()

			if tt.setUserID {
				expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
				mockService.On("GetByUserID", mock.Anything, expectedUserID).Return(tt.mockResponse, tt.mockError)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response []models.CategoryResponse
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

	categoryID := uuid.New()
	userID := uuid.New()
	validID := categoryID.String()

	tests := []struct {
		name           string
		categoryID     string
		setUserID      bool
		requestBody    interface{}
		mockResponse   models.CategoryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:       "success - update name",
			categoryID: validID,
			setUserID:  true,
			requestBody: models.UpdateCategoryRequest{
				Name: stringPtr("Alimentação Atualizada"),
			},
			mockResponse: models.CategoryResponse{
				ID:          categoryID,
				UserID:      userID,
				Name:        "Alimentação Atualizada",
				Description: "Gastos com comida",
				Color:       "#FF5733",
				Icon:        "food",
				IsActive:    true,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid UUID",
			categoryID:     "invalid-uuid",
			setUserID:      true,
			requestBody:    models.UpdateCategoryRequest{Name: stringPtr("Test")},
			mockResponse:   models.CategoryResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request body",
			categoryID:     validID,
			setUserID:      true,
			requestBody:    map[string]interface{}{"name": 123},
			mockResponse:   models.CategoryResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized - no user_id",
			categoryID:     validID,
			setUserID:      false,
			requestBody:    models.UpdateCategoryRequest{Name: stringPtr("Test")},
			mockResponse:   models.CategoryResponse{},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "service error",
			categoryID: validID,
			setUserID:  true,
			requestBody: models.UpdateCategoryRequest{
				Name: stringPtr("Alimentação Atualizada"),
			},
			mockResponse:   models.CategoryResponse{},
			mockError:      errors.New("category not found"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.PUT("/categories/:id", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.Update(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/categories/"+tt.categoryID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.setUserID && (tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusInternalServerError) {
				parsedID, err := uuid.Parse(tt.categoryID)
				if err == nil {
					id := pgtype.UUID{Bytes: parsedID, Valid: true}
					expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
					mockService.On("Update", mock.Anything, id, expectedUserID, mock.Anything).Return(tt.mockResponse, tt.mockError)
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response models.CategoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, response.ID)
			}
			if tt.setUserID && (tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusInternalServerError) {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	categoryID := uuid.New()
	userID := uuid.New()
	validID := categoryID.String()

	tests := []struct {
		name           string
		categoryID     string
		setUserID      bool
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			categoryID:     validID,
			setUserID:      true,
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid UUID",
			categoryID:     "invalid-uuid",
			setUserID:      true,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized - no user_id",
			categoryID:     validID,
			setUserID:      false,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			categoryID:     validID,
			setUserID:      true,
			mockError:      errors.New("category not found"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.DELETE("/categories/:id", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.Delete(c)
			})

			req := httptest.NewRequest("DELETE", "/categories/"+tt.categoryID, nil)
			w := httptest.NewRecorder()

			if tt.setUserID && (tt.expectedStatus == http.StatusNoContent || tt.expectedStatus == http.StatusInternalServerError) {
				parsedID, err := uuid.Parse(tt.categoryID)
				if err == nil {
					id := pgtype.UUID{Bytes: parsedID, Valid: true}
					expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
					mockService.On("Delete", mock.Anything, id, expectedUserID).Return(tt.mockError)
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusNoContent {
				assert.Empty(t, w.Body.String())
			}
			if tt.setUserID && (tt.expectedStatus == http.StatusNoContent || tt.expectedStatus == http.StatusInternalServerError) {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
