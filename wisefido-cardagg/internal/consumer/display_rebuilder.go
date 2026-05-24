// display_rebuilder.go — display 重派单一入口。
//
// 各层各算，无 winner proxy，无 trigger parent。sensor event 已经按 entry 精确路由到
// 唯一 card，caller 写完 own state 后调 `Rebuild(cardID)`：读 fresh hash + 卡级 snapshot →
// BuildCardDisplay → 写 display field。每张 card 独立派生，互不影响。

package consumer

import (
	"context"
	"time"

	"go.uber.org/zap"

	"owl-common/card"
	"wisefido-cardagg/internal/service"
)

type DisplayRebuilder struct {
	cache  *service.SpatialCache
	reader *card.Reader
	writer *card.Writer
	logger *zap.Logger
}

func NewDisplayRebuilder(cache *service.SpatialCache, reader *card.Reader, writer *card.Writer, logger *zap.Logger) *DisplayRebuilder {
	return &DisplayRebuilder{cache: cache, reader: reader, writer: writer, logger: logger}
}

// Rebuild 派生指定 card 的 display 并写 hash。
// caller 写完 own state 后调一次即可——内部读 fresh hash 自然拿到合并后的全状态。
func (r *DisplayRebuilder) Rebuild(ctx context.Context, cardID string) {
	if cardID == "" {
		return
	}
	status, _ := r.reader.ReadCardStatus(ctx, cardID)
	if status == nil {
		status = &card.CardStatus{CardID: cardID}
	}
	display := BuildCardDisplay(status, r.cache)
	if display == nil {
		return
	}
	if err := r.writer.WriteCardStatus(ctx, &card.CardStatus{
		CardID:  cardID,
		Display: display,
	}); err != nil {
		r.logger.Warn("rebuild display write", zap.String("card", cardID), zap.Error(err))
	}
}

// RebuildAll 遍历 cards 表（cache 内）逐张 Rebuild。
// 适用：cardagg 启动 boot-time 一次；config:card:stream op=reset；periodic safety net。
func (r *DisplayRebuilder) RebuildAll(ctx context.Context) {
	t0 := time.Now()
	cardIDs := r.cache.AllCardIDs()
	for _, cid := range cardIDs {
		r.Rebuild(ctx, cid.String())
	}
	r.logger.Info("RebuildAll done",
		zap.Int("rebuilt", len(cardIDs)),
		zap.Duration("elapsed", time.Since(t0)))
}

// RunPeriodic ctx done 之前每 interval 跑一次 RebuildAll。
// interval <= 0 时跑一次后退出（等价 boot-time）。
func (r *DisplayRebuilder) RunPeriodic(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		r.RebuildAll(ctx)
		return
	}
	r.logger.Info("display periodic rebuilder started", zap.Duration("interval", interval))
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			r.logger.Info("display periodic rebuilder stopped")
			return
		case <-ticker.C:
			r.RebuildAll(ctx)
		}
	}
}
