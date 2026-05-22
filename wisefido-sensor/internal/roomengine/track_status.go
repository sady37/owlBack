// track_status.go — sensor v2 Layer 1 → Layer 2 契约（TrackStatus）+ 流发布
//
// sensor_v2 §10.1.1 + §10.2.2 / PR-3a。
//
// 契约语义：
//   - 每帧每 track 一条 TrackStatus，published 到 sensor:track:status:stream（30s TTL）
//   - 瞬态投影，不入库；CLAUDE.md 规则 #2.1
//   - Verdict 用字符串（"pending"/"real"/"anchored"/"ghost"）；不暴露 int
//   - PersonID 仅当 SuiteCensusManager 已升格人时填，candidate 阶段一律空（防污染下游）
//   - **PersonID 短暂消失是正常状态**：track 失锁 + ClearAnchorOnLostTrack 清空 AnchorTrackID 后，
//     即使同 firmware trackID 重新激活，SuiteCensus 也会走 candidate 路径重新积累 anchor，
//     这段窗口 (≥ 60s 失锁阈值 + visitor 2min 锚定时长) 内 PersonID 会从 "non-empty → '' → non-empty" 跳变。
//     下游（zoneengine v2 / fall verifier）**不得据此触发 lost 类告警** — person 生命周期的真实状态
//     以 SuiteCensus.SuitePerson 为权威源（看 LastActiveMs），TrackStatus.PersonID 仅是本帧投影。
//
// 不在 PR-3 范围（占位字段，留给后续 PR 填写）：
//   - InRoomZoneID / InBedZoneID / InBathroomZoneID 当前由 cell.AreaType 近似派生；
//     真正 zoneengine v2 (Phase B) 接入后会替换为 ZoneEvent 权威 ZoneID。
//   - GhostPenalty 从 TrackState.GhostPenalty 直接镜像（v1 累积器，v2 PR-6 重写后语义会变）。

package roomengine

import (
	"context"
	"encoding/json"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// TrackStatus Layer 1 → Layer 2 契约。
// 详 sensor_v2 §10.1.1。
type TrackStatus struct {
	TrackID  int    `json:"track_id"`
	DeviceID string `json:"device_addr"` // radar IPv6 /128
	RoomID   string `json:"room_id"`   // room spatial_prefix

	Verdict      TrackVerdict `json:"-"`             // 内部 int，wire 用 verdict 字段（字符串）
	GhostPenalty int          `json:"ghost_penalty"` // 0-100；≥80 → Ghost

	X, Y, Z      int      `json:"-"` // 单独平铺到 wire（x/y/z）
	Pose         int      `json:"pose"`
	StillSec     int      `json:"still_sec"`
	CellAreaType AreaType `json:"-"` // 单独平铺到 wire（cell_area_type 字符串）
	EnterTarget  string   `json:"enter_target"`

	InBedZoneID      string `json:"in_bed_zone_id,omitempty"`
	InRoomZoneID     string `json:"in_room_zone_id,omitempty"`
	InBathroomZoneID string `json:"in_bathroom_zone_id,omitempty"`

	PersonID   string `json:"person_id,omitempty"`
	PersonRole string `json:"person_role,omitempty"`

	UpdatedAtMs int64 `json:"updated_at_ms"`
}

// areaTypeWireName cell area type 序列化字符串（v2 sensor 内部完整字典）。
//
// **与 areaTypeProtocolStr (engine.go) 是不同的序列化函数，对应不同的下游消费方**：
//
//   - areaTypeWireName    → 用于 sensor:track:status:stream（本流，PR-3 引入）
//     完整 v2 内部语义：toilet / shower / sit 等独立映射。
//     消费方：dev playback (track-status-tail) / 未来 zoneengine v2 / fall verifier。
//
//   - areaTypeProtocolStr → 用于 iot:event:stream / iot:alarm:stream（cardagg 消费）
//     向 owl-common/observation.EnumAreaType 投影（best-effort，5 类协议级 area）。
//     消费方：cardagg AI verdict 路由 / alarm 落库。
//
// 设计意图：v2 内部表达力（高维）和 cardagg 协议向后兼容（低维）解耦——
// 改本字典不会牵动 cardagg，反之亦然。新增消费方时按"哪条流"决定用哪套。
func areaTypeWireName(t AreaType) string {
	switch t {
	case AreaEnter:
		return "enter"
	case AreaBed:
		return "bed"
	case AreaSit:
		return "sit"
	case AreaActive:
		return "active"
	case AreaDeny:
		return "deny"
	case AreaShower:
		return "shower"
	case AreaToilet:
		return "toilet"
	}
	return "unknown"
}

// ToStreamMap 序列化为 Redis stream values。
// wire schema 稳定字段：
//   verdict / ghost_penalty / track_id / device_id / room_id /
//   x / y / z / pose / still_sec / cell_area_type / enter_target /
//   in_bed_zone_id / in_room_zone_id / in_bathroom_zone_id /
//   person_id / person_role / updated_at_ms
//
// 数值字段保持 int 原型；ZoneID/PersonID 为 ""（零值）时省略以降低 stream payload 体积。
func (ts *TrackStatus) ToStreamMap() map[string]interface{} {
	m := map[string]interface{}{
		"verdict":        VerdictName(ts.Verdict),
		"ghost_penalty":  ts.GhostPenalty,
		"track_id":       ts.TrackID,
		"device_id":      ts.DeviceID,
		"room_id":        ts.RoomID,
		"x":              ts.X,
		"y":              ts.Y,
		"z":              ts.Z,
		"pose":           ts.Pose,
		"still_sec":      ts.StillSec,
		"cell_area_type": areaTypeWireName(ts.CellAreaType),
		"enter_target":   ts.EnterTarget,
		"updated_at_ms":  ts.UpdatedAtMs,
	}
	if ts.InBedZoneID != "" {
		m["in_bed_zone_id"] = ts.InBedZoneID
	}
	if ts.InRoomZoneID != "" {
		m["in_room_zone_id"] = ts.InRoomZoneID
	}
	if ts.InBathroomZoneID != "" {
		m["in_bathroom_zone_id"] = ts.InBathroomZoneID
	}
	if ts.PersonID != "" {
		m["person_id"] = ts.PersonID
	}
	if ts.PersonRole != "" {
		m["person_role"] = ts.PersonRole
	}
	return m
}

// PublishTrackStatus 推送单条 TrackStatus 到 sensor:track:status:stream。
// 失败仅 warn，不阻塞调用方。redisClient==nil（测试/playback 模式）直接跳过。
func PublishTrackStatus(ctx context.Context, client *redis.Client, logger *zap.Logger, ts *TrackStatus) {
	if client == nil || ts == nil {
		return
	}
	def := rediscommon.StreamSensorTrackStatus
	if _, err := rediscommon.PublishToStream(ctx, client, def.Name, ts.ToStreamMap(), def.MaxLen, def.RetentionSeconds); err != nil {
		if logger != nil {
			logger.Warn("track_status_publish_failed",
				zap.String("stream", def.Name),
				zap.Int("track_id", ts.TrackID),
				zap.String("device_id", ts.DeviceID),
				zap.String("room_id", ts.RoomID),
				zap.Error(err),
			)
		}
	}
}

// MarshalJSON 自定义编码：把内部 TrackVerdict / 坐标 / AreaType 转成 wire schema 的字段名。
// 便于离线日志 / 单测 / dev-tool 用 JSON 直接读 stream payload。
func (ts *TrackStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(ts.ToStreamMap())
}
