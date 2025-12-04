package report

import (
	"net/http"

	"github.com/GabrielFerrarez19/gofinance-api/internal/models"
	"github.com/GabrielFerrarez19/gofinance-api/internal/queue"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service   *Service
	publisher *queue.Publisher
}

func NewHandler(service *Service, publisher *queue.Publisher) *Handler {
	return &Handler{
		service:   service,
		publisher: publisher,
	}
}

// CreateReport godoc
// @Summary Create report
// @Description Generates a new report for the authenticated user
// @Tags reports
// @Accept json
// @Produce json
// @Param report body models.CreateReportRequest true "Report payload"
// @Success 201 {object} models.ReportResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reports [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	// Validar e bind do JSON da requisição
	var req models.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obter ID do usuário do contexto (setado pelo middleware de autenticação)
	raw, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Converter user_id para pgtype.UUID
	userID := pgtype.UUID{Bytes: raw.(uuid.UUID), Valid: true}

	// Criar relatório inicialmente sem dados (ou com dados vazios)
	// O job assíncrono irá preencher os dados depois
	report, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create report")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Preparar payload do job para processamento assíncrono
	// O job irá buscar transações e calcular estatísticas
	jobPayload := models.ReportJobPayload{
		UserID:      userID.String(),
		ReportID:    report.ID.String(),
		Type:        string(req.Type),
		Title:       req.Title,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	// Publicar job no RabbitMQ para processamento assícrono
	// Isso permite que a reposta seja retornada imediatamente ao usuário
	// enquanto o relatório é gerado em background
	if err := h.publisher.Publish(c.Request.Context(), "report_generation", jobPayload, nil); err != nil {
		log.Error().Err(err).Msg("Failed to publish report job")
		// Continua mesmo se falhar (o relatório ja foi criado, pode ser processado depois)
		// Em produção, você pode querer implementar retry ou modificar o usuario
	} else {
		log.Info().
			Str("report_id", report.ID.String()).
			Msg("Report job publisher to RabbitMQ successfully")
	}

	// Retorna resposta com o relatório criado
	// O campo "data" estarã vazio inicialmente e será preenchido pelo job
	c.JSON(http.StatusCreated, report)
}

// GetReport godoc
// @Summary Get report by ID
// @Description Returns the details of a specific report
// @Tags reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} models.ReportResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /reports/{id} [get]
// @Security BearerAuth
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	idUUID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	raw, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := pgtype.UUID{Bytes: raw.(uuid.UUID), Valid: true}
	id := pgtype.UUID{Bytes: idUUID, Valid: true}
	out, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}

// ListReports godoc
// @Summary List reports
// @Description Lists all reports for the authenticated user
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {array} models.ReportResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reports [get]
// @Security BearerAuth
func (h *Handler) ListByUser(c *gin.Context) {
	raw, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := pgtype.UUID{Bytes: raw.(uuid.UUID), Valid: true}

	out, err := h.service.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}
