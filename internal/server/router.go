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

func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", r.authHandler.Login)
			authGroup.POST("/refresh", r.authHandler.RefreshToken)
			authGroup.POST("/logout", r.authHandler.Logout)
			authGroup.GET("/me", r.authHandler.Me)
		}

		users := api.Group("/users")
		users.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist))
		{
			users.POST("", r.userHandler.CreatedUser)
			users.GET("", r.userHandler.ListUsers)
			users.GET("/:id", r.userHandler.GetUser)
			users.PUT("/:id", r.userHandler.UpdateUser)
			users.DELETE("/:id", r.userHandler.DeleteUser)
		}

		accounts := api.Group("/accounts")
		accounts.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist))
		{
			accounts.POST("", r.accountHandler.Create)
			accounts.GET("", r.accountHandler.ListByUser)
			accounts.GET("/:id", r.accountHandler.GetByID)
			accounts.PUT("/:id", r.accountHandler.Update)
			accounts.DELETE("/:id", r.accountHandler.Delete)
		}

		transactions := api.Group("/transactions")
		transactions.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist))
		{
			transactions.POST("", r.txHandler.Create)
			transactions.GET("", r.txHandler.ListByUser)
			transactions.GET("/account/:account_id", r.txHandler.ListByAccount)
			transactions.GET("/:id", r.txHandler.GetByID)
			transactions.PUT("/:id", r.txHandler.Update)
			transactions.DELETE("/:id", r.txHandler.Delete)
		}

		categories := api.Group("/categories")
		categories.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist))
		{
			categories.POST("", r.ctHandler.Create)
			categories.GET("", r.ctHandler.ListByUser)
			categories.GET("/:id", r.ctHandler.GetByID)
			categories.PUT("/:id", r.ctHandler.Update)
			categories.DELETE("/:id", r.ctHandler.Delete)
		}

		reports := api.Group("/reports")
		reports.Use(auth.AuthMiddleware(r.jwtManager, r.blacklist))
		{
			reports.POST("", r.rpHandler.Create)
			reports.GET("", r.rpHandler.GetByID)
			reports.GET("/:id", r.rpHandler.ListByUser)
		}
	}

	return router
}
