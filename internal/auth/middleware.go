package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/GabrielFerrarez19/gofinance-api/internal/cache"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware é um middleware que protege rotas exigindo autenticação JWT
// Valida o token no header Authorization e verifica se está na blacklist
// Se válido, adiciona informações do usuário ao contexto para uso nos handlers
func AuthMiddleware(jwtManager *JWTManager, blacklist *cache.TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obter header Authorization da requisição
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort() // Interrompe a execução da requisição
			return
		}

		// Extrair token do formato "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// Token não está no formato Bearer
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		// Validar token JWT
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			log.Error().Err(err).Msg("invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Verificar se token está na blacklist (foi revogado)
		// Gera ID único do token baseado no userID e timestamp de expiração
		tokenId := fmt.Sprintf("%s:%d", claims.UserID.String(), claims.ExpiresAt.Unix())
		isBlacklisted, err := blacklist.IsTokenBlacklisted(c.Request.Context(), tokenId)
		if err != nil {
			log.Error().Err(err).Msg("failed to check token blacklist")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		if isBlacklisted {
			// Token foi revogado (logout)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		// Token válido: adicionar informações do usuário ao contexto
		// Essas informações estarão disponíveis nos handlers via c.Get()
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_full_name", claims.FullName)
		c.Set("claims", claims)

		// Continuar para o próximo handler
		c.Next()
	}
}

// OptionalAuthMiddleware é um middleware opcional de autenticação
// Se um token válido for fornecido, adiciona informações ao contexto
// Se não houver token ou for inválido, continua normalmente (não bloqueia)
// Útil para rotas que funcionam com ou sem autenticação
func OptionalAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Sem token, continuar normalmente
			c.Next()
			return
		}

		tokenstring := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenstring == authHeader {
			// Token não está no formato Bearer, continuar normalmente
			c.Next()
			return
		}

		// Tentar validar token
		claims, err := jwtManager.ValidateToken(tokenstring)
		if err != nil {
			// Token inválido, continuar normalmente (não é obrigatório)
			c.Next()
			return
		}

		// Token válido: adicionar informações ao contexto
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_full_name", claims.FullName)
		c.Set("claims", claims)

		c.Next()
	}
}

// GetUserFromContext extrai informações do usuário do contexto do Gin
// Usado nos handlers para obter dados do usuário autenticado
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
