package domain

import (
	"database/sql"
	"time"
)

// Card 卡片领域模型（v2: 对应新 cards 表 schema）
//
// 三层身份：
//   - spatial_prefix INET — 业务身份（北极星）
//   - card_id UUID        — DB PK 稳定性 / FK
//   - dns_short_name      — 永久人类可读名（bed-stable，无 PHI）
//
// 删除的 v1 字段：tenant_id / branch_id / bed_id / unit_id / card_address / timezone /
// devices JSONB / residents JSONB / unhandled_alarm_0..4 / pop_alarm_level / pop_alarm_type /
// pop_alarm_event_id（counter / pop 由 alarm_events 实时聚合）。
type Card struct {
	CardID        string         `db:"card_id"`        // UUID PK
	SpatialPrefix string         `db:"spatial_prefix"` // INET CIDR 字符串（业务身份；mask 决定 card_type）
	CardType      string         `db:"card_type"`      // 'tenant'|'branch'|'site'|'unit'|'public'|'room'|'bed'|'device'
	CardName      sql.NullString `db:"card_name"`
	DNSShortName  sql.NullString `db:"dns_short_name"` // 永久 DNS 名，如 "u42-r03-b01.tenant1.owl"
	ResidentID    sql.NullString `db:"resident_id"`    // INET HoA；NULL = 空床
	IsActive      bool           `db:"is_active"`
	EnabledAt     sql.NullTime   `db:"enabled_at"`
	DisabledAt    sql.NullTime   `db:"disabled_at"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

// CardWithUnitInfo 卡片及其关联的 Unit 信息（用于 Repository 层返回）
// v2: Unit 信息由 spatial_prefix 通过 unit_id (INET /80) join units 表反查；
//     不再随 cards 表共存。
type CardWithUnitInfo struct {
	Card *Card
	Unit *Unit
}

// CardOverviewItem 卡片概览项（v2: 字段全部基于 LPM 实时聚合，非 JSONB snapshot）
type CardOverviewItem struct {
	// 基础信息
	CardID        string  `json:"card_id"`
	SpatialPrefix string  `json:"spatial_prefix"`         // INET CIDR
	CardType      string  `json:"card_type"`              // 'bed' | 'unit' | 'public' | ...
	CardName      string  `json:"card_name"`
	DNSShortName  string  `json:"dns_short_name,omitempty"`
	ResidentID    *string `json:"resident_id,omitempty"`

	// Unit 信息（LPM 派生）
	UnitID        string `json:"unit_id,omitempty"`
	UnitType      string `json:"unit_type"`
	IsPublicSpace bool   `json:"is_public_space"`
	IsSharedUnit  bool   `json:"is_shared_unit"`

	// 聚合数据（来自 view + 实时 query）
	Devices   []CardDevice   `json:"devices"`
	Residents []CardResident `json:"residents"`

	// 告警计数（来自 alarm_events 实时聚合，参 view card_alarm_summary）
	// v2: cards 表已无 counter/pop 列；这些字段由 service 层 query view 填充。
	UnhandledAlarm0 int `json:"unhandled_alarm_0"`
	UnhandledAlarm1 int `json:"unhandled_alarm_1"`
	UnhandledAlarm2 int `json:"unhandled_alarm_2"`
	UnhandledAlarm3 int `json:"unhandled_alarm_3"`
	UnhandledAlarm4 int `json:"unhandled_alarm_4"`
	IconAlarmLevel  int `json:"icon_alarm_level"`
	PopAlarm        int `json:"pop_alarm"`

	// 权限
	ResidentAccess bool `json:"resident_access"`

	// 护理人员
	CaregiverGroups []CaregiverGroup `json:"caregiver_groups,omitempty"`
	Caregivers      []CardCaregiver  `json:"caregivers,omitempty"`

	// 计数
	DeviceCount    int `json:"device_count"`
	ResidentCount  int `json:"resident_count"`
	CaregiverCount int `json:"caregiver_count"`
}

// CardDevice 卡片关联的设备信息
type CardDevice struct {
	DeviceID    string      `json:"device_id"`
	UID         string      `json:"uid"`
	DeviceUID   string      `json:"device_uid,omitempty"`
	DeviceCode  string      `json:"device_code,omitempty"`
	DeviceName  string      `json:"device_name"`
	DeviceType  interface{} `json:"device_type"`
	DeviceModel string      `json:"device_model,omitempty"`
	Status      string      `json:"status,omitempty"`
}

// CardResident 卡片关联的住户信息
type CardResident struct {
	ResidentID   string `json:"resident_id"`
	LastName     string `json:"last_name,omitempty"`
	Nickname     string `json:"nickname"`
	ServiceLevel string `json:"service_level,omitempty"`
}

// CardCaregiver 卡片关联的护理人员信息
type CardCaregiver struct {
	UserID        string `json:"user_id"`
	Nickname      string `json:"nickname"`
	UserAccount   string `json:"user_account"`
	UserRole      string `json:"user_role"`
	UserBranchTag string `json:"user_branch_tag,omitempty"`
}

// CaregiverGroup 护理人员组信息
type CaregiverGroup struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}
