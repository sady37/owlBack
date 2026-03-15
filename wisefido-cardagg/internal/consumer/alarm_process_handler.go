package consumer

import (
	"context"
	"encoding/json"
	"time"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

type AlarmProcessHandler struct {
	alarms *service.AlarmService
	logger *zap.Logger
}

func NewAlarmProcessHandler(alarms *service.AlarmService, logger *zap.Logger) *AlarmProcessHandler {
	return &AlarmProcessHandler{alarms: alarms, logger: logger}
}

type alarmProcessData struct {
	TenantID       string `json:"tenant_id"`
	CardID         string `json:"card_id"`
	DeviceID       string `json:"device_id"`
	AlarmLevel     string `json:"alarm_level"`
	AlarmType      string `json:"alarm_type"`
	AlarmTimestamp int64  `json:"alarm_timestamp"`
	ProcessType    string `json:"process_type"`
	EventID        string `json:"event_id"`
}

type cloudEventsEnvelope struct {
	Data json.RawMessage `json:"data"`
}

func (h *AlarmProcessHandler) Handle(ctx context.Context, msg interface{}) error {
	streamMsg, ok := msg.(map[string]interface{})
	if !ok {
		return nil
	}

	var topTs int64
	switch v := streamMsg["timestamp"].(type) {
	case int64:
		topTs = v
	case float64:
		topTs = int64(v)
	}
	if topTs > 0 && time.Now().Unix()-topTs > 30 {
		return nil
	}

	dataStr, ok := streamMsg["data"].(string)
	if !ok {
		return nil
	}

	var env cloudEventsEnvelope
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		h.logger.Warn("parse cloud events", zap.Error(err))
		return nil
	}

	var d alarmProcessData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		h.logger.Warn("parse alarm process data", zap.Error(err))
		return nil
	}

	if d.CardID == "" || d.AlarmLevel == "" {
		return nil
	}

	h.logger.Info("alarm process",
		zap.String("cid", d.CardID),
		zap.String("level", d.AlarmLevel),
		zap.String("type", d.AlarmType),
		zap.String("event_id", d.EventID))

	return h.alarms.HandleAlarmProcess(ctx, d.CardID, d.TenantID, d.AlarmLevel, d.AlarmType, d.EventID)
}
