package consumer

import (
	"owl-common/alarm"
	"owl-common/redis"

	"go.uber.org/zap"
)

func streamEventName(m *redis.IoTStreamMessage, data map[string]interface{}) string {
	if data != nil {
		if v, _ := data[alarm.FieldEventName].(string); v != "" {
			return v
		}
		if v, _ := data[redis.DataCategoryKey].(string); v != "" {
			return v
		}
		if v, _ := data["category"].(string); v != "" {
			return v
		}
	}
	if m != nil {
		return m.Category
	}
	return ""
}

func streamLogFields(stream string, m *redis.IoTStreamMessage, eventName string) []zap.Field {
	return []zap.Field{
		zap.String("stream", stream),
		zap.String("topic_type", m.TopicType),
		zap.String("category", m.Category),
		zap.String("event", eventName),
		zap.String("card_id", m.CardID),
		zap.String("device_id", m.DeviceID),
		zap.String("device_uid", m.DeviceUID),
		zap.String("tenant_id", m.TenantID),
		zap.Int64("timestamp", m.Timestamp),
	}
}
