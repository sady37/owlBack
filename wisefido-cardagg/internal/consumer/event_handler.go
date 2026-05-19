// event_handler.go — iot:event:stream 消费者（cardagg 视角）。
//
// 当前唯一用途：filter category=number_people 把 firmware type=3 人数事件喂给
// BedPeopleTracker，给 VisitorDeriver bed-level 路径用（doc/card_display.md §4.4）。
//
// 不入业务库 / 不持久化：iot 模块已负责 event_log 写入；cardagg 此 handler 仅副作用更新
// in-memory tracker。
package consumer

import (
	"context"

	"owl-common/alarm"
	"owl-common/observation"
	owlredis "owl-common/redis"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

type EventHandler struct {
	bedPeople *service.BedPeopleTracker
	logger    *zap.Logger
}

func NewEventHandler(bedPeople *service.BedPeopleTracker, logger *zap.Logger) *EventHandler {
	return &EventHandler{bedPeople: bedPeople, logger: logger}
}

func (h *EventHandler) Handle(ctx context.Context, msg *owlredis.IoTStreamMessage) error {
	if msg == nil || !msg.DeviceAddr.IsValid() {
		return nil
	}
	if msg.Category != alarm.NumberPeople {
		return nil
	}
	fields := owlredis.FirstDataValue(msg.DataValue)
	if fields == nil {
		return nil
	}
	count := intFromAny(fields[observation.FieldNumberPeople])
	if count < 0 {
		return nil
	}
	h.bedPeople.Update(msg.DeviceAddr.String(), count, msg.Timestamp)
	return nil
}

func intFromAny(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	default:
		return -1
	}
}
