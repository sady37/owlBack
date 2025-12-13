package models

// IoTDataMessage iot:data:stream 消息格式
// 这是从 wisefido-data-transformer 发布到 iot:data:stream 的消息格式
type IoTDataMessage struct {
	IoTTimeSeriesID int64  `json:"iot_timeseries_id"`
	DeviceID        string `json:"device_id"`
	TenantID        string `json:"tenant_id"`
	DeviceType      string `json:"device_type"` // "Radar" 或 "Sleepace"（从设备表查询）
	Timestamp       int64  `json:"timestamp"`
	DataType        string `json:"data_type"`   // "observation" or "alarm"
	Category        string `json:"category"`    // FHIR Category
}

