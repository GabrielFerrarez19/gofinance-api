package models

import (
	"time"

	"github.com/google/uuid"
)

type ReportType string

const (
	ReportTypeMonthly ReportType = "monthly"
	ReportTypeYearly  ReportType = "yearly"
	ReportTypeCustom  ReportType = "custom"
)

type Report struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	Type        ReportType `json:"type" db:"type"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     time.Time  `json:"end_date" db:"end_date"`
	Data        string     `json:"data" db:"data"` // JSON com os dados do relatório
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateReportRequest struct {
	Type        ReportType `json:"type" binding:"required,oneof=monthly yearly custom"`
	Title       string     `json:"title" binding:"required,min=2,max=100"`
	Description string     `json:"description" binding:"max=500"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     time.Time  `json:"end_date" binding:"required"`
}

type ReportSummary struct {
	TotalIncome      float64           `json:"total_income"`
	TotalExpense     float64           `json:"total_expense"`
	NetBalance       float64           `json:"net_balance"`
	TransactionCount int               `json:"transaction_count"`
	TopCategories    []CategorySummary `json:"top_categories"`
}

type CategorySummary struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
	Percentage   float64   `json:"percentage"`
}

type ReportResponse struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Type        ReportType `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	Data        any        `json:"data"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ReportJobPayload representa os dados necessários para gerar um relatório
type ReportJobPayload struct {
	UserID      string    `json:"user_id"`     // ID do usuário
	ReportID    string    `json:"report_id"`   // ID do relatório a ser gerado
	Type        string    `json:"type"`        // Tipo de relatório
	Title       string    `json:"title"`       // Título do relatório
	Description string    `json:"description"` // Descrição do relatório
	StartDate   time.Time `json:"start_date"`  // Data inicial
	EndDate     time.Time `json:"end_date"`    // Data final
}
