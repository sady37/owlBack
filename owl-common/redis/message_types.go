package redis

import (
	"time"

	"github.com/google/uuid"
)

// LocationInfo 位置信息（仅用于 ConfigChangeMessage 等配置事件，不用于 iot:*:stream）
// iot:*:stream 中不再携带位置信息，需用时由消费方按 card_id 查询 Card 获取。
type LocationInfo struct {
	BranchID   *string `json:"branch_id,omitempty"`
	BuildingID *string `json:"building_id,omitempty"`
	UnitID     *string `json:"unit_id,omitempty"`
	RoomID     *string `json:"room_id,omitempty"`
	BedID      *string `json:"bed_id,omitempty"`
}

// IoTStreamMessage iot:*:stream 消息格式（统一格式，见 Reside_stream_stand.md）
// 顶层：device_id, device_type, card_id, tenant_id, timestamp, topic_type, category, data_value
// 不包含 device_uid、addressInfo；device_uid 需时放在 data_value 项内，位置信息按 card_id 查 Card
type IoTStreamMessage struct {
	DeviceID   string        `json:"device_id,omitempty"`
	DeviceType string        `json:"device_type"`
	CardID     string        `json:"card_id,omitempty"` // 未绑卡可空
	TenantID   string        `json:"tenant_id"`
	Timestamp  int64         `json:"timestamp"`
	TopicType  string        `json:"topic_type"` // "monitor", "stat", "event", "alarm"
	Category   string        `json:"category"`   // 数据类别（track, vital, sleep 等，多类用 . 拼接）
	DataValue  []interface{} `json:"data_value"` // 数组，每项含 category 及对应字段
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

// BuildDeviceStatusMessage 构建设备状态消息（发送到 iot:deviceStatus:stream）
// statuses: 设备状态 JSON（map[string]int，其中 0=异常/离线, 1=正常/在线）
// 支持字段（使用 const.StatusField* 常量）：
//   - online: 0=离线, 1=在线
//   - angle_abnormal: 0=正常, 1=倾角异常
//   - signal_poor: 0=正常, 1=信号差
//   - detached: 0=正常, 1=传感器脱落
//   - device_failure: 0=正常, 1=设备故障
//
// 例：map[string]int{"online": 1, "signal_poor": 1} 一般有故障时才会上报，正常状态可不包含或值为 0 但online=1 状态建议始终上报以便监控设备在线率,
func BuildDeviceStatusMessage(deviceID, deviceType, cardID, tenantID string, timestamp int64, statuses map[string]int) IoTStreamMessage {
	dataValue := []interface{}{
		map[string]interface{}{
			"category": "deviceStatus",
			"statuses": statuses,
		},
	}

	return IoTStreamMessage{
		DeviceID:   deviceID,
		DeviceType: deviceType,
		CardID:     cardID,
		TenantID:   tenantID,
		Timestamp:  timestamp,
		TopicType:  "status",
		Category:   "deviceStatus",
		DataValue:  dataValue,
	}
}

// Config device alarm setting 事件类型：设备告警配置变更
const (
	ConfigDeviceAlarmSettingUpdated = "config.alarmDevice"
)

// Config alarm process 事件类型：告警处理状态变更
const (
	ConfigAlarmProcess    = "config.alarmProcess"
	ConfigAlarmProcessAck = "config.alarm.process.ack" // 已弃用，改用 ConfigAlarmProcess
)

// Alarm process actions 告警处理 action 类型
const (
	AlarmProcessActionAck = "ack" // 用户已确认报警
)

// BuildAlarmDeviceMessage 构建设备告警配置变更消息（对应 config:alarmDevice:stream）
// deviceID: 设备ID
// deviceUID: 设备UID
// settingType: 配置类型（如 "enabled", "threshold_updated" 等）
// settingData: 配置数据（如启用状态、阈值等）
func BuildAlarmDeviceMessage(
	source, tenantID, deviceID, deviceUID, settingType string,
	settingData map[string]interface{},
) ConfigChangeMessage {
	now := time.Now()

	data := map[string]interface{}{
		"tenant_id":    tenantID,
		"device_id":    deviceID,
		"device_uid":   deviceUID,
		"setting_type": settingType,
		"timestamp_ms": now.UnixMilli(),
	}

	// 合并配置数据
	for k, v := range settingData {
		data[k] = v
	}

	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          uuid.New().String(),
		Source:      source,
		Type:        ConfigDeviceAlarmSettingUpdated, // 事件类型
		Time:        now.UTC().Format(time.RFC3339),
		Data:        data,
	}
}

// BuildAlarmProcessMessage 构建报警处理消息（发送到 config:alarmProcess:stream，供cardagg消费以更新显示）
// cardID: 卡片ID（用于cardagg找到最新数据）
// deviceID: 设备ID
// alarmLevel: 报警级别（EMERG, ALERT, CRIT, ERR, WARNING, NOTICE 等）
// alarmType: 报警类型（Fall, RadarAbnormalHeartRate 等）
// alarmTimestamp: 报警触发时间戳（秒级，用于比较防止旧数据覆盖）
// processType: 报警处理类型（acknowledged、resolved 等）
func BuildAlarmProcessMessage(
	source, tenantID, cardID, deviceID, alarmLevel, alarmType, processType string,
	alarmTimestamp int64,
) ConfigChangeMessage {
	now := time.Now()

	// 最小化数据：仅包含cardagg更新显示需要的字段
	data := map[string]interface{}{
		"tenant_id":       tenantID,
		"card_id":         cardID,
		"device_id":       deviceID,
		"alarm_level":     alarmLevel,
		"alarm_type":      alarmType,
		"alarm_timestamp": alarmTimestamp, // 报警触发时间，用于比较是否是同一个报警
		"process_type":    processType,    // 报警处理类型在 data 中
	}

	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          uuid.New().String(),
		Source:      source,
		Type:        ConfigAlarmProcess, // 事件类型
		Time:        now.UTC().Format(time.RFC3339),
		Data:        data,
	}
}

// Config card 事件类型：卡片变化通知
const (
	ConfigCardChanged            = "config.card"
	ConfigCardDeviceStoreChanged = "config.card.device_store" // device_store 变化信号
)

// DeviceItemForMessage 消息中包含的device信息（轻量级）
type DeviceItemForMessage struct {
	DeviceID   string      `json:"device_id"`
	DeviceUID  string      `json:"device_uid"`
	DeviceCode string      `json:"device_code,omitempty"`
	DeviceName string      `json:"device_name,omitempty"`
	DeviceType interface{} `json:"device_type,omitempty"`
}

// BuildCardChangeMessage 构建卡片变化通知消息
// 消费者收到消息后根据 branch_id 全量查询卡片，自动同步
func BuildCardChangeMessage(source, tenantID, cardID, unitID, branchID string) ConfigChangeMessage {
	return BuildCardChangeMessageWithExtra(source, tenantID, cardID, unitID, branchID, nil)
}

// BuildCardChangeMessageWithExtra 构建卡片变化通知消息（支持额外字段）
// 用于处理 card 变化导致的卡片变更（messageType 默认为 ConfigCardChanged）
// 若要处理 device_store 变化，使用 BuildCardChangeMessageWithExtraAndType 并指定 messageType 为 ConfigCardDeviceStoreChanged
//
// extraData: 额外字段（如果为 nil 则不添加）
func BuildCardChangeMessageWithExtra(source, tenantID, cardID, unitID, branchID string, extraData map[string]interface{}) ConfigChangeMessage {
	return BuildCardChangeMessageWithExtraAndType(source, tenantID, cardID, unitID, branchID, ConfigCardChanged, extraData)
}

// BuildCardChangeMessageWithExtraAndType 构建卡片变化通知消息（支持额外字段和自定义消息类型）
// 用于处理 device_store 变化导致的卡片变更或其他配置变化
// messageType: 消息类型，如 ConfigCardChanged（"config.card"）或 ConfigCardDeviceStoreChanged（"config.card.device_store"）
// 当处理 device_store 变化时，card_id 应为空，extraData 可包含：
//   - device_id: 受影响的设备ID
//   - change_type: 变化类型 (device_updated 或 device_deleted)
func BuildCardChangeMessageWithExtraAndType(source, tenantID, cardID, unitID, branchID, messageType string, extraData map[string]interface{}) ConfigChangeMessage {
	if messageType == "" {
		messageType = ConfigCardChanged
	}

	now := time.Now()
	data := map[string]interface{}{
		"tenant_id":    tenantID,
		"card_id":      cardID,
		"branch_id":    branchID,
		"unit_id":      unitID,
		"timestamp_ms": now.UnixMilli(),
	}

	// 合并额外字段
	if extraData != nil {
		for k, v := range extraData {
			data[k] = v
		}
	}

	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          uuid.New().String(),
		Source:      source,
		Type:        messageType,
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
	if len(deviceInfo) > 0 {
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
// 顶层不含 device_uid、addressInfo；data_value 为数组，每项含 category。
func BuildIoTStreamMessage(
	deviceID, deviceType, cardID, tenantID string,
	timestamp int64,
	topicType, category string,
	dataValue []interface{},
) IoTStreamMessage {
	return IoTStreamMessage{
		DeviceID:   deviceID,
		DeviceType: deviceType,
		CardID:     cardID,
		TenantID:   tenantID,
		Timestamp:  timestamp,
		TopicType:  topicType,
		Category:   category,
		DataValue:  dataValue,
	}
}

// CardChangeData 卡片变更事件数据（config:card:stream 中的数据）
type CardChangeData struct {
	TenantID    string `json:"tenant_id"`
	CardID      string `json:"card_id"`
	BranchID    string `json:"branch_id"`
	UnitID      string `json:"unit_id"`
	TimestampMs int64  `json:"timestamp_ms"`
}
