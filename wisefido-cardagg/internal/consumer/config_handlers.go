// config_handlers.go — config:alarmDevice:stream / config:alarmProcess:stream 消费者。
//
// alarmDevice：alarm 启用配置变更 → enablement cache 失效
// alarmProcess：alarm 人工 ack/处置 → 重算 AlarmState 写 card:status

package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"owl-common/card"

	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

const configMaxAgeSec = 30

type cloudEventsEnvelope struct {
	Data json.RawMessage `json:"data"`
}

type alarmDeviceData struct {
	DeviceAddr  string `json:"device_addr"`
	SettingType string `json:"setting_type"`
}

type alarmProcessData struct {
	TenantID    string `json:"tenant_id"`
	CardID      string `json:"card_id"`
	DeviceAddr    string `json:"device_addr"`
	AlarmLevel  string `json:"alarm_level"`
	AlarmType   string `json:"alarm_type"`
	ProcessType string `json:"process_type"`
	EventID     string `json:"event_id"`
}

// AlarmDeviceHandler config:alarmDevice:stream — 设备 alarm 配置变更通知。
type AlarmDeviceHandler struct {
	enablement *service.AlarmEnablementCache
	logger     *zap.Logger
}

func NewAlarmDeviceHandler(enablement *service.AlarmEnablementCache, logger *zap.Logger) *AlarmDeviceHandler {
	return &AlarmDeviceHandler{enablement: enablement, logger: logger}
}

func (h *AlarmDeviceHandler) Handle(ctx context.Context, raw map[string]interface{}) error {
	dataStr, _ := raw["data"].(string)
	if dataStr == "" {
		return nil
	}
	var env cloudEventsEnvelope
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		return nil
	}
	var d alarmDeviceData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		return nil
	}
	if d.DeviceAddr == "" {
		return nil
	}
	h.enablement.Invalidate(d.DeviceAddr)
	h.logger.Info("alarm device config invalidated",
		zap.String("device_addr", d.DeviceAddr),
		zap.String("setting", d.SettingType))
	return nil
}

// AlarmProcessHandler config:alarmProcess:stream — 人工 ack/处置 alarm 后重算 AlarmState。
type AlarmProcessHandler struct {
	db     *sql.DB
	writer *card.Writer
	logger *zap.Logger
}

func NewAlarmProcessHandler(db *sql.DB, writer *card.Writer, logger *zap.Logger) *AlarmProcessHandler {
	return &AlarmProcessHandler{db: db, writer: writer, logger: logger}
}

func (h *AlarmProcessHandler) Handle(ctx context.Context, raw map[string]interface{}) error {
	var topTs int64
	switch v := raw["timestamp"].(type) {
	case int64:
		topTs = v
	case float64:
		topTs = int64(v)
	}
	if topTs > 0 && time.Now().Unix()-topTs > configMaxAgeSec {
		return nil
	}
	dataStr, _ := raw["data"].(string)
	if dataStr == "" {
		return nil
	}
	var env cloudEventsEnvelope
	if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
		return nil
	}
	var d alarmProcessData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		return nil
	}
	if d.CardID == "" || d.EventID == "" {
		return nil
	}
	cas, err := card.QueryCardAlarmState(ctx, h.db, d.CardID)
	if err != nil {
		h.logger.Warn("query alarm state", zap.String("cid", d.CardID), zap.Error(err))
		return nil
	}
	if cas == nil {
		return nil
	}
	return h.writer.WriteCardStatus(ctx, &card.CardStatus{
		CardID:     d.CardID,
		AlarmState: cas.ToAlarmState(),
	})
}
