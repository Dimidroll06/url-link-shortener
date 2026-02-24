package ports

import (
	"context"

	"Dimidroll06/url-link-shortener/internal/core/domain"
)

type URLCache interface {
	Get(ctx context.Context, code string) (*domain.URL, error)
	Set(ctx context.Context, url *domain.URL, ttl int) error
	Delete(ctx context.Context, code string) error
	Exists(ctx context.Context, code string) (bool, error)
}

type StatsCache interface {
	IncrementAccess(ctx context.Context, code string) (int64, error)
	GetAccessCount(ctx context.Context, code string) (int64, error)
	Reset(ctx context.Context, code string) error
}
