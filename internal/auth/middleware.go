package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			log.Error().Err(err).Msg("invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_full_name", claims.FullName)
		c.Set("claims", claims)

		c.Next()
	}
}

func OptionalAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenstring := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenstring == authHeader {
			c.Next()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenstring)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_full_name", claims.FullName)
		c.Set("claims", claims)

		c.Next()
	}
}

func GetUserFromContext(c *gin.Context) (*models.UserResponse, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("user not found in context")
	}
	email, _ := c.Get("user_email")
	fullName, _ := c.Get("user_full_name")

	return &models.UserResponse{
		ID:       userID.(uuid.UUID),
		Email:    email.(string),
		FullName: fullName.(string),
	}, nil
}
