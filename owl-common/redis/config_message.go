package redis

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LocationInfo 位置信息（可选，与 iot:*:stream 格式一致）
// 注意：位置信息只在 event/alarm 消息中包含，auth/monitor/stat 不包含
type LocationInfo struct {
	BranchID   *string `json:"branch_id,omitempty"`
	BuildingID *string `json:"building_id,omitempty"`
	UnitID     *string `json:"unit_id,omitempty"`
	RoomID     *string `json:"room_id,omitempty"`
	BedID      *string `json:"bed_id,omitempty"`
}

// IoTStreamMessage iot:*:stream 消息格式（统一格式）
// 用于 monitor, stat, event, alarm 等 stream
// 注意：位置信息字段（LocationInfo）只在 event/alarm 中包含
type IoTStreamMessage struct {
	DeviceID   string                 `json:"device_id,omitempty"`
	DeviceUID  string                 `json:"device_uid"`
	DeviceType string                 `json:"device_type"`
	TenantID   string                 `json:"tenant_id"`
	Timestamp  int64                  `json:"timestamp"`
	TopicType  string                 `json:"topic_type"` // "monitor", "stat", "event", "alarm"
	Category   string                 `json:"category"`   // 数据类别（track, vital, sleep, fall 等）
	DataValue  map[string]interface{} `json:"data_value"`
	// 位置信息字段（可选，只在 event/alarm 中包含）
	LocationInfo
}

// ConfigChangeMessage 配置变更消息格式（CloudEvents 标准格式）
// 参考 CNCF CloudEvents 规范：https://github.com/cloudevents/spec
// 用于所有配置变更场景：设备在线状态、告警配置、其他配置等
//
// CloudEvents 标准字段：
//   - specversion: CloudEvents 规范版本（固定为 "1.0"）
//   - id: 事件唯一标识（UUID）
//   - source: 事件来源（服务 URI，如 "wisefido-qinglan"）
//   - type: 事件类型（格式：{domain}.{category}.{action}，如 "config.device.online"）
//   - time: 事件时间（RFC3339 UTC 格式，如 "2024-01-22T10:00:00Z"）
//   - data: 事件数据（业务数据，包含 device_id, device_uid, tenant_id 等）
type ConfigChangeMessage struct {
	SpecVersion string                 `json:"specversion"` // CloudEvents 规范版本，固定为 "1.0"
	ID          string                 `json:"id"`          // 事件唯一标识（UUID）
	Source      string                 `json:"source"`      // 事件来源（服务 URI，如 "wisefido-qinglan"）
	Type        string                 `json:"type"`        // 事件类型（格式：{domain}.{category}.{action}，如 "config.device.online"）
	Time        string                 `json:"time"`        // 事件时间（RFC3339 UTC 格式，如 "2024-01-22T10:00:00Z"）
	Data        map[string]interface{} `json:"data"`        // 事件数据（业务数据）
}

// BuildAlarmCloudMessage 构建 alarm_cloud 配置变更消息（CloudEvents 标准格式）
// 注意：alarm_cloud 事件已废弃，不再发布
func BuildAlarmCloudMessage(source, tenantID string) ConfigChangeMessage {
	now := time.Now()
	eventID := uuid.New().String()

	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          eventID,
		Source:      source,
		Type:        "config.alarm.cloud.updated", // 事件类型：{domain}.{category}.{action}
		Time:        now.UTC().Format(time.RFC3339),
		Data: map[string]interface{}{
			"tenant_id": tenantID,
		},
	}
}

// BuildAlarmDeviceMessage 构建 alarm_device 配置变更消息（位置信息可选，可为空，CloudEvents 标准格式）
func BuildAlarmDeviceMessage(source, tenantID, deviceID, deviceUID, deviceCode, deviceType string, locationInfo *LocationInfo) ConfigChangeMessage {
	now := time.Now()
	eventID := uuid.New().String()

	data := map[string]interface{}{
		"tenant_id":   tenantID,
		"device_id":   deviceID,
		"device_uid":  deviceUID,
		"device_code": deviceCode,
		"device_type": deviceType,
	}

	// 位置信息可选，如果提供则添加到 data
	if locationInfo != nil {
		if locationInfo.BranchID != nil {
			data["branch_id"] = *locationInfo.BranchID
		}
		if locationInfo.BuildingID != nil {
			data["building_id"] = *locationInfo.BuildingID
		}
		if locationInfo.UnitID != nil {
			data["unit_id"] = *locationInfo.UnitID
		}
		if locationInfo.RoomID != nil {
			data["room_id"] = *locationInfo.RoomID
		}
		if locationInfo.BedID != nil {
			data["bed_id"] = *locationInfo.BedID
		}
	}

	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          eventID,
		Source:      source,
		Type:        "config.alarm.device.updated", // 事件类型：{domain}.{category}.{action}
		Time:        now.UTC().Format(time.RFC3339),
		Data:        data,
	}
}

// BuildDeviceOnlineStatusMessage 构建设备在线状态消息（CloudEvents 标准格式）
// onlineStatus: online/offline/unsubscribed
func BuildDeviceOnlineStatusMessage(source, tenantID, deviceID, deviceUID, deviceCode, deviceType, onlineStatus string) ConfigChangeMessage {
	now := time.Now()
	eventID := uuid.New().String()

	data := map[string]interface{}{
		"tenant_id":   tenantID,
		"device_id":   deviceID,
		"device_uid":  deviceUID,
		"device_code": deviceCode,
		"device_type": deviceType,
		"status":      onlineStatus, // 状态值
	}

	// 构建事件类型：config.device.{status}
	eventType := fmt.Sprintf("config.device.%s", onlineStatus)

	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          eventID,
		Source:      source,
		Type:        eventType, // 事件类型：config.device.online/offline/unsubscribed
		Time:        now.UTC().Format(time.RFC3339),
		Data:        data,
	}
}

// Auth Stream 相关常量定义
const (
	// AuthTopicType Auth Stream 的 topic_type 值
	AuthTopicType = "auth"

	// AuthCategoryRequest Auth Stream 的 category 值：认证请求
	AuthCategoryRequest = "auth_request"

	// AuthCategoryResponse Auth Stream 的 category 值：认证响应
	AuthCategoryResponse = "auth_response"
)

// AuthMessage Auth 事件的 Redis Stream 消息格式（统一格式）
// 注意：auth 消息不包含位置信息（只在 event/alarm 中包含）
type AuthMessage struct {
	DeviceID   string                 `json:"device_id,omitempty"`
	DeviceUID  string                 `json:"device_uid"`
	DeviceType string                 `json:"device_type"`
	TenantID   string                 `json:"tenant_id"`
	Timestamp  int64                  `json:"timestamp"`
	TopicType  string                 `json:"topic_type"` // "auth"
	Category   string                 `json:"category"`   // "auth_request" 或 "auth_response"
	DataValue  map[string]interface{} `json:"data_value"`
	// 注意：auth 消息不包含 LocationInfo 字段（位置信息只在 event/alarm 中包含）
}

// BuildAuthRequestMessage 构建认证请求消息
func BuildAuthRequestMessage(deviceUID, deviceType, remoteAddr string, deviceInfo map[string]interface{}) AuthMessage {
	now := time.Now().Unix()

	// 构建 data_value
	dataValue := map[string]interface{}{
		"category":    AuthCategoryRequest,
		"device_uid":  deviceUID,
		"remote_addr": remoteAddr,
	}

	// 添加设备信息（如果存在）
	if deviceInfo != nil && len(deviceInfo) > 0 {
		// 如果 deviceInfo 中有 log 字段，将其作为嵌套对象
		if logInfo, ok := deviceInfo["log"]; ok {
			dataValue["log"] = logInfo
			delete(deviceInfo, "log")
		}
		// 其他字段复制到 data_value
		for k, v := range deviceInfo {
			dataValue[k] = v
		}
	}

	return AuthMessage{
		DeviceUID:  deviceUID,
		DeviceType: deviceType,
		TenantID:   "00000000-0000-0000-0000-000000000000", // 认证请求时使用默认 tenant_id
		Timestamp:  now,
		TopicType:  AuthTopicType,
		Category:   AuthCategoryRequest,
		DataValue:  dataValue,
	}
}

// BuildAuthResponseMessage 构建认证响应消息
// 注意：auth 消息不包含位置信息（只在 event/alarm 中包含）
func BuildAuthResponseMessage(
	deviceUID, deviceType, tenantID, authStatus, mqttServer string,
	mqttPort int,
	logInfo interface{},
) AuthMessage {
	now := time.Now().Unix()

	// 构建 data_value
	dataValue := map[string]interface{}{
		"category":    AuthCategoryResponse,
		"device_uid":  deviceUID,
		"auth_status": authStatus,
	}

	// 仅在成功时添加 MQTT 配置信息
	if authStatus == "success" {
		if mqttServer != "" {
			dataValue["mqtt_server"] = mqttServer
		}
		if mqttPort > 0 {
			dataValue["mqtt_port"] = mqttPort
		}
	}

	// 添加日志信息
	if logInfo != nil {
		dataValue["log"] = logInfo
	}

	return AuthMessage{
		DeviceUID:  deviceUID,
		DeviceType: deviceType,
		TenantID:   tenantID,
		Timestamp:  now,
		TopicType:  AuthTopicType,
		Category:   AuthCategoryResponse,
		DataValue:  dataValue,
	}
}

// BuildIoTStreamMessage 构建 iot:*:stream 消息（monitor, stat, event, alarm）
// locationInfo 可选，只在 event/alarm 中包含
func BuildIoTStreamMessage(
	deviceID, deviceUID, deviceType, tenantID string,
	timestamp int64,
	topicType, category string,
	dataValue map[string]interface{},
	locationInfo *LocationInfo,
) IoTStreamMessage {
	msg := IoTStreamMessage{
		DeviceID:   deviceID,
		DeviceUID:  deviceUID,
		DeviceType: deviceType,
		TenantID:   tenantID,
		Timestamp:  timestamp,
		TopicType:  topicType,
		Category:   category,
		DataValue:  dataValue,
	}

	// 位置信息可选，只在 event/alarm 中包含
	if (topicType == "event" || topicType == "alarm") && locationInfo != nil {
		msg.LocationInfo = *locationInfo
	}

	return msg
}
