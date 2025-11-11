package account

import (
	"net/http"

	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateAccount godoc
// @Summary Create account
// @Description Creates a new account owned by the authenticated user
// @Tags accounts
// @Accept json
// @Produce json
// @Param account body models.CreateAccountRequest true "Account payload"
// @Success 201 {object} models.AccountResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawUserID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := pgtype.UUID{Bytes: rawUserID.(uuid.UUID), Valid: true}

	acc, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}
	c.JSON(http.StatusCreated, acc)
}

// GetAccount godoc
// @Summary Get account by ID
// @Description Returns account details by its identifier
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} models.AccountResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /accounts/{id} [get]
// @Security BearerAuth
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := pgtype.UUID{Bytes: parsed, Valid: true}

	acc, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, acc)
}

// ListAccounts godoc
// @Summary List accounts
// @Description Returns all accounts for the authenticated user
// @Tags accounts
// @Accept json
// @Produce json
// @Success 200 {array} models.AccountResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts [get]
// @Security BearerAuth
func (h *Handler) ListByUser(c *gin.Context) {
	rawUserID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := pgtype.UUID{Bytes: rawUserID.(uuid.UUID), Valid: true}

	acc, err := h.service.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list"})
		return
	}

	c.JSON(http.StatusOK, acc)
}

// UpdateAccount godoc
// @Summary Update account
// @Description Updates an existing account
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param account body models.UpdateAccountRequest true "Account payload"
// @Success 200 {object} models.AccountResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /accounts/{id} [put]
// @Security BearerAuth
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := pgtype.UUID{Bytes: parsed, Valid: true}
	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, acc)
}

// DeleteAccount godoc
// @Summary Delete account
// @Description Deletes an account by its identifier
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /accounts/{id} [delete]
// @Security BearerAuth
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := pgtype.UUID{Bytes: parsed, Valid: true}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
