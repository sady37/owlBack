// display_rebuilder.go — display 重派单一入口。
//
// 每个 /88 room 的 state 独立存在自己 prefix 的 card:state（projector 按 SubjectEntity 路由，
// 不折叠进 owning card）。room.state 到达时 projector 写完该房、再触发 owning card Rebuild。
// Rebuild(cardID) 遍历 cache.EntryByCard 名下所有 room，按 (占用/risk/卫生间) 选"代表房"
// 写入 status.RoomState → BuildCardDisplay → 写 display。以 cache 名册为准，不靠"谁最后上报"，
// 消除单值 room_state 的 last-writer-wins 横跳。

package consumer

import (
	"context"
	"net/netip"
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
	// 聚合：以 cache 名册遍历 card 名下所有 room，选"代表房"覆盖 status.RoomState
	// （覆盖旧折叠值——各房独立存在自己 prefix，不再 last-writer-wins）。
	if pfx, err := netip.ParsePrefix(cardID); err == nil {
		status.RoomState = r.pickPriorityRoom(ctx, pfx)
	}
	display := BuildCardDisplay(status, r.cache)
	if display == nil {
		return
	}
	if err := r.writer.WriteCardStatus(ctx, &card.CardStatus{
		CardID:    cardID,
		Display:   display,
		RoomState: status.RoomState,
	}); err != nil {
		r.logger.Warn("rebuild display write", zap.String("card", cardID), zap.Error(err))
	}
}

// pickPriorityRoom 遍历 card 名下所有 /88 room（各房 state 存在自己 prefix），按 roomIconFor
// 值序（占用卫生间 > 占用其他房 > 空房）选"代表房"，并列取 total_people 更新更晚的。
// 无 room（纯床卡）返回 nil，display 端 floor 到床。
func (r *DisplayRebuilder) pickPriorityRoom(ctx context.Context, cardPfx netip.Prefix) *card.RoomState {
	var best *card.RoomState
	bestScore := -1
	var bestTs int64
	for _, e := range r.cache.EntryByCard(cardPfx) {
		if !e.IsRoom() {
			continue
		}
		prev, err := r.reader.ReadCardStatus(ctx, e.Prefix.String())
		if err != nil || prev == nil || prev.RoomState == nil {
			continue
		}
		rs := prev.RoomState
		score := roomIconFor(roomIsBathroom(r.cache, rs.RoomID), rs.RiskLevel, rs.TotalPeople)
		if score > bestScore || (score == bestScore && rs.TotalPeopleTs > bestTs) {
			best, bestScore, bestTs = rs, score, rs.TotalPeopleTs
		}
	}
	return best
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
