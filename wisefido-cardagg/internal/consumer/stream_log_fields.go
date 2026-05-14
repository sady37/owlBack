package consumer

import (
	"net/netip"

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
	addrStr := ""
	tenantStr := ""
	if m != nil && m.DeviceAddr.IsValid() {
		addrStr = m.DeviceAddr.String()
		tenantStr = netip.PrefixFrom(m.DeviceAddr, 48).Masked().String()
	}
	return []zap.Field{
		zap.String("stream", stream),
		zap.String("topic_type", m.TopicType),
		zap.String("category", m.Category),
		zap.String("event", eventName),
		zap.String("card_id", m.SubjectEntity),
		zap.String("device_addr", addrStr),
		zap.String("tenant_pref", tenantStr),
		zap.Int64("timestamp", m.Timestamp),
	}
}
