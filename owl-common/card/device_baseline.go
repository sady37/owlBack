package card

import (
	"net/netip"
	"strings"
)

// Phase A 后：subject_entity 由 publisher（device-gateway 端）填写，未绑卡用 device_id 占位。
// 历史的 StreamHeadCardID / EffectiveCardID 助手已删除（不再需要"包头 cardID 兜底回填"），
// 业务路径直接读 envelope.SubjectEntity。

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

// DeviceBaseline 设备身份与策略的统一类型：CardDB 联合查询 Scan、网关流包头、Redis/JSON、进程内缓存均用此结构。
// 约定：branch_id / building_id / floor 可选；同房判定以 tenant_id + unit_id（若 room 仅在 unit 内唯一则必填）+ room_id 为准。
//
// device_ipv6 单程票（doc/device_ipv6_migration_checklist.md）：
//   - DeviceAddr 是路由层主键；DeviceID/DeviceUID/TenantID 等保留作 admin/外部 API 边界 (R-002/R-003)
//   - 未绑卡时 CardID 为空，publisher 在构造 envelope.SubjectEntity 时留空（R-009 红线）
type DeviceBaseline struct {
	// 硬件身份
	DeviceID   string     `json:"device_id,omitempty"`
	DeviceUID  string     `json:"device_uid,omitempty"`
	DeviceCode string     `json:"device_code,omitempty"`
	DeviceType string     `json:"device_type,omitempty"`
	DeviceAddr netip.Addr `json:"device_addr,omitempty"` // /128 INET，路由层主键

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

// SubjectEntityForBaseline 计算 envelope.SubjectEntity：已绑=card_id；未绑=device_id 占位。
// device-gateway publisher 在构造 envelope 时统一调用本 helper（替代 Phase A 之前的 StreamHeadCardID）。
func SubjectEntityForBaseline(b *DeviceBaseline) string {
	if b == nil {
		return ""
	}
	if s := strings.TrimSpace(b.CardID); s != "" {
		return s
	}
	return strings.TrimSpace(b.DeviceID)
}

// SameRoomScopeKey 同房粗判：tenant|unit|room（unit 在 room 全局唯一时可传空 unit，由数据模型保证）。
func (b *DeviceBaseline) SameRoomScopeKey() string {
	if b == nil {
		return ""
	}
	return strings.TrimSpace(b.TenantID) + "|" + strings.TrimSpace(b.UnitID) + "|" + strings.TrimSpace(b.RoomID)
}
