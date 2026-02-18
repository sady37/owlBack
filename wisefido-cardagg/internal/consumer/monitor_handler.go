package consumer

import (
	"context"
	"encoding/json"
	"time"

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
	iotMsg := &redis.IoTStreamMessage{}

	streamMsg, ok := msg.(map[string]interface{})
	if !ok {
		h.logger.Warn("Invalid message type", zap.Any("type", msg))
		return nil
	}

	jsonBytes, err := json.Marshal(streamMsg)
	if err != nil {
		h.logger.Warn("Failed to marshal stream message", zap.Error(err))
		return nil
	}

	if err := json.Unmarshal(jsonBytes, iotMsg); err != nil {
		h.logger.Warn("Failed to unmarshal iot message", zap.Error(err))
		return nil
	}

	// MQTT 级别的时间过滤：丢弃超过 30 秒的旧消息
	now := time.Now().Unix()
	msgAge := now - iotMsg.Timestamp
	if msgAge > 30 {
		h.logger.Warn("Ignoring stale monitor message",
			zap.String("card_id", iotMsg.CardID),
			zap.String("device_id", iotMsg.DeviceID),
			zap.Int64("message_age_seconds", msgAge),
			zap.Int64("msg_timestamp", iotMsg.Timestamp))
		return nil
	}

	return h.service.ProcessMonitor(ctx, iotMsg)
}
