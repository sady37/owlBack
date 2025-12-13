package models

import (
	"encoding/json"
	"time"
)

// StreamMessage Redis Streams 消息
type StreamMessage struct {
	StreamID   string
	StreamName string
	Values     map[string]interface{}
}

// RawDeviceData 原始设备数据（从 Redis Streams 解析）
type RawDeviceData struct {
	DeviceID     string                 `json:"device_id"`
	TenantID     string                 `json:"tenant_id"`
	SerialNumber string                 `json:"serial_number"`
	UID          string                 `json:"uid"`
	DeviceType   string                 `json:"device_type"` // "Radar" or "SleepPad"
	RawData      map[string]interface{} `json:"raw_data"`
	Timestamp    int64                  `json:"timestamp"`
	Topic        string                 `json:"topic,omitempty"`
}

// StandardizedData 标准化后的数据（写入 PostgreSQL）
type StandardizedData struct {
	TenantID    string
	DeviceID    string
	Timestamp   time.Time
	DataType    string // "observation" or "alarm"
	Category    string // FHIR Category
	
	// 轨迹数据
	TrackingID  *int
	RadarPosX   *int // cm
	RadarPosY   *int // cm
	RadarPosZ   *int // cm
	
	// 姿态
	PostureSNOMEDCode *string
	PostureDisplay    *string
	
	// 事件
	EventType       *string
	EventSNOMEDCode *string
	EventDisplay    *string
	AreaID          *int
	
	// 生命体征
	HeartRateCode        *string
	HeartRateDisplay     *string
	HeartRate            *int
	RespiratoryRateCode  *string
	RespiratoryRateDisplay *string
	RespiratoryRate      *int
	
	// 睡眠状态
	SleepStateSNOMEDCode *string
	SleepStateDisplay    *string
	
	// 床状态
	BedStatusSNOMEDCode *string
	BedStatusDisplay    *string
	
	// 原始数据（JSONB）
	RawOriginal json.RawMessage
}

// ParseRawDeviceData 从 Redis Streams 消息解析原始设备数据
func ParseRawDeviceData(streamID, streamName string, values map[string]interface{}) (*RawDeviceData, error) {
	// 从 Values 中提取 data 字段（JSON 字符串）
	dataStr, ok := values["data"].(string)
	if !ok {
		return nil, ErrInvalidDataFormat
	}
	
	var rawData RawDeviceData
	if err := json.Unmarshal([]byte(dataStr), &rawData); err != nil {
		return nil, err
	}
	
	return &rawData, nil
}

// ErrInvalidDataFormat 数据格式错误
var ErrInvalidDataFormat = &DataFormatError{Message: "invalid data format"}

// DataFormatError 数据格式错误类型
type DataFormatError struct {
	Message string
}

func (e *DataFormatError) Error() string {
	return e.Message
}

