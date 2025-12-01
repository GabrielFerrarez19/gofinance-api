package report

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) ListByPeriod(ctx context.Context, userID pgtype.UUID, from, to pgtype.Timestamptz) ([]sqlc.Transaction, error) {
	args := m.Called(ctx, userID, from, to)
	if args.Get(0) == nil {
		return []sqlc.Transaction{}, args.Error(1)
	}
	return args.Get(0).([]sqlc.Transaction), args.Error(1)
}

type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) Create(ctx context.Context, arg sqlc.CreateReportParams) (sqlc.Report, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return sqlc.Report{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Report), args.Error(1)
}

func (m *MockReportRepository) GetByID(ctx context.Context, id pgtype.UUID) (sqlc.Report, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return sqlc.Report{}, args.Error(1)
	}
	return args.Get(0).(sqlc.Report), args.Error(1)
}

func (m *MockReportRepository) ListByUser(ctx context.Context, user_id pgtype.UUID) ([]sqlc.Report, error) {
	args := m.Called(ctx, user_id)
	if args.Get(0) == nil {
		return []sqlc.Report{}, args.Error(1)
	}
	return args.Get(0).([]sqlc.Report), args.Error(1)
}


func createTestReport() sqlc.Report {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	startDate := now
	endDate := now.Add(30 * 24 * time.Hour)

	data, _ := json.Marshal(map[string]interface{}{
		"summary": map[string]interface{}{
			"total_income":  1000.0,
			"total_expense": 500.0,
			"net_balance":  500.0,
		},
	})

	return sqlc.Report{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		UserID:      pgtype.UUID{Bytes: userID, Valid: true},
		Type:        "monthly",
		Title:       "Relatório Mensal",
		Description: pgtype.Text{String: "Relatório de janeiro", Valid: true},
		StartDate:   pgtype.Timestamptz{Time: startDate, Valid: true},
		EndDate:     pgtype.Timestamptz{Time: endDate, Valid: true},
		Data:        data,
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}
}

func TestService_Create(t *testing.T) {
	userID := uuid.New()
	uid := pgtype.UUID{Bytes: userID, Valid: true}
	startDate := time.Now()
	endDate := startDate.Add(30 * 24 * time.Hour)

	tests := []struct {
		name           string
		req            models.CreateReportRequest
		mockReport     sqlc.Report
		mockTxList     []sqlc.Transaction
		mockTxError    error
		mockRepoError  error
		wantErr        bool
		description    string
	}{
		{
			name: "success",
			req: models.CreateReportRequest{
				Type:        models.ReportTypeMonthly,
				Title:       "Relatório Mensal",
				Description: "Relatório de janeiro",
				StartDate:   startDate,
				EndDate:     endDate,
			},
			mockReport:    createTestReport(),
			mockTxList:    []sqlc.Transaction{},
			mockTxError:   nil,
			mockRepoError: nil,
			wantErr:       false,
			description:   "must create report successfully",
		},
		{
			name: "start date after end date",
			req: models.CreateReportRequest{
				Type:        models.ReportTypeMonthly,
				Title:       "Relatório Mensal",
				Description: "Relatório de janeiro",
				StartDate:   endDate,
				EndDate:     startDate,
			},
			mockReport:    sqlc.Report{},
			mockTxList:    []sqlc.Transaction{},
			mockTxError:   nil,
			mockRepoError: nil,
			wantErr:       true,
			description:   "should return an error when start date is after end date",
		},
		{
			name: "transaction repository error",
			req: models.CreateReportRequest{
				Type:        models.ReportTypeMonthly,
				Title:       "Relatório Mensal",
				Description: "Relatório de janeiro",
				StartDate:   startDate,
				EndDate:     endDate,
			},
			mockReport:    sqlc.Report{},
			mockTxList:    []sqlc.Transaction{},
			mockTxError:   errors.New("database error"),
			mockRepoError: nil,
			wantErr:       true,
			description:   "should return an error when transaction repository fails",
		},
		{
			name: "report repository error",
			req: models.CreateReportRequest{
				Type:        models.ReportTypeMonthly,
				Title:       "Relatório Mensal",
				Description: "Relatório de janeiro",
				StartDate:   startDate,
				EndDate:     endDate,
			},
			mockReport:    sqlc.Report{},
			mockTxList:    []sqlc.Transaction{},
			mockTxError:   nil,
			mockRepoError: errors.New("database error"),
			wantErr:       true,
			description:   "should return an error when report repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockReportRepository)
			mockTxRepo := new(MockTransactionRepository)
			service := &Service{repo: mockRepo, txRepo: mockTxRepo}

			if !tt.req.StartDate.After(tt.req.EndDate) {
				from := pgtype.Timestamptz{Time: tt.req.StartDate, Valid: true}
				to := pgtype.Timestamptz{Time: tt.req.EndDate, Valid: true}
				mockTxRepo.On("ListByPeriod", mock.Anything, uid, from, to).Return(tt.mockTxList, tt.mockTxError)
				if tt.mockTxError == nil {
					mockRepo.On("Create", mock.Anything, mock.Anything).Return(tt.mockReport, tt.mockRepoError)
				}
			}

			result, err := service.Create(context.Background(), uid, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.name == "start date after end date" {
					assert.Contains(t, err.Error(), "start date cannot be after end date")
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
			}
			if !tt.req.StartDate.After(tt.req.EndDate) {
				mockRepo.AssertExpectations(t)
				if tt.mockTxError == nil {
					mockTxRepo.AssertExpectations(t)
				}
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	userID := uuid.New()
	reportID := uuid.New()
	id := pgtype.UUID{Bytes: reportID, Valid: true}
	uid := pgtype.UUID{Bytes: userID, Valid: true}

	testReport := createTestReport()
	testReport.UserID = uid
	testReport.ID = id

	tests := []struct {
		name         string
		id           pgtype.UUID
		userID       pgtype.UUID
		mockReport   sqlc.Report
		mockError    error
		wantErr      bool
		description  string
	}{
		{
			name:        "success",
			id:          id,
			userID:      uid,
			mockReport:  testReport,
			mockError:   nil,
			wantErr:     false,
			description: "must get report by ID successfully",
		},
		{
			name:        "report not found",
			id:          id,
			userID:      uid,
			mockReport:  sqlc.Report{},
			mockError:   errors.New("report not found"),
			wantErr:     true,
			description: "should return an error when report not found",
		},
		{
			name:        "report does not belong to user",
			id:          id,
			userID:      pgtype.UUID{Bytes: uuid.New(), Valid: true},
			mockReport:  createTestReport(),
			mockError:   nil,
			wantErr:     true,
			description: "should return an error when report does not belong to user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockReportRepository)
			mockTxRepo := new(MockTransactionRepository)
			service := &Service{repo: mockRepo, txRepo: mockTxRepo}

			mockRepo.On("GetByID", mock.Anything, tt.id).Return(tt.mockReport, tt.mockError)

			result, err := service.GetByID(context.Background(), tt.id, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.name == "report does not belong to user" {
					assert.Contains(t, err.Error(), "report does not belong to user")
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListByUser(t *testing.T) {
	userID := uuid.New()
	uid := pgtype.UUID{Bytes: userID, Valid: true}

	tests := []struct {
		name         string
		userID       pgtype.UUID
		mockReports  []sqlc.Report
		mockError    error
		wantErr      bool
		wantLength   int
		description  string
	}{
		{
			name:        "success",
			userID:      uid,
			mockReports: []sqlc.Report{createTestReport()},
			mockError:   nil,
			wantErr:     false,
			wantLength:  1,
			description: "must list reports by user successfully",
		},
		{
			name:        "repository error",
			userID:      uid,
			mockReports: []sqlc.Report{},
			mockError:   errors.New("database error"),
			wantErr:     true,
			wantLength:  0,
			description: "should return an error when repository fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockReportRepository)
			mockTxRepo := new(MockTransactionRepository)
			service := &Service{repo: mockRepo, txRepo: mockTxRepo}

			mockRepo.On("ListByUser", mock.Anything, tt.userID).Return(tt.mockReports, tt.mockError)

			result, err := service.ListByUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLength)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

