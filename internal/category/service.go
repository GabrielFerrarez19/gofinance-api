package category

import (
	"context"
	"errors"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, userID pgtype.UUID, req models.CreateCategoryRequest) (models.CategoryResponse, error) {
	_, err := s.repo.GetByUserIDAndName(ctx, userID, req.Name)
	if err == nil {
		return models.CategoryResponse{}, errors.New("category already exists with this name")
	}
	category, err := s.repo.Create(ctx, sqlc.CreateCategoryParams{
		UserID:      userID,
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Color:       req.Color,
		Icon:        pgtype.Text{String: req.Icon, Valid: req.Icon != ""},
	})
	if err != nil {
		return models.CategoryResponse{}, err
	}
	return toCategoryResponse(category), nil
}

func (s *Service) GetByID(ctx context.Context, id pgtype.UUID) (models.CategoryResponse, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.CategoryResponse{}, err
	}
	return toCategoryResponse(category), nil
}

func (s *Service) GetByUserID(ctx context.Context, userID pgtype.UUID) ([]models.CategoryResponse, error) {
	categories, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return []models.CategoryResponse{}, err
	}

	result := make([]models.CategoryResponse, len(categories))
	for i, cat := range categories {
		result[i] = toCategoryResponse(cat)
	}

	return result, nil
}

func (s *Service) Update(ctx context.Context, id pgtype.UUID, userID pgtype.UUID, req models.UpdateCategoryRequest) (models.CategoryResponse, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Info().Msgf("categoria %s", string(category.Name))
		log.Info().Msgf("erro %s", err)

		return models.CategoryResponse{}, errors.New("category not found")
	}

	if category.UserID.Bytes != userID.Bytes {
		return models.CategoryResponse{}, errors.New("category does not belong to user")
	}

	updateParams := sqlc.UpdateCategoryParams{
		ID: id,
	}

	if req.Name != nil {
		updateParams.Name = *req.Name
	}

	if req.Description != nil {
		updateParams.Description = pgtype.Text{String: *req.Description, Valid: true}
	}

	if req.Color != nil {
		updateParams.Color = *req.Color
	}

	if req.Icon != nil {
		updateParams.Icon = pgtype.Text{String: *req.Icon, Valid: true}
	}

	if req.IsActive != nil {
		updateParams.IsActive = pgtype.Bool{Bool: *req.IsActive, Valid: true}
	}

	updated, err := s.repo.q.UpdateCategory(ctx, updateParams)
	if err != nil {
		return models.CategoryResponse{}, err
	}

	return toCategoryResponse(updated), nil
}

func (s *Service) Delete(ctx context.Context, id pgtype.UUID, userID pgtype.UUID) error {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("category not found")
	}

	if category.UserID.Bytes != userID.Bytes {
		return errors.New("category does not belong to user")
	}

	return s.repo.SoftDelete(ctx, id)
}

func toCategoryResponse(cat sqlc.Category) models.CategoryResponse {
	return models.CategoryResponse{
		ID:          cat.ID.Bytes,
		UserID:      cat.UserID.Bytes,
		Name:        cat.Name,
		Description: cat.Description.String,
		Color:       cat.Color,
		Icon:        cat.Icon.String,
		IsActive:    cat.IsActive.Bool,
		CreatedAt:   cat.CreatedAt.Time,
		UpdatedAt:   cat.UpdatedAt.Time,
	}
}
