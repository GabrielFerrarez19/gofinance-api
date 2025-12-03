package server

import (
	_ "github.com/GabrielFerrarez19/gofinance-api/docs"

	"github.com/GabrielFerrarez19/gofinance-api/internal/account"
	"github.com/GabrielFerrarez19/gofinance-api/internal/auth"
	"github.com/GabrielFerrarez19/gofinance-api/internal/cache"
	"github.com/GabrielFerrarez19/gofinance-api/internal/category"
	"github.com/GabrielFerrarez19/gofinance-api/internal/report"
	"github.com/GabrielFerrarez19/gofinance-api/internal/transaction"
	"github.com/GabrielFerrarez19/gofinance-api/internal/user"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router encapsula todos os handlers e configurações de roteamento
// Centraliza a configuração de todas as rotas da API
type Router struct {
	userHandler    *user.Handler
	authHandler    *auth.Handler
	jwtManager     *auth.JWTManager
	accountHandler *account.Handler
	txHandler      *transaction.Handler
	ctHandler      *category.Handler
	rpHandler      *report.Handler
	blacklist      *cache.TokenBlacklist
}

// NewRouter cria uma nova instância do Router com todos os handlers necessários
func NewRouter(userHandler *user.Handler, authHandler *auth.Handler, jwtManager *auth.JWTManager, accountHandler *account.Handler, txHandler *transaction.Handler, ctHandler *category.Handler, rpHandler *report.Handler, blacklist *cache.TokenBlacklist) *Router {
	return &Router{
		userHandler:    userHandler,
		authHandler:    authHandler,
		jwtManager:     jwtManager,
		accountHandler: accountHandler,
		txHandler:      txHandler,
		ctHandler:      ctHandler,
		rpHandler:      rpHandler,
		blacklist:      blacklist,
	}
}

// SetupRoutes configura todas as rotas da API
// Aplica middlewares globais e organiza rotas por grupos (domínios)
func (r *Router) SetupRoutes() *gin.Engine {
	// Definir modo de release (desabilita debug mode)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middlewares globais aplicados a todas as rotas
	router.Use(gin.Logger())     // Logger: registra todas as requisições HTTP
	router.Use(gin.Recovery())   // Recovery: recupera de panics e retorna erro 500
	router.Use(CORSMiddleware()) // CORS: permite requisições cross-origin

	// Rota para documentação Swagger
	// Acessível em /docs/index.html
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	// Usado para verificar se a API está funcionando (útil para load balancers)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Agrupar todas as rotas da API sob o prefixo /api/v1
	api := router.Group("/api/v1")
	{
		// Rotas de autenticação (públicas - não requerem JWT)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", r.authHandler.Login)          // Login e obtenção de tokens
			authGroup.POST("/refresh", r.authHandler.RefreshToken) // Renovação de access token
			authGroup.POST("/logout", r.authHandler.Logout)        // Logout e invalidação de tokens
			authGroup.GET("/me", r.authHandler.Me)                 // Informações do usuário autenticado
		}

		// Rotas de usuários (protegidas - requerem JWT)
		users := api.Group("/users")
		users.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist)) // Middleware de autenticação
		{
			users.POST("", r.userHandler.CreatedUser)      // Criar usuário
			users.GET("", r.userHandler.ListUsers)         // Listar todos os usuários
			users.GET("/:id", r.userHandler.GetUser)       // Obter usuário por ID
			users.PUT("/:id", r.userHandler.UpdateUser)    // Atualizar usuário
			users.DELETE("/:id", r.userHandler.DeleteUser) // Deletar usuário (soft delete)
		}

		// Rotas de contas (protegidas - requerem JWT)
		accounts := api.Group("/accounts")
		accounts.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist))
		{
			accounts.POST("", r.accountHandler.Create)       // Criar conta
			accounts.GET("", r.accountHandler.ListByUser)    // Listar contas do usuário
			accounts.GET("/:id", r.accountHandler.GetByID)   // Obter conta por ID
			accounts.PUT("/:id", r.accountHandler.Update)    // Atualizar conta
			accounts.DELETE("/:id", r.accountHandler.Delete) // Deletar conta
		}

		// Rotas de transações (protegidas - requerem JWT)
		transactions := api.Group("/transactions")
		transactions.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist))
		{
			transactions.POST("", r.txHandler.Create)                           // Criar transação
			transactions.GET("", r.txHandler.ListByUser)                        // Listar transações do usuário
			transactions.GET("/account/:account_id", r.txHandler.ListByAccount) // Listar transações de uma conta
			transactions.GET("/:id", r.txHandler.GetByID)                       // Obter transação por ID
			transactions.PUT("/:id", r.txHandler.Update)                        // Atualizar transação
			transactions.DELETE("/:id", r.txHandler.Delete)                     // Deletar transação
		}

		// Rotas de categorias (protegidas - requerem JWT)
		categories := api.Group("/categories")
		categories.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist))
		{
			categories.POST("", r.ctHandler.Create)       // Criar categoria
			categories.GET("", r.ctHandler.ListByUser)    // Listar categorias do usuário
			categories.GET("/:id", r.ctHandler.GetByID)   // Obter categoria por ID
			categories.PUT("/:id", r.ctHandler.Update)    // Atualizar categoria
			categories.DELETE("/:id", r.ctHandler.Delete) // Deletar categoria
		}

		// Rotas de relatórios (protegidas - requerem JWT)
		reports := api.Group("/reports")
		reports.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist))
		{
			reports.POST("", r.rpHandler.Create)        // Criar relatório
			reports.GET("", r.rpHandler.GetByID)        // Obter relatório
			reports.GET("/:id", r.rpHandler.ListByUser) // Listar relatórios do usuário
		}
	}

	return router
}
