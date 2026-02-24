package handlers

import (
	"Dimidroll06/url-link-shortener/internal/core/domain"
	"time"
)

type CreateRequest struct {
	URL string `json:"url" binding:"required"`
}

type CreateResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type StatsResponse struct {
	ShortCode     string `json:"short_code"`
	OriginalURL   string `json:"original_url"`
	TotalAccesses int64  `json:"total_accesses"`
	CreatedAt     string `json:"created_at"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func toCreateResponse(url *domain.URL, baseURL string) CreateResponse {
	return CreateResponse{
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		ShortURL:    baseURL + "/" + url.ShortCode,
	}
}

func toStatsResponse(url *domain.URL, accesses int64) StatsResponse {
	expiresAt := ""
	if url.ExpiresAt != nil {
		expiresAt = url.ExpiresAt.Format(time.RFC3339)
	}

	return StatsResponse{
		ShortCode:     url.ShortCode,
		OriginalURL:   url.OriginalURL,
		TotalAccesses: accesses,
		CreatedAt:     url.CreatedAt.Format(time.RFC3339),
		ExpiresAt:     expiresAt,
	}
}
