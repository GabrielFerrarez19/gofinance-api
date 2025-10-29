package auth

import (
	"context"
	"time"

	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/GabrielFerrarez19/gofinance-api/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type Service struct {
	userService *user.Service
	jwtManager  *JWTManager
}

func NewService(userService *user.Service, jwtManager *JWTManager) *Service {
	return &Service{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User         models.UserResponse `json:"user"`
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token"`
	ExpiresIn    int64               `json:"expires_in"`
	TokenType    string              `json:"token_type"`
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.userService.ValidatorPassword(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.FullName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tokens")
		return nil, err
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	tokenPair, err := s.jwtManager.RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	claims, err := s.jwtManager.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		return nil, err
	}

	// Corrigido: converter uuid.UUID para pgtype.UUID
	pgxUUID := pgtype.UUID{Bytes: claims.UserID, Valid: true}
	user, err := s.userService.GetUserByID(ctx, pgxUUID)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	// TODO: Implementar blacklist de tokens com Redis
	log.Info().Msg("User logged out")
	return nil
}

func (s *Service) SetAuthCookies(c *gin.Context, response *LoginResponse) {
	// Access Token Cookie (HttpOnly, Secure em produção)
	c.SetCookie(
		"access_token",
		response.AccessToken,
		int(time.Duration(response.ExpiresIn)*time.Second),
		"/",
		"",
		false, // Secure: false para desenvolvimento
		true,  // HttpOnly: true para segurança
	)

	// Refresh Token Cookie (HttpOnly, Secure em produção)
	c.SetCookie(
		"refresh_token",
		response.RefreshToken,
		int(7*24*time.Hour.Seconds()), // 7 dias
		"/",
		"",
		false, // Secure: false para desenvolvimento
		true,  // HttpOnly: true para segurança
	)
}

func (s *Service) ClearAuthCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}
