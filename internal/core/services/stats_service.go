package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"Dimidroll06/url-link-shortener/internal/core/domain"
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
		return nil, err
	}

	accessCount, err := s.statsCache.GetAccessCount(ctx, code)
	if err != nil {
		s.logger.Error("get access count", zap.Error(err))
		return nil, fmt.Errorf("get access count: %w", err)
	}

	recentAccesses, err := s.statsRepo.GetRecentAccesses(ctx, code, 10)
	if err != nil {
		s.logger.Error("get recent accesses", zap.Error(err))
		return nil, fmt.Errorf("get recent accesses: %w", err)
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
		return err
	}

	stats := domain.NewURLStats(url.ID, ipAddress, userAgent, referer)

	if err := s.statsRepo.RecordAccess(ctx, stats); err != nil {
		s.logger.Error("record access in repository", zap.Error(err))
		return fmt.Errorf("record access: %w", err)
	}

	return nil
}
