package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GabrielFerrarez19/gofinance-api/internal/database"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/GabrielFerrarez19/gofinance-api/internal/transaction"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type ReportJob struct {
	db *database.DB
}

// NewReportJob cria uma nova instância do job de relatórios
func NewReportJob(db *database.DB) *ReportJob {
	return &ReportJob{
		db: db,
	}
}

// Process processa uma mensagem de geração de relatório
// Busca transações no primeiro, calcula estatísticas e atualiza o relatório
func (j *ReportJob) Process(ctx context.Context, body []byte, headers amqp.Table) (bool, error) {
	// Deserializar payload
	var payload models.ReportJobPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Str("report_id", payload.ReportID).
		Str("user_id", payload.UserID).
		Msg("Processing report generation job")

	// Converter IDs para UUID
	userUUID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return false, fmt.Errorf("invalid user_id: %w", err)
	}

	reportUUID, err := uuid.Parse(payload.ReportID)
	if err != nil {
		return false, fmt.Errorf("invalid report_id: %w", err)
	}

	userID := pgtype.UUID{Bytes: userUUID, Valid: true}
	reportID := pgtype.UUID{Bytes: reportUUID, Valid: true}

	req := models.CreateReportRequest{
		Type:        models.ReportType(payload.Type),
		Title:       payload.Title,
		Description: payload.Description,
		StartDate:   payload.StartDate,
		EndDate:     payload.EndDate,
	}

	// Gerar dados do relatório (buscar transações e calcular estatísticas)
	// Isso pode ser demorado para periodos grandes, por isso é feito assincronamente
	dataJSON, err := j.buildReportData(ctx, userID, req)
	if err != nil {
		return false, fmt.Errorf("failed to build report data: %w", err)
	}

	// Atualizar relatório com os dados gerados
	// Usar trnsação para garantir consistência
	tx, err := j.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("failed ro begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Atualizar campo data do relatório
	_, err = tx.Exec(ctx,
		"UPDATE reports SET data = $1, updated_at = NOW() WHERE id = $2",
		dataJSON, reportID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to update report: %w", err)
	}

	// Confirmar transação
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Str("report_id", payload.ReportID).
		Msg("Report generate successfully")

	return true, nil
}

// buildReportData gera os dados do relatório baseado nas transações do período
// Reutiliza a lógica do serviço de relátorios
func (j *ReportJob) buildReportData(ctx context.Context, userID pgtype.UUID, req models.CreateReportRequest) ([]byte, error) {
	// Criar repositório de transações temporário para buscar dados
	txRepo := transaction.NewRepository(j.db.Pool)

	// Buscar transações no periodo
	from := pgtype.Timestamptz{Time: req.StartDate, Valid: true}
	to := pgtype.Timestamptz{Time: req.EndDate, Valid: true}

	txList, err := txRepo.ListByPeriod(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Calcular estatísticas
	summary := models.ReportSummary{
		TotalIncome:      0,
		TotalExpense:     0,
		NetBalance:       0,
		TransactionCount: len(txList),
		TopCategories:    []models.CategorySummary{},
	}

	categoryTotals := make(map[uuid.UUID]*models.CategorySummary)

	for _, tx := range txList {
		amount := numericToFloat64(tx.Amount)

		if tx.Type == "income" {
			summary.TotalIncome += amount
		} else if tx.Type == "expense" {
			summary.TotalExpense += amount
		}

		if tx.CategoryID.Valid {
			catID, err := pgUUIDToUUID(tx.CategoryID)
			if err == nil {
				summaryEntry, exists := categoryTotals[catID]
				if !exists {
					summaryEntry = &models.CategorySummary{
						CategoryID:   catID,
						CategoryName: "",
						Amount:       0,
						Percentage:   0,
					}
					categoryTotals[catID] = summaryEntry
				}
				summaryEntry.Amount += amount
			}
		}
	}
	summary.NetBalance = summary.TotalIncome - summary.TotalExpense
	totalAbsolute := summary.TotalExpense + summary.TotalExpense

	for _, catSummary := range categoryTotals {
		if totalAbsolute > 0 {
			catSummary.Percentage = (catSummary.Amount / totalAbsolute) * 100
		}
		summary.TopCategories = append(summary.TopCategories, *catSummary)
	}

	// Serializar dados para JSON
	payload := map[string]any{
		"summary":     summary,
		"by_category": summary.TopCategories,
	}

	return json.Marshal(payload)
}

// Helper functions (copiar do report/service.go se necessário)
func numericToFloat64(num pgtype.Numeric) float64 {
	f64, _ := num.Float64Value()
	return f64.Float64
}

func pgUUIDToUUID(p pgtype.UUID) (uuid.UUID, error) {
	if !p.Valid {
		return uuid.Nil, fmt.Errorf("invalid UUID")
	}
	return uuid.FromBytes(p.Bytes[:])
}
