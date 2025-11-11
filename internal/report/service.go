package report

import (
	"context"
	"encoding/json"
	"errors"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, userID pgtype.UUID, req models.CreateReportRequest)(models.ReportResponse,error){
	dataJSON, err := buildReportData(ctx,userID,req)
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

func buildReportData(ctx context.Context, userID pgtype.UUID, req models.CreateReportRequest)([]byte, error){
	payload := models.ReportSummary{
		TotalIncome: 0,
		TotalExpense: 0,
		NetBalance: 0,
		TransactionCount: 0,
		TopCategories: []models.CategorySummary{},
	}
	return json.Marshal(payload)
}

func toReportResponse(report sqlc.Report) models.ReportResponse{
	var data interface{}
	if len(report.Data) > 0{
		_ = json.Unmarshal(report.Data, &data)
	}

	return models.ReportResponse{
		ID: report.ID.Bytes,
		Type: models.ReportType(report.Type),
		Title: report.Title,
		Description: report.Description.String,
		StartDate: report.StartDate.Time,
		EndDate: report.CreatedAt.Time,
		Data: data,
		CreatedAt: report.CreatedAt.Time,
		UpdatedAt: report.UpdatedAt.Time,
	}
}