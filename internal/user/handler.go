package user

import (
	"net/http"

	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// Handler processa requisições HTTP relacionadas a usuários
// Valida dados de entrada, chama o serviço e retorna respostas HTTP apropriadas
type Handler struct {
	service ServiceInterface // Serviço com lógica de negócio
}

// NewHandler cria uma nova instância do handler de usuários
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{
		service: service,
	}
}

// CreatedUser cria um novo usuário no sistema
// @Summary Create a new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User data"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
// @Security BearerAuth
func (h *Handler) CreatedUser(c *gin.Context) {
	// Validar e bind do JSON da requisição para o modelo
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Chamar serviço para criar usuário
	user, err := h.service.CreateUser(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Retornar usuário criado com status 201 Created
	c.JSON(http.StatusCreated, user)
}

// GetUser busca um usuário pelo ID
// @Summary Get user by ID
// @Description Get user information by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [get]
// @Security BearerAuth
func (h *Handler) GetUser(c *gin.Context) {
	// Obter ID da URL
	idStr := c.Param("id")

	// Converter string para UUID (pgtype.UUID)
	var id pgtype.UUID
	err := id.Scan(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Buscar usuário pelo ID
	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser atualiza os dados de um usuário existente
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body models.UpdateUserRequest true "User data"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [put]
// @Security BearerAuth
func (h *Handler) UpdateUser(c *gin.Context) {
	// Obter ID da URL
	idStr := c.Param("id")

	// Converter string para UUID
	var id pgtype.UUID
	err := id.Scan(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Validar e bind do JSON da requisição
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Atualizar usuário
	user, err := h.service.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser realiza soft delete de um usuário
// @Summary Delete user
// @Description Delete user (soft delete)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeleteUser(c *gin.Context) {
	// Obter ID da URL
	idStr := c.Param("id")

	// Converter string para UUID
	var id pgtype.UUID
	err := id.Scan(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Deletar usuário (soft delete)
	err = h.service.DeletedUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Retornar 204 No Content (sucesso sem corpo de resposta)
	c.Status(http.StatusNoContent)
}

// ListUsers retorna todos os usuários do sistema
// @Summary List all users
// @Description Get list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.UserResponse
// @Failure 500 {object} map[string]string
// @Router /users [get]
// @Security BearerAuth
func (h *Handler) ListUsers(c *gin.Context) {
	// Buscar todos os usuários
	users, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.ValidatorPassword(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successfully",
		"user":    user,
		"token":   "jwt-token-here", // Implementar JWT depois
	})
}
