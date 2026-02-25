package ports

import (
	"context"

	"Dimidroll06/url-link-shortener/internal/core/domain"
)

type URLServiceInterface interface {
	Create(ctx context.Context, originalURL string) (*domain.URL, error)
	GetByShortCode(ctx context.Context, code string) (*domain.URL, error)
	GetStats(ctx context.Context, code string) (int64, error)
	Delete(ctx context.Context, code string) error
}

type StatsServiceInterface interface {
	GetDetailedStats(ctx context.Context, code string) (map[string]interface{}, error)
	RecordAccess(ctx context.Context, code, ipAddress, userAgent, referer string) error
}
