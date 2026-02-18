package store

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var ErrMiss = errors.New("cache miss")

type KV interface {
	Get(ctx context.Context, key string) (string, error)
	MGet(ctx context.Context, keys []string) ([]string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	ScanKeys(ctx context.Context, pattern string) ([]string, error)
}

type RedisKV struct {
	c *redis.Client
}

func NewRedisKV(c *redis.Client) *RedisKV { return &RedisKV{c: c} }

func (r *RedisKV) Get(ctx context.Context, key string) (string, error) {
	val, err := r.c.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrMiss
		}
		return "", err
	}
	return val, nil
}

func (r *RedisKV) MGet(ctx context.Context, keys []string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	vals, err := r.c.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	result := make([]string, len(vals))
	for i, v := range vals {
		if v != nil {
			result[i] = v.(string)
		}
	}
	return result, nil
}

func (r *RedisKV) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.c.Set(ctx, key, value, ttl).Err()
}

func (r *RedisKV) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.c.Del(ctx, keys...).Err()
}

func (r *RedisKV) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		k, next, err := r.c.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, k...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}
