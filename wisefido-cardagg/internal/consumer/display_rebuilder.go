// display_rebuilder.go — cardagg 周期性全量重 build `card:state.display`。
//
// Why: cardagg 只在 state 变化时通过 SensorStateProjector republish display；
// 长时间无事件的卡，display 会保留旧 schema/算法的 JSON 跨重启滞留，违反
// 「hard reload 永远拿到正确 state」契约。周期 rebuild 兜底，确保 display
// 始终贴当前 cardagg 代码逻辑。
//
// 周期间隔通过 RunPeriodicRebuild 参数注入；生产建议 5min，开发期可缩到 30s
// 加速验证。

package consumer

import (
	"context"
	"strings"
	"time"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"owl-common/card"
	"wisefido-cardagg/internal/service"
)

// RebuildAllDisplays 扫所有 card:state:* hash，重新派生 display 写回。Idempotent。
// 用 SCAN 不用 KEYS——避免大 keyspace 阻塞 Redis；batch 200。
func RebuildAllDisplays(
	ctx context.Context,
	client *redislib.Client,
	reader *card.Reader,
	writer *card.Writer,
	picker *UnitPicker,
	meta *service.DeviceMetaCache,
	logger *zap.Logger,
) {
	const scanBatch = 200
	const matchPattern = "card:state:*"
	const prefix = "card:state:"

	t0 := time.Now()
	var cursor uint64
	scanned := 0
	rebuilt := 0

	for {
		keys, next, err := client.Scan(ctx, cursor, matchPattern, scanBatch).Result()
		if err != nil {
			logger.Error("RebuildAllDisplays: scan failed", zap.Error(err))
			return
		}
		for _, key := range keys {
			cardID := strings.TrimPrefix(key, prefix)
			scanned++

			status, err := reader.ReadCardStatus(ctx, cardID)
			if err != nil || status == nil {
				continue
			}

			hasBed := false
			isBath := false
			if meta != nil {
				m := meta.GetOrLoad(ctx, cardID)
				hasBed = m.HasBed()
				isBath = m.IsBathroom()
			}
			display := BuildCardDisplay(status, hasBed, isBath)
			if display == nil {
				continue
			}

			if err := writer.WriteCardStatus(ctx, &card.CardStatus{
				CardID:  cardID,
				Display: display,
			}); err != nil {
				logger.Warn("RebuildAllDisplays: write failed",
					zap.String("card", cardID), zap.Error(err))
				continue
			}
			rebuilt++

			// leaf 卡（/88 /96）触发父 /80 unit display 重算。
			if picker != nil {
				picker.RefreshParent(ctx, cardID)
			}
		}
		if next == 0 {
			break
		}
		cursor = next
	}

	logger.Info("RebuildAllDisplays done",
		zap.Int("scanned", scanned),
		zap.Int("rebuilt", rebuilt),
		zap.Duration("elapsed", time.Since(t0)))
}

// RunPeriodicRebuild ctx done 之前每 interval 跑一次 RebuildAllDisplays。
// interval <= 0 时跑一次后退出（等价于 boot-time rebuild）。
func RunPeriodicRebuild(
	ctx context.Context,
	interval time.Duration,
	client *redislib.Client,
	reader *card.Reader,
	writer *card.Writer,
	picker *UnitPicker,
	meta *service.DeviceMetaCache,
	logger *zap.Logger,
) {
	if interval <= 0 {
		RebuildAllDisplays(ctx, client, reader, writer, picker, meta, logger)
		return
	}
	logger.Info("display periodic rebuilder started", zap.Duration("interval", interval))
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Info("display periodic rebuilder stopped")
			return
		case <-ticker.C:
			RebuildAllDisplays(ctx, client, reader, writer, picker, meta, logger)
		}
	}
}
