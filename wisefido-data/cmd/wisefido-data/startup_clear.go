package main

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// clearRedisOnStartup data 启动时清两类 Hash：
//   - card:state:*   让 cardagg/sensor 从空白重建；与 wisefido-sensor publishInitialResetState 配套
//   - device:status:* 让 cardagg DeviceStatusTracker 从看门狗 / 心跳重建（避开存量旧字段名遗留）
//
// 用 SCAN 批量取 keys 再 DEL（避免 KEYS 锁库）。失败仅 log warning 不阻塞启动。
// 启动顺序由 start-owlback-full.sh 保证（data → 10s → cardagg/sensor）。
func clearRedisOnStartup(ctx context.Context, rdb *redis.Client, logger *zap.Logger) {
	clearHashPattern(ctx, rdb, logger, "card:state:*")
	clearHashPattern(ctx, rdb, logger, "device:status:*")
}

func clearHashPattern(ctx context.Context, rdb *redis.Client, logger *zap.Logger, pattern string) {
	const batchSize = 100
	iter := rdb.Scan(ctx, 0, pattern, batchSize).Iterator()
	keys := make([]string, 0, batchSize)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		logger.Warn("data startup clear: scan failed", zap.Error(err), zap.String("pattern", pattern))
		return
	}
	if len(keys) == 0 {
		logger.Info("data startup clear: nothing to clear", zap.String("pattern", pattern))
		return
	}
	if err := rdb.Del(ctx, keys...).Err(); err != nil {
		logger.Warn("data startup clear: del failed", zap.Error(err), zap.String("pattern", pattern), zap.Int("count", len(keys)))
		return
	}
	logger.Info("data startup clear: cleared hashes", zap.String("pattern", pattern), zap.Int("count", len(keys)))
}
