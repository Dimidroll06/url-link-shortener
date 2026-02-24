package cache

import (
	"context"

	"Dimidroll06/url-link-shortener/internal/core/ports"
)

type MockStatsCache struct {
	IncrementAccessFunc func(ctx context.Context, code string) (int64, error)
	GetAccessCountFunc  func(ctx context.Context, code string) (int64, error)
	ResetFunc           func(ctx context.Context, code string) error
}

func NewMockStatsCache() *MockStatsCache {
	return &MockStatsCache{
		IncrementAccessFunc: func(ctx context.Context, code string) (int64, error) { return 0, nil },
		GetAccessCountFunc:  func(ctx context.Context, code string) (int64, error) { return 0, nil },
		ResetFunc:           func(ctx context.Context, code string) error { return nil },
	}
}

var _ ports.StatsCache = (*MockStatsCache)(nil)

func (m *MockStatsCache) IncrementAccess(ctx context.Context, code string) (int64, error) {
	return m.IncrementAccessFunc(ctx, code)
}

func (m *MockStatsCache) GetAccessCount(ctx context.Context, code string) (int64, error) {
	return m.GetAccessCountFunc(ctx, code)
}

func (m *MockStatsCache) Reset(ctx context.Context, code string) error {
	return m.ResetFunc(ctx, code)
}
