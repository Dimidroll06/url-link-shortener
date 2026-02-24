package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	servererrors "Dimidroll06/url-link-shortener/internal/core/errors"
	"Dimidroll06/url-link-shortener/internal/core/services"
)

type URLHandler struct {
	urlService   *services.URLService
	statsService *services.StatsService
	logger       *zap.Logger
	baseURL      string
}

func NewURLHandler(
	urlSvc *services.URLService,
	statsSvc *services.StatsService,
	logger *zap.Logger,
	baseURL string,
) *URLHandler {
	return &URLHandler{
		urlService:   urlSvc,
		statsService: statsSvc,
		logger:       logger,
		baseURL:      baseURL,
	}
}

func (h *URLHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/shorten", h.Create)
		api.GET("/:code/stats", h.GetStats)
	}

	r.GET("/:code", h.Redirect)
}

// Create godoc
// @Summary Create short URL
// @Accept json
// @Produce json
// @Param url body CreateRequest true "Original URL"
// @Success 201 {object} CreateResponse
// @Router /api/v1/shorten [post]
func (h *URLHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	url, err := h.urlService.Create(c.Request.Context(), req.URL)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	resp := toCreateResponse(url, h.baseURL)
	c.JSON(http.StatusCreated, resp)
}

// Redirect godoc
// @Summary Redirect to original URL
// @Param code path string true "Short code"
// @Success 302 "Redirect to original URL"
// @Router /:code [get]
func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	url, err := h.urlService.GetByShortCode(c.Request.Context(), code)
	if err != nil {
		// 🔥 Ключевой момент: 404 для всех "невалидных" состояний
		if errors.Is(err, servererrors.ErrURLNotFound) ||
			errors.Is(err, servererrors.ErrURLExpired) ||
			errors.Is(err, servererrors.ErrURLInactive) {
			h.logger.Debug("redirect: link not available",
				zap.String("code", code),
				zap.Error(err),
			)
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "link not found"})
			return
		}

		h.logger.Error("redirect: service error",
			zap.String("code", code),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	c.Redirect(http.StatusFound, url.OriginalURL)
}

// GetStats godoc
// @Summary Get URL statistics
// @Param code path string true "Short code"
// @Success 200 {object} StatsResponse
// @Router /api/v1/{code}/stats [get]
func (h *URLHandler) GetStats(c *gin.Context) {
	code := c.Param("code")

	url, err := h.urlService.GetByShortCode(c.Request.Context(), code)
	if err != nil {
		h.handleStatsError(c, err)
		return
	}

	accesses, err := h.urlService.GetStats(c.Request.Context(), code)
	if err != nil {
		h.logger.Error("get stats failed",
			zap.String("code", code),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get statistics"})
		return
	}

	resp := toStatsResponse(url, accesses)
	c.JSON(http.StatusOK, resp)
}

func (h *URLHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, servererrors.ErrInvalidURL),
		errors.Is(err, servererrors.ErrInvalidURLScheme),
		errors.Is(err, servererrors.ErrURLTooLong):
		h.logger.Warn("validation error", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

	case errors.Is(err, servererrors.ErrShortCodeExists):
		h.logger.Warn("short code conflict", zap.Error(err))
		c.JSON(http.StatusConflict, ErrorResponse{Error: "short code already exists"})

	case errors.Is(err, servererrors.ErrCacheUnavailable):
		h.logger.Error("repository error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})

	default:
		h.logger.Error("unexpected service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}

func (h *URLHandler) handleStatsError(c *gin.Context, err error) {
	if errors.Is(err, servererrors.ErrURLNotFound) ||
		errors.Is(err, servererrors.ErrURLExpired) ||
		errors.Is(err, servererrors.ErrURLInactive) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "link not found"})
		return
	}

	h.logger.Error("stats handler error", zap.Error(err))
	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
}
