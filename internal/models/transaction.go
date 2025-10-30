package models

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionIncome   TransactionType = "income"
	TransactionExpense  TransactionType = "expense"
	TransactionTransfer TransactionType = "transfer"
)

type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusCompleted TransactionStatus = "completed"
	StatusCancelled TransactionStatus = "cancelled"
)

type CreateTransactionRequest struct {
	AccountID   uuid.UUID          `json:"account_id" binding:"required"`
	CategoryID  *uuid.UUID         `json:"category_id,omitempty"`
	Type        TransactionType    `json:"type" binding:"required,oneof=income expense transfer"`
	Amount      float64            `json:"amount" binding:"required,gt=0"`
	Description string             `json:"description" binding:"required,max=500"`
	Status      *TransactionStatus `json:"status,omitempty" binding:"omitempty,oneof=pending completed cancelled"`
	Date        time.Time          `json:"date" binding:"required"`
}

type UpdateTransactionRequest struct {
	AccountID   *uuid.UUID         `json:"account_id,omitempty"`
	CategoryID  *uuid.UUID         `json:"category_id,omitempty"`
	Type        *TransactionType   `json:"type,omitempty" binding:"omitempty,oneof=income expense transfer"`
	Amount      *float64           `json:"amount,omitempty" binding:"omitempty,gt=0"`
	Description *string            `json:"description,omitempty" binding:"omitempty,max=500"`
	Status      *TransactionStatus `json:"status,omitempty" binding:"omitempty,oneof=pending completed cancelled"`
	Date        *time.Time         `json:"date,omitempty"`
}

type TransactionResponse struct {
	ID          uuid.UUID         `json:"id"`
	UserID      uuid.UUID         `json:"user_id"`
	AccountID   uuid.UUID         `json:"account_id"`
	CategoryID  *uuid.UUID        `json:"category_id,omitempty"`
	Type        TransactionType   `json:"type"`
	Amount      float64           `json:"amount"`
	Description string            `json:"description"`
	Status      TransactionStatus `json:"status"`
	Date        time.Time         `json:"date"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
