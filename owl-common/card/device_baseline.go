package card

import "strings"

// BaselineField 基准结构在 JSON / Redis map / 配置里的统一键名，与各网关、流包头对齐。
const (
	BaselineFieldTenantID           = "tenant_id"
	BaselineFieldDeviceID           = "device_id"
	BaselineFieldDeviceUID          = "device_uid"
	BaselineFieldDeviceCode         = "device_code"    //device--server communite id,not uid
	BaselineFieldDeviceType         = "device_type"
	BaselineFieldBranchID           = "branch_id"
	BaselineFieldBuildingID         = "building_id"
	BaselineFieldFloor              = "floor"
	BaselineFieldUnitID             = "unit_id"
	BaselineFieldRoomID             = "room_id"
	BaselineFieldBedID              = "bed_id"
	BaselineFieldCardID             = "card_id"
	BaselineFieldAllowAccess        = "allow_access"
	BaselineFieldBusinessAccess     = "business_access"
	BaselineFieldMonitoringEnabled  = "monitoring_enabled"
	BaselineFieldRevision           = "revision"
	BaselineFieldUpdatedAtMS        = "updated_at_ms"
)

// DeviceBaseline 设备身份与策略的统一类型：CardDB 联合查询 Scan、网关流包头、Redis/JSON、进程内缓存均用此结构（原 DeviceCardMapping 已并入）。
// 约定：branch_id / building_id / floor 可选；同房判定以 tenant_id + unit_id（若 room 仅在 unit 内唯一则必填）+ room_id 为准。
// 未绑卡时 card_id 为空，EffectiveCardID() 回退为 device_id（UUID），与 cardagg IotPreparedHandler 未绑卡占位一致。
type DeviceBaseline struct {
	// 硬件身份
	DeviceID   string `json:"device_id,omitempty"`
	DeviceUID  string `json:"device_uid,omitempty"`
	DeviceCode string `json:"device_code,omitempty"`
	DeviceType string `json:"device_type,omitempty"`

	// 组织（tenant 业务上必填；branch/building/floor 可省略）
	TenantID   string `json:"tenant_id,omitempty"`
	BranchID   string `json:"branch_id,omitempty"`
	BuildingID string `json:"building_id,omitempty"`
	Floor      string `json:"floor,omitempty"`

	// 空间附加
	UnitID string `json:"unit_id,omitempty"`
	RoomID string `json:"room_id,omitempty"`
	BedID  string `json:"bed_id,omitempty"`

	// 展示 / 卡聚合
	CardID string `json:"card_id,omitempty"`

	// 策略（device_store + devices）
	AllowAccess         bool   `json:"allow_access"`
	BusinessAccess      string `json:"business_access,omitempty"`
	MonitoringEnabled   bool   `json:"monitoring_enabled,omitempty"`

	// 同步元（可选，用于 cardChange / 对账）
	Revision    int64 `json:"revision,omitempty"`
	UpdatedAtMS int64 `json:"updated_at_ms,omitempty"`
}

// StreamHeadCardID 流包头 card_id：已绑卡用 card_id，未绑卡用 device_id（UUID），与 cardagg 未绑卡占位一致。
func StreamHeadCardID(cardID, deviceID string) string {
	if s := strings.TrimSpace(cardID); s != "" {
		return s
	}
	return strings.TrimSpace(deviceID)
}

// EffectiveCardID 同 StreamHeadCardID（DeviceBaseline 上的便捷方法）。
func (b *DeviceBaseline) EffectiveCardID() string {
	if b == nil {
		return ""
	}
	return StreamHeadCardID(b.CardID, b.DeviceID)
}

// SameRoomScopeKey 同房粗判：tenant|unit|room（unit 在 room 全局唯一时可传空 unit，由数据模型保证）。
func (b *DeviceBaseline) SameRoomScopeKey() string {
	if b == nil {
		return ""
	}
	return strings.TrimSpace(b.TenantID) + "|" + strings.TrimSpace(b.UnitID) + "|" + strings.TrimSpace(b.RoomID)
}
