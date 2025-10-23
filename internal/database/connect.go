package database

import (
	"fmt"

	"github.com/GabrielFerrarez19/gofinance-api/internal/config"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}
