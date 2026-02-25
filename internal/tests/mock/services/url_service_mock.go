package services

import (
	"context"

	"Dimidroll06/url-link-shortener/internal/core/domain"
)

type MockURLService struct {
	CreateFunc         func(ctx context.Context, originalURL string) (*domain.URL, error)
	GetByShortCodeFunc func(ctx context.Context, code string) (*domain.URL, error)
	GetStatsFunc       func(ctx context.Context, code string) (int64, error)
	DeleteFunc         func(ctx context.Context, code string) error
}

func NewMockURLService() *MockURLService {
	return &MockURLService{
		CreateFunc: func(ctx context.Context, originalURL string) (*domain.URL, error) {
			return nil, nil
		},
		GetByShortCodeFunc: func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, nil
		},
		GetStatsFunc: func(ctx context.Context, code string) (int64, error) {
			return 0, nil
		},
		DeleteFunc: func(ctx context.Context, code string) error {
			return nil
		},
	}
}

func (m *MockURLService) Create(ctx context.Context, originalURL string) (*domain.URL, error) {
	return m.CreateFunc(ctx, originalURL)
}

func (m *MockURLService) GetByShortCode(ctx context.Context, code string) (*domain.URL, error) {
	return m.GetByShortCodeFunc(ctx, code)
}

func (m *MockURLService) GetStats(ctx context.Context, code string) (int64, error) {
	return m.GetStatsFunc(ctx, code)
}

func (m *MockURLService) Delete(ctx context.Context, code string) error {
	return m.DeleteFunc(ctx, code)
}
