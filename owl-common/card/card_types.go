package card

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
	UnitID       string
	UnitName     string
	BranchID     string
	BranchName   string // 院区名称
	Building     string
	IsPublic     bool   // 对应数据库字段 is_public
	IsSharedUnit bool   // 对应数据库字段 is_shared_unit
	UnitType     string //"facility" | "home"
	Timezone     string // IANA, e.g. "America/Los_Angeles"
}

// DeviceInfo device information
type DeviceInfo struct {
	DeviceID          string
	DeviceUID         string // devices.device_uid，与 cards.devices JSON、card-overview 对齐
	DeviceCode        string // device_store.device_code，与 card-overview 对齐
	DeviceName        string
	DeviceType        any // 数字 1/2 或字符串，JSON 兼容
	DeviceModel       string
	UnitID            string // 所属unit
	BoundBedID        *string
	BoundRoomID       *string // Room ID where device is bound (if bound to room)
	MonitoringEnabled bool
	Status            string // "online" | "offline" | "error" | "disabled"（API 动态）
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

// VitalFocusCardInfo 卡片静态+动态；同一张 card 必属同一 unit、branch，以下仅存一份。
type VitalFocusCardInfo struct {
	// 基础信息（来自 cards 表）
	CardID            string  `json:"card_id"`
	CardType          string  `json:"card_type"` // "ActiveBedCard" 或 "UnitCard"
	CardName          string  `json:"card_name"`
	CardAddress       string  `json:"card_address"`


	// 同一 card 必属同一 unit/branch，只存一份
	TenantID          string  `json:"tenant_id"`
	BranchID   string  `json:"branch_id,omitempty"`   // 院区 ID（来自 unit）
	BranchName string  `json:"branch_name,omitempty"` // 院区名（来自 unit）
	UnitID     string  `json:"unit_id,omitempty"`     // 单元 ID
	UnitName   string  `json:"unit_name,omitempty"`   // 单元名（来自 unit）	
	Timezone   string  `json:"timezone,omitempty"`    // IANA，如 America/Los_Angeles

	// 一个卡片可能关联多个房间
	Rooms []RoomIdentifier `json:"rooms,omitempty"` // 房间列表（可能有多个房间）
	BedID      *string `json:"bed_id,omitempty"`      // 仅 ActiveBedCard 有值，UnitCard 为空
	BedName    *string `json:"bed_name,omitempty"`    // 床名称（仅 ActiveBedCard 有值，UnitCard 为空）

	// 住户和设备（来自 cards.residents 和 cards.devices JSONB）
	Residents []ResidentInfo `json:"residents"`
	Devices   []DeviceInfo   `json:"devices"`

	// 报警显示控制（来自 cards 表）
	IconAlarmLevel *int `json:"icon_alarm_level,omitempty"` // 图标报警级别阈值（默认 3）
	PopAlarmEmerge *int `json:"pop_alarm_emerge,omitempty"` // 弹出报警级别阈值（默认 0）
}

//-------------- 动态数据：仅 card_id 作为索引，不含其它静态数据。
// 从 iot:*:stream 获取实时数据，供前端读取。

// DeviceTrack 设备轨迹数据（track 数据，更新频率 1Hz）
type DeviceTrack struct {
	DeviceID   string        `json:"device_id"`   // 设备ID
	DeviceType string        `json:"device_type"` // "Radar" | "SleepPad"
	Timestamp  int64         `json:"timestamp"`   // 设备时间戳
	Category   string        `json:"category"`    // "track2" 等
	DataValue  []interface{} `json:"data_value"`  // track 数据数组，保持原始格式
}

// DeviceVital 设备生命体征数据（vital 数据，更新频率 2Hz）
type DeviceVital struct {
	DeviceID   string        `json:"device_id"`   // 设备ID
	DeviceType string        `json:"device_type"` // "Radar" | "SleepPad"
	Timestamp  int64         `json:"timestamp"`   // 设备时间戳
	Category   string        `json:"category"`    // "vital1" 等
	DataValue  []interface{} `json:"data_value"`  // vital 数据数组，保持原始格式
}

// DeviceStatus 设备状态信息
type DeviceStatus struct {
	DeviceID   string         `json:"device_id"`          // 设备 ID
	DeviceType string         `json:"device_type"`        // "Radar" | "SleepPad" 等
	Timestamp  int64          `json:"timestamp"`          // 最后一次更新时间戳（Unix毫秒）
	Statuses   map[string]int `json:"statuses,omitempty"` // 设备状态 map
}

type BedState struct {
	BedID        string `json:"bed_id,omitempty"`        // 床 ID
	BedName      string `json:"bed_name,omitempty"`      // 床名称
	CurrentState string `json:"current_state,omitempty"` // "in_bed" | "out_of_bed"
	Timestamp    int64  `json:"timestamp,omitempty"`     // 最后一次更新时间戳（Unix毫秒）
}

type RoomState struct {
	RoomID      string `json:"room_id,omitempty"`      // 房间 ID
	RoomName    string `json:"room_name,omitempty"`    // 房间名
	PeopleCount int    `json:"people_count,omitempty"` // 房间人数
	StayTime    int    `json:"stay_time,omitempty"`    // 驻留时间（分钟）
	Timestamp   int64  `json:"timestamp,omitempty"`    // 最后一次更新时间戳（Unix毫秒）
}

// ActiveAlarmState 活跃报警状态
type ActiveAlarmState struct {
	Timestamp     int64  `json:"timestamp,omitempty"` // 最后一次更新时间（防止旧数据覆盖新数据）
	ActiveEmerg   int    `json:"emerg,omitempty"`     // 未处理数量
	ActiveAlert   int    `json:"alert,omitempty"`     // 未处理数量
	ActiveCrit    int    `json:"crit,omitempty"`      // 未处理数量
	ActiveErr     int    `json:"err,omitempty"`       // 未处理数量
	ActiveWarning int    `json:"warning,omitempty"`
	NowAlarm      string `json:"alarmLevel.alarmType,omitempty"` // 当前未处理的最高级别报警，可选项,防止低级别报警刷掉高级别报警。
}

// CardRealTime 卡片实时数据快照（Redis缓存格式）
// 用于 vital-focus:card:{card_id}:realtime
// 包含设备轨迹和生命体征数据，高频更新（TTL较短）
type CardRealTime struct {
	CardID    string        `json:"card_id"`
	Timestamp int64         `json:"timestamp"`
	TrackData []interface{} `json:"track_data,omitempty"` // track 数据，1Hz 更新（保存为通用格式）
	VitalData []interface{} `json:"vital_data,omitempty"` // vital 数据，2Hz 更新（保存为通用格式）
}

// CardStatus 卡片状态快照（Redis缓存格式）
// 用于 vital-focus:card:{card_id}:status
// 包含长期保存的状态信息，TTL较长
type CardStatus struct {
	CardID       string                   `json:"card_id"`
	Timestamp    int64                    `json:"timestamp"`
	DeviceStatus map[string]*DeviceStatus `json:"device_status,omitempty"` // 按设备ID索引的设备状态 map
	BedState     *BedState                `json:"bed_state,omitempty"`     // 床状态（UnitCard 可能没有）
	RoomState    *RoomState               `json:"room_state,omitempty"`    // 房间状态（UnitCard 可能没有）
	ActiveAlarms *ActiveAlarmState        `json:"active_alarms,omitempty"` // 活跃报警：未处理数量 + 最高级别
}

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
