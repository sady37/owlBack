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
	DeviceUID         string  `json:"device_uid"`  // devices.device_uid
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

//-------------- 动态数据：仅 card_id 作为索引，不含其它静态数据。
// 从 iot:*:stream 获取实时数据，供前端读取。
// DeviceRealTime 是纯观测层——设备说了什么，高频可丢，不做业务判断。

// ========== data_value category 常量 ==========

const (
	CategoryTrack  = "track"  // 轨迹数据（Radar）
	CategoryVital  = "vital"  // 生理数据（Radar + Sleepad 共有）
	CategoryMotion = "motion" // 动作数据（Sleepad: bodyMove, turnOver, sitUp）
	CategoryStatus = "status" // 状态观测（Sleepad: bedStatus, initStatus）
	CategoryDevice = "device" // 设备质量观测（signalQuality 等）
)

// ========== 观测数据结构体（按 category 对应） ==========

// TrackData 轨迹数据（category="track"）—— Radar
type TrackData struct {
	Category      string `json:"category"`                 // "track"
	TargetID      int    `json:"target_id,omitempty"`      // 目标ID: 0-7=有人，88=无人
	PositionX     int    `json:"position_x,omitempty"`     // X坐标（cm）
	PositionY     int    `json:"position_y,omitempty"`     // Y坐标（cm）
	PositionZ     int    `json:"position_z,omitempty"`     // Z坐标（cm）
	RemainingTime int    `json:"remaining_time,omitempty"` // 剩余时间（秒）
	AreaID        int    `json:"area_id,omitempty"`        // 区域ID
	Pose          int    `json:"pose,omitempty"`           // 姿态数值（0-11，见 PoseNumToDisplay）
	Event         int    `json:"event,omitempty"`          // 事件类型: 0=无, 1=进房, 2=离房, 3=进区, 4=离区
}

// VitalData 生理数据（category="vital"）—— Radar + Sleepad
type VitalData struct {
	Category        string `json:"category"`                   // "vital"
	VitalFlag       int    `json:"vital_flag,omitempty"`       // 标识符: 固定为0表示实时呼吸心率
	RespiratoryRate int    `json:"respiratory_rate,omitempty"` // 呼吸率（次/分钟）
	HeartRate       int    `json:"heart_rate,omitempty"`       // 心率（次/分钟）
	SleepStatus     int    `json:"sleep_status,omitempty"`     // 睡眠状态（0=未定义, 1=浅睡, 2=深睡, 3=清醒）
	Stability       int    `json:"stability,omitempty"`        // 稳定性（0=未定义, 1=较大动作, 2=较小动作, 3=无干扰）
}

// MotionData 动作数据（category="motion"）—— Sleepad
type MotionData struct {
	Category  string `json:"category"`             // "motion"
	BodyMove  int    `json:"body_move,omitempty"`  // 体动 0/1
	TurnOver  int    `json:"turn_over,omitempty"`  // 翻身 0/1
	SitUp     int    `json:"sit_up,omitempty"`     // 坐起 0/1
	LeftRight int    `json:"left_right,omitempty"` // 左右翻 0/1/2
}

// StatusData 状态观测（category="status"）—— Sleepad
type StatusData struct {
	Category   string `json:"category"`              // "status"
	BedStatus  int    `json:"bed_status,omitempty"`  // 在床状态 0=离床 1=在床
	InitStatus int    `json:"init_status,omitempty"` // 初始化状态
}

// DeviceData 设备质量观测（category="device"）
type DeviceData struct {
	Category      string `json:"category"`                 // "device"
	SignalQuality int    `json:"signal_quality,omitempty"` // 信号质量
	Battery       int    `json:"battery,omitempty"`        // 电量百分比
	SensorState   int    `json:"sensor_state,omitempty"`   // 传感器状态
}

// ========== 设备实时数据聚合 ==========

// DeviceRealTime 单个设备的实时观测数据
// Data 原样存储 qinglan 输出的 data_value（每项含 category 字段），cardagg 不分拣
// 前端直接消费 device-first 结构，按 category 过滤取值
type DeviceRealTime struct {
	DeviceID   string                   `json:"device_id"`      // 设备ID
	DeviceType string                   `json:"device_type"`    // "Radar" | "SleepPad"
	Timestamp  int64                    `json:"timestamp"`      // 设备最后更新时间戳
	Data       []map[string]interface{} `json:"data,omitempty"` // 原样存储，保留 category
}

// DeviceStatus 设备状态信息
type DeviceStatus struct {
	DeviceID   string         `json:"device_id"`          // 设备ID
	DeviceType string         `json:"device_type"`        // "Radar" | "SleepPad" 等
	Timestamp  int64          `json:"timestamp"`          // 最后一次更新时间戳（Unix毫秒）
	Statuses   map[string]int `json:"statuses,omitempty"` // 设备状态 map, 0=正常/在线, 1=异常/离线
}

// EventState 事件状态摘要（统一替代 BedState/RoomState）
// Category 区分事件类型，单个 EventState 表示卡片当前最新事件
type EventState struct {
	UpdatedAt    int64  `json:"updated_at,omitempty"`    // 消息处理时间（排序/去重）
	Category     string `json:"category,omitempty"`      // "BedState" | "RoomState"
	CurrentState string `json:"current_state,omitempty"` // "in_bed" | "out_of_bed" 等
	StateValue   string `json:"state_value,omitempty"`   // string（people_count 等）
	StartTime    int64  `json:"start_time,omitempty"`    // 事件开始时间
	DurationSec  int    `json:"duration_sec,omitempty"`  // 初始秒数 = UpdatedAt - StartTime（前端在此基础上自增）
}

// AlarmState 报警状态摘要（与 cards 表 alarm 字段对齐）
type AlarmState struct {
	UpdatedAt     int64  `json:"updated_at,omitempty"`     // 最后更新时间
	ActiveEmerg   int    `json:"active_emerg,omitempty"`   // 未处理数量 EMERGENCY
	ActiveAlert   int    `json:"active_alert,omitempty"`   // 未处理数量 ALERT
	ActiveCrit    int    `json:"active_crit,omitempty"`    // 未处理数量 CRITICAL
	ActiveErr     int    `json:"active_err,omitempty"`     // 未处理数量 ERROR
	ActiveWarning int    `json:"active_warning,omitempty"` // 未处理数量 WARNING
	PopAlarm      string `json:"pop_alarm,omitempty"`      // 当前弹出报警 "EMERG.Fall"
	EventID       string `json:"event_id,omitempty"`       // pop alarm 对应的 alarm_events.event_id
}

// CardRealTime 卡片实时数据（按设备聚合）
// 存储在 Redis: vital-focus:card:{card_id}:realtime
// 发布到 Redis Stream: card:realtime:stream
type CardRealTime struct {
	CardID    string                     `json:"card_id"`           // 卡片ID
	Timestamp int64                      `json:"timestamp"`         // 最后更新时间戳
	Devices   map[string]*DeviceRealTime `json:"devices,omitempty"` // 按设备ID索引的实时数据
}

// CardStatus 卡片状态数据（bed/room/alarms/device status）
// 存储在 Redis Stream: card:status:stream
type CardStatus struct {
	CardID       string                   `json:"card_id"`                 // 卡片ID
	Timestamp    int64                    `json:"timestamp"`               // 最后更新时间戳
	UpdateType   string                   `json:"update_type,omitempty"`   // 更新类型: "DeviceStatus" | "EventState" | "AlarmState"
	DeviceStatus map[string]*DeviceStatus `json:"device_status,omitempty"` // 按设备ID索引的设备状态 map
	EventState   *EventState              `json:"event_state,omitempty"`   // 最新事件（替代 BedState/RoomState）
	AlarmState   *AlarmState              `json:"alarm_state,omitempty"`   // 报警状态
	Message      map[string]interface{}   `json:"message,omitempty"`       // 可选：完整原始消息内容
}
