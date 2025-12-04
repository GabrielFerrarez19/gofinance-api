package database

import (
	"context"
	"fmt"
	"time"

	"github.com/GabrielFerrarez19/gofinance-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB encapsula o pool de conexões do PostgreSQL
// Usa pgxpool para gerenciar múltiplas conexões de forma eficiente
type DB struct {
	*pgxpool.Pool
}

// NewConnection cria uma nova conexão com o banco de dados PostgreSQL
// Configura um pool de conexões para melhor performance e gerenciamento de recursos
func NewConnection(cfg *config.Config) (*DB, error) {
	// Montar DSN (Data Source Name) para conexão com PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	// Parse da configuração do pool de conexões
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configurar pool de conexões
	config.MaxConns = 25        // Máximo de conexões simultâneas no pool
	config.MinConns = 5         // Mínimo de conexões mantidas no pool (para reduzir latência)
	config.MaxConnLifetime = 5 * time.Minute // Tempo máximo de vida de uma conexão

	// Criar pool de conexões com a configuração definida
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Testar conexão fazendo um ping no banco de dados
	// Garante que a conexão está funcionando antes de retornar
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{pool}, nil
}

// Close fecha todas as conexões do pool de banco de dados
// Deve ser chamado ao finalizar a aplicação para liberar recursos
func (db *DB) Close() {
	db.Pool.Close()
}

// Health verifica se a conexão com o banco de dados está saudável
// Útil para health checks e monitoramento
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
