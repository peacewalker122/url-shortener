package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker/v2"

	"url-shortener/internal/shared/config"
)

type URLCache struct {
	client  *redis.Client
	breaker *gobreaker.CircuitBreaker[any]
	ttl     time.Duration
}

type NoopCache struct{}

func NewURLCache(cfg config.Config) (*URLCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  cfg.RedisDialTO,
		ReadTimeout:  cfg.RedisReadTO,
		WriteTimeout: cfg.RedisWriteTO,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	settings := gobreaker.Settings{
		Name:        "redis-cache",
		MaxRequests: cfg.CBMaxRequests,
		Interval:    cfg.CBInterval,
		Timeout:     cfg.CBTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < cfg.CBMinRequests {
				return false
			}

			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= cfg.CBFailureRatio
		},
	}

	breaker := gobreaker.NewCircuitBreaker[any](settings)

	return &URLCache{client: client, breaker: breaker, ttl: cfg.RedisTTL}, nil
}

func (c *URLCache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}

	return c.client.Close()
}

func (c *URLCache) Get(ctx context.Context, shortCode string) (string, bool, error) {
	if c == nil || c.client == nil {
		return "", false, nil
	}

	key := cacheKey(shortCode)

	value, err := c.breaker.Execute(func() (any, error) {
		result, redisErr := c.client.Get(ctx, key).Result()
		if errors.Is(redisErr, redis.Nil) {
			return "", nil
		}
		if redisErr != nil {
			return nil, redisErr
		}
		return result, nil
	})
	if err != nil {
		return "", false, err
	}

	stringValue, ok := value.(string)
	if !ok || stringValue == "" {
		return "", false, nil
	}

	return stringValue, true, nil
}

func (c *URLCache) Set(ctx context.Context, shortCode string, longURL string) error {
	if c == nil || c.client == nil {
		return nil
	}

	key := cacheKey(shortCode)

	_, err := c.breaker.Execute(func() (any, error) {
		return nil, c.client.Set(ctx, key, longURL, c.ttl).Err()
	})

	return err
}

func (NoopCache) Get(context.Context, string) (string, bool, error) {
	return "", false, nil
}

func (NoopCache) Set(context.Context, string, string) error {
	return nil
}

func cacheKey(shortCode string) string {
	return "url:" + shortCode
}
