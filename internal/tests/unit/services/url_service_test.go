package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"Dimidroll06/url-link-shortener/internal/core/domain"
	servererrors "Dimidroll06/url-link-shortener/internal/core/errors"
	"Dimidroll06/url-link-shortener/internal/core/services"
	mockCache "Dimidroll06/url-link-shortener/internal/tests/mock/cache"
	mockRepo "Dimidroll06/url-link-shortener/internal/tests/mock/repository"
)

func TestURLService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		originalURL string
		setupMocks  func(*mockRepo.MockURLRepository, *mockCache.MockURLCache, *mockCache.MockStatsCache)
		wantURL     bool
		wantErr     error
	}{
		{
			name:        "success_create_new_url",
			originalURL: "https://example.com",
			setupMocks: func(repo *mockRepo.MockURLRepository, cache *mockCache.MockURLCache, statsCache *mockCache.MockStatsCache) {
				repo.ExistsByShortCodeFunc = func(ctx context.Context, code string) (bool, error) {
					return false, nil
				}
				repo.CreateFunc = func(ctx context.Context, url *domain.URL) error {
					return nil
				}
				cache.SetFunc = func(ctx context.Context, url *domain.URL, ttl int) error {
					return nil
				}
			},
			wantURL: true,
			wantErr: nil,
		},
		{
			name:        "error_empty_url",
			originalURL: "",
			setupMocks: func(repo *mockRepo.MockURLRepository, cache *mockCache.MockURLCache, statsCache *mockCache.MockStatsCache) {
			},
			wantURL: false,
			wantErr: servererrors.ErrInvalidURL,
		},
		{
			name:        "error_invalid_scheme",
			originalURL: "ftp://example.com",
			setupMocks: func(repo *mockRepo.MockURLRepository, cache *mockCache.MockURLCache, statsCache *mockCache.MockStatsCache) {
			},
			wantURL: false,
			wantErr: servererrors.ErrInvalidURLScheme,
		},
		{
			name:        "error_url_too_long",
			originalURL: "https://" + strings.Repeat("a", 2050),
			setupMocks: func(repo *mockRepo.MockURLRepository, cache *mockCache.MockURLCache, statsCache *mockCache.MockStatsCache) {
			},
			wantURL: false,
			wantErr: servererrors.ErrURLTooLong,
		},
		{
			name:        "error_repository_failure",
			originalURL: "https://example.com",
			setupMocks: func(repo *mockRepo.MockURLRepository, cache *mockCache.MockURLCache, statsCache *mockCache.MockStatsCache) {
				repo.ExistsByShortCodeFunc = func(ctx context.Context, code string) (bool, error) {
					return false, nil
				}
				repo.CreateFunc = func(ctx context.Context, url *domain.URL) error {
					return errors.New("db connection failed")
				}
			},
			wantURL: false,
			wantErr: servererrors.ErrCacheUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mockRepo.NewMockURLRepository()
			mockStatsCache := mockCache.NewMockStatsCache()
			mockCache := mockCache.NewMockURLCache()
			logger := zap.NewNop()

			tt.setupMocks(mockRepo, mockCache, mockStatsCache)

			svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
			url, err := svc.Create(context.Background(), tt.originalURL)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, url)
			} else {
				assert.NoError(t, err)
				if tt.wantURL {
					assert.NotNil(t, url)
					assert.Equal(t, tt.originalURL, url.OriginalURL)
					assert.NotEmpty(t, url.ShortCode)
					assert.True(t, url.IsActive)
				}
			}
		})
	}
}

func TestURLService_GetByShortCode(t *testing.T) {
	t.Parallel()

	t.Run("success_cache_hit", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		expectedURL := &domain.URL{
			ID:          "test-id",
			OriginalURL: "https://example.com",
			ShortCode:   "abc123",
			CreatedAt:   time.Now().UTC(),
			IsActive:    true,
		}

		mockCache.GetFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			assert.Equal(t, "abc123", code)
			return expectedURL, nil
		}

		mockRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			t.Error("repository should not be called on cache hit")
			return nil, nil
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		url, err := svc.GetByShortCode(context.Background(), "abc123")

		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
	})

	t.Run("success_cache_miss_repository_hit", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		expectedURL := &domain.URL{
			ID:          "test-id",
			OriginalURL: "https://example.com",
			ShortCode:   "abc123",
			CreatedAt:   time.Now().UTC(),
			IsActive:    true,
		}

		mockCache.GetFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLNotFound
		}

		mockRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return expectedURL, nil
		}

		mockCache.SetFunc = func(ctx context.Context, url *domain.URL, ttl int) error {
			return nil
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		url, err := svc.GetByShortCode(context.Background(), "abc123")

		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
	})

	t.Run("error_url_not_found", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		mockCache.GetFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLNotFound
		}

		mockRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLNotFound
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		url, err := svc.GetByShortCode(context.Background(), "abc123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrURLNotFound)
		assert.Nil(t, url)
	})

	t.Run("error_expired_url", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		expiredURL := &domain.URL{
			ID:          "test-id",
			OriginalURL: "https://example.com",
			ShortCode:   "abc123",
			CreatedAt:   time.Now().AddDate(0, 0, -31),
			ExpiresAt:   func() *time.Time { t := time.Now().AddDate(0, 0, -1); return &t }(),
			IsActive:    true,
		}

		mockCache.GetFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return expiredURL, nil
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		url, err := svc.GetByShortCode(context.Background(), "abc123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrURLExpired)
		assert.Nil(t, url)
	})

	t.Run("error_inactive_url", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		inactiveURL := &domain.URL{
			ID:          "test-id",
			OriginalURL: "https://example.com",
			ShortCode:   "abc123",
			CreatedAt:   time.Now().UTC(),
			IsActive:    false,
		}

		mockCache.GetFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return inactiveURL, nil
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		url, err := svc.GetByShortCode(context.Background(), "abc123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrURLInactive)
		assert.Nil(t, url)
	})
}

func TestURLService_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success_delete", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		mockRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return &domain.URL{ShortCode: code, IsActive: true}, nil
		}

		mockRepo.DeleteFunc = func(ctx context.Context, code string) error {
			return nil
		}

		mockCache.DeleteFunc = func(ctx context.Context, code string) error {
			return nil
		}

		mockStatsCache.ResetFunc = func(ctx context.Context, code string) error {
			return nil
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		err := svc.Delete(context.Background(), "abc123")

		assert.NoError(t, err)
	})

	t.Run("error_url_not_found", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		mockRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLNotFound
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		err := svc.Delete(context.Background(), "abc123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrURLNotFound)
	})

	t.Run("error_repository_failure", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		mockRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return &domain.URL{ShortCode: code}, nil
		}

		mockRepo.DeleteFunc = func(ctx context.Context, code string) error {
			return errors.New("db error")
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		err := svc.Delete(context.Background(), "abc123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrCacheUnavailable)
	})
}

func TestURLService_GetStats(t *testing.T) {
	t.Parallel()

	t.Run("success_get_stats", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		mockRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return &domain.URL{ShortCode: code, IsActive: true}, nil
		}

		mockStatsCache.GetAccessCountFunc = func(ctx context.Context, code string) (int64, error) {
			return 42, nil
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		count, err := svc.GetStats(context.Background(), "abc123")

		assert.NoError(t, err)
		assert.Equal(t, int64(42), count)
	})

	t.Run("error_url_not_found", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockStatsCache := mockCache.NewMockStatsCache()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		mockRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLNotFound
		}

		svc := services.NewURLService(mockRepo, mockCache, mockStatsCache, logger, 30)
		count, err := svc.GetStats(context.Background(), "abc123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrURLNotFound)
		assert.Equal(t, int64(0), count)
	})

	t.Run("error_stats_cache_unavailable", func(t *testing.T) {
		t.Parallel()

		mockRepo := mockRepo.NewMockURLRepository()
		mockCache := mockCache.NewMockURLCache()
		logger := zap.NewNop()

		mockRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return &domain.URL{ShortCode: code, IsActive: true}, nil
		}

		svc := services.NewURLService(mockRepo, mockCache, nil, logger, 30)
		count, err := svc.GetStats(context.Background(), "abc123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrStatsUnavailable)
		assert.Equal(t, int64(0), count)
	})
}
