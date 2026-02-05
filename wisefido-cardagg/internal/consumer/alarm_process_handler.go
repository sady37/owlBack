package consumer

import (
	"context"
	"encoding/json"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// AlarmProcessHandler 报警处理消息处理器（来自 config:alarm.process:stream）
type AlarmProcessHandler struct {
	service *service.EventAlarmService
	logger  *zap.Logger
}

// NewAlarmProcessHandler 创建报警处理消息处理器
func NewAlarmProcessHandler(svc *service.EventAlarmService, logger *zap.Logger) *AlarmProcessHandler {
	return &AlarmProcessHandler{
		service: svc,
		logger:  logger,
	}
}

// AlarmProcessMessageData 报警处理消息数据结构
type AlarmProcessMessageData struct {
	TenantID       string `json:"tenant_id"`
	CardID         string `json:"card_id"`
	DeviceID       string `json:"device_id"`
	AlarmLevel     string `json:"alarm_level"`
	AlarmType      string `json:"alarm_type"`
	AlarmTimestamp int64  `json:"alarm_timestamp"`
}

// Handle 处理 config:alarm.process:stream 中的 alarmProcess 消息
// 消息格式与 EventAlarmHandler 相同，通过 Map 解析
func (h *AlarmProcessHandler) Handle(ctx context.Context, msg interface{}) error {
	// 从 Redis Stream message 中解析消息数据
	streamMsg, ok := msg.(map[string]interface{})
	if !ok {
		h.logger.Warn("Invalid message type for alarm process", zap.Any("type", msg))
		return nil
	}

	// 提取 data 字段
	dataField, ok := streamMsg["data"].(map[string]interface{})
	if !ok {
		h.logger.Warn("Invalid data field in alarm process message", zap.Any("data", streamMsg["data"]))
		return nil
	}

	// 解析数据
	alarmData := &AlarmProcessMessageData{}
	jsonBytes, err := json.Marshal(dataField)
	if err != nil {
		h.logger.Warn("Failed to marshal alarm process data", zap.Error(err))
		return nil
	}

	if err := json.Unmarshal(jsonBytes, alarmData); err != nil {
		h.logger.Warn("Failed to unmarshal alarm process data",
			zap.String("data", string(jsonBytes)),
			zap.Error(err))
		return nil
	}

	// 验证必填字段
	if alarmData.CardID == "" {
		h.logger.Warn("Missing card_id in alarm process message")
		return nil
	}

	if alarmData.AlarmLevel == "" {
		h.logger.Warn("Missing alarm_level in alarm process message")
		return nil
	}

	// 调用 service 处理
	return h.service.HandleAlarmProcessMessage(
		ctx,
		alarmData.CardID,
		alarmData.AlarmLevel,
		alarmData.AlarmType,
		alarmData.AlarmTimestamp,
	)
}
