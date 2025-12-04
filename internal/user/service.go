package user

import (
	"context"
	"errors"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// ServiceInterface define o contrato para a camada de serviço de usuários
// Permite criar mocks para testes e facilita inversão de dependência
type ServiceInterface interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (models.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (models.UserResponse, error)
	ValidatorPassword(ctx context.Context, email, password string) (models.UserResponse, error)
	UpdateUser(ctx context.Context, id pgtype.UUID, req models.UpdateUserRequest) (models.UserResponse, error)
	DeletedUser(ctx context.Context, id pgtype.UUID) error
	ListUsers(ctx context.Context) ([]models.UserResponse, error)
}

// Service contém a lógica de negócio para usuários
// Processa dados, aplica regras de negócio e converte entre modelos de domínio e banco
type Service struct {
	repo RepositoryInterface // Repositório para acesso a dados
}

// NewService cria uma nova instância do serviço de usuários
func NewService(repo RepositoryInterface) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateUser cria um novo usuário no sistema
// Aplica hash na senha antes de armazenar (segurança)
// Converte entre modelos de request/response e modelos do banco
func (s *Service) CreateUser(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	// Gerar hash da senha usando bcrypt
	// bcrypt é um algoritmo de hash seguro e lento (protege contra brute force)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserResponse{}, err
	}

	// Criar usuário no banco de dados
	user, err := s.repo.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword), // Armazenar hash, nunca a senha em texto plano
	})
	if err != nil {
		return models.UserResponse{}, err
	}

	// Converter modelo do banco para modelo de resposta
	return models.UserResponse{
		ID:        user.ID.Bytes,
		FullName:  user.FullName,
		Email:     user.Email,
		IsActive:  user.IsActive.Bool,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

// GetUserByID busca um usuário pelo ID
// Converte o modelo do banco para modelo de resposta
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

// GetUserByEmail busca um usuário pelo email
// Útil para validações e buscas por email
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

// ValidatorPassword valida as credenciais de um usuário (email e senha)
// Usado durante o processo de login
// Retorna erro genérico "invalid credentials" para não revelar se email existe
func (s *Service) ValidatorPassword(ctx context.Context, email string, password string) (models.UserResponse, error) {
	// Buscar usuário pelo email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// Retornar erro genérico para não revelar se o email existe
		return models.UserResponse{}, errors.New("invalid credentials")
	}

	// Comparar senha fornecida com hash armazenado
	// bcrypt.CompareHashAndPassword verifica se a senha corresponde ao hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Senha incorreta
		return models.UserResponse{}, errors.New("invalid credentials")
	}

	// Credenciais válidas: retornar dados do usuário
	return models.UserResponse{
		ID:        user.ID.Bytes,
		FullName:  user.FullName,
		Email:     user.Email,
		IsActive:  user.IsActive.Bool,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

// UpdateUser atualiza os dados de um usuário existente
// Não atualiza senha (requer endpoint específico com validação)
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

// DeletedUser realiza soft delete de um usuário
// Marca como deletado ao invés de remover fisicamente (permite recuperação)
func (s *Service) DeletedUser(ctx context.Context, id pgtype.UUID) error {
	return s.repo.DeletedUser(ctx, id)
}

// ListUsers retorna todos os usuários ativos do sistema
// Converte lista de modelos do banco para modelos de resposta
func (s *Service) ListUsers(ctx context.Context) ([]models.UserResponse, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	// Converter cada usuário do banco para modelo de resposta
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
