package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"Dimidroll06/url-link-shortener/internal/core/domain"
	"Dimidroll06/url-link-shortener/internal/core/ports"
)

type URLService struct {
	repo           ports.URLRepository
	cache          ports.URLCache
	statsCache     ports.StatsCache
	logger         *zap.Logger
	cacheTTL       int
	expirationDays int
}

func NewURLService(
	repo ports.URLRepository,
	cache ports.URLCache,
	statsCache ports.StatsCache,
	logger *zap.Logger,
	expirationDays int,
) *URLService {
	return &URLService{
		repo:           repo,
		cache:          cache,
		statsCache:     statsCache,
		logger:         logger,
		cacheTTL:       int(1 * (time.Hour / time.Second)),
		expirationDays: expirationDays,
	}
}

func (s *URLService) Create(ctx context.Context, originalURL string) (*domain.URL, error) {
	if err := s.validateURL(originalURL); err != nil {
		return nil, err
	}

	i := 0 // попытки
	shortCode := s.generateShortCode(originalURL)
	exists, err := s.repo.ExistsByShortCode(ctx, shortCode)
	for exists && i <= 10 {
		if i == 10 {
			s.logger.Error("too many attempts to generate unique short code",
				zap.String("original_url", originalURL),
				zap.Int("attempts", i),
			)
			return nil, fmt.Errorf("too many attempts to generate unique short code")
		}

		if err != nil {
			s.logger.Error("check short code exists", zap.Error(err))
			return nil, fmt.Errorf("check uniqueness: %w", err)
		}

		shortCode = s.generateShortCode(fmt.Sprintf("%s-%d", originalURL, time.Now().UnixNano()))
		exists, err = s.repo.ExistsByShortCode(ctx, shortCode)
		i++
	}

	url, err := domain.NewURL(originalURL, shortCode, s.expirationDays)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, url); err != nil {
		s.logger.Error("create url in repository", zap.Error(err))
		return nil, fmt.Errorf("create url: %w", err)
	}

	if err := s.cache.Set(ctx, url, s.cacheTTL); err != nil {
		s.logger.Warn("cache set failed", zap.Error(err))
	}

	s.logger.Info("url created",
		zap.String("short_code", url.ShortCode),
		zap.String("original_url", originalURL),
	)

	return url, nil
}

func (s *URLService) GetByShortCode(ctx context.Context, code string) (*domain.URL, error) {
	url, err := s.cache.Get(ctx, code)
	if err == nil {
		if err := url.Validate(); err != nil {
			s.cache.Delete(ctx, code)
			return nil, err
		}

		go s.recordAccess(code)

		return url, nil
	}

	if !errors.Is(err, domain.ErrURLNotFound) {
		s.logger.Warn("cache get error", zap.Error(err))
	}

	url, err = s.repo.GetByShortCode(ctx, code)
	if err != nil {
		s.logger.Error("get url from repository", zap.Error(err))
		return nil, err
	}

	if err := url.Validate(); err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, url, s.cacheTTL); err != nil {
		s.logger.Warn("cache set failed", zap.Error(err))
	}

	go s.recordAccess(code)

	s.logger.Debug("url retrieved", zap.String("short_code", code))

	return url, nil
}

func (s *URLService) Delete(ctx context.Context, code string) error {
	_, err := s.repo.GetByShortCode(ctx, code)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, code); err != nil {
		s.logger.Error("delete url from repository", zap.Error(err))
		return fmt.Errorf("delete url: %w", err)
	}

	if err := s.cache.Delete(ctx, code); err != nil {
		s.logger.Warn("cache delete failed", zap.Error(err))
	}

	if err := s.statsCache.Reset(ctx, code); err != nil {
		s.logger.Warn("stats cache reset failed", zap.Error(err))
	}

	s.logger.Info("url deleted", zap.String("short_code", code))

	return nil
}

func (s *URLService) GetStats(ctx context.Context, code string) (int64, error) {
	_, err := s.repo.GetByShortCode(ctx, code)
	if err != nil {
		return 0, err
	}

	count, err := s.statsCache.GetAccessCount(ctx, code)
	if err != nil {
		s.logger.Error("get stats from cache", zap.Error(err))
		return 0, fmt.Errorf("get stats: %w", err)
	}

	return count, nil
}

func (s *URLService) validateURL(url string) error {
	if url == "" {
		return domain.ErrInvalidURL
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("invalid url scheme: must start with http:// or https://")
	}

	if len(url) > 2048 {
		return fmt.Errorf("url too long: maximum 2048 characters")
	}

	return nil
}

func (s *URLService) generateShortCode(url string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", url, time.Now().UnixNano())))
	encoded := base64.URLEncoding.EncodeToString(hash[:])

	if len(encoded) > 8 {
		return encoded[:8]
	}
	return encoded
}

func (s *URLService) recordAccess(code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.statsCache == nil {
		return
	}

	count, err := s.statsCache.IncrementAccess(ctx, code)
	if err != nil {
		s.logger.Error("increment access counter", zap.Error(err))
		return
	}

	s.logger.Debug("access recorded",
		zap.String("short_code", code),
		zap.Int64("count", count),
	)
}
