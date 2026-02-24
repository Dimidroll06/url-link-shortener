package services_test

import (
	"context"
	"errors"
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

func TestStatsService_GetDetailedStats(t *testing.T) {
	t.Parallel()

	t.Run("success_get_detailed_stats", func(t *testing.T) {
		t.Parallel()

		mockStatsCache := mockCache.NewMockStatsCache()
		mockStatsRepo := mockRepo.NewMockStatsRepository()
		mockURLRepo := mockRepo.NewMockURLRepository()
		logger := zap.NewNop()

		expectedURL := &domain.URL{
			ID:          "test-id",
			OriginalURL: "https://example.com",
			ShortCode:   "abc123",
			CreatedAt:   time.Now().UTC(),
			IsActive:    true,
		}

		expectedStats := []*domain.URLStats{
			{ID: "stat-1", URLID: "test-id", AccessedAt: time.Now().UTC()},
			{ID: "stat-2", URLID: "test-id", AccessedAt: time.Now().UTC()},
		}

		mockURLRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			assert.Equal(t, "abc123", code)
			return expectedURL, nil
		}

		mockStatsCache.GetAccessCountFunc = func(ctx context.Context, code string) (int64, error) {
			return 42, nil
		}

		mockStatsRepo.GetRecentAccessesFunc = func(ctx context.Context, code string, limit int) ([]*domain.URLStats, error) {
			assert.Equal(t, "abc123", code)
			assert.Equal(t, 10, limit)
			return expectedStats, nil
		}

		svc := services.NewStatsService(mockStatsRepo, mockStatsCache, mockURLRepo, logger)

		result, err := svc.GetDetailedStats(context.Background(), "abc123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedURL, result["url"])
		assert.Equal(t, int64(42), result["total_accesses"])
		assert.Equal(t, expectedStats, result["recent_accesses"])
	})

	t.Run("error_url_not_found", func(t *testing.T) {
		t.Parallel()

		mockStatsCache := mockCache.NewMockStatsCache()
		mockStatsRepo := mockRepo.NewMockStatsRepository()
		mockURLRepo := mockRepo.NewMockURLRepository()
		logger := zap.NewNop()

		mockURLRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLNotFound
		}

		svc := services.NewStatsService(mockStatsRepo, mockStatsCache, mockURLRepo, logger)
		result, err := svc.GetDetailedStats(context.Background(), "abc123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrURLNotFound)
		assert.Nil(t, result)
	})

	t.Run("error_stats_cache_unavailable", func(t *testing.T) {
		t.Parallel()

		mockStatsCache := mockCache.NewMockStatsCache()
		mockStatsRepo := mockRepo.NewMockStatsRepository()
		mockURLRepo := mockRepo.NewMockURLRepository()
		logger := zap.NewNop()

		mockURLRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return &domain.URL{ShortCode: code, IsActive: true}, nil
		}

		mockStatsCache.GetAccessCountFunc = func(ctx context.Context, code string) (int64, error) {
			return 0, errors.New("redis connection failed")
		}

		svc := services.NewStatsService(mockStatsRepo, mockStatsCache, mockURLRepo, logger)
		result, err := svc.GetDetailedStats(context.Background(), "abc123")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrStatsUnavailable)
		assert.Nil(t, result)
	})

	t.Run("error_repository_failure", func(t *testing.T) {
		t.Parallel()

		mockStatsCache := mockCache.NewMockStatsCache()
		mockStatsRepo := mockRepo.NewMockStatsRepository()
		mockURLRepo := mockRepo.NewMockURLRepository()
		logger := zap.NewNop()

		mockURLRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return &domain.URL{ShortCode: code, IsActive: true}, nil
		}

		mockStatsCache.GetAccessCountFunc = func(ctx context.Context, code string) (int64, error) {
			return 42, nil
		}

		mockStatsRepo.GetRecentAccessesFunc = func(ctx context.Context, code string, limit int) ([]*domain.URLStats, error) {
			return nil, errors.New("db query failed")
		}

		svc := services.NewStatsService(mockStatsRepo, mockStatsCache, mockURLRepo, logger)
		result, err := svc.GetDetailedStats(context.Background(), "abc123")

		assert.Error(t, err)
		assert.NotErrorIs(t, err, servererrors.ErrStatsUnavailable)
		assert.Nil(t, result)
	})
}

func TestStatsService_RecordAccess(t *testing.T) {
	t.Parallel()

	t.Run("success_record_access", func(t *testing.T) {
		t.Parallel()

		mockStatsCache := mockCache.NewMockStatsCache()
		mockStatsRepo := mockRepo.NewMockStatsRepository()
		mockURLRepo := mockRepo.NewMockURLRepository()
		logger := zap.NewNop()

		mockURLRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return &domain.URL{ID: "test-id", ShortCode: code, IsActive: true}, nil
		}

		mockStatsRepo.RecordAccessFunc = func(ctx context.Context, stats *domain.URLStats) error {
			assert.Equal(t, "test-id", stats.URLID)
			assert.Equal(t, "192.168.1.1", stats.IPAddress)
			assert.Equal(t, "Mozilla/5.0", stats.UserAgent)
			assert.Equal(t, "https://referrer.com", stats.Referer)
			return nil
		}

		svc := services.NewStatsService(mockStatsRepo, mockStatsCache, mockURLRepo, logger)
		err := svc.RecordAccess(context.Background(), "abc123", "192.168.1.1", "Mozilla/5.0", "https://referrer.com")

		assert.NoError(t, err)
	})

	t.Run("error_url_not_found", func(t *testing.T) {
		t.Parallel()

		mockStatsCache := mockCache.NewMockStatsCache()
		mockStatsRepo := mockRepo.NewMockStatsRepository()
		mockURLRepo := mockRepo.NewMockURLRepository()
		logger := zap.NewNop()

		mockURLRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLNotFound
		}

		svc := services.NewStatsService(mockStatsRepo, mockStatsCache, mockURLRepo, logger)
		err := svc.RecordAccess(context.Background(), "abc123", "192.168.1.1", "Mozilla/5.0", "")

		assert.Error(t, err)
		assert.ErrorIs(t, err, servererrors.ErrURLNotFound)
	})

	t.Run("error_repository_failure", func(t *testing.T) {
		t.Parallel()

		mockStatsCache := mockCache.NewMockStatsCache()
		mockStatsRepo := mockRepo.NewMockStatsRepository()
		mockURLRepo := mockRepo.NewMockURLRepository()
		logger := zap.NewNop()

		mockURLRepo.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return &domain.URL{ID: "test-id", ShortCode: code, IsActive: true}, nil
		}

		mockStatsRepo.RecordAccessFunc = func(ctx context.Context, stats *domain.URLStats) error {
			return errors.New("db insert failed")
		}

		svc := services.NewStatsService(mockStatsRepo, mockStatsCache, mockURLRepo, logger)
		err := svc.RecordAccess(context.Background(), "abc123", "192.168.1.1", "Mozilla/5.0", "")

		assert.Error(t, err)
		assert.NotErrorIs(t, err, servererrors.ErrURLNotFound)
	})
}
