package card

// ----
// CardWithContent card with devices and residents as structured data (for comparison)
type CardWithContent struct {
	CardID      string
	CardType    string
	BedID       *string
	UnitID      string
	CardName    string
	CardAddress string
	Timezone    string
	ResidentID  *string
	Devices     []DeviceInfo   `json:"devices,omitempty"`
	Residents   []ResidentInfo `json:"residents,omitempty"`
}

// ExpectedCard represents an expected card as structured data (for comparison, without card_id)
type ExpectedCard struct {
	CardType    string
	BedID       *string
	UnitID      string
	CardName    string
	CardAddress string
	Timezone    string
	ResidentID  *string
	Devices     []DeviceInfo   `json:"devices,omitempty"`
	Residents   []ResidentInfo `json:"residents,omitempty"`
}

// CardUpdateStats statistics for card updates
type CardUpdateStats struct {
	ExistingCount  int // Number of existing cards before update
	DeletedCount   int // Number of cards deleted
	CreatedCount   int // Number of cards created
	UpdatedCount   int // Number of cards updated (deleted + created)
	UnchangedCount int // Number of cards that remained unchanged
}

// ========== ID-Name 标识符结构体（避免重复定义）==========

// BedIdentifier 床标识符（ID-Name对）
type BedIdentifier struct {
	BedID   string `json:"bed_id"`
	BedName string `json:"bed_name,omitempty"` // 床名称
}

// RoomIdentifier 房间标识符（ID-Name对）
type RoomIdentifier struct {
	RoomID   string `json:"room_id"`
	RoomName string `json:"room_name,omitempty"` // 房间名称
}

// UnitIdentifier 单元标识符（ID-Name对）
type UnitIdentifier struct {
	UnitID   string `json:"unit_id"`
	UnitName string `json:"unit_name,omitempty"` // 单元名称
}

// BranchIdentifier 院区标识符（ID-Name对）
type BranchIdentifier struct {
	BranchID   string `json:"branch_id"`
	BranchName string `json:"branch_name"` // 院区名称
}

type DeviceIdentifier struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name,omitempty"` // 设备名称
}

// ========== 静态数据结构体 ==========

// ActiveBedRow 活跃床位信息（用于卡片创建）
// 包含床位标识、名称和关联的住户信息
type ActiveBedRow struct {
	BedID      string  `json:"bed_id"`                // 床位 ID
	BedName    string  `json:"bed_name"`              // 床位名称
	RoomName   string  `json:"room_name"`             // 房间名称
	ResidentID *string `json:"resident_id,omitempty"` // 住户 ID（可为空）
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
	Timezone     string `json:"timezone,omitempty"`  // IANA, e.g. "America/Los_Angeles"
}

// DeviceInfo device information
type DeviceInfo struct {
	DeviceID          string  `json:"device_id"`
	DeviceUID         string  `json:"-"`           // devices.device_uid，内部用，不向前端暴露（HIPAA）
	DeviceCode        string  `json:"device_code"` // device_store.device_code
	DeviceName        string  `json:"device_name"`
	DeviceType        string  `json:"device_type"` // 或字符串，JSON 兼容
	DeviceModel       string  `json:"device_model"`
	UnitID            string  `json:"unit_id"` // 所属unit
	BoundBedID        *string `json:"bound_bed_id,omitempty"`
	BoundRoomID       *string `json:"bound_room_id,omitempty"` // Room ID where device is bound
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

// CardStatic 卡片静态+动态；同一张 card 必属同一 unit、branch，以下仅存一份。
type CardStatic struct {
	// 基础信息（来自 cards 表）
	CardID      string `json:"card_id"`
	CardType    string `json:"card_type"` // "ActiveBedCard" 或 "UnitCard"
	CardName    string `json:"card_name"`
	CardAddress string `json:"card_address"`
	TenantID    string `json:"tenant_id"`

	// 同一 card 必属同一 unit/branch，只存一份
	Unit *UnitInfo `json:"unit,omitempty"`

	// 一个卡片可能关联多个房间
	Rooms   []RoomIdentifier `json:"rooms,omitempty"`    // 房间列表（可能有多个房间）
	BedID   *string          `json:"bed_id,omitempty"`   // 仅 ActiveBedCard 有值，UnitCard 为空
	BedName *string          `json:"bed_name,omitempty"` // 床名称（仅 ActiveBedCard 有值，UnitCard 为空）

	// 住户和设备（来自 cards.residents 和 cards.devices JSONB）
	Residents []ResidentInfo `json:"residents"`
	Devices   []DeviceInfo   `json:"devices"`

	// 护理人员（从 resident_caregivers 聚合）
	CaregiverGroups []string        `json:"caregiver_groups,omitempty"` // 护理组标签（去重）
	Caregivers      []CaregiverInfo `json:"caregivers,omitempty"`       // 指定护理人员

	// 报警显示控制（来自 cards 表，值=syslog级别阈值，<=该值的报警触发对应行为）
	// 图标变红阈值：默认2 → EMERG(0)/ALERT(1)/CRITICAL(2)红色，ERR(3)/WARNING(4)橙色
	IconAlarmLevel *int `json:"icon_alarm_level,omitempty"`
	// 弹出报警阈值：默认2 → EMERG(0)/ALERT(1)/CRITICAL(2)弹出
	PopAlarm *int `json:"pop_alarm,omitempty"`
}

// CaregiverInfo 护理人员信息（从 users 表查询）
type CaregiverInfo struct {
	UserID      string `json:"user_id"`
	Nickname    string `json:"nickname,omitempty"`
	UserAccount string `json:"user_account,omitempty"`
	Role        string `json:"role,omitempty"`
}


// ========== Realtime Monitor 聚合（card:realtime:stream） ==========
//{ card_id, "<device_id>": { "<track_id>": {...}, ... } }
/*
card1:device1:{track0(有效值）,track1,track2..}
card1:device2:{track1(有效值）}
card2:device3:{track2(有效值）}

type Track struct {
	TrackID  int    `json:"track_id"`  // Radar 0-8, Sleepad 0=left
	LogicID  string `json:"logic_id,omitempty"` // Card 分配的逻辑目标 ID
	Ts       int64  `json:"ts"`

	PositionX       *int   `json:"position_x,omitempty"`
	PositionY       *int   `json:"position_y,omitempty"`
	PositionZ       *int   `json:"position_z,omitempty"`
	AreaType        string `json:"area_type,omitempty"`
	TrackConfidence int    `json:"track_confidence,omitempty"`

	Pose           int `json:"pose,omitempty"`
	PoseConfidence int `json:"pose_confidence,omitempty"`

	HeartRate        int     `json:"heart_rate,omitempty"`
	RespiratoryRate  int     `json:"respiratory_rate,omitempty"`
	BloodOxygen      int     `json:"blood_oxygen,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	VitalConfidence  int     `json:"vital_confidence,omitempty"`

	// 周期≤5min 上报视为实时；Sleepad 持续更新，保留供后续使用
	BedStatus int `json:"bed_status,omitempty"`

	BodyMove         int `json:"body_move,omitempty"`
	TurnOver         int `json:"turn_over,omitempty"`
	SitUp            int `json:"sit_up,omitempty"`
	MotionConfidence int `json:"motion_confidence,omitempty"`
}
*/

// TrackFields 单条 track 的平铺字段 + "ts"（仅含有效字段）；与 observation.TrackID 约定一致，88 不入库
type TrackFields map[string]interface{}

// DeviceRealtimeTracks 单设备在 monitor 流中的格式：track_id（如 "0","9"）-> TrackFields
// cardagg Flush 输出 device_uid -> 此类型，wisefido-data 原样缓存
type DeviceRealtimeTracks map[string]TrackFields

// CardRealTime 卡片实时数据（monitor 格式）
// 发布到 Redis Stream: card:realtime:stream（type=monitor）
// 缓存/下发：card_id -> { card_id, device_uid -> DeviceRealtimeTracks }
type CardRealTime struct {
	CardID    string                          `json:"card_id"`
	Timestamp int64                           `json:"timestamp,omitempty"`
	Devices   map[string]DeviceRealtimeTracks `json:"devices,omitempty"` // device_uid -> track_id -> fields+ts
}

// ========== Card Status (card:state / card:status:stream) ==========

// DeviceStatus 单个设备在线状态（仅 offline；其它如 signal_poor 走告警）
type DeviceStatus struct {
	DeviceUID  string `json:"-"`                   // 内部/Redis 用，不向前端暴露（HIPAA）
	DeviceID   string `json:"device_id"`            // 前端展示用
	DeviceType string `json:"device_type"`          // "Radar" | "Sleepad"
	UpdatedAt  int64  `json:"updated_at,omitempty"` // 最后更新时间（Unix 毫秒），与 TargetState/AlarmState 一致
	Offline    int    `json:"offline"`              // 0=在线, 1=离线
}

// BedState 在/离床状态（护理者最核心的二元判断）
// UpdatedAt/StartTime 未设置可为 0；仅 DurationSec 用 -1 表示未设置（0 秒为合法值）。
const BedStateDurationNotSet int = -1 // DurationSec 未设置

type BedState struct {
	UpdatedAt        int64 `json:"updated_at,omitempty"`   // 本次状态变化时间（毫秒），未设置=0
	BedStatus        int   `json:"bed_status"`             // 0=在床(in_bed), 1=离床(out_of_bed), 8=未改变
	TrackNumber      int   `json:"track_number"`           // 在床轨迹数量（max=2, sleepad left=0,right=1）
	StartTime        int64 `json:"start_time,omitempty"`   // 当前 BedStatus 开始时间（毫秒），未设置=0
	DurationSec      int   `json:"duration_sec"`           // 事件发生时已持续秒数，未设置=-1
	BedConfidence    int   `json:"bed_confidence"`    // 在床置信度 100 分制：60=雷达基准, 90=Sleepad 基准, 100=双设备（窗口内比较）相同
	BedEvent         int   `json:"bed_event"`         // 8=无/保持不变, 0=InBed, 1=LeftBed（与事件一致，初始化用 8 避免混淆）
	SleepStage       int   `json:"sleep_stage"`       // 睡眠阶段：0=初始, 1=清醒, 2=浅睡, 4=深睡, 8=未知
	SleepConfidence  int   `json:"sleep_confidence"`  // 睡眠阶段置信度 100 分制：0=未知, 60=Radar, 90=Sleepad, 100=sleepad+radar 相同
}

// SleepStage 合法值
const (
	SleepStageInitial = 0 // 初始（未收到 sleepStage 事件前）
	SleepStageAwake   = 1
	SleepStageLight   = 2
	SleepStageDeep    = 4
	SleepStageUnknown = 8
)

// RoomState 房间空间层（不含卫生间）：人数、进出、站立、风险。Card 初始化时默认创建；非卫生间雷达 stat.track 更新。
// HasRisk 由本房间数据各自计算。
type RoomState struct {
	UpdatedAt             int64          `json:"updated_at,omitempty"`
	TotalPeople           int            `json:"total_people"`                       // 房间人数（各 room 雷达 track 相加，且不低于床上人数）
	AreaPeople            map[string]int `json:"area_people,omitempty"`             // 各设备人数，key=deviceID，value=该雷达 NumberPeople；TotalPeople=sum(AreaPeople)，用于加减校正
	LastEnterTime         int64          `json:"last_enter_time,omitempty"`         // 最近进入时间（毫秒）
	LastExitTime          int64          `json:"last_exit_time,omitempty"`          // 最近离开时间（毫秒）
	StandingContinuousMin int            `json:"standing_continuous_min,omitempty"` // 连续站立分钟数（多分钟 track 累计）
	HasMulti              bool           `json:"has_multi"`
	HasRisk               bool           `json:"has_risk"` // 本房间风险，各自计算
}

// BathRoomState 卫生间空间层。仅当存在 isBathroomRadar（绑定 room 为 Bathroom）时创建/更新；由该雷达 stat.track 更新。
// HasRisk 由卫生间数据各自计算。
type BathRoomState struct {
	DeviceUID              string `json:"-"`                   // 内部用，不向前端暴露（HIPAA）
	DeviceID               string `json:"device_id,omitempty"` // 卫生间雷达 device_id，前端展示用
	UpdatedAt              int64  `json:"updated_at,omitempty"`
	TotalPeople            int    `json:"total_people"`     // 卫生间人数
	LastEnterTime          int64  `json:"last_enter_time,omitempty"`
	LastExitTime           int64  `json:"last_exit_time,omitempty"`
	StaySec                int    `json:"stay_sec,omitempty"`                // 卫生间滞留时长（秒）
	StandingContinuousMin  int    `json:"standing_continuous_min,omitempty"` // 连续站立分钟数
	HasMulti               bool   `json:"has_multi"`
	HasRisk                bool   `json:"has_risk"` // 卫生间风险，各自计算
}

// TargetState 单 Target 汇总（老人维度）：活动与生理时间、弱信号。Pose/SleepStage/Area/PersonNumber 已移除（多源覆盖无意义，由报警/BedState/RoomState 表达）。
type TargetState struct {
	UpdatedAt           int64  `json:"updated_at,omitempty"`
	TrackID             int    `json:"track_id"`             // track ID
	LogicID             string `json:"logic_id,omitempty"`  // Logic 层 ID（track ID）
	LastActiveTs        int64  `json:"last_active_ts,omitempty"`        // 最后活动时间（毫秒），最后生理指标时间（毫秒），state.track 行走距离>2米 & 行走时长>20秒 否则认为静止
	WeakBiometricSignal int    `json:"weak_biometric_signal"`            // 信号弱风险度，可多维度相加（如 HH+RR），0=好 越高越差，累加+每天 7 点清空
	VisitorStartTs       int64  `json:"VisitorStartTs,omitempty"`        // 当前访问时间 -1表示无
	TodayMaxVisitorMin   int   `json:"TodayMaxVisitorMin,omitempty"`   // 今日访客持续的最长时间
	HasVisitorToday      bool  `json:"HasVisitorToday,omitempty"`     // 今日访客标记
}

// AlarmState 告警摘要（与 cards 表 alarm 字段对齐）
type AlarmState struct {
	UpdatedAt     int64  `json:"updated_at,omitempty"`     // 状态最后更新时间（毫秒）
	TriggeredAt   int64  `json:"triggered_at,omitempty"`   // pop 告警真实触发时间（alarm_events.triggered_at 毫秒），用于前端 Alarm Time 展示
	ActiveEmerg   int    `json:"active_emerg"`   // EMERGENCY 未处理数
	ActiveAlert   int    `json:"active_alert"`   // ALERT 未处理数
	ActiveCrit    int    `json:"active_crit"`    // CRITICAL 未处理数
	ActiveErr     int    `json:"active_err"`     // ERROR 未处理数
	ActiveWarning int    `json:"active_warning"` // WARNING 未处理数
	PopAlarm      string `json:"pop_alarm,omitempty"`      // 当前弹出报警 "EMERG.Fall"
	EventID       string `json:"event_id,omitempty"`       // pop_alarm 对应的 alarm_events.event_id
}

// CardStatus 卡片状态数据
// 写入 card:state:{card_id} Hash + 发布到 card:status:stream
// 初始化：默认创建 RoomState；为雷达分配角色 isBathroomRadar（绑定 room=Bathroom）时创建/更新 BathRoomState。
// 当前仅有一个 Target（老人维度），用 Target 单对象。
type CardStatus struct {
	CardID        string                   `json:"card_id"`
	Target        *TargetState             `json:"target,omitempty"`         // 单 Target（当前仅一个）
	RoomState     *RoomState               `json:"room_state,omitempty"`     // 房间（不含卫生间），默认创建
	BathRoomState *BathRoomState           `json:"bathroom_state,omitempty"` // 仅当有卫生间雷达时存在
	BedState      *BedState                `json:"bed_state,omitempty"`
	AlarmState    *AlarmState              `json:"alarm_state,omitempty"`
	DeviceStatus  map[string]*DeviceStatus `json:"device_status,omitempty"`
	Message       map[string]interface{}   `json:"message,omitempty"`
}
