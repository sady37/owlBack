package main

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// clearRedisOnStartup data 启动时清两类 Hash：
//   - card:state:*    cardagg 收到 op=reset 后从 DB 重 build；clear 是保 consistency 的前提
//   - device:status:* cardagg DeviceStatusTracker 看门狗 + 心跳重建
//
// 设计意图（非 HA 场景）：重启罕见，clean slate + DB 重读 比 in-place 增量同步更稳。
// 用 SCAN 批量取 keys 再 DEL（避免 KEYS 锁库）。失败仅 log warning 不阻塞启动。
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
