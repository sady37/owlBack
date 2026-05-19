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

// ShortCodeOf — 6 位 base36 短码：SHA256(spatial_prefix) → mod 36^6 → 0-padded
// 用于 UI Card 显示 + DDNS FQDN（替代冗长的 hierarchy 编号）。
// 性质：
//   - 纯函数：同 spatial_prefix 永得同 code，无 DB 状态
//   - 空间 36^6 = 21.8 亿，万卡级 tenant 碰撞概率 < 1e-4（生日悖论）
//   - 不可逆：UI 显示用，不替代 spatial_prefix 主键
//
// 例: "fd00:0:3:111:3::/80" → "j0yvp3"
func ShortCodeOf(spatialPrefix string) string {
	if spatialPrefix == "" {
		return ""
	}
	h := sha256Sum([]byte(spatialPrefix))
	// 前 8 byte 作 uint64 big-endian
	var v uint64
	for i := 0; i < 8; i++ {
		v = (v << 8) | uint64(h[i])
	}
	const base36 = "0123456789abcdefghijklmnopqrstuvwxyz"
	const modulo = uint64(36 * 36 * 36 * 36 * 36 * 36) // 36^6
	v %= modulo
	out := []byte("000000")
	for i := 5; i >= 0; i-- {
		out[i] = base36[v%36]
		v /= 36
	}
	return string(out)
}

// sha256Sum — 简单包装，避免在 type 文件直接 import crypto（避免循环风险）。
// 实现见 short_code.go
func sha256Sum(b []byte) [32]byte { return sha256SumImpl(b) }

// UnitType 居住单位类型 — 与 units.unit_type SMALLINT 一一对应。
const (
	UnitTypeUnknown = 0
	UnitTypePrivate = 1
	UnitTypeShare   = 2
	UnitTypePublic  = 3
)

// UnitProperty 业务属性 — 与 units.unit_property SMALLINT 一一对应。B2C/B2B 分流由此决定。
const (
	UnitPropertyHome     = 0 // B2C
	UnitPropertyFacility = 1 // B2B
)

// UnitInfo Unit information
type UnitInfo struct {
	UnitID       string `json:"unit_id,omitempty"`
	UnitName     string `json:"unit_name,omitempty"`
	BranchID     string `json:"branch_id,omitempty"`
	BranchName   string `json:"branch_name,omitempty"`
	Building     string `json:"building,omitempty"`
	BuildingName string `json:"building_name,omitempty"` // sites.site_name (FE Detail 页 Branch/Building/Unit/Room/Bed 路径)
	RoomName     string `json:"room_name,omitempty"`     // card 锚点所在 room (room/bed 级 card 才非空)
	BedName      string `json:"bed_name,omitempty"`      // card 锚点所在 bed (bed 级 card 才非空)
	IsPublic     bool   `json:"is_public,omitempty"`
	IsSharedUnit bool   `json:"is_shared_unit,omitempty"`
	UnitType     int    `json:"unit_type,omitempty"`     // UnitType*: 0=Unknown, 1=Private, 2=Share, 3=Public
	UnitProperty int    `json:"unit_property,omitempty"` // UnitProperty*: 0=Home, 1=Facility
	Timezone     string `json:"timezone,omitempty"`      // IANA
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
	DNSShortName     string            `json:"dns_short_name,omitempty"` // 6 位 base36，URL 友好替代 IPv6 hoa（与 cards.dns_short_name 同语义）
	LastName         string            `json:"last_name,omitempty"`
	FirstName        string            `json:"first_name,omitempty"`
	Nickname         string            `json:"nickname,omitempty"`
	ServiceLevel     string            `json:"service_level,omitempty"`
	ServiceLevelInfo *ServiceLevelInfo `json:"service_level_info,omitempty"`
	BedID            *string           `json:"bed_id,omitempty"`
}

// CardStatic 卡片静态+动态视图（v2：基于 cards 表 + LPM 实时查询，非 JSONB snapshot）
// CardType v2 枚举：'tenant'|'branch'|'site'|'unit'|'public'|'room'|'bed'|'device'
type CardStatic struct {
	// 基础信息（cards 表权威）
	CardID        string `json:"card_id"`
	CardType      string `json:"card_type"`
	CardName      string `json:"card_name"`                 // 'nickname' 有人 / 'NoOne' 无人
	DNSShortName  string `json:"dns_short_name,omitempty"`  // v2: SHA256(spatial_prefix) → 6 位 base36（参 ShortCodeOf）；DDNS FQDN + UI Card 显示同源
	SpatialPrefix string `json:"spatial_prefix"`            // INET CIDR 字符串

	// 人类可读地址（按最长 mask 派生：Unit-Room-Bed）；空间结构变才变
	// v2: cards 表无 card_address 列，运行时 LEFT JOIN units/rooms/beds 拼出
	CardAddress string `json:"card_address,omitempty"` // e.g. "Unit 201 / Room A / Bed 1"

	// CoverageLabel — UI 卡行 2 自适应标签:
	//   bed card  → ""（FE 自行用 unit.room_name + bed_name 拼）
	//   room card → ""（FE 自行用 unit.room_name）
	//   unit card → 装的 device 跨 distinct room 数：
	//                1 room → 该 room name (如 "Bathroom")；≥2 → "Whole Unit"；0 → ""
	CoverageLabel string `json:"coverage_label,omitempty"`

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

const BedStateDurationNotSet int = -1

// BedState 在/离床状态。
// 注：OOB 不构成独立 risk（单源不够）；OOB 作为输入证据喂到 sensor fall 链路
// （BedsideFall / BedroomLostFall / Still fall），不再有 BedState.RiskLevel 字段。
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

// RiskLevel 风险等级 — RoomState.RiskLevel 用；FE 据此配色。
// 由 sensor zoneengine 按 room_type + risk-time + 多人陪伴评估。
const (
	RiskNormal    = 0 // 绿
	RiskMuted     = 1 // 灰
	RiskAttention = 2 // 黄
	RiskRisk      = 3 // 红
)

// RoomType 房间语义类型；只标"特殊高风险位置"——其余 room（bedroom/living/dining/lobby/etc.）
// 统一走 Default，BE 不细分。
//   - Default(0)：默认 standing-only risk 阈值
//   - Bathroom(1)：stay+standing 严苛阈值 + bathroom_fall + bathroom ghost adjudicator
//   - Kitchen(2)：占位给未来 lost-fall 例外（老人不下厨）；当前与 Default 同处理
//
// 注意：public 不是 room 类型，是 unit 类型（unit_type=3）；不要混进来。
const (
	RoomTypeDefault  = 0
	RoomTypeBathroom = 1
	RoomTypeKitchen  = 2
)

// RoomState per-physical-room (/88) 聚合状态。所有 kind 共用一个 shape，risk 语义由 Kind 区分。
//
// 注：unit /80 没有 unit_state — unit 卡 FE 自己按 /80 prefix 列子 /88 room，挑最近 UpdatedAt
// 的那个显示。sensor 内部 unit-scope risk（如 NightAbsence of Room）走自己的 counter，不入此 hash。
type RoomState struct {
	RoomType          int   `json:"room_type,omitempty"` // RoomType*: 0=Default, 1=Bathroom, 2=Kitchen
	UpdatedAt         int64 `json:"updated_at,omitempty"`
	TotalPeople       int   `json:"total_people"` // 当前证据推断的人数（radar number_people + sleepad in_bed），不是绝对真实总人数；监控盲区/识别误差会偏移
	LastEnterTime     int64 `json:"last_enter_time,omitempty"`
	LastExitTime      int64 `json:"last_exit_time,omitempty"`
	LastExitToOutside bool  `json:"last_exit_to_outside,omitempty"` // 最近 Vacant 由 EnterArea==outside 触发；不参与 SceneState 派生，仅留作 risk/alarm 原始信号
	StaySec           int   `json:"stay_sec,omitempty"`
	RiskLevel         int   `json:"risk_level,omitempty"` // 0=Normal 1=Muted 2=Attention 3=Risk；kind-specific 阈值 + night/multi-people 由 sensor 评估
}

// TargetState 单 Target 汇总（老人维度）。
//
// per-device 维度（v2 拍板 [[target_state_per_device]]）：sensor 每个 radar /128 维护自己
// 一份；cardagg max-merge 到单 card.target Hash 写入。
//
// StandingContinuousMin 2026-05-18 已从 RoomState 挪到此处：单 device 内（属同 radar 的物理
// 占用）的连续站立分钟，sensor 累加封顶 8；cardagg max-merge across devices in card。
type TargetState struct {
	UpdatedAt             int64  `json:"updated_at,omitempty"`
	TrackID               int    `json:"track_id"`
	LogicID               string `json:"logic_id,omitempty"`
	LastActiveTs          int64  `json:"last_active_ts,omitempty"`
	StandingContinuousMin int    `json:"standing_continuous_min,omitempty"`
	WeakBiometricSignal   int    `json:"weak_biometric_signal"`
	VisitorStartTs        int64  `json:"visitor_start_ts,omitempty"`
	TodayMaxVisitorMin    int    `json:"today_max_visitor_min,omitempty"`
	HasVisitorToday       bool   `json:"has_visitor_today,omitempty"`
}

// AlarmState 告警摘要（v2：cards 表无 alarm 列，counter/pop 由 alarm_events 实时聚合得出）
//
// 计算规则按 sensor_v2.md §6.7 三级告警状态机（决定 17）—— alarm_status 与级别分级处理：
//
//	┌──────────────────────┬─────────────────────────────────────────────────────────────┐
//	│ 级别                  │ 计入 Active* counter 的 alarm_status 集合                    │
//	├──────────────────────┼─────────────────────────────────────────────────────────────┤
//	│ Critical (Emerg/     │ {active, acked, auto_resolved}                              │
//	│  Alert/Crit, lvl0-2) │ — 必须人工 ack 才离开 Pending+AlarmBell；                    │
//	│                      │   auto_resolved 不离开（等待人工 handle）；                   │
//	│                      │   终态 = acked_auto_resolved → 不计入                        │
//	├──────────────────────┼─────────────────────────────────────────────────────────────┤
//	│ Error (lvl 3)        │ {active} 仅                                                  │
//	│ Warning (lvl 4)      │ {active} 仅                                                  │
//	│                      │ — auto_resolved 直接归 Resolved（不强制人工 ack）            │
//	└──────────────────────┴─────────────────────────────────────────────────────────────┘
//
// PopAlarm 选择规则（§6.7.2）：
//   - 仅看 alarm_status='active' 的行（acked/auto_resolved 不参与 popAlarm 选秀）
//   - 高级别 cover 低（level int ASC：Emerg=0 优先）
//   - 同级别 新 cover 旧（triggered_at DESC）
//   - 结果取 1 条 → "<LEVEL>.<event_type>"（如 "EMERG.Fall"）
//   - 无 active 行 → PopAlarm="" + EventID=""
//
// UI 渲染三件套（owlFront Overview.vue 消费）：
//   - PopAlarm bar：仅当 PopAlarm != "" 显示（active 才上 bar）
//   - AlarmBell 着色：由 Active* counter 任一 > 0 触发（Critical auto_resolved 保持 Bell 红）
//   - Pending 列表：详情页查询，覆盖 active + acked + auto_resolved（未到 acked_auto_resolved 终态）
type AlarmState struct {
	UpdatedAt   int64 `json:"updated_at,omitempty"`
	TriggeredAt int64 `json:"triggered_at,omitempty"`

	// Active*：未达 Resolved 终态的告警计数（Critical 含 acked/auto_resolved，Warning/Error 仅 active）
	ActiveEmerg   int `json:"active_emerg"`   // lvl 0 — Critical 语义
	ActiveAlert   int `json:"active_alert"`   // lvl 1 — Critical 语义
	ActiveCrit    int `json:"active_crit"`    // lvl 2 — Critical 语义
	ActiveErr     int `json:"active_err"`     // lvl 3 — auto_resolved 即归 Resolved，不计入
	ActiveWarning int `json:"active_warning"` // lvl 4 — auto_resolved 即归 Resolved，不计入

	// PopAlarm：仅 status=active 行参与；空字符串表示当前无 active 告警（bar 不显示）
	PopAlarm string `json:"pop_alarm"` // "<LEVEL>.<event_type>" 如 "EMERG.Fall" / ""
	EventID  string `json:"event_id"`  // popAlarm 对应 alarm_events.event_id（UI dedup）
}

// ========== CardDisplay：FE Overview dumb-render 契约 ==========
//
// cardagg 投影到 card:state.display Hash 子字段的显示态；FE 单读它就能渲 Overview，
// 不需要也不应读 room_state/bed_state/target/alarm_state（除 alarm_state.pop_alarm 用作
// Section1.up/down.right 自派生）。
//
// DisplayTime 算法：FE 用 `serverTimeRef - <Section>AnchorMs` 算时长，不用本地时钟。
// serverTimeRef 由 SSE `event: tick` 每 ~30s 推送的 server_ts 维护。
//
// 字段对应 UI 4 区：
//   - Section1.up.right / Section1.down.right：alarmBell + alarmEvent — 不在 display，FE 从 alarm_state 派生
//   - Section1.down.left：active room 名 — 不在 display，FE 仅 card.bed_id==null 时按业务渲（暂留空）
//   - Section2.left：Section2LeftMode + bed_status + sleep_stage / room_*
//   - Section2.middle/right：vital/posture — 不在 display，来自 realtime stream
//   - Section3.up.left：ActiveState + ActiveAnchorMs
//   - Section3.up.right：SceneState + SceneAnchorMs
//   - Section3.down：VisitorState + VisitorAnchorMs
//
// 详 owlBack/doc/card_display.md。

const (
	Section2LeftModeNone       = 0 // 无可显
	Section2LeftModeSleepStage = 1 // 显 sleep_stage / bed_status icon set
	Section2LeftModeRoomStatus = 2 // 显 room/bathroom icon set
)

const (
	RoomIconKindRoom     = 0
	RoomIconKindBathroom = 1
)

const (
	ActiveStateInactive = 0 // FE 渲 "Active <DisplayTime> ago"
	ActiveStateNow      = 1 // FE 渲 "Active now"（DisplayTime 不参与）
)

const (
	SceneStateOOR    = 0 // 房间空（total_people==0 + 无人在床）
	SceneStateInRoom = 1
	SceneStateInBath = 2 // 含 Stay / Stand 子状态由 SceneAnchorMs 携带
	SceneStateInBed  = 3
	SceneStateOOB    = 4 // 离床（在房）
)

const (
	VisitorStateNone  = 0 // "No visitor today"
	VisitorStateNow   = 1 // "Visitor now · <DisplayTime>"
	VisitorStateToday = 2 // "Visitor today · <DisplayTime>"
)

// VitalTrendLevel — Section3.down.right 横条配色（S4/W1 WeakBio 风险描述符 UI）。
// 由 cardagg card_display_builder 按 Target.WeakBiometricSignal score 阈值派生
// （详 [[target_state_weak_bio_signal_design]]）：
//
//	0-29  → None (hide 横条)
//	30-59 → Gray (Attention)
//	60-79 → Yellow (Watch)
//	80-100→ Red (Alert) — 不独立触发 alarm，由风险放大消费者按 ≥80 提级
const (
	VitalTrendLevelNone   = 0
	VitalTrendLevelGray   = 1
	VitalTrendLevelYellow = 2
	VitalTrendLevelRed    = 3
)

// CardDisplay — 写入 card:state hash 的 `display` JSON 字段。
type CardDisplay struct {
	UpdatedAt int64 `json:"updated_at"`

	// Section2.left
	Section2LeftMode int `json:"section2_left_mode"`           // Section2LeftMode*
	BedStatus        int `json:"bed_status,omitempty"`         // 0=InBed, 1=NotInBed, 8=Unknown；FE 仅在 CardStatic.bed_id != null 时消费
	SleepStage       int `json:"sleep_stage,omitempty"`        // 0/1/2/4/8 (复用 SleepStage*；仅 BedStatus=0 时有效)
	RoomPersonCount  int `json:"room_person_count,omitempty"`  // 右上角 badge
	RoomIconKind     int `json:"room_icon_kind,omitempty"`     // RoomIconKind*
	RoomRiskLevel    int `json:"room_risk_level,omitempty"`    // 0/1/2/3 → FE 配色（复用 RiskLevel*）

	// Section3.up.left
	ActiveState    int   `json:"active_state"`
	ActiveAnchorMs int64 `json:"active_anchor_ms,omitempty"`

	// Section3.up.right
	SceneState    int   `json:"scene_state"`
	SceneAnchorMs int64 `json:"scene_anchor_ms,omitempty"`

	// Section3.down.left (Visitor / Bed timing)
	VisitorState    int   `json:"visitor_state"`
	VisitorAnchorMs int64 `json:"visitor_anchor_ms,omitempty"`

	// Section3.down.right (WeakBio 横条配色；VitalTrendLevel* 0=hide/1=Gray/2=Yellow/3=Red)
	VitalTrendLevel int `json:"vital_trend_level,omitempty"`
}

// CardStatus 卡片状态数据
//
// Display 字段 by cardagg card_display_projector 投影；FE Overview 卡列表 dumb render 读此字段，
// 不再派生 picker/scene/tier。RoomState/BedState 给 Detail/RadarCanvas 等诊断视图保留。
type CardStatus struct {
	CardID     string                 `json:"card_id"`
	Target     *TargetState           `json:"target,omitempty"`
	RoomState  *RoomState             `json:"room_state,omitempty"`
	BedState   *BedState              `json:"bed_state,omitempty"`
	AlarmState *AlarmState            `json:"alarm_state,omitempty"`
	Display    *CardDisplay           `json:"display,omitempty"`
	Message    map[string]interface{} `json:"message,omitempty"`
}

// ========== v2 Card Type 枚举（与 cards.card_type CHECK 一致） ==========

const (
	CardTypeTenant = "tenant" // /48
	CardTypeBranch = "branch" // /56
	CardTypeSite   = "site"   // /64
	CardTypeUnit   = "unit"   // /80
	CardTypePublic = "public" // /80 公共区域
	CardTypeRoom   = "room"   // /88
	CardTypeBed    = "bed"    // /96 ⭐ 居住空间最细层
	CardTypeDevice = "device" // /128 fallback
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
	case CardTypeBed:
		return 96
	case CardTypeDevice:
		return 128
	}
	return 0
}

// CardTypeForMasklen 反向解析（/96 默认 bed；/80 默认 unit；调用方按业务自选 public）
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
		return CardTypeBed
	case 128:
		return CardTypeDevice
	}
	return ""
}
