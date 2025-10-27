package models

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeTransfer TransactionType = "transfer"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

type Transaction struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	UserID      uuid.UUID         `json:"user_id" db:"user_id"`
	AccountID   uuid.UUID         `json:"account_id" db:"account_id"`
	CategoryID  *uuid.UUID        `json:"category_id" db:"category_id"`
	Type        TransactionType   `json:"type" db:"type"`
	Amount      float64           `json:"amount" db:"amount"`
	Description string            `json:"description" db:"description"`
	Status      TransactionStatus `json:"status" db:"status"`
	Date        time.Time         `json:"time" db:"time"`
	CreatedAt   time.Time         `json:"crated_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at" dn:"deleted_at"`
}

type CreateTransactionRequest struct {
	AccountId   uuid.UUID       `json:"account_id" binding:"required"`
	CategoryID  *uuid.UUID      `json:"category_id,omitempty"`
	Type        TransactionType `json:"type" binding:"required,oneof=income expense transfer"`
	Amount      float64         `json:"amount" binding:"required,gt=0"`
	Description string          `json:"description" binding:"required,min=1,max=500"`
	Date        time.Time       `json:"date" binding:"required"`
}

type UpdateTransactionRequest struct {
	AccountId   *uuid.UUID         `json:"account_id" binding:"required"`
	CategoryID  *uuid.UUID         `json:"category_id,omitempty"`
	Type        *TransactionType   `json:"type" binding:"required,oneof=income expense transfer"`
	Amount      *float64           `json:"amount" binding:"required,gt=0"`
	Description *string            `json:"description" binding:"required,min=1,max=500"`
	Status      *TransactionStatus `json:"status,omitempty" binding:"omitempty,oneof=pending completed cancelled"`
	Date        *time.Time         `json:"date" binding:"required"`
}
