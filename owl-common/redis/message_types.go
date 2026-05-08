package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"owl-common/observation"

	"github.com/google/uuid"
)

// =============================================================================
// device_type AI 派生身份（历史 ".AI<NodeID>" 后缀方案已废弃）
//
// 当前约定：device_type 始终是源 sensor 物理类型（"Radar" / "Sleepad"），
// AI 派生身份由 dataValue 内 observation.Track.Source 一等公民字段表达。
// 例：source = "ai_ghost_penalty" / "ai_track_real"。
// =============================================================================

// =============================================================================
// 块 1：流与 dataValue 约定
// 流名称唯一出处：stream_names.go（StreamMonitor.Name、StreamEvent.Name 等）
// =============================================================================

const (
	DataCategoryKey = "dataCategory" // dataValue 每项“数据类别”键名；observation 字段名见 observation 包 Field*
	DataValueKey    = "dataValue"    // iot:*:stream 顶层键名，与 IoTStreamMessage 的 json 一致
	EventNameKey    = "event_name"   // dataValue 每项事件名，alarm/event 流必填；NewSingleItemMessage 自动补
)

// =============================================================================
// 块 2：IoTStreamMessage（iot:*:stream 唯一消息体）
// =============================================================================

// IoTStreamMessage envelope（Phase A v2，TDPv2 三维度对齐）：
//
//   - Producer / SubjectEntity / SpatialPath 显式 first-class
//   - 历史 CardID 字段已弃用，subject 走 SubjectEntity
//   - DeviceUID 仍保留（Phase D 移除）；wire 不再依赖物理 MAC/SN 做身份
//
// 5W where 演进（envelope_protocol_evolution.md）：
//   - 旧字段 SemanticLocation（单一 room_id/unit_id 字符串）保留兜底，新代码不要再写
//   - 新字段 SpatialPath / ScopePrefix（IPv6 风格） + RoomID / UnitID（first-class UUID）
//   - 路径格式 "tenant.branch.building.floor.unit.room.bed" 7 段，缺级用空（".."）
//   - ScopePrefix 1-7：producer 写到自己有的精度；consumer 用 prefix-match 决定订阅
//
// SubjectEntity 填写责任：
//   - device-gateway（qinglan/sleepace）：必填——已绑卡=card_id；未绑卡=device_id 占位
//   - sensor agent（wisefido-sensor 等 layer 1+）：留空，cardagg 反查
type IoTStreamMessage struct {
	// TDPv2 三维度
	Producer       string `json:"producer"`                  // "device:<UUID>" / "sensor.caregiver01"
	SubjectEntity  string `json:"subject_entity,omitempty"`  // 已绑=card_id；未绑=device_id；AI 留空
	SequenceNumber uint64 `json:"sequence_number,omitempty"` // producer 内单调

	// 5W where（first-class，IPv6 风格）—— 新代码必填，旧 SemanticLocation 保留兜底
	UnitID      string `json:"unit_id,omitempty"`      // 5W where: unit
	RoomID      string `json:"room_id,omitempty"`      // 5W where: room
	BedID       string `json:"bed_id,omitempty"`       // 5W where: bed（可选；床上传感器才填）
	SpatialPath string `json:"spatial_path,omitempty"` // "tenant.branch.building.floor.unit.room.bed"，缺级空
	ScopePrefix int    `json:"scope_prefix,omitempty"` // 1-7：producer 写到的精度；consumer prefix-match 用

	// Deprecated: 用 RoomID/UnitID/SpatialPath 替代；过渡期内继续兜底反序列化。
	SemanticLocation string `json:"semantic_location,omitempty"`

	// 物理身份（Phase D 远期淡出）
	DeviceUID  string `json:"device_uid"`  // 主键，业务逻辑
	DeviceID   string `json:"device_id"`   // server UUID
	DeviceType string `json:"device_type"`

	TenantID  string        `json:"tenant_id"`
	Timestamp int64         `json:"timestamp"`
	TopicType string        `json:"topic_type"`
	Category  string        `json:"category"`
	DataValue []interface{} `json:"dataValue"` // monitor 流为 []Track 形状，字段名见 observation 包
}

type IotHead struct {
	Producer       string `json:"producer"`
	SubjectEntity  string `json:"subject_entity,omitempty"`
	SequenceNumber uint64 `json:"sequence_number,omitempty"`

	// 5W where 新字段（first-class）
	UnitID      string `json:"unit_id,omitempty"`
	RoomID      string `json:"room_id,omitempty"`
	BedID       string `json:"bed_id,omitempty"`
	SpatialPath string `json:"spatial_path,omitempty"`
	ScopePrefix int    `json:"scope_prefix,omitempty"`

	// Deprecated: 用 RoomID/UnitID/SpatialPath
	SemanticLocation string `json:"semantic_location,omitempty"`

	DeviceUID  string `json:"device_uid"`
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	TenantID   string `json:"tenant_id"`
	Timestamp  int64  `json:"timestamp"`
	TopicType  string `json:"topic_type"`
	Category   string `json:"category"`
}


func streamStr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%v", s)
	}
}

func (m *IoTStreamMessage) ToStreamMap() map[string]interface{} {
	dataValueJSON, _ := json.Marshal(m.DataValue)
	out := map[string]interface{}{
		"producer":        m.Producer,
		"subject_entity":  m.SubjectEntity,
		"sequence_number": fmt.Sprintf("%d", m.SequenceNumber),
		"device_uid":      m.DeviceUID,
		"device_type":     m.DeviceType,
		"tenant_id":       m.TenantID,
		"timestamp":       fmt.Sprintf("%d", m.Timestamp),
		"topic_type":      m.TopicType,
		"category":        m.Category,
		DataValueKey:      string(dataValueJSON),
	}
	if m.DeviceID != "" {
		out["device_id"] = m.DeviceID
	}
	// 5W where（first-class）；只写非空字段，省 wire 体积
	if m.UnitID != "" {
		out["unit_id"] = m.UnitID
	}
	if m.RoomID != "" {
		out["room_id"] = m.RoomID
	}
	if m.BedID != "" {
		out["bed_id"] = m.BedID
	}
	if m.SpatialPath != "" {
		out["spatial_path"] = m.SpatialPath
	}
	if m.ScopePrefix > 0 {
		out["scope_prefix"] = fmt.Sprintf("%d", m.ScopePrefix)
	}
	// Deprecated 兜底：旧消费者仍读 semantic_location（取 RoomID 优先，否则 UnitID）
	if loc := m.SemanticLocation; loc != "" {
		out["semantic_location"] = loc
	} else if m.RoomID != "" {
		out["semantic_location"] = m.RoomID
	} else if m.UnitID != "" {
		out["semantic_location"] = m.UnitID
	}
	return out
}

func FromStreamMap(values map[string]interface{}) (*IoTStreamMessage, error) {
	msg := &IoTStreamMessage{}
	msg.Producer = streamStr(values["producer"])
	msg.SubjectEntity = streamStr(values["subject_entity"])
	msg.SemanticLocation = streamStr(values["semantic_location"])
	// 5W where（first-class）；新字段优先，旧 SemanticLocation 兜底
	msg.UnitID = streamStr(values["unit_id"])
	msg.RoomID = streamStr(values["room_id"])
	msg.BedID = streamStr(values["bed_id"])
	msg.SpatialPath = streamStr(values["spatial_path"])
	if sp := streamStr(values["scope_prefix"]); sp != "" {
		if v, err := strconv.Atoi(sp); err == nil {
			msg.ScopePrefix = v
		}
	}
	if seq := streamStr(values["sequence_number"]); seq != "" {
		if v, err := strconv.ParseUint(seq, 10, 64); err == nil {
			msg.SequenceNumber = v
		}
	}
	msg.DeviceUID = streamStr(values["device_uid"])
	msg.DeviceID = streamStr(values["device_id"])
	if msg.DeviceUID == "" {
		msg.DeviceUID = msg.DeviceID
	}
	msg.DeviceType = streamStr(values["device_type"])
	msg.TenantID = streamStr(values["tenant_id"])
	if ts := streamStr(values["timestamp"]); ts != "" {
		msg.Timestamp, _ = strconv.ParseInt(ts, 10, 64)
	}
	msg.TopicType = streamStr(values["topic_type"])
	msg.Category = streamStr(values["category"])
	if dataStr := streamStr(values[DataValueKey]); dataStr != "" {
		if err := json.Unmarshal([]byte(dataStr), &msg.DataValue); err != nil {
			return nil, fmt.Errorf("unmarshal dataValue: %w", err)
		}
	}
	if msg.DataValue == nil {
		msg.DataValue = []interface{}{}
	}
	return msg, nil
}

// =============================================================================
// SpatialPath helpers 已迁移至独立 package owl-common/spatial（2026-05-07）
//
// 调用方请用：
//   import "owl-common/spatial"
//   spatial.BuildPath / spatial.MatchPrefix / spatial.MatchAddress / ...
//   spatial.Address{Tenant:"T", Unit:"U", Room:"R", Bed:"B"}.Path()
//
// 见 doc/datagram_envelope.md 与 spatial/spatial.go。
// =============================================================================

// BuildDeviceProducer 设备 Producer 标识：用 server UUID（device_id），不用 device_uid (MAC/SN)。
// 见 doc/TODO.md Phase A: 内网仍有 plain MQTT 1883；MAC 泄露 → 物理追踪 + 位置关联。
func BuildDeviceProducer(deviceID string) string {
	if deviceID == "" {
		return ""
	}
	return "device:" + deviceID
}

// NewIoTStreamMessageWithData 构造 envelope（Phase A v2）。
// Producer 留给 caller 自行 msg.Producer = ... 设置（device gateway 用 BuildDeviceProducer(deviceID)；
// sensor agent 用 "sensor.<name>"），避免参数列表暴涨。
// subjectEntity 由 caller 明确填：device-gateway 已绑=card_id；未绑=device_id 占位；sensor agent 留空。
func NewIoTStreamMessageWithData(tenantID, subjectEntity, deviceUID, deviceID, deviceType string, ts int64, topicType, category string, data map[string]interface{}) *IoTStreamMessage {
	// Stage 1a/1b：在 publish 边界 strip dataCategory + event_name，envelope.Category 是事件类型唯一权威
	if data != nil {
		delete(data, DataCategoryKey)
		delete(data, EventNameKey)
	}
	dataValue := []interface{}{data}
	if data == nil {
		dataValue = nil
	}
	msg := &IoTStreamMessage{
		Producer:      BuildDeviceProducer(deviceID), // 默认 device-gateway 语义；sensor agent 调用方再覆盖
		SubjectEntity: subjectEntity,
		DeviceUID:     deviceUID,
		DeviceType:    deviceType,
		TenantID:      tenantID,
		Timestamp:     ts,
		TopicType:     topicType,
		Category:      category,
		DataValue:     dataValue,
	}
	if deviceID != "" {
		msg.DeviceID = deviceID
	}
	return msg
}

func BuildIoTStreamMessage(deviceUID, deviceID, deviceType, subjectEntity, tenantID string, timestamp int64, topicType, category string, dataValue []interface{}) IoTStreamMessage {
	return IoTStreamMessage{
		Producer:      BuildDeviceProducer(deviceID),
		SubjectEntity: subjectEntity,
		DeviceUID:     deviceUID,
		DeviceID:      deviceID,
		DeviceType:    deviceType,
		TenantID:      tenantID,
		Timestamp:     timestamp,
		TopicType:     topicType,
		Category:      category,
		DataValue:     dataValue,
	}
}

// FirstDataValue 取 dataValue 首项为 map，无或类型不对返回 nil。单条 payload 时 cardagg 等用此取 data。
func FirstDataValue(dataValue []interface{}) map[string]interface{} {
	if len(dataValue) == 0 {
		return nil
	}
	v, ok := dataValue[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return v
}

// NewSingleItemMessage 构造单条 data 的 IoTStreamMessage。先有 dataCategory（data 内），包头 Category 优先取 data 的 dataCategory，缺省时用参数 category；保证 payload 含 dataCategory。deviceID 可选，非空时写入包头。
// Producer 默认按 device-gateway 语义自动填充 ("device:<deviceID>")；sensor agent 调用方在拿到结果后覆盖 msg.Producer。
func NewSingleItemMessage(tenantID, subjectEntity, deviceUID, deviceID, deviceType string, ts int64, topicType, category string, data map[string]interface{}) *IoTStreamMessage {
	if data == nil {
		data = make(map[string]interface{})
	}
	withCat := make(map[string]interface{}, len(data)+1)
	for k, v := range data {
		withCat[k] = v
	}
	cat := category
	if cat == "" {
		// 兼容老 caller：category 参数为空时从 data["dataCategory"] 提一次进 envelope（提完即剥）
		if v, ok := withCat[DataCategoryKey].(string); ok {
			cat = v
		}
	}
	// Stage 1b：dataCategory + event_name 由 NewIoTStreamMessageWithData 边界 strip；envelope.Category 唯一权威
	return NewIoTStreamMessageWithData(tenantID, subjectEntity, deviceUID, deviceID, deviceType, ts, topicType, cat, withCat)
}

// =============================================================================
// 块 3：配置变更（CloudEvents、设备状态、告警配置、卡片变更）
// =============================================================================

type LocationInfo struct {
	BranchID   *string `json:"branch_id,omitempty"`
	BuildingID *string `json:"building_id,omitempty"`
	UnitID     *string `json:"unit_id,omitempty"`
	RoomID     *string `json:"room_id,omitempty"`
	BedID      *string `json:"bed_id,omitempty"`
}

type ConfigChangeMessage struct {
	SpecVersion string                 `json:"specversion"`
	ID          string                 `json:"id"`
	Source      string                 `json:"source"`
	Type        string                 `json:"type"`
	Time        string                 `json:"time"`
	Data        map[string]interface{} `json:"data"`
}

const (
	ConfigDeviceAlarmSettingUpdated = "config.alarmDevice"
	ConfigAlarmProcess              = "config.alarmProcess"
	ConfigAlarmProcessAck           = "config.alarm.process.ack" // 已弃用
	ConfigCardChanged = "config.card"
)

const (
	AlarmStatusActive       = "active"
	AlarmStatusAcked        = "acked"
	AlarmStatusResolved     = "resolved"
	AlarmStatusAutoResolved = "auto_resolved"
	AlarmStatusExpired      = "expired"
	AlarmProcessActionAck  = "ack"
)

func BuildDeviceStatusMessage(deviceUID, deviceID, deviceType, subjectEntity, tenantID string, timestamp int64, statuses map[string]int) IoTStreamMessage {
	// Stage 1b：dataValue 不再带 dataCategory；envelope.Category 是事件类型唯一权威
	return IoTStreamMessage{
		Producer:      BuildDeviceProducer(deviceID),
		SubjectEntity: subjectEntity,
		DeviceUID:     deviceUID,
		DeviceID:      deviceID,
		DeviceType:    deviceType,
		TenantID:      tenantID,
		Timestamp:     timestamp,
		TopicType:     "status",
		Category:      "deviceStatus",
		DataValue: []interface{}{
			map[string]interface{}{"statuses": statuses},
		},
	}
}

func BuildAlarmDeviceMessage(source, tenantID, deviceID, deviceUID, settingType string, settingData map[string]interface{}) ConfigChangeMessage {
	now := time.Now()
	data := map[string]interface{}{
		"tenant_id":    tenantID,
		"device_id":    deviceID,
		"device_uid":   deviceUID,
		"setting_type": settingType,
		"timestamp_ms": now.UnixMilli(),
	}
	for k, v := range settingData {
		data[k] = v
	}
	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          uuid.New().String(),
		Source:      source,
		Type:        ConfigDeviceAlarmSettingUpdated,
		Time:        now.UTC().Format(time.RFC3339),
		Data:        data,
	}
}

func BuildAlarmProcessMessage(source, tenantID, cardID, deviceID, alarmLevel, alarmType, processType, eventID string, alarmTimestamp int64) ConfigChangeMessage {
	now := time.Now()
	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          uuid.New().String(),
		Source:      source,
		Type:        ConfigAlarmProcess,
		Time:        now.UTC().Format(time.RFC3339),
		Data: map[string]interface{}{
			"tenant_id":       tenantID,
			"card_id":         cardID,
			"device_id":       deviceID,
			"alarm_level":     alarmLevel,
			"alarm_type":      alarmType,
			"alarm_timestamp": alarmTimestamp,
			"process_type":    processType,
			"event_id":        eventID,
		},
	}
}

type DeviceItemForMessage struct {
	DeviceID   string      `json:"device_id"`
	DeviceUID  string      `json:"device_uid"`
	DeviceCode string      `json:"device_code,omitempty"`
	DeviceName string      `json:"device_name,omitempty"`
	DeviceType interface{} `json:"device_type,omitempty"`
}

type CardChangeData struct {
	TenantID    string `json:"tenant_id"`
	CardID      string `json:"card_id"`
	BranchID    string `json:"branch_id"`
	UnitID      string `json:"unit_id"`
	TimestampMs int64  `json:"timestamp_ms"`
}

func BuildCardChangeMessage(source, tenantID, cardID, unitID, branchID string) ConfigChangeMessage {
	return BuildCardChangeMessageWithExtraAndType(source, tenantID, cardID, unitID, branchID, ConfigCardChanged, nil)
}

func BuildCardChangeMessageWithExtra(source, tenantID, cardID, unitID, branchID string, extraData map[string]interface{}) ConfigChangeMessage {
	return BuildCardChangeMessageWithExtraAndType(source, tenantID, cardID, unitID, branchID, ConfigCardChanged, extraData)
}

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
	for k, v := range extraData {
		data[k] = v
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

// =============================================================================
// 块 5：Auth 流
// =============================================================================

const (
	AuthTopicType        = "auth"
	AuthCategoryRequest  = "auth_request"
	AuthCategoryResponse = "auth_response"
)

type AuthMessage struct {
	Producer      string                 `json:"producer,omitempty"`
	SubjectEntity string                 `json:"subject_entity,omitempty"`

	DeviceUID  string                 `json:"device_uid"`
	DeviceID   string                 `json:"device_id,omitempty"`
	DeviceType string                 `json:"device_type"`
	TenantID   string                 `json:"tenant_id"`
	Timestamp  int64                  `json:"timestamp"`
	TopicType  string                 `json:"topic_type"`
	Category   string                 `json:"category"`
	DataValue  map[string]interface{} `json:"dataValue"`
}

// BuildAuthRequestMessage tenantID 须由 Device gateway 在查 device_store 后传入：在库则用该行 tenant_id，未入库则用平台 trash(000…000)。
func BuildAuthRequestMessage(deviceUID, deviceType, tenantID, remoteAddr string, deviceInfo map[string]interface{}) AuthMessage {
	now := time.Now().Unix()
	dataValue := map[string]interface{}{
		"category":    AuthCategoryRequest,
		"device_uid":  deviceUID,
		"remote_addr": remoteAddr,
	}
	if len(deviceInfo) > 0 {
		if logInfo, ok := deviceInfo["log"]; ok {
			dataValue["log"] = logInfo
			delete(deviceInfo, "log")
		}
		for k, v := range deviceInfo {
			dataValue[k] = v
		}
	}
	return AuthMessage{
		DeviceUID:  deviceUID,
		DeviceType: deviceType,
		TenantID:   tenantID,
		Timestamp:  now,
		TopicType:  AuthTopicType,
		Category:   AuthCategoryRequest,
		DataValue:  dataValue,
	}
}

func BuildAuthResponseMessage(deviceUID, deviceType, tenantID, authStatus, mqttServer string, mqttPort int, logInfo interface{}) AuthMessage {
	now := time.Now().Unix()
	dataValue := map[string]interface{}{
		"category":    AuthCategoryResponse,
		"device_uid":  deviceUID,
		"auth_status": authStatus,
	}
	if authStatus == "success" {
		if mqttServer != "" {
			dataValue["mqtt_server"] = mqttServer
		}
		if mqttPort > 0 {
			dataValue["mqtt_port"] = mqttPort
		}
	}
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

// =============================================================================
// 块 6：iot:monitor 与 Track（Track 唯一定义在 observation 包）
// =============================================================================

type Track = observation.Track

type EventItem = observation.EventItem
/*
iot:monitor:stream
type MonitorTrack struck {
		DeviceUID:  deviceUID,
		DeviceType: deviceType,
		TenantID:   tenantID,
		Timestamp:  now,
		TopicType:  AuthTopicType,
		Category:   AuthCategoryResponse,
        observation.Track
		observation.Track
		observation.Track3
}

iot:event:stream
type MonitorTrack struck {
		DeviceUID:  deviceUID,
		DeviceType: deviceType,
		TenantID:   tenantID,
		Timestamp:  now,
		TopicType:  AuthTopicType,
		Category:   AuthCategoryResponse,
        observation.EventItem
}





*/