package user

import (
	"database/sql"

	"github.com/GabrielFerrarez19/gofinance-api/internal/database/sqlc"
)

type Repository struct {
	db *sql.DB
	q  *sqlc.Queries
}
