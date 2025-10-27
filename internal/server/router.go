package server

import (
	"github.com/GabrielFerrarez19/gofinance-api/internal/user"
	"github.com/gin-gonic/gin"
)

type Router struct {
	userHandler *user.Handler
}

func NewRouter(userHandler *user.Handler) *Router {
	return &Router{
		userHandler: userHandler,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "StatusOK"})
	})

	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("", r.userHandler.CreatedUser)
			users.GET("", r.userHandler.ListUsers)
			users.GET("/:id", r.userHandler.GetUser)
			users.PUT("/:id", r.userHandler.UpdateUser)
			users.DELETE("/:id", r.userHandler.DeleteUser)
		}

		auth := api.Group("/auth")
		{
			auth.POST("/login", r.userHandler.Login)
		}
	}
	return router
}
