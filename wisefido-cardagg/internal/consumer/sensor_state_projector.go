// sensor_state_projector.go — sensor:derived:stream 消费者。
//
// 把 sensor 派生的 card.BedState / RoomState / TargetState JSON 投影到
// card:state:<cardID> hash 对应字段。cardagg 是 card:state 的唯一 writer
// （CLAUDE.md 规则 #1.3）。
//
// 字段 ownership（sensor 只发自己 owner 的，非 sensor owner 字段在 zero value，cardagg
// 按字段级 merge 保留 prev）：
//
//   BedState 双 category 写入（sensor 全 owner，分 category 写不同字段族）：
//     bed.state       sensor owner: UpdatedAt / BedStatus / BedEvent / StartTime / DurationSec /
//                                   TrackNumber (S5b) / BedConfidence (S5b)
//                                   （zoneengine bed FSM + 10s dedup 后写；projector 保留 SleepStage/SleepConfidence prev）
//     bed.sleepstage  sensor owner: UpdatedAt / SleepStage / SleepConfidence
//                                   （S4 SleepStageConsumer confidence ladder 写；
//                                    projector 保留其他字段 prev）
//     **2026-05-19 owner 全栈订正**：v1 时代标"非 sensor owner"的 4 个字段
//     （SleepStage/SleepConfidence/TrackNumber/BedConfidence）全部归 sensor，分 category 路径写
//     （详 [[cardagg_v1_to_v2_migration_audit]] S4+S5b）
//
//   RoomState  sensor owner: 全部字段（v2 拍板后无第三方写者）
//
//   TargetState 走 TargetMerger 路径：sensor 按 /128 device 发；cardagg 按 owning card
//              max-merge 多 device 写单 hash。Visitor 三字段由 VisitorDeriver 注入 merger。
//
// sensor stream_publisher 发 BedState/RoomState 时 SubjectEntity = /96 bed / /88 room
// 物理实体地址，cardagg 收到后这地址就是 card_id 写 hash。Target 路径 SubjectEntity 是
// /128 device，cardagg 通过 TargetMerger 反查 owning card。

package consumer

import (
	"context"
	"encoding/json"

	"owl-common/card"
	owlredis "owl-common/redis"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

const (
	CategoryBedState      = "bed.state"
	CategoryBedSleepStage = "bed.sleepstage"
	CategoryRoomState     = "room.state"
	CategoryTargetState   = "target.state"
)

type SensorStateProjector struct {
	writer *card.Writer
	reader *card.Reader
	picker *UnitPicker
	merger *service.TargetMerger // nil 时 target.state 分支退化为直写（per-device 不合并）
	logger *zap.Logger
}

func NewSensorStateProjector(writer *card.Writer, reader *card.Reader, picker *UnitPicker, logger *zap.Logger) *SensorStateProjector {
	return &SensorStateProjector{writer: writer, reader: reader, picker: picker, logger: logger}
}

// SetTargetMerger 注入 per-device → owning card max-merge 合并器。未注入时 target.state 直写。
func (p *SensorStateProjector) SetTargetMerger(m *service.TargetMerger) {
	if p == nil {
		return
	}
	p.merger = m
}

func (p *SensorStateProjector) Handle(ctx context.Context, msg *owlredis.IoTStreamMessage) error {
	if msg.SubjectEntity == "" {
		return nil
	}
	traceID := service.BuildParentSpan(msg.Producer, msg.SequenceNumber)
	payload, _ := json.Marshal(owlredis.FirstDataValue(msg.DataValue))

	// destCardID 默认 = msg.SubjectEntity（bed/room state 路径 = 物理实体地址直接对应 card）；
	// target.state 分支由 TargetMerger 反查 owning card 后覆盖。
	destCardID := msg.SubjectEntity
	status := &card.CardStatus{CardID: destCardID}

	switch msg.Category {
	case CategoryBedState:
		var bs card.BedState
		if err := json.Unmarshal(payload, &bs); err != nil {
			p.logger.Warn("bed.state unmarshal", zap.String("trace_id", traceID), zap.String("cid", destCardID), zap.Error(err))
			return nil
		}
		// 字段级 merge：保留 prev 的 SleepStage / SleepConfidence / TrackNumber / BedConfidence
		// （SleepStage/SleepConfidence 走 bed.sleepstage 独立 category；TrackNumber/BedConfidence S5b 决策中）
		prevBed := p.readPrevBedState(ctx, destCardID)
		status.BedState = mergeBedStateSensorOwner(prevBed, &bs)
	case CategoryBedSleepStage:
		var bs card.BedState
		if err := json.Unmarshal(payload, &bs); err != nil {
			p.logger.Warn("bed.sleepstage unmarshal", zap.String("trace_id", traceID), zap.String("cid", destCardID), zap.Error(err))
			return nil
		}
		// 字段级 merge：仅更 SleepStage / SleepConfidence / UpdatedAt，保留 BedStatus 等
		prevBed := p.readPrevBedState(ctx, destCardID)
		status.BedState = mergeBedStateSleepStage(prevBed, &bs)
	case CategoryRoomState:
		var rs card.RoomState
		if err := json.Unmarshal(payload, &rs); err != nil {
			p.logger.Warn("room.state unmarshal", zap.String("trace_id", traceID), zap.String("cid", destCardID), zap.Error(err))
			return nil
		}
		// RoomState v2 拍板后全部字段 sensor owner，整段覆盖即正确
		status.RoomState = &rs
	case CategoryTargetState:
		var ts card.TargetState
		if err := json.Unmarshal(payload, &ts); err != nil {
			p.logger.Warn("target.state unmarshal", zap.String("trace_id", traceID), zap.String("cid", destCardID), zap.Error(err))
			return nil
		}
		// SubjectEntity 是 /128 device addr；通过 TargetMerger 反查 owning card + max-merge。
		// merger 未注入时退化为直写（per-device 卡上单独显示，FE 上不合并）。
		if p.merger != nil {
			ownerCardID, merged := p.merger.OnDeviceTarget(ctx, msg.SubjectEntity, &ts)
			if ownerCardID == "" || merged == nil {
				return nil // 未绑卡 device 直接 drop
			}
			destCardID = ownerCardID
			status.CardID = destCardID
			status.Target = merged
		} else {
			status.Target = &ts
		}
	default:
		return nil
	}

	// 派生 CardDisplay 与 state 一并写入。读 prev 取其他字段（room+bed 共存场景）。
	prev, err := p.reader.ReadCardStatus(ctx, destCardID)
	if err == nil && prev != nil {
		merged := mergeForDisplay(prev, status)
		status.Display = BuildCardDisplay(merged)
	} else {
		status.Display = BuildCardDisplay(status)
	}

	if err := p.writer.WriteCardStatus(ctx, status); err != nil {
		return err
	}
	// 子卡（/88 /96）写完，立刻重算父 /80 unit display；/80 自身写则 no-op（unit-level state 走 RefreshSelf）。
	if p.picker != nil {
		p.picker.RefreshParent(ctx, destCardID)
	}
	return nil
}

// readPrevBedState 读 card:state.bed_state JSON 字段。读失败/不存在返 nil。
func (p *SensorStateProjector) readPrevBedState(ctx context.Context, cardID string) *card.BedState {
	if p.reader == nil {
		return nil
	}
	prev, err := p.reader.ReadCardStatus(ctx, cardID)
	if err != nil || prev == nil {
		return nil
	}
	return prev.BedState
}

// mergeBedStateSensorOwner 字段级合并 bed.state category：sensor zoneengine bed FSM owner 字段
// （UpdatedAt / BedStatus / BedEvent / StartTime / DurationSec / TrackNumber (S5b) /
// BedConfidence (S5b)）从 incoming 取，其他字段保留 prev。
//
// 与 mergeBedStateSleepStage 配套：bed.state 路径不动 SleepStage/SleepConfidence；
// bed.sleepstage 路径不动 BedStatus 等。详 sensor_state_projector.go §"BedState 双 category 写入"。
func mergeBedStateSensorOwner(prev, incoming *card.BedState) *card.BedState {
	if incoming == nil {
		return prev
	}
	out := &card.BedState{}
	if prev != nil {
		*out = *prev
	}
	out.UpdatedAt = incoming.UpdatedAt
	out.BedStatus = incoming.BedStatus
	out.BedEvent = incoming.BedEvent
	out.StartTime = incoming.StartTime
	out.DurationSec = incoming.DurationSec
	out.TrackNumber = incoming.TrackNumber
	out.BedConfidence = incoming.BedConfidence
	return out
}

// mergeBedStateSleepStage 字段级合并 bed.sleepstage category：仅 SleepStage / SleepConfidence
// 从 incoming 取；UpdatedAt 取 max(prev, incoming) 防 sleepstage event 携带较早采样时间倒推
// bed transition anchor。其他字段保留 prev（不动 BedStatus 等 bed.state owner 字段）。
//
// 与 mergeBedStateSensorOwner 配套；分 category 写让两个 publisher 路径不互相覆盖。
func mergeBedStateSleepStage(prev, incoming *card.BedState) *card.BedState {
	if incoming == nil {
		return prev
	}
	out := &card.BedState{}
	if prev != nil {
		*out = *prev
	}
	if incoming.UpdatedAt > out.UpdatedAt {
		out.UpdatedAt = incoming.UpdatedAt
	}
	out.SleepStage = incoming.SleepStage
	out.SleepConfidence = incoming.SleepConfidence
	return out
}

// mergeForDisplay 合并 prev 已存字段 + 本次 update，避免本次只更 bed 时 display 丢 room 数据。
func mergeForDisplay(prev, cur *card.CardStatus) *card.CardStatus {
	out := &card.CardStatus{CardID: cur.CardID}
	out.RoomState = prev.RoomState
	out.BedState = prev.BedState
	out.AlarmState = prev.AlarmState
	out.Target = prev.Target
	if cur.RoomState != nil {
		out.RoomState = cur.RoomState
	}
	if cur.BedState != nil {
		out.BedState = cur.BedState
	}
	if cur.Target != nil {
		out.Target = cur.Target
	}
	return out
}
