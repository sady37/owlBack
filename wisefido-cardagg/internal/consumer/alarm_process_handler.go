package consumer

import (
	"context"
	"encoding/json"
	"time"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

// AlarmProcessHandler 报警处理消息处理器（来自 config:alarmProcess:stream）
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

// AlarmProcessMessageData data 层的业务字段
type AlarmProcessMessageData struct {
	TenantID       string `json:"tenant_id"`
	CardID         string `json:"card_id"`
	DeviceID       string `json:"device_id"`
	AlarmLevel     string `json:"alarm_level"`
	AlarmType      string `json:"alarm_type"`
	AlarmTimestamp int64  `json:"alarm_timestamp"` // hand_time，业务逻辑用（当前未比对）
	ProcessType    string `json:"process_type"`
	EventID        string `json:"event_id"`
}

// cloudEventsEnvelope 顶层 CloudEvents 信封，用于解析 data JSON 字符串
type cloudEventsEnvelope struct {
	Data json.RawMessage `json:"data"`
}

// Handle 处理 config:alarmProcess:stream 中的 alarmProcess 消息
// Redis Stream 两层结构：
//   - 顶层: {"data": "<CloudEvents JSON>", "timestamp": <Redis发布时间>}
//   - data JSON: {"specversion":"1.0", ..., "data": {业务字段}}
func (h *AlarmProcessHandler) Handle(ctx context.Context, msg interface{}) error {
	streamMsg, ok := msg.(map[string]interface{})
	if !ok {
		h.logger.Warn("Invalid message type for alarm process", zap.Any("type", msg))
		return nil
	}

	// 1. 提取顶层 timestamp（Redis 发布时间，由 ConvertRedisValues 已转为 int64）
	var topTimestamp int64
	switch v := streamMsg["timestamp"].(type) {
	case int64:
		topTimestamp = v
	case float64:
		topTimestamp = int64(v)
	}

	// 顶层 timestamp 做 staleness 过滤（Redis 发布时间 vs 当前时间）
	if topTimestamp > 0 {
		msgAge := time.Now().Unix() - topTimestamp
		if msgAge > 30 {
			h.logger.Warn("Ignoring stale alarm process message (by top-level timestamp)",
				zap.Int64("top_timestamp", topTimestamp),
				zap.Int64("message_age_seconds", msgAge))
			return nil
		}
	}

	// 2. 提取顶层 data（JSON 字符串），解析 CloudEvents 信封
	dataStr, ok := streamMsg["data"].(string)
	if !ok {
		h.logger.Warn("Invalid data field type in alarm process message", zap.Any("data", streamMsg["data"]))
		return nil
	}

	var envelope cloudEventsEnvelope
	if err := json.Unmarshal([]byte(dataStr), &envelope); err != nil {
		h.logger.Warn("Failed to parse CloudEvents envelope",
			zap.String("data", dataStr), zap.Error(err))
		return nil
	}

	// 3. 解析内层 data → AlarmProcessMessageData
	alarmData := &AlarmProcessMessageData{}
	if err := json.Unmarshal(envelope.Data, alarmData); err != nil {
		h.logger.Warn("Failed to parse alarm process data",
			zap.String("inner_data", string(envelope.Data)), zap.Error(err))
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

	h.logger.Info("Alarm process message received",
		zap.Int64("top_timestamp", topTimestamp),
		zap.String("card_id", alarmData.CardID),
		zap.String("alarm_level", alarmData.AlarmLevel),
		zap.String("alarm_type", alarmData.AlarmType),
		zap.String("process_type", alarmData.ProcessType),
		zap.String("event_id", alarmData.EventID),
		zap.Int64("alarm_timestamp", alarmData.AlarmTimestamp))

	// 调用 service 处理
	return h.service.HandleAlarmProcessMessage(
		ctx,
		alarmData.CardID,
		alarmData.TenantID,
		alarmData.AlarmLevel,
		alarmData.AlarmType,
		alarmData.EventID,
		alarmData.AlarmTimestamp,
	)
}
