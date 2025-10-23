package main

import (
	"github.com/GabrielFerrarez19/gofinance-api/internal/config"
	"github.com/GabrielFerrarez19/gofinance-api/internal/database"
	"github.com/GabrielFerrarez19/gofinance-api/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal().Err(err).Msg("Error loading settings")
	}

	logger.InitLogger(cfg.AppEnv)

	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Database connected successfully")

	log.Info().Msgf("GoFinance running at the port %s (env: %s)", cfg.AppPort, cfg.AppEnv)
	log.Info().Msgf("Data Base: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	select {}
}
