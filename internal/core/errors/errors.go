package errors

import "errors"

var (
	ErrInvalidURL        = errors.New("invalid URL")
	ErrInvalidURLScheme  = errors.New("invalid URL scheme: must start with http:// or https://")
	ErrURLTooLong        = errors.New("url too long: maximum 2048 characters")
	ErrURLNotFound       = errors.New("url not found")
	ErrInvalidShortCode  = errors.New("invalid short code")
	ErrShortCodeExists   = errors.New("short code already exists")
	ErrShortCodeNotFound = errors.New("short code not found")
	ErrCacheUnavailable  = errors.New("cache unavailable")
	ErrStatsUnavailable  = errors.New("statistics unavailable")
	ErrURLExpired        = errors.New("url expired")
	ErrURLInactive       = errors.New("url inactive")
)
