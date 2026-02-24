package cache

import (
	"context"
	"fmt"
	"time"

	"Dimidroll06/url-link-shortener/internal/core/ports"

	"github.com/redis/go-redis/v9"
)

type StatsCacheImpl struct {
	client *redis.Client
	prefix string
}

func NewStatsCache(client *redis.Client, prefix string) ports.StatsCache {
	return &StatsCacheImpl{
		client: client,
		prefix: prefix,
	}
}

func (c *StatsCacheImpl) IncrementAccess(ctx context.Context, code string) (int64, error) {
	key := c.makeKey(code)
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("cache increment: %w", err)
	}
	c.client.Expire(ctx, key, 24*time.Hour)
	return count, nil
}

func (c *StatsCacheImpl) GetAccessCount(ctx context.Context, code string) (int64, error) {
	key := c.makeKey(code)
	count, err := c.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("cache get count: %w", err)
	}
	return count, nil
}

func (c *StatsCacheImpl) Reset(ctx context.Context, code string) error {
	key := c.makeKey(code)
	return c.client.Del(ctx, key).Err()
}

func (c *StatsCacheImpl) makeKey(code string) string {
	return fmt.Sprintf("%s:stats:%s", c.prefix, code)
}
