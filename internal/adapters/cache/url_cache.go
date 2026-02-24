package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"Dimidroll06/url-link-shortener/internal/core/domain"
	"Dimidroll06/url-link-shortener/internal/core/ports"

	"github.com/redis/go-redis/v9"
)

type URLCacheImpl struct {
	client *redis.Client
	prefix string
}

func NewURLCache(client *redis.Client, prefix string) ports.URLCache {
	return &URLCacheImpl{
		client: client,
		prefix: prefix,
	}
}

func (c *URLCacheImpl) Get(ctx context.Context, code string) (*domain.URL, error) {
	key := c.makeKey(code)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrURLNotFound
		}
		return nil, fmt.Errorf("cache get: %w", err)
	}

	url := &domain.URL{}
	if err := json.Unmarshal(data, url); err != nil {
		return nil, fmt.Errorf("cache unmarshal: %w", err)
	}

	return url, nil
}

func (c *URLCacheImpl) Set(ctx context.Context, url *domain.URL, ttl int) error {
	key := c.makeKey(url.ShortCode)

	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	err = c.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("cache set: %w", err)
	}

	return nil
}

func (c *URLCacheImpl) Delete(ctx context.Context, code string) error {
	key := c.makeKey(code)
	return c.client.Del(ctx, key).Err()
}

func (c *URLCacheImpl) Exists(ctx context.Context, code string) (bool, error) {
	key := c.makeKey(code)
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists: %w", err)
	}
	return result > 0, nil
}

func (c *URLCacheImpl) makeKey(code string) string {
	return fmt.Sprintf("%s:url:%s", c.prefix, code)
}
