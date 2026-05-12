package card

// ========== ID-Name 标识符结构体（避免重复定义）==========

// BedIdentifier 床标识符（ID-Name对）
type BedIdentifier struct {
	BedID   string `json:"bed_id"`
	BedName string `json:"bed_name,omitempty"`
}

// RoomIdentifier 房间标识符（ID-Name对）
type RoomIdentifier struct {
	RoomID   string `json:"room_id"`
	RoomName string `json:"room_name,omitempty"`
}

// UnitIdentifier 单元标识符（ID-Name对）
type UnitIdentifier struct {
	UnitID   string `json:"unit_id"`
	UnitName string `json:"unit_name,omitempty"`
}

// BranchIdentifier 院区标识符（ID-Name对）
type BranchIdentifier struct {
	BranchID   string `json:"branch_id"`
	BranchName string `json:"branch_name"`
}

type DeviceIdentifier struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name,omitempty"`
}

// ========== 静态数据结构体 ==========

// ActiveBedRow 活跃床位信息（用于卡片创建）
type ActiveBedRow struct {
	BedID      string  `json:"bed_id"`
	BedName    string  `json:"bed_name"`
	RoomName   string  `json:"room_name"`
	ResidentID *string `json:"resident_id,omitempty"`
}

// UnitInfo Unit information
type UnitInfo struct {
	UnitID       string `json:"unit_id,omitempty"`
	UnitName     string `json:"unit_name,omitempty"`
	BranchID     string `json:"branch_id,omitempty"`
	BranchName   string `json:"branch_name,omitempty"`
	Building     string `json:"building,omitempty"`
	IsPublic     bool   `json:"is_public,omitempty"`
	IsSharedUnit bool   `json:"is_shared_unit,omitempty"`
	UnitType     string `json:"unit_type,omitempty"` // "facility" | "home"
	Timezone     string `json:"timezone,omitempty"`  // IANA
}

// DeviceInfo device information（来自 devices/device_store JOIN，非 JSONB snapshot）
type DeviceInfo struct {
	DeviceID          string  `json:"device_id"`
	DeviceUID         string  `json:"-"`           // HIPAA 不向前端暴露
	DeviceCode        string  `json:"device_code"` // device_store.device_code
	DeviceName        string  `json:"device_name"`
	DeviceType        string  `json:"device_type"`
	DeviceModel       string  `json:"device_model"`
	UnitID            string  `json:"unit_id"`
	BoundBedID        *string `json:"bound_bed_id,omitempty"`
	BoundRoomID       *string `json:"bound_room_id,omitempty"`
	MonitoringEnabled bool    `json:"monitoring_enabled"`
	Status            string  `json:"status"` // "online" | "offline" | "error" | "disabled"
}

type ServiceLevelInfo struct {
	LevelCode     string `json:"level_code"`
	DisplayName   string `json:"display_name"`
	DisplayNameCN string `json:"display_name_cn,omitempty"`
	ColorTag      string `json:"color_tag"`
	ColorHex      string `json:"color_hex"`
	Priority      int    `json:"priority"`
}

// ResidentInfo resident information
type ResidentInfo struct {
	ResidentID       string            `json:"resident_id"`
	LastName         string            `json:"last_name,omitempty"`
	FirstName        string            `json:"first_name,omitempty"`
	Nickname         string            `json:"nickname,omitempty"`
	ServiceLevel     string            `json:"service_level,omitempty"`
	ServiceLevelInfo *ServiceLevelInfo `json:"service_level_info,omitempty"`
	BedID            *string           `json:"bed_id,omitempty"`
}

// CardStatic 卡片静态+动态视图（v2：基于 cards 表 + LPM 实时查询，非 JSONB snapshot）
// CardType v2 枚举：'tenant'|'branch'|'site'|'unit'|'public'|'room'|'active_bed'|'device'
type CardStatic struct {
	// 基础信息（cards 表权威）
	CardID        string `json:"card_id"`
	CardType      string `json:"card_type"`
	CardName      string `json:"card_name"`
	DNSShortName  string `json:"dns_short_name,omitempty"` // v2 永久名（bed-stable）
	SpatialPrefix string `json:"spatial_prefix"`            // INET CIDR 字符串

	// 同一 card 必属同一 unit/branch（reverse-derive via LPM）
	Unit *UnitInfo `json:"unit,omitempty"`

	// 房间/床位（reverse-derive via LPM）
	Rooms   []RoomIdentifier `json:"rooms,omitempty"`
	BedID   *string          `json:"bed_id,omitempty"`
	BedName *string          `json:"bed_name,omitempty"`

	// 住户和设备（实时查询，非 JSONB snapshot）
	Residents []ResidentInfo `json:"residents"`
	Devices   []DeviceInfo   `json:"devices"`

	// 护理人员（从 resident_caregivers 聚合）
	CaregiverGroups []string        `json:"caregiver_groups,omitempty"`
	Caregivers      []CaregiverInfo `json:"caregivers,omitempty"`

	// 报警显示控制（阈值 = syslog 级别；图标/弹出条件）
	IconAlarmLevel *int `json:"icon_alarm_level,omitempty"`
	PopAlarm       *int `json:"pop_alarm,omitempty"`
}

// CaregiverInfo 护理人员信息
type CaregiverInfo struct {
	UserID      string `json:"user_id"`
	Nickname    string `json:"nickname,omitempty"`
	UserAccount string `json:"user_account,omitempty"`
	Role        string `json:"role,omitempty"`
}

// ========== Realtime Monitor 聚合（card:realtime:stream） ==========

// TrackFields 单条 track 的平铺字段（仅含有效字段 + ts）
type TrackFields map[string]interface{}

// DeviceRealtimeTracks 单设备 monitor 流：track_id (str) -> TrackFields
type DeviceRealtimeTracks map[string]TrackFields

// CardRealTime 卡片实时数据（monitor 格式）
type CardRealTime struct {
	CardID    string                          `json:"card_id"`
	Timestamp int64                           `json:"timestamp,omitempty"`
	Devices   map[string]DeviceRealtimeTracks `json:"devices,omitempty"` // device_uid -> track_id -> fields+ts
}

// ========== Card Status (card:state / card:status:stream) ==========

// DeviceStatus 单个设备运行时真相（独立 Hash device:status:{device_id}）
type DeviceStatus struct {
	DeviceUID  string `json:"-"`
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`

	UpdatedAt  int64 `json:"updated_at,omitempty"`
	LastSeenMs int64 `json:"last_seen_ms,omitempty"`

	Offline        int `json:"offline"`
	SignalPoor     int `json:"signal_poor,omitempty"`
	AngleAbnormal  int `json:"angle_abnormal,omitempty"`
	SensorDetached int `json:"sensor_detached,omitempty"`
}

// BedState 在/离床状态
const BedStateDurationNotSet int = -1

type BedState struct {
	UpdatedAt       int64 `json:"updated_at,omitempty"`
	BedStatus       int   `json:"bed_status"`
	TrackNumber     int   `json:"track_number"`
	StartTime       int64 `json:"start_time,omitempty"`
	DurationSec     int   `json:"duration_sec"`
	BedConfidence   int   `json:"bed_confidence"`
	BedEvent        int   `json:"bed_event"`
	SleepStage      int   `json:"sleep_stage"`
	SleepConfidence int   `json:"sleep_confidence"`
}

const (
	SleepStageInitial = 0
	SleepStageAwake   = 1
	SleepStageLight   = 2
	SleepStageDeep    = 4
	SleepStageUnknown = 8
)

// RoomState 房间空间层
type RoomState struct {
	UpdatedAt             int64          `json:"updated_at,omitempty"`
	TotalPeople           int            `json:"total_people"`
	AreaPeople            map[string]int `json:"area_people,omitempty"`
	LastEnterTime         int64          `json:"last_enter_time,omitempty"`
	LastExitTime          int64          `json:"last_exit_time,omitempty"`
	StandingContinuousMin int            `json:"standing_continuous_min,omitempty"`
	HasMulti              bool           `json:"has_multi"`
	HasRisk               bool           `json:"has_risk"`
}

// BathRoomState 卫生间空间层
type BathRoomState struct {
	DeviceUID             string `json:"-"`
	DeviceID              string `json:"device_id,omitempty"`
	RoomID                string `json:"room_id,omitempty"`
	RoomName              string `json:"room_name,omitempty"`
	UpdatedAt             int64  `json:"updated_at,omitempty"`
	TotalPeople           int    `json:"total_people"`
	LastEnterTime         int64  `json:"last_enter_time,omitempty"`
	LastExitTime          int64  `json:"last_exit_time,omitempty"`
	StaySec               int    `json:"stay_sec,omitempty"`
	StandingContinuousMin int    `json:"standing_continuous_min,omitempty"`
	HasMulti              bool   `json:"has_multi"`
	HasRisk               bool   `json:"has_risk"`
	StayFSMPhase          string `json:"stay_fsm_phase,omitempty"`
	StayArmEnterAt        int64  `json:"stay_arm_enter_at,omitempty"`
	StayResolveExitAt     int64  `json:"stay_resolve_exit_at,omitempty"`
}

// TargetState 单 Target 汇总（老人维度）
type TargetState struct {
	UpdatedAt           int64  `json:"updated_at,omitempty"`
	TrackID             int    `json:"track_id"`
	LogicID             string `json:"logic_id,omitempty"`
	LastActiveTs        int64  `json:"last_active_ts,omitempty"`
	WeakBiometricSignal int    `json:"weak_biometric_signal"`
	VisitorStartTs      int64  `json:"VisitorStartTs,omitempty"`
	TodayMaxVisitorMin  int    `json:"TodayMaxVisitorMin,omitempty"`
	HasVisitorToday     bool   `json:"HasVisitorToday,omitempty"`
}

// AlarmState 告警摘要（v2：cards 表无 alarm 列，counter/pop 由 alarm_events 实时聚合得出）
type AlarmState struct {
	UpdatedAt     int64 `json:"updated_at,omitempty"`
	TriggeredAt   int64 `json:"triggered_at,omitempty"`
	ActiveEmerg   int   `json:"active_emerg"`
	ActiveAlert   int   `json:"active_alert"`
	ActiveCrit    int   `json:"active_crit"`
	ActiveErr     int   `json:"active_err"`
	ActiveWarning int   `json:"active_warning"`
	PopAlarm      string `json:"pop_alarm"` // "EMERG.Fall" / ""
	EventID       string `json:"event_id"`
}

// CardStatus 卡片状态数据
type CardStatus struct {
	CardID        string                 `json:"card_id"`
	Target        *TargetState           `json:"target,omitempty"`
	RoomState     *RoomState             `json:"room_state,omitempty"`
	BathRoomState *BathRoomState         `json:"bathroom_state,omitempty"`
	BedState      *BedState              `json:"bed_state,omitempty"`
	AlarmState    *AlarmState            `json:"alarm_state,omitempty"`
	Message       map[string]interface{} `json:"message,omitempty"`
}

// ========== v2 Card Type 枚举（与 cards.card_type CHECK 一致） ==========

const (
	CardTypeTenant    = "tenant"     // /48
	CardTypeBranch    = "branch"     // /56
	CardTypeSite      = "site"       // /64
	CardTypeUnit      = "unit"       // /80
	CardTypePublic    = "public"     // /80 公共区域
	CardTypeRoom      = "room"       // /88
	CardTypeActiveBed = "active_bed" // /96 ⭐ 最常见
	CardTypeDevice    = "device"     // /128 fallback
)

// MasklenForCardType v2 cards.card_type ↔ masklen 强绑定（CHECK 约束）
func MasklenForCardType(cardType string) int {
	switch cardType {
	case CardTypeTenant:
		return 48
	case CardTypeBranch:
		return 56
	case CardTypeSite:
		return 64
	case CardTypeUnit, CardTypePublic:
		return 80
	case CardTypeRoom:
		return 88
	case CardTypeActiveBed:
		return 96
	case CardTypeDevice:
		return 128
	}
	return 0
}

// CardTypeForMasklen 反向解析（/96 默认为 active_bed；/80 默认为 unit；调用方按业务自选 public）
func CardTypeForMasklen(masklen int) string {
	switch masklen {
	case 48:
		return CardTypeTenant
	case 56:
		return CardTypeBranch
	case 64:
		return CardTypeSite
	case 80:
		return CardTypeUnit
	case 88:
		return CardTypeRoom
	case 96:
		return CardTypeActiveBed
	case 128:
		return CardTypeDevice
	}
	return ""
}
