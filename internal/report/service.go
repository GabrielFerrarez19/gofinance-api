package report

import (
	"context"
	"encoding/json"
	"errors"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/GabrielFerrarez19/gofinance-api/internal/transaction"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	repo *Repository
	txRepo *transaction.Repository
}

func NewService(repo *Repository, txRepo *transaction.Repository) *Service {
	return &Service{
		repo: repo,
		txRepo: txRepo,
	}
}

func (s *Service) Create(ctx context.Context, userID pgtype.UUID, req models.CreateReportRequest)(models.ReportResponse,error){
	if req.StartDate.After(req.EndDate){
		return models.ReportResponse{}, errors.New("start date cannot be after end date")
	}
	dataJSON, err := s.buildReportData(ctx,userID,req)
	if err != nil{
		return models.ReportResponse{}, err
	}

	report, err := s.repo.Create(ctx, sqlc.CreateReportParams{
		UserID: userID,
		Type: string(req.Type),
		Title: req.Title,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
		StartDate: pgtype.Timestamptz{Time: req.StartDate, Valid: true},
		EndDate: pgtype.Timestamptz{Time: req.EndDate, Valid: true},
		Data: dataJSON,
	})
	if err != nil{
		return models.ReportResponse{}, err
	}

	return  toReportResponse(report), nil

}

func (s *Service) GetByID(ctx context.Context, id pgtype.UUID, userID pgtype.UUID)(models.ReportResponse, error){
	report, err := s.repo.GetByID(ctx,id)
	if err != nil{
		return models.ReportResponse{}, err
	}

	if report.UserID.Bytes != userID.Bytes{
		return models.ReportResponse{}, errors.New("report does not belong to user")
	}

	return toReportResponse(report), nil
}

func (s *Service) ListByUser(ctx context.Context, userID pgtype.UUID)([]models.ReportResponse, error){
	reports, err := s.repo.ListByUser(ctx,userID)
	if err != nil{
		return nil,err
	}

	result := make([]models.ReportResponse, len(reports))
	for i,r := range reports{
		result[i] = toReportResponse(r)
	}

	return result, nil
} 

func (s *Service) buildReportData(ctx context.Context, userID pgtype.UUID, req models.CreateReportRequest)([]byte, error){
	from := pgtype.Timestamptz{Time: req.StartDate, Valid: true}
	to := pgtype.Timestamptz{Time: req.EndDate, Valid: true}

	txList, err := s.txRepo.ListByPeriod(ctx,userID,from,to)
	if err != nil{
		return nil, err
	}

	summary := models.ReportSummary{
		TotalIncome: 0,
		TotalExpense: 0,
		NetBalance: 0,
		TransactionCount: len(txList),
		TopCategories: []models.CategorySummary{},
	}

	categoryTotals := make(map[uuid.UUID]*models.CategorySummary)

	for _, tx := range txList{
		amount := numericToFloat64(tx.Amount) 

		if tx.Type == "income"{
			summary.TotalIncome += amount
		}else if tx.Type == "expense"{
			summary.TotalExpense += amount
		}

		if tx.CategoryID.Valid{
			catID , err := pgUUIDToUUID(tx.CategoryID)
			if err != nil{
				summaryEntry, exists := categoryTotals[catID]
				if !exists{
					summaryEntry = &models.CategorySummary{
						CategoryID: catID,
						CategoryName: "",
						Amount: 0,
						Percentage: 0,
					}
					categoryTotals[catID] = summaryEntry
				}
				summaryEntry.Amount += amount
			}
		}
	}

	summary.NetBalance = summary.TotalIncome - summary.TotalExpense

	totalAbsolute := summary.TotalIncome + summary.TotalExpense

	for _, catSumary := range categoryTotals{
		if totalAbsolute > 0 {
			catSumary.Percentage = (catSumary.Amount / totalAbsolute) * 100
		}
		summary.TopCategories = append(summary.TopCategories, *catSumary)
	}

	payload := map[string]any{
		"summary": summary,
		"by_category": summary.TopCategories,
	}

	return json.Marshal(payload)
}

func toReportResponse(report sqlc.Report) models.ReportResponse{
	var payload any
	if len(report.Data) > 0{
		_ = json.Unmarshal(report.Data, &payload)
	}

	return models.ReportResponse{
		ID: report.ID.Bytes,
		Type: models.ReportType(report.Type),
		Title: report.Title,
		Description: report.Description.String,
		StartDate: report.StartDate.Time,
		EndDate: report.CreatedAt.Time,
		Data: payload,
		CreatedAt: report.CreatedAt.Time,
		UpdatedAt: report.UpdatedAt.Time,
	}
}


func numericToFloat64(num pgtype.Numeric) float64 {
    // Tenta extrair o valor como float64
    f64, _ := num.Float64Value()
    return f64.Float64
}


func pgUUIDToUUID(p pgtype.UUID) (uuid.UUID, error) {
    if !p.Valid {
        return uuid.Nil, errors.New("invalid UUID")
    }
    return uuid.FromBytes(p.Bytes[:])
}