package report

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

func (m *MockService) Create(ctx context.Context, userID pgtype.UUID, req models.CreateReportRequest) (models.ReportResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return models.ReportResponse{}, args.Error(1)
	}
	return args.Get(0).(models.ReportResponse), args.Error(1)
}

func (m *MockService) GetByID(ctx context.Context, id pgtype.UUID, userID pgtype.UUID) (models.ReportResponse, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return models.ReportResponse{}, args.Error(1)
	}
	return args.Get(0).(models.ReportResponse), args.Error(1)
}

func (m *MockService) ListByUser(ctx context.Context, userID pgtype.UUID) ([]models.ReportResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return []models.ReportResponse{}, args.Error(1)
	}
	return args.Get(0).([]models.ReportResponse), args.Error(1)
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
	var req models.CreateReportRequest
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

	raw, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := pgtype.UUID{Bytes: raw.(uuid.UUID), Valid: true}
	id := pgtype.UUID{Bytes: idUUID, Valid: true}
	out, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	startDate := time.Now()
	endDate := startDate.Add(30 * 24 * time.Hour)

	tests := []struct {
		name           string
		requestBody    interface{}
		setUserID      bool
		mockResponse   models.ReportResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: models.CreateReportRequest{
				Type:        models.ReportTypeMonthly,
				Title:       "Relatório Mensal",
				Description: "Relatório de janeiro",
				StartDate:   startDate,
				EndDate:     endDate,
			},
			setUserID: true,
			mockResponse: models.ReportResponse{
				ID:          uuid.New(),
				UserID:      userID,
				Type:        models.ReportTypeMonthly,
				Title:       "Relatório Mensal",
				Description: "Relatório de janeiro",
				StartDate:   startDate,
				EndDate:     endDate,
				Data:        map[string]interface{}{"summary": map[string]interface{}{}},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"type": "monthly",
			},
			setUserID:      true,
			mockResponse:   models.ReportResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthorized - no user_id",
			requestBody: models.CreateReportRequest{
				Type:        models.ReportTypeMonthly,
				Title:       "Relatório Mensal",
				Description: "Relatório de janeiro",
				StartDate:   startDate,
				EndDate:     endDate,
			},
			setUserID:      false,
			mockResponse:   models.ReportResponse{},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "service error",
			requestBody: models.CreateReportRequest{
				Type:        models.ReportTypeMonthly,
				Title:       "Relatório Mensal",
				Description: "Relatório de janeiro",
				StartDate:   startDate,
				EndDate:     endDate,
			},
			setUserID:      true,
			mockResponse:   models.ReportResponse{},
			mockError:      errors.New("start date cannot be after end date"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.POST("/reports", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.Create(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/reports", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.setUserID && (tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusInternalServerError) {
				expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
				mockService.On("Create", mock.Anything, expectedUserID, mock.Anything).Return(tt.mockResponse, tt.mockError)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				var response models.ReportResponse
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

	reportID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		id             string
		setUserID      bool
		mockResponse   models.ReportResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:      "success",
			id:        reportID.String(),
			setUserID: true,
			mockResponse: models.ReportResponse{
				ID:          reportID,
				UserID:      userID,
				Type:        models.ReportTypeMonthly,
				Title:       "Relatório Mensal",
				Description: "Relatório de janeiro",
				StartDate:   time.Now(),
				EndDate:     time.Now(),
				Data:        map[string]interface{}{"summary": map[string]interface{}{}},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "invalid-uuid",
			setUserID:      true,
			mockResponse:   models.ReportResponse{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized - no user_id",
			id:             reportID.String(),
			setUserID:      false,
			mockResponse:   models.ReportResponse{},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "service error",
			id:        reportID.String(),
			setUserID: true,
			mockResponse: models.ReportResponse{},
			mockError:      errors.New("report does not belong to user"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.GET("/reports/:id", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.GetByID(c)
			})

			req := httptest.NewRequest("GET", "/reports/"+tt.id, nil)
			w := httptest.NewRecorder()

			if tt.setUserID && (tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusInternalServerError) {
				parsedID, err := uuid.Parse(tt.id)
				if err == nil {
					id := pgtype.UUID{Bytes: parsedID, Valid: true}
					expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
					mockService.On("GetByID", mock.Anything, id, expectedUserID).Return(tt.mockResponse, tt.mockError)
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response models.ReportResponse
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

func TestHandler_ListByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	tests := []struct {
		name           string
		setUserID      bool
		mockResponse   []models.ReportResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:      "success",
			setUserID: true,
			mockResponse: []models.ReportResponse{
				{
					ID:          uuid.New(),
					UserID:      userID,
					Type:        models.ReportTypeMonthly,
					Title:       "Relatório Mensal Janeiro",
					Description: "Relatório de janeiro",
					StartDate:   time.Now(),
					EndDate:     time.Now(),
					Data:        map[string]interface{}{"summary": map[string]interface{}{}},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				{
					ID:          uuid.New(),
					UserID:      userID,
					Type:        models.ReportTypeMonthly,
					Title:       "Relatório Mensal Fevereiro",
					Description: "Relatório de fevereiro",
					StartDate:   time.Now(),
					EndDate:     time.Now(),
					Data:        map[string]interface{}{"summary": map[string]interface{}{}},
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
			mockResponse:   []models.ReportResponse{},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			setUserID:      true,
			mockResponse:   []models.ReportResponse{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.GET("/reports", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("user_id", userID)
				}
				handler.ListByUser(c)
			})

			req := httptest.NewRequest("GET", "/reports", nil)
			w := httptest.NewRecorder()

			if tt.setUserID {
				expectedUserID := pgtype.UUID{Bytes: userID, Valid: true}
				mockService.On("ListByUser", mock.Anything, expectedUserID).Return(tt.mockResponse, tt.mockError)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response []models.ReportResponse
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



