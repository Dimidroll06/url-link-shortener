package domain

import (
	"Dimidroll06/url-link-shortener/internal/core/errors"
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID          string     `json:"id"`
	OriginalURL string     `json:"original_url"`
	ShortCode   string     `json:"short_code"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
}

func NewURL(originalURL, shortCode string, expirationDays int) (*URL, error) {
	if originalURL == "" {
		return nil, errors.ErrInvalidURL
	}
	if shortCode == "" {
		return nil, errors.ErrInvalidShortCode
	}

	url := &URL{
		ID:          uuid.New().String(),
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		CreatedAt:   time.Now().UTC(),
		IsActive:    true,
	}

	if expirationDays > 0 {
		expiresAt := time.Now().UTC().AddDate(0, 0, expirationDays)
		url.ExpiresAt = &expiresAt
	}

	return url, nil
}

func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*u.ExpiresAt)
}

func (u *URL) Validate() error {
	if u.OriginalURL == "" {
		return errors.ErrInvalidURL
	}
	if u.ShortCode == "" {
		return errors.ErrInvalidShortCode
	}
	if u.IsExpired() {
		return errors.ErrURLExpired
	}
	if !u.IsActive {
		return errors.ErrURLInactive
	}
	return nil
}
