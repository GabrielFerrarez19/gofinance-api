package transaction

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

// CreateTransaction godoc
// @Summary Create transaction
// @Description Creates a new transaction for the authenticated user
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body models.CreateTransactionRequest true "Transaction payload"
// @Success 201 {object} models.TransactionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /transactions [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
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

// GetTransaction godoc
// @Summary Get transaction by ID
// @Description Returns transaction details by its identifier
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} models.TransactionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /transactions/{id} [get]
// @Security BearerAuth
func (h *Handler) GetByID(c *gin.Context) {
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

// ListTransactionsByAccount godoc
// @Summary List transactions by account
// @Description Lists transactions associated with a specific account
// @Tags transactions
// @Accept json
// @Produce json
// @Param account_id path string true "Account ID"
// @Success 200 {array} models.TransactionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /transactions/account/{account_id} [get]
// @Security BearerAuth
func (h *Handler) ListByAccount(c *gin.Context) {
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

// ListTransactionsByUser godoc
// @Summary List user transactions
// @Description Lists all transactions for the authenticated user
// @Tags transactions
// @Accept json
// @Produce json
// @Success 200 {array} models.TransactionResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /transactions [get]
// @Security BearerAuth
func (h *Handler) ListByUser(c *gin.Context) {
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

// UpdateTransaction godoc
// @Summary Update transaction
// @Description Updates an existing transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Param transaction body models.UpdateTransactionRequest true "Transaction payload"
// @Success 200 {object} models.TransactionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /transactions/{id} [put]
// @Security BearerAuth
func (h *Handler) Update(c *gin.Context) {
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

// DeleteTransaction godoc
// @Summary Delete transaction
// @Description Deletes a transaction by its identifier
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /transactions/{id} [delete]
// @Security BearerAuth
func (h *Handler) Delete(c *gin.Context) {
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
