package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/GabrielFerrarez19/gofinance-api/internal/cache"
	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/GabrielFerrarez19/gofinance-api/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

// Service contém a lógica de negócio para autenticação
// Coordena validação de credenciais, geração de tokens e gerenciamento de sessões
type Service struct {
	userService *user.Service         // Serviço de usuários para validar credenciais
	jwtManager  *JWTManager           // Gerenciador de tokens JWT
	blacklist   *cache.TokenBlacklist // Blacklist para tokens revogados
}

// NewService cria uma nova instância do serviço de autenticação
func NewService(userService *user.Service, jwtManager *JWTManager, blacklist *cache.TokenBlacklist) *Service {
	return &Service{
		userService: userService,
		jwtManager:  jwtManager,
		blacklist:   blacklist,
	}
}

// LoginRequest representa os dados de login recebidos na requisição
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"` // Email do usuário (obrigatório, formato email)
	Password string `json:"password" binding:"required"`    // Senha do usuário (obrigatória)
}

// LoginResponse representa a resposta de login bem-sucedido
// Contém informações do usuário e tokens de autenticação
type LoginResponse struct {
	User         models.UserResponse `json:"user"`          // Dados do usuário autenticado
	AccessToken  string              `json:"access_token"`  // Token de acesso
	RefreshToken string              `json:"refresh_token"` // Token de renovação
	ExpiresIn    int64               `json:"expires_in"`    // Tempo de expiração em segundos
	TokenType    string              `json:"token_type"`    // Tipo do token (Bearer)
}

// Login autentica um usuário e retorna tokens JWT
// Valida email e senha, depois gera um par de tokens (access + refresh)
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Validar credenciais do usuário (email e senha)
	user, err := s.userService.ValidatorPassword(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	// Gerar par de tokens JWT para o usuário autenticado
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

// RefreshToken renova os tokens de autenticação usando um refresh token válido
// Valida o refresh token e gera um novo par de tokens
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Renovar tokens usando o refresh token
	tokenPair, err := s.jwtManager.RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Validar o novo access token para extrair claims
	claims, err := s.jwtManager.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		return nil, err
	}

	// Buscar dados atualizados do usuário no banco de dados
	// Converter uuid.UUID para pgtype.UUID (formato usado pelo banco)
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

// Logout invalida um token adicionando-o à blacklist
// O token será rejeitado em requisições futuras até expirar naturalmente
func (s *Service) Logout(ctx context.Context, token string) error {
	// Validar token para extrair claims (ID do usuário, expiração, etc)
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return err
	}

	// Calcular tempo restante até a expiração do token
	expiresIn := time.Until(claims.ExpiresAt.Time)
	if expiresIn <= 0 {
		// Token já expirado, não precisa adicionar à blacklist
		return nil
	}

	// Criar ID único para o token (userID:timestamp)
	tokenID := fmt.Sprintf("%s:%d", claims.UserID.String(), claims.ExpiresAt.Unix())
	// Adicionar token à blacklist até expirar
	return s.blacklist.AddToken(ctx, tokenID, expiresIn)
}

// SetAuthCookies define cookies HTTP com os tokens de autenticação
// Cookies HttpOnly protegem contra ataques XSS (JavaScript não pode acessar)
func (s *Service) SetAuthCookies(c *gin.Context, response *LoginResponse) {
	// Cookie para Access Token
	// HttpOnly: true impede acesso via JavaScript (segurança)
	// Secure: false em desenvolvimento, true em produção (HTTPS apenas)
	c.SetCookie(
		"access_token",
		response.AccessToken,
		int(time.Duration(response.ExpiresIn)*time.Second), // TTL do cookie
		"/",   // Path onde o cookie é válido
		"",    // Domain (vazio = domínio atual)
		false, // Secure: false para desenvolvimento (true em produção)
		true,  // HttpOnly: true para segurança
	)

	// Cookie para Refresh Token
	// Vida mais longa (7 dias) para permitir renovação de tokens
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

// ClearAuthCookies remove os cookies de autenticação
// Usado durante logout para limpar tokens do navegador
func (s *Service) ClearAuthCookies(c *gin.Context) {
	// Definir cookies com valor vazio e TTL negativo para removê-los
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}
