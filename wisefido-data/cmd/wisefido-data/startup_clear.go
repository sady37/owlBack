package main

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// clearRedisOnStartup data 启动时初始化两类 Hash：
//   - card:state:*    HDEL 动态字段 + bed_state 写 LeftBed (=1) 默认；保 hash 形态供 cardagg 后续 merge
//   - device:status:* DEL 全清（cardagg DeviceStatusTracker 看门狗 + 心跳重建）
//
// bed_state default LeftBed: "在床需证据"原则——startup_clear 后没有 fresh BedState evidence
// 时缺省视为不在床，避免 hash 残留 0 值被误读为 InBed。
// 用 SCAN 批量取 keys（避免 KEYS 锁库）。失败仅 log warning 不阻塞启动。
func clearRedisOnStartup(ctx context.Context, rdb *redis.Client, logger *zap.Logger) {
	resetCardStateHashes(ctx, rdb, logger)
	clearHashPattern(ctx, rdb, logger, "device:status:*")
}

// resetCardStateHashes 对每个 card:state:{cardID} hash：HDEL 动态字段 + HSET bed_state default LeftBed。
func resetCardStateHashes(ctx context.Context, rdb *redis.Client, logger *zap.Logger) {
	const batchSize = 100
	iter := rdb.Scan(ctx, 0, "card:state:*", batchSize).Iterator()
	keys := make([]string, 0, batchSize)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		logger.Warn("data startup clear: scan failed", zap.Error(err))
		return
	}
	if len(keys) == 0 {
		logger.Info("data startup clear: card:state nothing to reset")
		return
	}
	const bedStateLeftBed = `{"bed_status":1}`
	pipe := rdb.Pipeline()
	for _, k := range keys {
		pipe.HDel(ctx, k, "target", "room_state", "alarm_state", "display", "message")
		pipe.HSet(ctx, k, "bed_state", bedStateLeftBed)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		logger.Warn("data startup clear: card:state pipeline failed", zap.Error(err), zap.Int("count", len(keys)))
		return
	}
	logger.Info("data startup clear: card:state reset (bed_state=LeftBed default)", zap.Int("count", len(keys)))
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
