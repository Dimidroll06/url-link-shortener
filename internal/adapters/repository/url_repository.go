package repository

import (
	"context"
	"errors"
	"time"

	"Dimidroll06/url-link-shortener/internal/core/domain"
	servererrors "Dimidroll06/url-link-shortener/internal/core/errors"
	"Dimidroll06/url-link-shortener/internal/core/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) ports.URLRepository {
	return &URLRepositoryImpl{db: db}
}

func (r *URLRepositoryImpl) Create(ctx context.Context, url *domain.URL) error {
	query := `
        INSERT INTO urls (id, original_url, short_code, created_at, expires_at, is_active)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(ctx, query,
		url.ID,
		url.OriginalURL,
		url.ShortCode,
		url.CreatedAt,
		nullTime(url.ExpiresAt),
		url.IsActive,
	)

	if err != nil {
		if isUniqueViolation(err, "urls_short_code_key") {
			return servererrors.ErrShortCodeExists
		}
		return err
	}

	return nil
}

func (r *URLRepositoryImpl) GetByShortCode(ctx context.Context, code string) (*domain.URL, error) {
	query := `
        SELECT id, original_url, short_code, created_at, expires_at, is_active
        FROM urls
        WHERE short_code = $1 AND is_active = true
    `

	url := &domain.URL{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrURLNotFound
		}
		return nil, err
	}

	return url, nil
}

func (r *URLRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.URL, error) {
	query := `
        SELECT id, original_url, short_code, created_at, expires_at, is_active
        FROM urls
        WHERE id = $1
    `

	url := &domain.URL{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrURLNotFound
		}
		return nil, err
	}

	return url, nil
}

func (r *URLRepositoryImpl) Update(ctx context.Context, url *domain.URL) error {
	query := `
        UPDATE urls
        SET original_url = $2, expires_at = $3, is_active = $4
        WHERE id = $1
    `

	result, err := r.db.Exec(ctx, query,
		url.ID,
		url.OriginalURL,
		nullTime(url.ExpiresAt),
		url.IsActive,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrURLNotFound
	}

	return nil
}

func (r *URLRepositoryImpl) Delete(ctx context.Context, code string) error {
	query := `
        UPDATE urls
        SET is_active = false
        WHERE short_code = $1
    `

	result, err := r.db.Exec(ctx, query, code)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrURLNotFound
	}

	return nil
}

func (r *URLRepositoryImpl) ExistsByShortCode(ctx context.Context, code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, code).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == constraint
	}
	return false
}
