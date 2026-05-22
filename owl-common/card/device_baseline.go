package card

import (
	"net/netip"
	"strings"
)

// subject_entity 契约：已绑卡 = card_id (INET CIDR text)；未绑卡 = 空（R-009 红线）。
// device-gateway publisher 直接 inline 处理（qinglan/sleepace mqtt_consumer），不再有 owl-common helper。
// 业务路径读 envelope.SubjectEntity；空字符串由 consumer 端 LPM 反查解析。

// BaselineField 基准结构在 JSON / Redis map / 配置里的统一键名，与各网关、流包头对齐。
const (
	BaselineFieldTenantID           = "tenant_id"
	BaselineFieldDeviceUID          = "device_uid"
	BaselineFieldDeviceAddr         = "device_addr"
	BaselineFieldDeviceCode         = "device_code"    //device--server communite id,not uid
	BaselineFieldDeviceType         = "device_type"
	BaselineFieldBranchID           = "branch_id"
	BaselineFieldBuildingID         = "building_id"
	BaselineFieldFloor              = "floor"
	BaselineFieldUnitID             = "unit_id"
	BaselineFieldRoomID             = "room_id"
	BaselineFieldBedID              = "bed_id"
	BaselineFieldCardID             = "card_id"
	BaselineFieldAccess             = "access"           // platform_admin 审批位（devices.access），合并自 v1 allow_access/business_access
	BaselineFieldMonitoringEnabled  = "monitoring_enabled"
	BaselineFieldRevision           = "revision"
	BaselineFieldUpdatedAtMS        = "updated_at_ms"
)

// DeviceBaseline 设备身份与策略的统一类型：CardDB 联合查询 Scan、网关流包头、Redis/JSON、进程内缓存均用此结构。
// 约定：branch_id / building_id / floor 可选；同房判定以 tenant_id + unit_id（若 room 仅在 unit 内唯一则必填）+ room_id 为准。
//
// Phase 2 一刀切：device_id UUID 合成层退役。
//   - DeviceAddr INET /128 路由层主键（业务侧寻址）
//   - DeviceUID  VARCHAR(50) logMAC identity 不变量（dfm PK）
//   - 未绑卡时 CardID 为空，publisher 在构造 envelope.SubjectEntity 时留空（R-009 红线）
type DeviceBaseline struct {
	// 硬件身份
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

	// 策略（devices 表）
	// Access = platform_admin 审批位（devices.access）。v1 双字段 AllowAccess(bool) + BusinessAccess(string='approved'|'pending')
	// 系同源同语义，合并为单一 bool，消除冗余表示。
	Access            bool `json:"access"`
	MonitoringEnabled bool `json:"monitoring_enabled,omitempty"`

	// 同步元（可选，用于 cardChange / 对账）
	Revision    int64 `json:"revision,omitempty"`
	UpdatedAtMS int64 `json:"updated_at_ms,omitempty"`
}

// SameRoomScopeKey 同房粗判：tenant|unit|room（unit 在 room 全局唯一时可传空 unit，由数据模型保证）。
func (b *DeviceBaseline) SameRoomScopeKey() string {
	if b == nil {
		return ""
	}
	return strings.TrimSpace(b.TenantID) + "|" + strings.TrimSpace(b.UnitID) + "|" + strings.TrimSpace(b.RoomID)
}
