package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	prefix string
}

var _ Cache = (*Redis)(nil)

func NewRedis(client *redis.Client, prefix string) *Redis {
	return &Redis{
		client: client,
		prefix: prefix,
	}
}

func (r *Redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache key %q: %w", r.key(key), err)
	}

	if err := r.client.Set(ctx, r.key(key), data, ttl).Err(); err != nil {
		return fmt.Errorf("set cache key %q: %w", r.key(key), err)
	}

	return nil
}

func (r *Redis) Get(ctx context.Context, key string, dest any) (bool, error) {
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get cache key %q: %w", r.key(key), err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("unmarshal cache key %q: %w", r.key(key), err)
	}

	return true, nil
}

func (r *Redis) Invalidate(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, r.key(key)).Err(); err != nil {
		return fmt.Errorf("invalidate cache key %q: %w", r.key(key), err)
	}

	return nil
}

func (r *Redis) InvalidateAll(ctx context.Context) error {
	var cursor uint64
	pattern := r.key("*")

	for {
		keys, next, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("scan cache keys %q: %w", pattern, err)
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("invalidate cache keys %q: %w", pattern, err)
			}
		}

		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

func (r *Redis) key(key string) string {
	return r.prefix + ":" + key
}
