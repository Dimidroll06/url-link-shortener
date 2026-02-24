package ports

import (
	"context"

	"Dimidroll06/url-link-shortener/internal/core/domain"
)

type URLRepository interface {
	Create(ctx context.Context, url *domain.URL) error
	GetByShortCode(ctx context.Context, code string) (*domain.URL, error)
	GetByID(ctx context.Context, id string) (*domain.URL, error)
	Update(ctx context.Context, url *domain.URL) error
	Delete(ctx context.Context, code string) error
	ExistsByShortCode(ctx context.Context, code string) (bool, error)
}

type StatsRepository interface {
	RecordAccess(ctx context.Context, stats *domain.URLStats) error
	GetTotalAccesses(ctx context.Context, code string) (int64, error)
	GetRecentAccesses(ctx context.Context, code string, limit int) ([]*domain.URLStats, error)
}
