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
}
