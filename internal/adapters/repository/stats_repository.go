package repository

import (
	"context"

	"Dimidroll06/url-link-shortener/internal/core/domain"
	"Dimidroll06/url-link-shortener/internal/core/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewStatsRepository(db *pgxpool.Pool) ports.StatsRepository {
	return &StatsRepositoryImpl{db: db}
}

func (r *StatsRepositoryImpl) RecordAccess(ctx context.Context, stats *domain.URLStats) error {
	query := `
        INSERT INTO url_stats (id, url_id, accessed_at, ip_address, user_agent, referer)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(ctx, query,
		stats.ID,
		stats.URLID,
		stats.AccessedAt,
		stats.IPAddress,
		stats.UserAgent,
		stats.Referer,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *StatsRepositoryImpl) GetTotalAccesses(ctx context.Context, code string) (int64, error) {
	query := `
        SELECT COUNT(*)
        FROM url_stats us
        JOIN urls u ON us.url_id = u.id
        WHERE u.short_code = $1
    `

	var count int64
	err := r.db.QueryRow(ctx, query, code).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *StatsRepositoryImpl) GetRecentAccesses(ctx context.Context, code string, limit int) ([]*domain.URLStats, error) {
	query := `
        SELECT us.id, us.url_id, us.accessed_at, us.ip_address, us.user_agent, us.referer
        FROM url_stats us
        JOIN urls u ON us.url_id = u.id
        WHERE u.short_code = $1
        ORDER BY us.accessed_at DESC
        LIMIT $2
    `

	rows, err := r.db.Query(ctx, query, code, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]*domain.URLStats, 0, limit)
	for rows.Next() {
		s := &domain.URLStats{}
		err := rows.Scan(
			&s.ID,
			&s.URLID,
			&s.AccessedAt,
			&s.IPAddress,
			&s.UserAgent,
			&s.Referer,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *StatsRepositoryImpl) GetByURLID(ctx context.Context, urlID string) ([]*domain.URLStats, error) {
	query := `
        SELECT id, url_id, accessed_at, ip_address, user_agent, referer
        FROM url_stats
        WHERE url_id = $1
        ORDER BY accessed_at DESC
    `

	rows, err := r.db.Query(ctx, query, urlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]*domain.URLStats, 0)
	for rows.Next() {
		s := &domain.URLStats{}
		err := rows.Scan(
			&s.ID,
			&s.URLID,
			&s.AccessedAt,
			&s.IPAddress,
			&s.UserAgent,
			&s.Referer,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}
