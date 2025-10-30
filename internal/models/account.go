package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AccountType string

const (
	AccountTypeChecking   AccountType = "checking"
	AccountTypeSavings    AccountType = "savings"
	AccountTypeCredit     AccountType = "credit"
	AccountTypeInvestment AccountType = "investment"
)

type Account struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	UserID      uuid.UUID   `json:"user_id" db:"user_id"`
	Name        string      `json:"name" db:"name"`
	Type        AccountType `json:"type" db:"type"`
	Balance     float64     `json:"balance" db:"balance"` // Corrigido: era bd:"balance"
	Currency    string      `json:"currency" db:"currency"`
	Description string      `json:"description" db:"description"`
	IsActive    bool        `json:"is_active" db:"is_active"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"` // Corrigido: era UpdateAt
	DeletedAt   *time.Time  `json:"deleted_at" db:"deleted_at"`
}

type CreateAccountRequest struct { // Corrigido: era CreatedAccountRequest
	Name        string         `json:"name" binding:"required,min=2,max=100"` // Corrigido: removido espaços
	Type        AccountType    `json:"type" binding:"required,oneof=checking savings credit investment"`
	Balance     pgtype.Numeric `json:"balance" binding:"gte=0"`
	Currency    string         `json:"currency" binding:"required,len=3"`
	Description pgtype.Text    `json:"description" binding:"max=500"`
}

type UpdateAccountRequest struct { // Corrigido: era UpdatedAccountRequest
	Name        *string      `json:"name" binding:"omitempty,min=2,max=100"` // Corrigido: removido espaços
	Type        *AccountType `json:"type" binding:"omitempty,oneof=checking savings credit investment"`
	Balance     *float64     `json:"balance" binding:"omitempty,gte=0"`
	Currency    *string      `json:"currency" binding:"omitempty,len=3"`
	Description *string      `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool        `json:"is_active,omitempty"` // Corrigido: sintaxe da tag JSON
}

type AccountResponse struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	Name        string      `json:"name"`
	Type        AccountType `json:"type"`
	Balance     float64     `json:"balance"`
	Currency    string      `json:"currency"`
	Description string      `json:"description"`
	IsActive    bool        `json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
