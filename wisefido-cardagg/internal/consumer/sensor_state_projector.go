// sensor_state_projector.go — sensor:derived:stream 消费者。
//
// 把 sensor 派生的 card.BedState / RoomState / TargetState JSON 投影到
// card:state:<cardID> hash 对应字段。cardagg 是 card:state 的唯一 writer
// （CLAUDE.md 规则 #1.3）。
//
// sensor stream_publisher 写到 e.ZoneID (/88 room / /96 bed)；no /80 unit aggregation here —
// unit-scope risk (NightAbsence of Room 等) 由 sensor 内部 counter 处理，不入 card:state。

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
	CategoryBedState    = "bed.state"
	CategoryRoomState   = "room.state"
	CategoryTargetState = "target.state"
)

type SensorStateProjector struct {
	writer *card.Writer
	reader *card.Reader
	picker *UnitPicker
	logger *zap.Logger
}

func NewSensorStateProjector(writer *card.Writer, reader *card.Reader, picker *UnitPicker, logger *zap.Logger) *SensorStateProjector {
	return &SensorStateProjector{writer: writer, reader: reader, picker: picker, logger: logger}
}

func (p *SensorStateProjector) Handle(ctx context.Context, msg *owlredis.IoTStreamMessage) error {
	if msg.SubjectEntity == "" {
		return nil
	}
	traceID := service.BuildParentSpan(msg.Producer, msg.SequenceNumber)
	payload, _ := json.Marshal(owlredis.FirstDataValue(msg.DataValue))
	status := &card.CardStatus{CardID: msg.SubjectEntity}

	switch msg.Category {
	case CategoryBedState:
		var bs card.BedState
		if err := json.Unmarshal(payload, &bs); err != nil {
			p.logger.Warn("bed.state unmarshal", zap.String("trace_id", traceID), zap.String("cid", msg.SubjectEntity), zap.Error(err))
			return nil
		}
		status.BedState = &bs
	case CategoryRoomState:
		var rs card.RoomState
		if err := json.Unmarshal(payload, &rs); err != nil {
			p.logger.Warn("room.state unmarshal", zap.String("trace_id", traceID), zap.String("cid", msg.SubjectEntity), zap.Error(err))
			return nil
		}
		status.RoomState = &rs
	case CategoryTargetState:
		var ts card.TargetState
		if err := json.Unmarshal(payload, &ts); err != nil {
			p.logger.Warn("target.state unmarshal", zap.String("trace_id", traceID), zap.String("cid", msg.SubjectEntity), zap.Error(err))
			return nil
		}
		status.Target = &ts
	default:
		return nil
	}

	// 派生 CardDisplay 与 state 一并写入。读 prev 取其他字段（room+bed 共存场景）。
	prev, err := p.reader.ReadCardStatus(ctx, msg.SubjectEntity)
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
		p.picker.RefreshParent(ctx, msg.SubjectEntity)
	}
	return nil
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
