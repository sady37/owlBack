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
	"net/netip"

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
	writer    *card.Writer
	reader    *card.Reader
	rebuilder *DisplayRebuilder
	merger    *service.TargetMerger // nil 时 target.state 分支退化为直写（per-device 不合并）
	cache     *service.SpatialCache
	logger    *zap.Logger
}

func NewSensorStateProjector(writer *card.Writer, reader *card.Reader, rebuilder *DisplayRebuilder, cache *service.SpatialCache, logger *zap.Logger) *SensorStateProjector {
	return &SensorStateProjector{writer: writer, reader: reader, rebuilder: rebuilder, cache: cache, logger: logger}
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
	case CategoryBedState, CategoryBedSleepStage:
		// bed.state / bed.sleepstage：sensor 发 /96 bed prefix；通过 entry 表精确等值反查 owning card_id。
		// entry 未注册（schema 缺对应 bed 行）→ 丢弃，不做 LPM 父卡兜底。
		if pfx, err := netip.ParsePrefix(msg.SubjectEntity); err == nil {
			if ent := p.cache.LookupEntry(pfx); ent != nil {
				destCardID = ent.CardID.String()
				status.CardID = destCardID
			} else {
				p.logger.Warn("bed.state entry not in cache", zap.String("subject", msg.SubjectEntity))
				return nil
			}
		}
		var bs card.BedState
		if err := json.Unmarshal(payload, &bs); err != nil {
			p.logger.Warn("bed state unmarshal", zap.String("category", msg.Category), zap.String("trace_id", traceID), zap.String("cid", destCardID), zap.Error(err))
			return nil
		}
		prevBed := p.readPrevBedState(ctx, destCardID)
		if msg.Category == CategoryBedState {
			// 字段级 merge：保留 prev 的 SleepStage / SleepConfidence / TrackNumber / BedConfidence
			status.BedState = mergeBedStateSensorOwner(prevBed, &bs)
		} else {
			// 字段级 merge：仅更 SleepStage / SleepConfidence / UpdatedAt，保留 BedStatus 等
			status.BedState = mergeBedStateSleepStage(prevBed, &bs)
		}
	case CategoryRoomState:
		// room.state：sensor 发 /88 room prefix；entry 表精确等值反查 owning card_id。
		if pfx, err := netip.ParsePrefix(msg.SubjectEntity); err == nil {
			if ent := p.cache.LookupEntry(pfx); ent != nil {
				destCardID = ent.CardID.String()
				status.CardID = destCardID
			} else {
				p.logger.Warn("room.state entry not in cache", zap.String("subject", msg.SubjectEntity))
				return nil
			}
		}
		var rs card.RoomState
		if err := json.Unmarshal(payload, &rs); err != nil {
			p.logger.Warn("room.state unmarshal", zap.String("trace_id", traceID), zap.String("cid", destCardID), zap.Error(err))
			return nil
		}
		// RoomState 字段级 merge（state-change-anchored）：值不变 ts 保 prev。
		prevRoom := p.readPrevRoomState(ctx, destCardID)
		status.RoomState = mergeRoomState(prevRoom, &rs)
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

	// 写 own state，display 派生交给 DisplayRebuilder（读 fresh hash 自然拿到合并状态）。
	if err := p.writer.WriteCardStatus(ctx, status); err != nil {
		return err
	}
	if p.rebuilder != nil {
		p.rebuilder.Rebuild(ctx, destCardID)
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

// mergeBedField state-change-anchored 字段级合并：值不变 → 保 prev value+ts；值变 → 用 incoming。
//
// 例：prev={BedStatus:0, Ts:T1}, incoming={BedStatus:0, Ts:T2(later)} → 返回 (0, T1) 保不变。
//
// 边界：incoming.Ts == 0 通常表示 placeholder（startup_initial_publish），跳过更新保 prev。
func mergeBedField(prevVal int, prevTs int64, inVal int, inTs int64) (int, int64) {
	if inTs == 0 {
		// incoming 没有 ts（placeholder / 占位），不动 prev
		return prevVal, prevTs
	}
	if prevVal == inVal {
		// 值未变：保留 prev ts（state-change-anchored）
		return prevVal, prevTs
	}
	return inVal, inTs
}

// mergeBedStateSensorOwner 字段级合并 bed.state category：sensor zoneengine bed FSM owner 字段
// （BedStatus / BedEvent / TrackNumber / BedConfidence）走 state-change-anchored merge。
// SleepStage/SleepConfidence 保留 prev（由 bed.sleepstage 独立 path 更新）。
//
// BedEvent 是事件不是状态，每次 incoming.BedEventTs > 0 都更新（事件类没有 "state-change-anchored"）。
func mergeBedStateSensorOwner(prev, incoming *card.BedState) *card.BedState {
	if incoming == nil {
		return prev
	}
	out := &card.BedState{}
	if prev != nil {
		*out = *prev
	}
	out.BedStatus, out.BedStatusTs = mergeBedField(out.BedStatus, out.BedStatusTs, incoming.BedStatus, incoming.BedStatusTs)
	out.TrackNumber, out.TrackNumberTs = mergeBedField(out.TrackNumber, out.TrackNumberTs, incoming.TrackNumber, incoming.TrackNumberTs)
	out.BedConfidence, out.BedConfidenceTs = mergeBedField(out.BedConfidence, out.BedConfidenceTs, incoming.BedConfidence, incoming.BedConfidenceTs)
	// BedEvent: 事件类，每次都更新（除非 incoming.BedEventTs=0）
	if incoming.BedEventTs > 0 {
		out.BedEvent = incoming.BedEvent
		out.BedEventTs = incoming.BedEventTs
	}
	return out
}

// mergeBedStateSleepStage 字段级合并 bed.sleepstage category：仅 SleepStage / SleepConfidence
// 走 state-change-anchored merge。其他 BedState owner 字段保留 prev（不动 BedStatus 等）。
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
	out.SleepStage, out.SleepStageTs = mergeBedField(out.SleepStage, out.SleepStageTs, incoming.SleepStage, incoming.SleepStageTs)
	out.SleepConfidence, out.SleepConfidenceTs = mergeBedField(out.SleepConfidence, out.SleepConfidenceTs, incoming.SleepConfidence, incoming.SleepConfidenceTs)
	return out
}

// mergeRoomState 字段级合并 room.state：所有字段走 state-change-anchored merge。
//
// **AloneSinceTs 特殊规则**（基于 TotalPeople 的 cross-field 派生）：
//   - new.TotalPeople != 1：AloneSinceTs = 0（独居语义失效）
//   - new.TotalPeople == 1 且 prev.TotalPeople != 1：AloneSinceTs = incoming（新独居开始）
//   - new.TotalPeople == 1 且 prev.TotalPeople == 1：保 prev.AloneSinceTs（独居延续）
//
// **LastEnterTs / LastExitTs 也是 state-anchored**：
//   - LastEnterTs 仅 prev.TotalPeople==0 && new.TotalPeople>0 时刷新（0→N+ transition）
//   - LastExitTs  仅 prev.TotalPeople>0 && new.TotalPeople==0 时刷新（N→0 transition）
//   - 其它情况保 prev（避免 visitor 来回时反复刷 anchor）
func mergeRoomState(prev, incoming *card.RoomState) *card.RoomState {
	if incoming == nil {
		return prev
	}
	out := &card.RoomState{}
	if prev != nil {
		*out = *prev
	}
	// RoomID 跟 latest event（incoming.subject_entity 标定的 room）；display 派生靠它决定 room_icon_kind。
	if incoming.RoomID != "" {
		out.RoomID = incoming.RoomID
	}

	prevPeople := 0
	if prev != nil {
		prevPeople = prev.TotalPeople
	}
	newPeople := incoming.TotalPeople

	// 普通 state-change-anchored 字段
	out.TotalPeople, out.TotalPeopleTs = mergeBedField(out.TotalPeople, out.TotalPeopleTs, incoming.TotalPeople, incoming.TotalPeopleTs)
	out.RiskLevel, out.RiskLevelTs = mergeBedField(out.RiskLevel, out.RiskLevelTs, incoming.RiskLevel, incoming.RiskLevelTs)

	// LastEnterTs: 仅在 0→N+ transition 时刷
	if prevPeople == 0 && newPeople > 0 && incoming.LastEnterTs > 0 {
		out.LastEnterTs = incoming.LastEnterTs
	}
	// LastExitTs: 仅在 N→0 transition 时刷
	if prevPeople > 0 && newPeople == 0 && incoming.LastExitTs > 0 {
		out.LastExitTs = incoming.LastExitTs
	}

	// AloneSinceTs: cross-field 派生（基于 TotalPeople）
	switch {
	case newPeople != 1:
		out.AloneSinceTs = 0 // 不独居，clear anchor
	case prevPeople != 1 && incoming.AloneSinceTs > 0:
		// 新进入独居（任意值 → 1）
		out.AloneSinceTs = incoming.AloneSinceTs
	}
	// case prevPeople==1 && newPeople==1: 已在 *out = *prev 保留 prev.AloneSinceTs

	// LastExitToOutside 是 bool，无 ts；incoming 总是覆盖
	out.LastExitToOutside = incoming.LastExitToOutside

	return out
}

// readPrevRoomState 读 card:state.room_state JSON。
func (p *SensorStateProjector) readPrevRoomState(ctx context.Context, cardID string) *card.RoomState {
	if p.reader == nil {
		return nil
	}
	prev, err := p.reader.ReadCardStatus(ctx, cardID)
	if err != nil || prev == nil {
		return nil
	}
	return prev.RoomState
}

