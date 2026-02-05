package consumer

import (
	"context"
	"encoding/json"

	"owl-common/redis"
	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// EventAlarmHandler event/alarm 消息处理器
type EventAlarmHandler struct {
	service *service.EventAlarmService
	logger  *zap.Logger
}

// NewEventAlarmHandler 创建 event/alarm 处理器
func NewEventAlarmHandler(svc *service.EventAlarmService, logger *zap.Logger) *EventAlarmHandler {
	return &EventAlarmHandler{
		service: svc,
		logger:  logger,
	}
}

// Handle 处理 event/alarm Redis Stream 消息
// 参考 MonitorHandler：使用 JSON 编解码标准处理
func (h *EventAlarmHandler) Handle(ctx context.Context, msg interface{}) error {
	// 从 Redis Stream message 中解析 IoTStreamMessage
	iotMsg := &redis.IoTStreamMessage{}

	// 构建一个新的 map，用于 JSON 编码和解码
	streamMsg, ok := msg.(map[string]interface{})
	if !ok {
		h.logger.Warn("Invalid message type", zap.Any("type", msg))
		return nil
	}

	// JSON 编码和解码，正确处理嵌套结构
	jsonBytes, err := json.Marshal(streamMsg)
	if err != nil {
		h.logger.Warn("Failed to marshal stream message", zap.Error(err))
		return nil
	}

	if err := json.Unmarshal(jsonBytes, iotMsg); err != nil {
		h.logger.Warn("Failed to unmarshal iot message", zap.Error(err))
		return nil
	}

	// 调用 service 处理（时序检查由 service 层统一处理）
	return h.service.HandleMessage(ctx, iotMsg)
}
