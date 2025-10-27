package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GabrielFerrarez19/gofinance-api/internal/config"
	"github.com/GabrielFerrarez19/gofinance-api/internal/database"
	"github.com/GabrielFerrarez19/gofinance-api/internal/logger"
	"github.com/GabrielFerrarez19/gofinance-api/internal/server"
	"github.com/GabrielFerrarez19/gofinance-api/internal/user"
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

	userRepo := user.NewRepository(db.Pool)

	userService := user.NewService(userRepo)

	userHandler := user.NewHandler(userService)

	router := server.NewRouter(userHandler)

	engine := router.SetupRoutes()

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: engine,
	}

	go func() {
		log.Info().Msgf("GoFinance running at the port %s (env: %s)", cfg.AppPort, cfg.AppEnv)
		log.Info().Msgf("Data Base: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}
	log.Info().Msg("Server exited")
}
