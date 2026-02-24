package services

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"Dimidroll06/url-link-shortener/internal/core/domain"
	servererrors "Dimidroll06/url-link-shortener/internal/core/errors"
	"Dimidroll06/url-link-shortener/internal/core/ports"
)

type StatsService struct {
	statsRepo  ports.StatsRepository
	statsCache ports.StatsCache
	urlRepo    ports.URLRepository
	logger     *zap.Logger
}

func NewStatsService(
	statsRepo ports.StatsRepository,
	statsCache ports.StatsCache,
	urlRepo ports.URLRepository,
	logger *zap.Logger,
) *StatsService {
	return &StatsService{
		statsRepo:  statsRepo,
		statsCache: statsCache,
		urlRepo:    urlRepo,
		logger:     logger,
	}
}

func (s *StatsService) GetDetailedStats(ctx context.Context, code string) (map[string]interface{}, error) {
	url, err := s.urlRepo.GetByShortCode(ctx, code)
	if err != nil {
		s.logger.Error("get url for stats failed",
			zap.String("short_code", code),
			zap.Error(err),
		)
		if errors.Is(err, servererrors.ErrURLNotFound) {
			return nil, servererrors.ErrURLNotFound
		}
		return nil, err
	}

	// 2. Получаем счётчик из кэша
	accessCount, err := s.statsCache.GetAccessCount(ctx, code)
	if err != nil {
		s.logger.Error("get access count failed",
			zap.String("short_code", code),
			zap.Error(err),
		)
		return nil, servererrors.ErrStatsUnavailable
	}

	// 3. Получаем последние переходы из БД
	recentAccesses, err := s.statsRepo.GetRecentAccesses(ctx, code, 10)
	if err != nil {
		s.logger.Error("get recent accesses failed",
			zap.String("short_code", code),
			zap.Error(err),
		)
		return nil, err
	}

	return map[string]interface{}{
		"url":             url,
		"total_accesses":  accessCount,
		"recent_accesses": recentAccesses,
	}, nil
}

func (s *StatsService) RecordAccess(ctx context.Context, code, ipAddress, userAgent, referer string) error {
	url, err := s.urlRepo.GetByShortCode(ctx, code)
	if err != nil {
		s.logger.Error("get url for record access failed",
			zap.String("short_code", code),
			zap.Error(err),
		)
		if errors.Is(err, servererrors.ErrURLNotFound) {
			return servererrors.ErrURLNotFound
		}
		return err
	}

	stats := domain.NewURLStats(url.ID, ipAddress, userAgent, referer)

	if err := s.statsRepo.RecordAccess(ctx, stats); err != nil {
		s.logger.Error("record access in repository failed",
			zap.String("short_code", code),
			zap.Error(err),
		)
		return err
	}

	return nil
}
