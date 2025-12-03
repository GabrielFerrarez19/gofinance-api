package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler processa requisições HTTP relacionadas a autenticação
// Gerencia login, logout, refresh token e informações do usuário autenticado
type Handler struct {
	service *Service // Serviço com lógica de autenticação
}

// NewHandler cria uma nova instância do handler de autenticação
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Login autentica um usuário e retorna tokens JWT
// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	// Validar e bind do JSON da requisição
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Autenticar usuário e gerar tokens
	response, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Definir cookies HTTP com os tokens (HttpOnly para segurança)
	h.service.SetAuthCookies(c, response)

	// Retornar resposta com tokens e dados do usuário
	c.JSON(http.StatusOK, response)
}

// RefreshToken renova os tokens de autenticação usando um refresh token
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body map[string]string true "Refresh token"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	// Estrutura para receber refresh token
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	// Validar requisição
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Renovar tokens usando refresh token
	response, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Atualizar cookies com novos tokens
	h.service.SetAuthCookies(c, response)

	c.JSON(http.StatusOK, response)
}

// Logout realiza logout do usuário e invalida tokens
// @Summary User logout
// @Description Logout user and invalidate tokens
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	// Limpar cookies de autenticação
	h.service.ClearAuthCookies(c)

	// Nota: Em uma implementação completa, o token também deveria ser adicionado à blacklist
	// Isso requer extrair o token do header Authorization

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Me retorna informações do usuário autenticado
// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	// Extrair informações do usuário do contexto (adicionadas pelo AuthMiddleware)
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
