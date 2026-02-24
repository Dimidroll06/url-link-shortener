package repository

import (
	"context"

	"Dimidroll06/url-link-shortener/internal/core/domain"
	"Dimidroll06/url-link-shortener/internal/core/ports"
)

type MockStatsRepository struct {
	RecordAccessFunc      func(ctx context.Context, stats *domain.URLStats) error
	GetTotalAccessesFunc  func(ctx context.Context, code string) (int64, error)
	GetRecentAccessesFunc func(ctx context.Context, code string, limit int) ([]*domain.URLStats, error)
}

func NewMockStatsRepository() *MockStatsRepository {
	return &MockStatsRepository{
		RecordAccessFunc:      func(ctx context.Context, stats *domain.URLStats) error { return nil },
		GetTotalAccessesFunc:  func(ctx context.Context, code string) (int64, error) { return 0, nil },
		GetRecentAccessesFunc: func(ctx context.Context, code string, limit int) ([]*domain.URLStats, error) { return nil, nil },
	}
}

var _ ports.StatsRepository = (*MockStatsRepository)(nil)

func (m *MockStatsRepository) RecordAccess(ctx context.Context, stats *domain.URLStats) error {
	return m.RecordAccessFunc(ctx, stats)
}

func (m *MockStatsRepository) GetTotalAccesses(ctx context.Context, code string) (int64, error) {
	return m.GetTotalAccessesFunc(ctx, code)
}

func (m *MockStatsRepository) GetRecentAccesses(ctx context.Context, code string, limit int) ([]*domain.URLStats, error) {
	return m.GetRecentAccessesFunc(ctx, code, limit)
}
