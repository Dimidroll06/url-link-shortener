package repository

import (
	"context"

	"Dimidroll06/url-link-shortener/internal/core/domain"
	"Dimidroll06/url-link-shortener/internal/core/ports"
)

type MockURLRepository struct {
	CreateFunc            func(ctx context.Context, url *domain.URL) error
	GetByShortCodeFunc    func(ctx context.Context, code string) (*domain.URL, error)
	GetByIDFunc           func(ctx context.Context, id string) (*domain.URL, error)
	UpdateFunc            func(ctx context.Context, url *domain.URL) error
	DeleteFunc            func(ctx context.Context, code string) error
	ExistsByShortCodeFunc func(ctx context.Context, code string) (bool, error)
}

func NewMockURLRepository() *MockURLRepository {
	return &MockURLRepository{
		CreateFunc:            func(ctx context.Context, url *domain.URL) error { return nil },
		GetByShortCodeFunc:    func(ctx context.Context, code string) (*domain.URL, error) { return nil, nil },
		GetByIDFunc:           func(ctx context.Context, id string) (*domain.URL, error) { return nil, nil },
		UpdateFunc:            func(ctx context.Context, url *domain.URL) error { return nil },
		DeleteFunc:            func(ctx context.Context, code string) error { return nil },
		ExistsByShortCodeFunc: func(ctx context.Context, code string) (bool, error) { return false, nil },
	}
}

var _ ports.URLRepository = (*MockURLRepository)(nil)

func (m *MockURLRepository) Create(ctx context.Context, url *domain.URL) error {
	return m.CreateFunc(ctx, url)
}

func (m *MockURLRepository) GetByShortCode(ctx context.Context, code string) (*domain.URL, error) {
	return m.GetByShortCodeFunc(ctx, code)
}

func (m *MockURLRepository) GetByID(ctx context.Context, id string) (*domain.URL, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockURLRepository) Update(ctx context.Context, url *domain.URL) error {
	return m.UpdateFunc(ctx, url)
}

func (m *MockURLRepository) Delete(ctx context.Context, code string) error {
	return m.DeleteFunc(ctx, code)
}

func (m *MockURLRepository) ExistsByShortCode(ctx context.Context, code string) (bool, error) {
	return m.ExistsByShortCodeFunc(ctx, code)
}
