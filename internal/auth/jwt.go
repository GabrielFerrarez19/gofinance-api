package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims representa os dados armazenados no token JWT
// Contém informações do usuário e metadados do token (expiração, emissão, etc)
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`   // ID único do usuário
	Email    string    `json:"email"`     // Email do usuário
	FullName string    `json:"full_name"` // Nome completo do usuário
	jwt.RegisteredClaims                  // Claims padrão do JWT (exp, iat, nbf, iss, sub)
}

// TokenPair representa um par de tokens (access e refresh)
// Access token: usado para autenticar requisições (vida curta)
// Refresh token: usado para renovar o access token (vida longa)
type TokenPair struct {
	AccessToken  string `json:"access_token"`  // Token de acesso (15 minutos)
	RefreshToken string `json:"refresh_token"` // Token de renovação (7 dias)
	ExpiresIn    int64  `json:"expires_in"`    // Tempo de expiração em segundos
}

// JWTManager gerencia a criação e validação de tokens JWT
// Centraliza toda a lógica relacionada a tokens de autenticação
type JWTManager struct {
	secretKey       string        // Chave secreta para assinar tokens
	accessTokenTTL  time.Duration // Tempo de vida do access token (15 minutos)
	refreshTokenTTL time.Duration // Tempo de vida do refresh token (7 dias)
}

// NewJWTManager cria uma nova instância do gerenciador JWT
// Recebe a chave secreta usada para assinar e validar tokens
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey:       secretKey,
		accessTokenTTL:  15 * time.Minute,  // Access token expira em 15 minutos
		refreshTokenTTL: 7 * 24 * time.Hour, // Refresh token expira em 7 dias
	}
}

// GenerateTokenPair gera um par de tokens (access e refresh) para um usuário
// Access token: usado para autenticar requisições, vida curta por segurança
// Refresh token: usado para renovar access tokens, vida longa mas mais seguro
func (j *JWTManager) GenerateTokenPair(userID uuid.UUID, email, fullname string) (*TokenPair, error) {
	// Criar claims para o Access Token
	// Access token tem vida curta (15 min) para reduzir risco se comprometido
	accessClaims := &Claims{
		UserID:   userID,
		Email:    email,
		FullName: fullname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)), // Expira em 15 minutos
			IssuedAt:  jwt.NewNumericDate(time.Now()),                        // Data de emissão
			NotBefore: jwt.NewNumericDate(time.Now()),                        // Não válido antes de agora
			Issuer:    "gofinance-api",                                       // Emissor do token
			Subject:   userID.String(),                                       // Sujeito (ID do usuário)
		},
	}

	// Criar e assinar o access token usando HS256 (HMAC SHA256)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return nil, err
	}

	// Criar claims para o Refresh Token
	// Refresh token tem vida longa (7 dias) mas só pode ser usado para renovar access tokens
	refreshClaims := &Claims{
		UserID:   userID,
		Email:    email,
		FullName: fullname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenTTL)), // Expira em 7 dias
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gofinance-api",
			Subject:   userID.String(),
		},
	}

	// Criar e assinar o refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(j.accessTokenTTL.Seconds()), // Tempo de expiração em segundos
	}, nil
}

// ValidateToken valida um token JWT e retorna os claims se válido
// Verifica assinatura, expiração e formato do token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// Parse do token verificando a assinatura
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		// Verificar se o método de assinatura é HMAC (esperado)
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// Retornar a chave secreta para validação
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	// Verificar se o token é válido e extrair os claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken renova um par de tokens usando um refresh token válido
// Valida o refresh token e gera um novo par de tokens (access + refresh)
func (j *JWTManager) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	// Validar o refresh token primeiro
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, err
	}
	// Gerar novo par de tokens com os dados do usuário
	return j.GenerateTokenPair(claims.UserID, claims.Email, claims.FullName)
}
