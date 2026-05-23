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
	"database/sql"
	"time"

	"go.uber.org/zap"

	"owl-common/card"
	"wisefido-cardagg/internal/service"
)

// RebuildAllFromDB 从 PG cards 表枚举所有卡，逐张重 build display 写回 Redis。
//
// 适用场景：
//   - cardagg 启动 boot-time 一次（main.go）— Redis 可能有上一代 schema 的 stale display
//   - 收到 op=reset 信号（data 重启了，in-memory cache 全失效）
//   - periodic safety net（5min 一次）— 兜底事件丢失
//
// DB 是真相源（vs 旧 RebuildAllDisplays 走 Redis SCAN）— 启动场景 Redis 可能为空 / stale，
// 必须以 DB 为准。
func RebuildAllFromDB(
	ctx context.Context,
	db *sql.DB,
	reader *card.Reader,
	writer *card.Writer,
	picker *UnitPicker,
	meta *service.DeviceMetaCache,
	logger *zap.Logger,
) {
	t0 := time.Now()
	rows, err := db.QueryContext(ctx, `SELECT host(card_id)||'/'||masklen(card_id), masklen(card_id) FROM cards`)
	if err != nil {
		logger.Warn("RebuildAllFromDB: query cards", zap.Error(err))
		return
	}
	defer rows.Close()

	rebuilt := 0
	for rows.Next() {
		var cardID string
		var mask int
		if err := rows.Scan(&cardID, &mask); err != nil {
			continue
		}
		// 读现有 Redis state（可能 nil）作为 RoomState/BedState/Target 输入；BuildCardDisplay
		// 容忍空 status——pickCardPriority 仍可基于 CardMeta.hasBed 决定 BedInUse=2。
		status, _ := reader.ReadCardStatus(ctx, cardID)
		if status == nil {
			status = &card.CardStatus{CardID: cardID}
		}
		// /80 走 picker 路径（unit display 合成）；/88 /96 走 BuildCardDisplay 单卡视角。
		if mask == 80 {
			if picker != nil {
				picker.RefreshSelf(ctx, cardID)
			}
			rebuilt++
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
			logger.Warn("RebuildAllFromDB: write", zap.String("card", cardID), zap.Error(err))
			continue
		}
		if picker != nil {
			picker.RefreshParent(ctx, cardID)
		}
		rebuilt++
	}
	logger.Info("RebuildAllFromDB done",
		zap.Int("rebuilt", rebuilt),
		zap.Duration("elapsed", time.Since(t0)))
}

// RunPeriodicRebuild ctx done 之前每 interval 跑一次 RebuildAllFromDB。
// interval <= 0 时跑一次后退出（等价于 boot-time rebuild）。
func RunPeriodicRebuild(
	ctx context.Context,
	interval time.Duration,
	db *sql.DB,
	reader *card.Reader,
	writer *card.Writer,
	picker *UnitPicker,
	meta *service.DeviceMetaCache,
	logger *zap.Logger,
) {
	if interval <= 0 {
		RebuildAllFromDB(ctx, db, reader, writer, picker, meta, logger)
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
			RebuildAllFromDB(ctx, db, reader, writer, picker, meta, logger)
		}
	}
}
