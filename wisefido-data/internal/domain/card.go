package domain

import (
	"database/sql"
	"encoding/json"
)

// Card 卡片领域模型（对应 cards 表）
type Card struct {
	// 主键
	CardID string `db:"card_id"` // UUID, PRIMARY KEY

	// 租户
	TenantID string `db:"tenant_id"` // UUID, NOT NULL

	// 卡片类型
	CardType string `db:"card_type"` // VARCHAR(20), NOT NULL, 'ActiveBedCard' | 'UnitCard'

	// 绑定目标（取决于 card_type）
	BedID  sql.NullString `db:"bed_id"`  // UUID, nullable (ActiveBedCard 卡片)
	UnitID sql.NullString `db:"unit_id"` // UUID, nullable (UnitCard 卡片)

	// 显示信息
	CardName    string `db:"card_name"`    // VARCHAR(255), NOT NULL
	CardAddress string `db:"card_address"` // VARCHAR(255), NOT NULL
	Timezone    string `db:"timezone"`     // IANA，如 "America/Los_Angeles"

	// 主要住户（ActiveBedCard
	ResidentID sql.NullString `db:"resident_id"` // UUID, nullable

	// 预计算的关联（应用层维护）
	Devices   json.RawMessage `db:"devices"`   // JSONB, NOT NULL, 设备 ID 数组
	Residents json.RawMessage `db:"residents"` // JSONB, NOT NULL, 住户 ID 数组

	// 未处理告警计数器（应用层维护）
	UnhandledAlarm0 int `db:"unhandled_alarm_0"` // INTEGER, NOT NULL, DEFAULT 0
	UnhandledAlarm1 int `db:"unhandled_alarm_1"` // INTEGER, NOT NULL, DEFAULT 0
	UnhandledAlarm2 int `db:"unhandled_alarm_2"` // INTEGER, NOT NULL, DEFAULT 0
	UnhandledAlarm3 int `db:"unhandled_alarm_3"` // INTEGER, NOT NULL, DEFAULT 0
	UnhandledAlarm4 int `db:"unhandled_alarm_4"` // INTEGER, NOT NULL, DEFAULT 0

	// 当前弹出报警（应用层维护）
	PopAlarmLevel   string         `db:"pop_alarm_level"`    // VARCHAR(20), 'EMERG','ALERT' etc
	PopAlarmType    string         `db:"pop_alarm_type"`     // VARCHAR(50), 'Fall','HeartRateAlert' etc
	PopAlarmEventId sql.NullString `db:"pop_alarm_event_id"` // UUID, nullable

	// UI 告警阈值
	IconAlarmLevel   int `db:"icon_alarm_level"`   // INTEGER, NOT NULL, DEFAULT 2 (<=2红色: EMERG/ALERT/CRITICAL)
	PopAlarm int `db:"pop_alarm"` // INTEGER, NOT NULL, DEFAULT 2 (<=2弹出: EMERG/ALERT/CRITICAL)
}

// CardWithUnitInfo 卡片及其关联的 Unit 信息（用于 Repository 层返回）
type CardWithUnitInfo struct {
	Card *Card
	Unit *Unit // 关联的 Unit 信息（用于权限过滤和 family_view 计算）
}

// CardOverviewItem 卡片概览项（用于 Service 层返回给 Handler）
// 注意：字段命名使用 snake_case（json tag），与 owlFront cardOverviewModel.ts 对齐
type CardOverviewItem struct {
	// 基础信息
	CardID      string  `json:"card_id"`
	TenantID    string  `json:"tenant_id"`
	CardType    string  `json:"card_type"` // "ActiveBedCard" | "UnitCard"
	BedID       *string `json:"bed_id,omitempty"`
	UnitID      *string `json:"unit_id,omitempty"`
	CardName    string  `json:"card_name"`
	CardAddress string  `json:"card_address"`
	ResidentID  *string `json:"resident_id,omitempty"`

	// Unit 信息
	UnitType      string `json:"unit_type"` // "Home" | "Facility"
	IsPublicSpace bool   `json:"is_public_space"`
	IsSharedUnit  bool   `json:"is_shared_unit"`

	// 聚合数据
	Devices   []CardDevice   `json:"devices"`
	Residents []CardResident `json:"residents"`

	// 告警信息
	UnhandledAlarm0  int `json:"unhandled_alarm_0"`
	UnhandledAlarm1  int `json:"unhandled_alarm_1"`
	UnhandledAlarm2  int `json:"unhandled_alarm_2"`
	UnhandledAlarm3  int `json:"unhandled_alarm_3"`
	UnhandledAlarm4  int `json:"unhandled_alarm_4"`
	IconAlarmLevel   int `json:"icon_alarm_level"`
	PopAlarm int `json:"pop_alarm"`

	// 权限相关
	ResidentAccess bool `json:"resident_access"` // 是否允许住户访问

	// 护理人员相关（可选，如果实现）
	CaregiverGroups []CaregiverGroup `json:"caregiver_groups,omitempty"`
	Caregivers      []CardCaregiver  `json:"caregivers,omitempty"`

	// 计数字段（用于快速显示）
	DeviceCount    int `json:"device_count"`
	ResidentCount  int `json:"resident_count"`
	CaregiverCount int `json:"caregiver_count"`
}

// CardDevice 卡片关联的设备信息
type CardDevice struct {
	DeviceID    string      `json:"device_id"`             // device_id (UUID, 主键)
	UID         string      `json:"uid"`                   // 兼容旧字段，与 device_uid 同义
	DeviceUID   string      `json:"device_uid,omitempty"`  // 设备唯一标识，与 vue CardOverviewDevice 对齐
	DeviceCode  string      `json:"device_code,omitempty"` // 设备编码：Sleepace 有 device_code，Radar 无；接口处空则回退为 device_uid（每设备均有）
	DeviceName  string      `json:"device_name"`
	DeviceType  interface{} `json:"device_type"` // 数字（1=sleepace, 2=radar）或字符串（向后兼容）
	DeviceModel string      `json:"device_model,omitempty"`
	Status      string      `json:"status,omitempty"` // "online" | "offline" | "error" | "disabled"
}

// CardResident 卡片关联的住户信息
type CardResident struct {
	ResidentID   string `json:"resident_id"`
	LastName     string `json:"last_name,omitempty"` // 住户姓氏（从residents表获取）
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

// CaregiverGroup 护理人员组信息（从resident_caregiver_groups表获取）
type CaregiverGroup struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}
