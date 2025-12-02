// @title GoFinance API
// @version 1.0
// @description API para gestão financeira pessoal.
// @BasePath /api/v1
// @schemes http https
// @contact.name GoFinance Team
// @contact.email dev@gofinance.com
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GabrielFerrarez19/gofinance-api/internal/account"
	"github.com/GabrielFerrarez19/gofinance-api/internal/auth"
	"github.com/GabrielFerrarez19/gofinance-api/internal/cache"
	"github.com/GabrielFerrarez19/gofinance-api/internal/category"
	"github.com/GabrielFerrarez19/gofinance-api/internal/config"
	"github.com/GabrielFerrarez19/gofinance-api/internal/database"
	"github.com/GabrielFerrarez19/gofinance-api/internal/logger"
	"github.com/GabrielFerrarez19/gofinance-api/internal/report"
	"github.com/GabrielFerrarez19/gofinance-api/internal/server"
	"github.com/GabrielFerrarez19/gofinance-api/internal/transaction"
	"github.com/GabrielFerrarez19/gofinance-api/internal/user"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading settings")
	}

	logger.InitLogger(cfg.AppEnv)

	// Connect to database
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Connect to Redis
	redisClient, err := cache.NewRedisConnection(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize cache services
	cacheService := cache.NewCacheService(redisClient)
	tokenBlacklist := cache.NewTokenBlacklist(redisClient)

	// Initialize JWT Manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Initialize repositories
	userRepo := user.NewRepository(db.Pool)
	accountRepo := account.NewRepository(db.Pool)
	txRepo := transaction.NewRepository(db.Pool)
	ctRepo := category.NewRepository(db.Pool)
	rpRepo := report.NewRepository(db.Pool)

	// Initialize services
	userService := user.NewService(userRepo)
	accountService := account.NewService(accountRepo)
	authService := auth.NewService(userService, jwtManager, tokenBlacklist)
	txService := transaction.NewService(txRepo)
	ctService := category.NewService(ctRepo)
	rpService := report.NewService(rpRepo, txRepo)

	// Initialize handlers
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(authService)
	accountHandler := account.NewHandler(accountService)
	txHandler := transaction.NewHandler(txService)
	ctHandler := category.NewHandler(ctService)
	rpHandler := report.NewHandler(rpService)

	// Initialize router
	router := server.NewRouter(userHandler, authHandler, jwtManager, accountHandler, txHandler, ctHandler, rpHandler, tokenBlacklist)
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

	// Graceful shutdown
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
