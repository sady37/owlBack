package consumer

import (
	"context"
	"encoding/json"

	"owl-common/redis"
	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// MonitorHandler monitor 消息处理器
type MonitorHandler struct {
	service *service.MonitorService
	logger  *zap.Logger
}

// NewMonitorHandler 创建 monitor 处理器
func NewMonitorHandler(svc *service.MonitorService, logger *zap.Logger) *MonitorHandler {
	return &MonitorHandler{
		service: svc,
		logger:  logger,
	}
}

// Handle 处理 monitor 消息
func (h *MonitorHandler) Handle(ctx context.Context, msg interface{}) error {
	// 从 Redis Stream message 中解析 IoTStreamMessage
	iotMsg := &redis.IoTStreamMessage{}

	// 构建一个新的 map，用于 JSON 编码和解码
	streamMsg, ok := msg.(map[string]interface{})
	if !ok {
		h.logger.Warn("Invalid message type", zap.Any("type", msg))
		return nil
	}

	// JSON 编码和解码，这样可以正确处理嵌套结构
	jsonBytes, err := json.Marshal(streamMsg)
	if err != nil {
		h.logger.Warn("Failed to marshal stream message", zap.Error(err))
		return nil
	}

	if err := json.Unmarshal(jsonBytes, iotMsg); err != nil {
		h.logger.Warn("Failed to unmarshal iot message", zap.Error(err))
		return nil
	}

	// 调用 service 处理
	return h.service.ProcessMonitor(ctx, iotMsg)
}
