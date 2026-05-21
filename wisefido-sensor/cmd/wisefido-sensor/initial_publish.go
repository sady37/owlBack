package main

import (
	"context"
	"database/sql"
	"time"

	"owl-common/card"

	"wisefido-sensor/internal/zoneengine"

	"go.uber.org/zap"
)

// publishInitialResetState 在 sensor 启动后主动 publish 所有 room/bed 的 "no real data" 占位态，
// 解决「sensor 重启后 cardagg card:state hash 保留 stale LastExitTime / StartTime（如 29h ago）」
// 残留问题。
//
// **设计要点（2026-05-20 修订）**：
//   - 所有时间字段（UpdatedAt / StartTime / LastExitTime）= 0，**不**伪造 nowMs 当 anchor
//   - cardagg `BuildCardDisplay` 看 UpdatedAt=0 → `bedHas/roomHas=false` → 不进 display 派生
//   - FE 渲染：active_anchor_ms=0 → "—"；scene_state=0 default → "OOR"；bed_status undefined → "No visitor today"
//   - 旧实现 nowMs anchor 让重启后所有卡同步显示 "Active 3m ago / InBed 3m / LeftBed 3m"（假活跃）
//
// 与 OnZoneEvent 路径互补——OnZoneEvent 只在 zone transition 时发，启动时无 transition；
// 本函数主动一次性发清空，避免重启前的 stale anchor 在 cardagg hash 里持续。
//
// cardagg SensorStateProjector 字段级 merge 规则保证安全：
//   - RoomState category=room.state：整段覆盖（v2 sensor 全 owner，无第三方写者）
//   - BedState  category=bed.state：sensor owner 字段（BedStatus/StartTime/TrackNumber 等）整段覆盖；
//     SleepStage/SleepConfidence 由独立 category=bed.sleepstage 同步清，避免残留
func publishInitialResetState(ctx context.Context, db *sql.DB, p *zoneengine.StreamPublisher, logger *zap.Logger) {
	// 注意：UpdatedAt / StartTime / LastExitTime 全部 0，cardagg builder 当 "no data" 处理。
	// 不再用 time.Now().UnixMilli() 否则所有卡同步显示假活跃。
	_ = time.Now // 保留 import 兼容（PublishBedSleepStage 内部可能用）

	// Rooms：扫所有 rooms（含无监控的，cardagg 端字段级覆盖，无 device 的 card 仍可正常显示 OOR）
	rRows, err := db.QueryContext(ctx, `SELECT room_id::text FROM rooms`)
	if err != nil {
		logger.Warn("initial_reset: query rooms failed", zap.Error(err))
		return
	}
	roomCount := 0
	for rRows.Next() {
		var cardID string
		if err := rRows.Scan(&cardID); err != nil {
			continue
		}
		// Per-field ts 模式（2026-05-20）：所有 ts=0，cardagg merge 端 state-change-anchored
		// 保 prev；下游 maxTs 派生 = 0 → FE 显示"—"无假活跃。
		// RoomType 已挪静态属性（CardMeta/CardStatic.Room.RoomType），此处不发。
		rs := &card.RoomState{
			TotalPeople: 0,
		}
		if err := p.PublishRoomState(ctx, cardID, rs); err != nil {
			logger.Warn("initial_reset: publish room.state failed",
				zap.String("cid", cardID), zap.Error(err))
			continue
		}
		roomCount++
	}
	rRows.Close()

	// Beds：bed.state owner 字段重置 + bed.sleepstage 同步清
	bRows, err := db.QueryContext(ctx, `SELECT bed_id::text FROM beds`)
	if err != nil {
		logger.Warn("initial_reset: query beds failed", zap.Error(err))
		return
	}
	bedCardIDs := make([]string, 0, 16)
	for bRows.Next() {
		var cardID string
		if err := bRows.Scan(&cardID); err != nil {
			continue
		}
		bedCardIDs = append(bedCardIDs, cardID)
	}
	bRows.Close()

	bedCount := 0
	for _, cardID := range bedCardIDs {
		// Per-field ts 模式：所有 ts=0，cardagg merge state-change-anchored 保 prev。
		bs := &card.BedState{
			BedStatus: 1, // 占位 LeftBed (cardagg merge 后真值由后续真实 transition 覆盖)
		}
		if err := p.PublishBedState(ctx, cardID, bs); err != nil {
			logger.Warn("initial_reset: publish bed.state failed",
				zap.String("cid", cardID), zap.Error(err))
			continue
		}
		bedCount++
	}

	// bed.sleepstage 走独立 category 清；projector 字段级 merge 否则 SleepStage 残留 prev
	sleepStageCount := 0
	for _, cardID := range bedCardIDs {
		if err := p.PublishBedSleepStage(ctx, cardID, card.SleepStageInitial, 0); err != nil {
			logger.Warn("initial_reset: publish bed.sleepstage failed",
				zap.String("cid", cardID), zap.Error(err))
			continue
		}
		sleepStageCount++
	}

	logger.Info("sensor startup: initial vacant state published (anchors reset to now)",
		zap.Int("rooms", roomCount),
		zap.Int("beds", bedCount),
		zap.Int("bed_sleepstage", sleepStageCount))
}
