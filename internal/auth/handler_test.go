package auth

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type AuthServiceInterface interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	SetAuthCookies(c *gin.Context, response *LoginResponse)
	ClearAuthCookies(c *gin.Context)
}

type MockService struct {
	mock.Mock
}

func (m *MockService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockService) SetAuthCookies(c *gin.Context, response *LoginResponse) {
	m.Called(c, response)
}

func (m *MockService) ClearAuthCookies(c *gin.Context) {
	m.Called(c)
}

type TestHandler struct {
	service AuthServiceInterface
}

func NewTestHandler(service AuthServiceInterface) *TestHandler {
	return &TestHandler{
		service: service,
	}
}

func (h *TestHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	h.service.SetAuthCookies(c, response)

	c.JSON(http.StatusOK, response)
}

func (h *TestHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	h.service.SetAuthCookies(c, response)

	c.JSON(http.StatusOK, response)
}

func (h *TestHandler) Logout(c *gin.Context) {
	h.service.ClearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func TestHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *LoginResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: LoginRequest{
				Email:    "joao@test.com",
				Password: "123456",
			},
			mockResponse: &LoginResponse{
				User: models.UserResponse{
					ID:       uuid.New(),
					FullName: "João Silva",
					Email:    "joao@test.com",
				},
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
				ExpiresIn:    3600,
				TokenType:    "Bearer",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"email": "joao@test.com",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			requestBody: LoginRequest{
				Email:    "joao@test.com",
				Password: "wrong_password",
			},
			mockResponse:   nil,
			mockError:      errors.New("invalid credentials"),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.POST("/auth/login", handler.Login)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusUnauthorized {
				mockService.On("Login", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)
				if tt.expectedStatus == http.StatusOK {
					mockService.On("SetAuthCookies", mock.Anything, mock.Anything).Return()
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
			}
			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusUnauthorized {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *LoginResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: map[string]string{
				"refresh_token": "valid_refresh_token",
			},
			mockResponse: &LoginResponse{
				User: models.UserResponse{
					ID:       uuid.New(),
					FullName: "João Silva",
					Email:    "joao@test.com",
				},
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
				ExpiresIn:    3600,
				TokenType:    "Bearer",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"token": "refresh_token",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid refresh token",
			requestBody: map[string]string{
				"refresh_token": "invalid_token",
			},
			mockResponse:   nil,
			mockError:      errors.New("invalid refresh token"),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			handler := NewTestHandler(mockService)

			router := gin.New()
			router.POST("/auth/refresh", handler.RefreshToken)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusUnauthorized {
				mockService.On("RefreshToken", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)
				if tt.expectedStatus == http.StatusOK {
					mockService.On("SetAuthCookies", mock.Anything, mock.Anything).Return()
				}
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
			}
			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusUnauthorized {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockService)
	handler := NewTestHandler(mockService)

	router := gin.New()
	router.POST("/auth/logout", handler.Logout)

	mockService.On("ClearAuthCookies", mock.Anything).Return()

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])
	mockService.AssertExpectations(t)
}

func TestHandler_Me(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	userEmail := "joao@test.com"
	userFullName := "João Silva"

	tests := []struct {
		name           string
		setUserContext bool
		expectedStatus int
	}{
		{
			name:           "success",
			setUserContext: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user in context",
			setUserContext: false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(&Service{})

			router := gin.New()
			router.GET("/auth/me", func(c *gin.Context) {
				if tt.setUserContext {
					c.Set("user_id", userID)
					c.Set("user_email", userEmail)
					c.Set("user_full_name", userFullName)
				}
				handler.Me(c)
			})

			req := httptest.NewRequest("GET", "/auth/me", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response models.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, userID, response.ID)
				assert.Equal(t, userEmail, response.Email)
				assert.Equal(t, userFullName, response.FullName)
			}
		})
	}
}
