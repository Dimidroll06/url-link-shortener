package cache

import (
	"context"

	"Dimidroll06/url-link-shortener/internal/core/domain"
	"Dimidroll06/url-link-shortener/internal/core/ports"
)

type MockURLCache struct {
	GetFunc    func(ctx context.Context, code string) (*domain.URL, error)
	SetFunc    func(ctx context.Context, url *domain.URL, ttl int) error
	DeleteFunc func(ctx context.Context, code string) error
	ExistsFunc func(ctx context.Context, code string) (bool, error)
}

func NewMockURLCache() *MockURLCache {
	return &MockURLCache{
		GetFunc:    func(ctx context.Context, code string) (*domain.URL, error) { return nil, nil },
		SetFunc:    func(ctx context.Context, url *domain.URL, ttl int) error { return nil },
		DeleteFunc: func(ctx context.Context, code string) error { return nil },
		ExistsFunc: func(ctx context.Context, code string) (bool, error) { return false, nil },
	}
}

var _ ports.URLCache = (*MockURLCache)(nil)

func (m *MockURLCache) Get(ctx context.Context, code string) (*domain.URL, error) {
	return m.GetFunc(ctx, code)
}

func (m *MockURLCache) Set(ctx context.Context, url *domain.URL, ttl int) error {
	return m.SetFunc(ctx, url, ttl)
}

func (m *MockURLCache) Delete(ctx context.Context, code string) error {
	return m.DeleteFunc(ctx, code)
}

func (m *MockURLCache) Exists(ctx context.Context, code string) (bool, error) {
	return m.ExistsFunc(ctx, code)
}
