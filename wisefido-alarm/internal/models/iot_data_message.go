package models

// IoTDataMessage IoT 数据消息（从 Redis Streams 读取）
type IoTDataMessage struct {
	TenantID  string  `json:"tenant_id"`
	DeviceID  string  `json:"device_id"`
	EventType *string `json:"event_type,omitempty"` // 事件类型，如 "BED_LEFT"
	// 其他字段根据需要添加
}

