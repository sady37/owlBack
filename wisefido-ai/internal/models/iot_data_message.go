package models

// IoTDataMessage IoT 数据消息（从 Redis Streams 读取）
// 这是从设备 streams 读取的消息格式（兼容性结构体）
// 注意：现在直接从设备 streams 读取，不再使用 iot:data:stream
type IoTDataMessage struct {
	IoTTimeSeriesID int64   `json:"iot_timeseries_id"`
	DeviceID        string  `json:"device_id"`
	TenantID        string  `json:"tenant_id"`
	DeviceType      string  `json:"device_type"` // "Radar" 或 "Sleepace"（从设备表查询）
	Timestamp       int64   `json:"timestamp"`
	DataType        string  `json:"data_type"`   // "observation" or "alarm"
	Category        string  `json:"category"`    // FHIR Category
	EventType       *string `json:"event_type,omitempty"` // 设备直接报警类型（如 "Fall"），可选
}

