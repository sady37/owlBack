package main

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// clearCardStateOnStartup data 启动时清空 card:state:* hash，让 cardagg/sensor 从空白起点重建。
//
// 用 SCAN 批量取 keys 再 DEL（避免 KEYS 锁库）。失败仅 log warning 不阻塞启动——data 自身职责
// 不依赖此清理；清失败最多就是 stale 显示，与原行为等价。
//
// 与 wisefido-sensor publishInitialResetState 配套：sensor 启动后会 publish 所有 room/bed
// vacant 初态重建 hash。启动顺序由 start-owlback-full.sh 保证（data → 10s → cardagg/sensor）。
func clearCardStateOnStartup(ctx context.Context, rdb *redis.Client, logger *zap.Logger) {
	const pattern = "card:state:*"
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
		logger.Info("data startup clear: no card:state keys to clear")
		return
	}
	if err := rdb.Del(ctx, keys...).Err(); err != nil {
		logger.Warn("data startup clear: del failed", zap.Error(err), zap.Int("count", len(keys)))
		return
	}
	logger.Info("data startup clear: cleared card:state hashes", zap.Int("count", len(keys)))
}
