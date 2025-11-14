package user

import (
	"context"
	"errors"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type ServiceInterface interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (models.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (models.UserResponse, error)
	ValidatorPassword(ctx context.Context, email, password string) (models.UserResponse, error)
	UpdateUser(ctx context.Context, id pgtype.UUID, req models.UpdateUserRequest) (models.UserResponse, error)
	DeletedUser(ctx context.Context, id pgtype.UUID) error
	ListUsers(ctx context.Context) ([]models.UserResponse, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateUser(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserResponse{}, err
	}

	user, err := s.repo.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return models.UserResponse{}, err
	}

	return models.UserResponse{
		ID:        user.ID.Bytes,
		FullName:  user.FullName,
		Email:     user.Email,
		IsActive:  user.IsActive.Bool,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func (s *Service) GetUserByID(ctx context.Context, id pgtype.UUID) (models.UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return models.UserResponse{}, err
	}

	return models.UserResponse{
		ID:        user.ID.Bytes,
		FullName:  user.FullName,
		Email:     user.Email,
		IsActive:  user.IsActive.Bool,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (models.UserResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return models.UserResponse{}, err
	}

	return models.UserResponse{
		ID:        user.ID.Bytes,
		FullName:  user.FullName,
		Email:     user.Email,
		IsActive:  user.IsActive.Bool,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func (s *Service) ValidatorPassword(ctx context.Context, email string, password string) (models.UserResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return models.UserResponse{}, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return models.UserResponse{}, errors.New("invalid credentials")
	}

	return models.UserResponse{
		ID:        user.ID.Bytes,
		FullName:  user.FullName,
		Email:     user.Email,
		IsActive:  user.IsActive.Bool,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, id pgtype.UUID, req models.UpdateUserRequest) (models.UserResponse, error) {
	user, err := s.repo.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       id,
		FullName: req.FullName,
		Email:    req.Email,
	})
	if err != nil {
		return models.UserResponse{}, err
	}

	return models.UserResponse{
		ID:        user.ID.Bytes,
		FullName:  user.FullName,
		Email:     user.Email,
		IsActive:  user.IsActive.Bool,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func (s *Service) DeletedUser(ctx context.Context, id pgtype.UUID) error {
	return s.repo.DeletedUser(ctx, id)
}

func (s *Service) ListUsers(ctx context.Context) ([]models.UserResponse, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	var response []models.UserResponse
	for _, user := range users {
		response = append(response, models.UserResponse{
			ID:        user.ID.Bytes,
			FullName:  user.FullName,
			Email:     user.Email,
			IsActive:  user.IsActive.Bool,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
		})
	}

	return response, nil
}
