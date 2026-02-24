package domain

import "errors"

var (
	ErrInvalidURL       = errors.New("invalid URL")
	ErrInvalidShortCode = errors.New("invalid short code")
	ErrURLNotFound      = errors.New("URL not found")
	ErrURLExpired       = errors.New("URL has expired")
	ErrURLInactive      = errors.New("URL is inactive")
	ErrShortCodeExists  = errors.New("short code already exists")
)
