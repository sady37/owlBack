package redis

import (
	"encoding/json"
	"fmt"
	"net/netip"
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
// =============================================================================

// =============================================================================
// 块 1：流与 dataValue 约定
// 流名称唯一出处：stream_names.go（StreamMonitor.Name、StreamEvent.Name 等）
// =============================================================================

const (
	DataCategoryKey = "dataCategory" // dataValue 每项"数据类别"键名；observation 字段名见 observation 包 Field*
	DataValueKey    = "dataValue"    // iot:*:stream 顶层键名，与 IoTStreamMessage 的 json 一致
	EventNameKey    = "event_name"   // dataValue 每项事件名，alarm/event 流必填；NewSingleItemMessage 自动补
)

// =============================================================================
// 块 2：IoTStreamMessage（iot:*:stream 唯一消息体）
// =============================================================================

// IoTStreamMessage envelope（device_ipv6 v2，单程票，2026-05-13）：
//
//   - DeviceAddr 是设备唯一身份；无 MAC/UUID 字段
//   - 空间归属由 consumer 用 prefix slice 派生：
//       addr.Prefix(80) = unit；addr.Prefix(88) = room；addr.Prefix(96) = bed
//   - tenant 同理：addr.Prefix(48)
//
// SubjectEntity 填写责任：
//   - device-gateway（qinglan/sleepace）：已绑卡填 card_id；未绑卡留空（R-009 红线）
//   - sensor agent（wisefido-sensor 等 layer 1+）：留空，cardagg 反查
//
// 详 doc/device_ipv6_migration_checklist.md
type IoTStreamMessage struct {
	// TDPv2 envelope
	Producer       string `json:"producer"`                  // canonical /128 IPv6（device-direct = device_addr；agent-derived = agent slot 内 IP）
	SubjectEntity  string `json:"subject_entity,omitempty"`  // card_id (UUID) when bound; 空 = 未绑卡 (R-009)
	SequenceNumber uint64 `json:"sequence_number,omitempty"` // producer 内单调

	// 5W where：唯一 device 标识；prefix slice 派生 unit/room/bed
	DeviceAddr netip.Addr `json:"device_addr"` // /128 INET, e.g. fd00:0:3:111:3:101:a2ac:d523
	DeviceType string     `json:"device_type"` // "Radar" / "Sleepad" — 物理类别

	// Envelope wrapper
	Timestamp int64         `json:"timestamp"`
	TopicType string        `json:"topic_type"`
	Category  string        `json:"category"`
	DataValue []interface{} `json:"dataValue"`
}

// IotHead 是 IoTStreamMessage 的 envelope-only 视图（不带 DataValue），用于不需 payload 的快速透传。
type IotHead struct {
	Producer       string     `json:"producer"`
	SubjectEntity  string     `json:"subject_entity,omitempty"`
	SequenceNumber uint64     `json:"sequence_number,omitempty"`
	DeviceAddr     netip.Addr `json:"device_addr"`
	DeviceType     string     `json:"device_type"`
	Timestamp      int64      `json:"timestamp"`
	TopicType      string     `json:"topic_type"`
	Category       string     `json:"category"`
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
		"device_addr":     m.DeviceAddr.String(), // 无效 addr 序列化为 "invalid IP"，consumer 端 ParseAddr 会失败
		"device_type":     m.DeviceType,
		"timestamp":       fmt.Sprintf("%d", m.Timestamp),
		"topic_type":      m.TopicType,
		"category":        m.Category,
		DataValueKey:      string(dataValueJSON),
	}
	return out
}

func FromStreamMap(values map[string]interface{}) (*IoTStreamMessage, error) {
	msg := &IoTStreamMessage{}
	msg.Producer = streamStr(values["producer"])
	msg.SubjectEntity = streamStr(values["subject_entity"])
	if seq := streamStr(values["sequence_number"]); seq != "" {
		if v, err := strconv.ParseUint(seq, 10, 64); err == nil {
			msg.SequenceNumber = v
		}
	}
	if addrStr := streamStr(values["device_addr"]); addrStr != "" {
		addr, err := netip.ParseAddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("parse device_addr %q: %w", addrStr, err)
		}
		msg.DeviceAddr = addr
	}
	msg.DeviceType = streamStr(values["device_type"])
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
// SpatialPath helpers 见独立 package owl-common/spatial（2026-05-07）
//
//   import "owl-common/spatial"
//   spatial.LongestPrefixMatch / spatial.DeriveDeviceAddr / ...
// =============================================================================

// BuildDeviceProducer 设备 Producer 标识：返回设备自己的 canonical IPv6 /128 字符串。
//
// 2026-05-15 platform-agent cutover：
//   - producer 列改 INET 类型；不再带 "device:" 前缀字符串
//   - device-direct alarm 的 producer == device_addr（同一 /128），用 producer==device_addr 区分
//     "设备直发" vs "agent 派生"（agent producer 在 fd00:0:fff0::/44 platform slot）
//   - 详见 owlBack/doc/platform_agent_addressing.md §4.2 / §5.2
func BuildDeviceProducer(addr netip.Addr) string {
	if !addr.IsValid() {
		return ""
	}
	return addr.String()
}

// NewIoTStreamMessageWithData 构造 envelope（device_ipv6 v2）。
// Producer 默认 "device:<addr>"；sensor agent 调用方再覆盖 msg.Producer = "sensor.<name>"。
// subjectEntity 由 caller 明确填：device-gateway 已绑=card_id；未绑/sensor agent 留空。
func NewIoTStreamMessageWithData(addr netip.Addr, subjectEntity, deviceType string, ts int64, topicType, category string, data map[string]interface{}) *IoTStreamMessage {
	// publish 边界 strip dataCategory + event_name；envelope.Category 是事件类型唯一权威
	if data != nil {
		delete(data, DataCategoryKey)
		delete(data, EventNameKey)
	}
	dataValue := []interface{}{data}
	if data == nil {
		dataValue = nil
	}
	return &IoTStreamMessage{
		Producer:      BuildDeviceProducer(addr),
		SubjectEntity: subjectEntity,
		DeviceAddr:    addr,
		DeviceType:    deviceType,
		Timestamp:     ts,
		TopicType:     topicType,
		Category:      category,
		DataValue:     dataValue,
	}
}

// BuildIoTStreamMessage 多 item dataValue 版本。
func BuildIoTStreamMessage(addr netip.Addr, deviceType, subjectEntity string, timestamp int64, topicType, category string, dataValue []interface{}) IoTStreamMessage {
	return IoTStreamMessage{
		Producer:      BuildDeviceProducer(addr),
		SubjectEntity: subjectEntity,
		DeviceAddr:    addr,
		DeviceType:    deviceType,
		Timestamp:     timestamp,
		TopicType:     topicType,
		Category:      category,
		DataValue:     dataValue,
	}
}

// FirstDataValue 取 dataValue 首项为 map，无或类型不对返回 nil。
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

// NewSingleItemMessage 构造单条 data 的 IoTStreamMessage，category 兜底从 data["dataCategory"] 提取。
func NewSingleItemMessage(addr netip.Addr, subjectEntity, deviceType string, ts int64, topicType, category string, data map[string]interface{}) *IoTStreamMessage {
	if data == nil {
		data = make(map[string]interface{})
	}
	withCat := make(map[string]interface{}, len(data)+1)
	for k, v := range data {
		withCat[k] = v
	}
	cat := category
	if cat == "" {
		if v, ok := withCat[DataCategoryKey].(string); ok {
			cat = v
		}
	}
	return NewIoTStreamMessageWithData(addr, subjectEntity, deviceType, ts, topicType, cat, withCat)
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
	ConfigCardChanged               = "config.card"
)

const (
	AlarmStatusActive       = "active"
	AlarmStatusAcked        = "acked"
	AlarmStatusResolved     = "resolved"
	AlarmStatusAutoResolved = "auto_resolved"
	AlarmStatusExpired      = "expired"
	AlarmProcessActionAck   = "ack"
)

func BuildDeviceStatusMessage(addr netip.Addr, subjectEntity, deviceType string, timestamp int64, statuses map[string]int) IoTStreamMessage {
	return IoTStreamMessage{
		Producer:      BuildDeviceProducer(addr),
		SubjectEntity: subjectEntity,
		DeviceAddr:    addr,
		DeviceType:    deviceType,
		Timestamp:     timestamp,
		TopicType:     "status",
		Category:      "deviceStatus",
		DataValue: []interface{}{
			map[string]interface{}{"statuses": statuses},
		},
	}
}

// BuildAlarmDeviceMessage 设备告警配置变更通知（config 流）。
// device_addr 是 v2 路由 key；qinglan/sleepace consumer 用 addr.Prefix(48) 派生 tenant，
// 物理 MAC 由 consumer 自己 DB 反查 (device_addr → devices.device_id → dfm.device_uid)
// 后下推到设备（IPv6 只携 MAC 末 32 bit，OUI 段需 DB 补全）。
func BuildAlarmDeviceMessage(source string, addr netip.Addr, settingType string, settingData map[string]interface{}) ConfigChangeMessage {
	now := time.Now()
	data := map[string]interface{}{
		"device_addr":  addr.String(),
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

// BuildAlarmProcessMessage 告警处置通知（config 流）；cardID 是路由 key（cards.card_id UUID 仍是 PK）。
//
// device_ipv6 单程票后不再带 device_addr/device_id —— consumer (cardagg/alarm_process_handler)
// 验证只读 cardID + alarmLevel + alarmType + eventID，device 字段是历史冗余。
func BuildAlarmProcessMessage(source, cardID, alarmLevel, alarmType, processType, eventID string, alarmTimestamp int64) ConfigChangeMessage {
	now := time.Now()
	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          uuid.New().String(),
		Source:      source,
		Type:        ConfigAlarmProcess,
		Time:        now.UTC().Format(time.RFC3339),
		Data: map[string]interface{}{
			"card_id":         cardID,
			"alarm_level":     alarmLevel,
			"alarm_type":      alarmType,
			"alarm_timestamp": alarmTimestamp,
			"process_type":    processType,
			"event_id":        eventID,
		},
	}
}

// DeviceItemForMessage 是 admin/api 后台对设备清单的展示项（外部 API 边界 R-002，不进 internal routing）。
// Phase 2 一刀切后：identity 字段只剩 device_uid；业务侧请用 device_addr 单独字段。
type DeviceItemForMessage struct {
	DeviceUID  string      `json:"device_uid"`
	DeviceCode string      `json:"device_code,omitempty"`
	DeviceName string      `json:"device_name,omitempty"`
	DeviceType interface{} `json:"device_type,omitempty"`
}

// ConfigChangedData — config:card:stream 唯一消息格式（type=ConfigCardChanged）。
//
// op:
//   - "reset"  → data startup 重对账信号；scope 三数组应为空
//   - "delete" → 卡物理删除（cardagg DeleteCardState + DDNS unregister 等清理）
//   - "update" → 其他任何变化（spatial create / monitor / bind / resident / name / alarm config）
//
// scope 三数组按 affected 范围给：
//   - Cards       : card_id CIDR  → cardagg.metaCache.Remove + UnitPicker.InvalidateUnit
//   - DeviceAddrs : device_addr /128 → cardagg.enablement.InvalidateDevices
//   - DeviceUIDs  : device_uid → qinglan/sleepace.InvalidateByDeviceUID
//
// "card 变更很少"设计哲学：不做字段级 incremental hint，evict 范围 + lazy reload。
type ConfigChangedData struct {
	Op           string   `json:"op"`            // reset / delete / update
	Cards        []string `json:"cards"`         // INET CIDR (/80, /88, /96)
	DeviceAddrs  []string `json:"device_addrs"`  // INET /128
	DeviceUIDs   []string `json:"device_uids"`   // device_uid (gateway 用)
}

func BuildConfigChangedMessage(source, op string, cards, deviceAddrs, deviceUIDs []string) ConfigChangeMessage {
	now := time.Now()
	return ConfigChangeMessage{
		SpecVersion: "1.0",
		ID:          uuid.New().String(),
		Source:      source,
		Type:        ConfigCardChanged,
		Time:        now.UTC().Format(time.RFC3339),
		Data: map[string]interface{}{
			"op":            op,
			"cards":         cards,
			"device_addrs":  deviceAddrs,
			"device_uids":   deviceUIDs,
		},
	}
}

// =============================================================================
// 块 5：Auth 流
//
// AuthMessage 保留 DeviceUID (MAC) — 外部 device 物理 auth 入口（R-002 边界）。
// 设备 boot 时只持有 MAC（出厂烧录），通过 auth handshake 才能拿到 device_addr。
// 这里是 IPv6 派生的"upstream 边界"，不能也用 device_addr。
// =============================================================================

const (
	AuthTopicType        = "auth"
	AuthCategoryRequest  = "auth_request"
	AuthCategoryResponse = "auth_response"
)

type AuthMessage struct {
	Producer      string                 `json:"producer,omitempty"`
	SubjectEntity string                 `json:"subject_entity,omitempty"`

	// Phase 2: identity 收口到 device_uid (logMAC)；handshake 后业务侧路由用 device_addr (handshake response 补)
	DeviceUID  string                 `json:"device_uid"`              // logMAC，外部物理 auth 入口
	DeviceAddr string                 `json:"device_addr,omitempty"`   // INET /128 canonical text，handshake 完成后 gateway 补
	DeviceType string                 `json:"device_type"`
	TenantID   string                 `json:"tenant_id"`
	Timestamp  int64                  `json:"timestamp"`
	TopicType  string                 `json:"topic_type"`
	Category   string                 `json:"category"`
	DataValue  map[string]interface{} `json:"dataValue"`
}

// BuildAuthRequestMessage tenantID 须由 Device gateway 在查 device_store 后传入：
// 在库则用该行 tenant_id（v2 = INET CIDR），未入库则用平台 trash(000…000)。
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
