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
	"github.com/GabrielFerrarez19/gofinance-api/internal/queue"
	"github.com/GabrielFerrarez19/gofinance-api/internal/queue/jobs"
	"github.com/GabrielFerrarez19/gofinance-api/internal/report"
	"github.com/GabrielFerrarez19/gofinance-api/internal/server"
	"github.com/GabrielFerrarez19/gofinance-api/internal/transaction"
	"github.com/GabrielFerrarez19/gofinance-api/internal/user"
	"github.com/rs/zerolog/log"
)

// main é o ponto de entrada da aplicação
// Responsável por inicializar todos os componentes e iniciar o servidor HTTP
func main() {
	// Carregar configurações da aplicação (variáveis de ambiente)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading settings")
	}

	// Inicializar logger com base no ambiente (development/production)
	logger.InitLogger(cfg.AppEnv)

	// Conectar ao banco de dados PostgreSQL
	// Usa pgxpool para gerenciar pool de conexões
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close() // Garantir fechamento da conexão ao finalizar

	// Conectar ao Redis para cache e blacklist de tokens
	redisClient, err := cache.NewRedisConnection(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close() // Garantir fechamento da conexão ao finalizar

	// Conectar ao RabbitMQ
	rabbitmqClient, err := queue.NewRabbitMQConnection(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer rabbitmqClient.Close()

	// Inicializar publisher e consumer
	publisher := queue.NewPublisher(rabbitmqClient)
	consumer := queue.NewConsumer(rabbitmqClient)

	// Inicializar serviços de cache
	// TokenBlacklist: serviço específico para gerenciar blacklist de tokens JWT
	tokenBlacklist := cache.NewTokenBlacklist(redisClient)

	// Inicializar gerenciador de JWT
	// Responsável por gerar e validar tokens de acesso e refresh
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Inicializar repositórios (camada de acesso a dados)
	// Cada repositório encapsula as operações de banco de dados para seu domínio
	userRepo := user.NewRepository(db.Pool)
	accountRepo := account.NewRepository(db.Pool)
	txRepo := transaction.NewRepository(db.Pool)
	ctRepo := category.NewRepository(db.Pool)
	rpRepo := report.NewRepository(db.Pool)

	// Inicializar serviços (camada de lógica de negócio)
	// Cada serviço contém a lógica de negócio e usa os repositórios para acessar dados
	userService := user.NewService(userRepo)
	accountService := account.NewService(accountRepo)
	// AuthService depende do UserService e JWTManager para autenticação
	authService := auth.NewService(userService, jwtManager, tokenBlacklist)
	txService := transaction.NewService(txRepo)
	ctService := category.NewService(ctRepo)
	// ReportService depende do TransactionRepository para gerar relatórios
	rpService := report.NewService(rpRepo, txRepo)

	// Inicializar jobs
	reportJob := jobs.NewReportJob(db)

	// Inicializar handlers (camada de apresentação/HTTP)
	// Cada handler processa requisições HTTP e chama os serviços correspondentes
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(authService)
	accountHandler := account.NewHandler(accountService)
	txHandler := transaction.NewHandler(txService)
	ctHandler := category.NewHandler(ctService)
	rpHandler := report.NewHandler(rpService, publisher)

	// Inicializar router e configurar rotas
	// O router gerencia todas as rotas da API e aplica middlewares
	router := server.NewRouter(userHandler, authHandler, jwtManager, accountHandler, txHandler, ctHandler, rpHandler, tokenBlacklist)
	engine := router.SetupRoutes()

	// Configurar servidor HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort, // Porta do servidor
		Handler: engine,            // Router do Gin como handler
	}

	// Iniciar servidor em goroutine separada
	// Isso permite que o código continue e implemente graceful shutdown
	go func() {
		log.Info().Msgf("GoFinance running at the port %s (env: %s)", cfg.AppPort, cfg.AppEnv)
		log.Info().Msgf("Data Base: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

		// ListenAndServe bloqueia até que o servidor seja fechado
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Iniciar consumer de relatórios em backgroud
	ctx := context.Background()
	go func() {
		if err := consumer.Consume(ctx, "report_generation", reportJob.Process); err != nil {
			log.Error().Err(err).Msg("Failed to start report consumer")
		}
	}()

	// Implementar graceful shutdown
	// Aguarda sinais do sistema (SIGINT, SIGTERM) para desligar o servidor de forma controlada
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Bloqueia até receber um sinal
	log.Info().Msg("Shutting down server...")

	// Criar contexto com timeout de 5 segundos para finalizar requisições em andamento
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown tenta fechar o servidor de forma graciosa
	// Aguarda até 5 segundos para que requisições em andamento terminem
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
