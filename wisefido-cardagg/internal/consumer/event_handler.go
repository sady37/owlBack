// event_handler.go — iot:event:stream 消费者（cardagg 视角）。
//
// 当前消费 category（其他 sensor 内部消费由 roomengine 直接订阅）：
//   - number_people  → BedPeopleTracker.Update（doc/card_display.md §4.4 visitor bed 路径）
//   - EnterRoom/ExitRoom → AIOverrideCache.ClearDevice
//     （新人进/房间空 → track_id 槽位会复用，旧 verdict 必须作废避免错挂；
//      [[cardagg_v1_to_v2_migration_audit]] S5a 清理触发 b/c）
//
// 不入业务库 / 不持久化：iot 模块已负责 event_log 写入；cardagg 此 handler 仅副作用更新 in-memory tracker。
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
	bedPeople   *service.BedPeopleTracker
	aiOverrides *service.AIOverrideCache // 可 nil
	logger      *zap.Logger
}

func NewEventHandler(bedPeople *service.BedPeopleTracker, logger *zap.Logger) *EventHandler {
	return &EventHandler{bedPeople: bedPeople, logger: logger}
}

// SetAIOverrides main wiring 注入；nil 时 EnterRoom/ExitRoom 不触发清理。
func (h *EventHandler) SetAIOverrides(c *service.AIOverrideCache) {
	if h == nil {
		return
	}
	h.aiOverrides = c
}

func (h *EventHandler) Handle(ctx context.Context, msg *owlredis.IoTStreamMessage) error {
	if msg == nil || !msg.DeviceAddr.IsValid() {
		return nil
	}
	switch msg.Category {
	case alarm.NumberPeople:
		fields := owlredis.FirstDataValue(msg.DataValue)
		if fields == nil {
			return nil
		}
		count := intFromAny(fields[observation.FieldNumberPeople])
		if count < 0 {
			return nil
		}
		h.bedPeople.Update(msg.DeviceAddr.String(), count, msg.Timestamp)
	case alarm.EnterRoom, alarm.ExitRoom:
		if h.aiOverrides != nil {
			h.aiOverrides.ClearDevice(msg.DeviceAddr.String())
		}
	}
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
