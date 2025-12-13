package redis

import (
	"context"
	"owl-common/config"
	
	"github.com/go-redis/redis/v8"
)

// Client Redis客户端类型别名
type Client = redis.Client

// NewRedisClient 创建Redis客户端
func NewRedisClient(cfg *config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

// Ping 测试Redis连接
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}

// Close 关闭Redis连接
func Close(client *redis.Client) error {
	return client.Close()
}

