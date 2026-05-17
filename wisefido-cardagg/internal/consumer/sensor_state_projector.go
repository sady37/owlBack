// sensor_state_projector.go — sensor:derived:stream 消费者。
//
// 把 sensor 派生的 card.BedState / RoomState / BathRoomState / TargetState JSON 投影到
// card:status:<cardID> hash 对应字段。cardagg 是 card:status 的唯一 writer
// （CLAUDE.md 规则 #1.3）。

package consumer

import (
	"context"
	"encoding/json"

	"owl-common/card"
	owlredis "owl-common/redis"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// sensor stream_publisher 写的 category 取值
const (
	CategoryBedState      = "bed.state"
	CategoryRoomState     = "room.state"
	CategoryBathroomState = "bathroom.state"
	CategoryTargetState   = "target.state"
)

type SensorStateProjector struct {
	writer *card.Writer
	logger *zap.Logger
}

func NewSensorStateProjector(writer *card.Writer, logger *zap.Logger) *SensorStateProjector {
	return &SensorStateProjector{writer: writer, logger: logger}
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
	case CategoryBathroomState:
		var br card.BathRoomState
		if err := json.Unmarshal(payload, &br); err != nil {
			p.logger.Warn("bathroom.state unmarshal", zap.String("trace_id", traceID), zap.String("cid", msg.SubjectEntity), zap.Error(err))
			return nil
		}
		status.BathRoomState = &br
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

	return p.writer.WriteCardStatus(ctx, status)
}
