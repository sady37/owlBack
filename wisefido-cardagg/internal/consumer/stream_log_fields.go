package consumer

import (
	"owl-common/redis"

	"go.uber.org/zap"
)

func streamEventName(m *redis.IoTStreamMessage, _ map[string]interface{}) string {
	// Stage 1b：envelope.Category 是事件类型唯一权威
	if m == nil {
		return ""
	}
	return m.Category
}

func streamLogFields(stream string, m *redis.IoTStreamMessage, eventName string) []zap.Field {
	return []zap.Field{
		zap.String("stream", stream),
		zap.String("topic_type", m.TopicType),
		zap.String("category", m.Category),
		zap.String("event", eventName),
		zap.String("card_id", m.SubjectEntity),
		zap.String("device_id", m.DeviceID),
		zap.String("device_uid", m.DeviceUID),
		zap.String("tenant_id", m.TenantID),
		zap.Int64("timestamp", m.Timestamp),
	}
}
