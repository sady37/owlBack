package aggregator

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// ErrCacheMiss 表示缓存不存在
var ErrCacheMiss = errors.New("cache miss")

// KVStore 抽象的 KV 存储（用于在单元测试中替换 Redis）
type KVStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}

// RedisKVStore 基于 go-redis 的 KV 实现
type RedisKVStore struct {
	client *redis.Client
}

func NewRedisKVStore(client *redis.Client) *RedisKVStore {
	return &RedisKVStore{client: client}
}

func (r *RedisKVStore) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrCacheMiss
		}
		return "", err
	}
	return val, nil
}

func (r *RedisKVStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}


