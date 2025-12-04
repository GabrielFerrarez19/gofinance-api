package user

import (
	"context"

	sqlc "github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RepositoryInterface define o contrato para operações de banco de dados de usuários
// Permite criar mocks para testes e facilita inversão de dependência
type RepositoryInterface interface {
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
	GetUserByEmail(ctx context.Context, email string) (sqlc.User, error)
	UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error)
	DeletedUser(ctx context.Context, id pgtype.UUID) error
	ListUsers(ctx context.Context) ([]sqlc.User, error)
}

// Repository implementa o acesso a dados de usuários
// Usa SQLC para executar queries type-safe geradas a partir de SQL
type Repository struct {
	db *pgxpool.Pool    // Pool de conexões PostgreSQL
	q  *sqlc.Queries   // Queries geradas pelo SQLC
}

// NewRepository cria uma nova instância do repositório de usuários
// Recebe o pool de conexões e inicializa as queries do SQLC
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
		q:  sqlc.New(db), // Inicializar queries do SQLC com o pool
	}
}

// CreateUser cria um novo usuário no banco de dados
// Recebe os parâmetros e delega a execução para o SQLC
func (r *Repository) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	return r.q.CreateUser(ctx, arg)
}

// GetUserByID busca um usuário pelo ID (UUID)
// Retorna erro se o usuário não for encontrado
func (r *Repository) GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	return r.q.GetUserByID(ctx, id)
}

// GetUserByEmail busca um usuário pelo email
// Usado principalmente para autenticação e validação de unicidade
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

// UpdateUser atualiza os dados de um usuário existente
// Recebe parâmetros com ID e campos a serem atualizados
func (r *Repository) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	return r.q.UpdateUser(ctx, arg)
}

// DeletedUser realiza soft delete de um usuário
// Marca o usuário como deletado (deleted_at) ao invés de remover fisicamente
func (r *Repository) DeletedUser(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteUser(ctx, id)
}

// ListUsers retorna todos os usuários ativos (não deletados)
// Retorna lista vazia se não houver usuários
func (r *Repository) ListUsers(ctx context.Context) ([]sqlc.User, error) {
	return r.q.ListUsers(ctx)
}
