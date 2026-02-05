package card

import (
	consts "owl-common/const"
)

// -------------- 静态数据：同一张 card 必属同一 unit/branch，unit/branch 在 card 层只存一份。
// ActiveBedInfo 床的静态信息（仅房间与住户，不重复 unit/branch）
type ActiveBedInfo struct {
	BedID      string  `json:"bed_id"`
	RoomID     string  `json:"room_id"`
	ResidentID *string `json:"resident_id,omitempty"`
}

// ActiveBedRow 用于 repository/creator：需 BedID 做 GetDevicesByBed/GetResidentByBed
type ActiveBedRow struct {
	UnitID           string  `json:"unit_id"`
	BoundDeviceCount int     `json:"BoundDevcieCount"`
	BedID            string  `json:"bed_id"`
	RoomID           string  `json:"room_id"`
	ResidentID       *string `json:"resident_id,omitempty"`
}

// UnitInfo Unit information
type UnitInfo struct {
	UnitID       string
	UnitName     string
	BranchID     string
	Building     string
	IsPublic     bool // 对应数据库字段 is_public
	IsSharedUnit bool // 对应数据库字段 is_shared_unit
	UnitType     string
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
	BoundBedID        *string
	BedName           *string // Bed name (if bound to bed)
	BoundRoomID       *string // Room ID where device is bound (if bound to room)
	RoomName          *string // Room name (if bound to room)
	UnitID            string
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
	UnitID           *string
	BedID            *string
}

// VitalFocusCardInfo 卡片静态+动态；同一张 card 必属同一 unit、branch，以下仅存一份。
type VitalFocusCardInfo struct {
	// 基础信息（来自 cards 表）
	CardID            string  `json:"card_id"`
	TenantID          string  `json:"tenant_id"`
	CardType          string  `json:"card_type"` // "ActiveBed" 或 "Location"
	CardName          string  `json:"card_name"`
	CardAddress       string  `json:"card_address"`
	PrimaryResidentID *string `json:"primary_resident_id,omitempty"` // ActiveBed 卡片的主住户

	// 同一 card 必属同一 unit/branch，只存一份
	BedID      *string `json:"bed_id,omitempty"`      // 仅 ActiveBed 有值，Location 为空
	UnitID     string  `json:"unit_id,omitempty"`     // 单元 ID
	BranchID   string  `json:"branch_id,omitempty"`   // 院区 ID（来自 unit）
	BranchName string  `json:"branch_name,omitempty"` // 院区名（来自 unit）
	Timezone   string  `json:"timezone,omitempty"`    // IANA，如 America/Los_Angeles

	// 住户和设备（来自 cards.residents 和 cards.devices JSONB）
	Residents []ResidentInfo `json:"residents"`
	Devices   []DeviceInfo   `json:"devices"`

	// 报警显示控制（来自 cards 表）
	IconAlarmLevel *int `json:"icon_alarm_level,omitempty"` // 图标报警级别阈值（默认 3）
	PopAlarmEmerge *int `json:"pop_alarm_emerge,omitempty"` // 弹出报警级别阈值（默认 0）
}

//-------------- 动态数据：仅 card_id 作为索引，不含其它静态数据。
// 从 iot:*:stream 获取实时数据，供前端读取。

// VitalSimplified vital 简化格式（monitor vital 直接转换，用于缓存/显示）
// Timestamp 为本条 MQTT 时间，用于与后续消息比较是否更新。
type VitalSimplified struct {
	DeviceID        string  `json:"device_id"`
	Timestamp       int64   `json:"timestamp"` // 本条 MQTT 时间戳
	RespiratoryRate *int    `json:"respiratory_rate,omitempty"`
	HeartRate       *int    `json:"heart_rate,omitempty"`
	SleepStatus     *string `json:"sleep_status,omitempty"`
	Stability       *string `json:"stability,omitempty"`
}

// 设备检测的人数
type DevicePosture struct {
	DeviceID  string `json:"device_id"`
	Timestamp int64  `json:"timestamp"` // 本次 MQTT 时间戳
	Postures  []int  `json:"postures"`  // 姿态列表，每设备/人一个，值 0-11
}

type BedState struct {
	BedID        string `json:"bed_id,omitempty"`          // 床 ID
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
	EMERG    int    `json:"emerg,omitempty"`     // 未处理 EMERG 数量
	ALERT    int    `json:"alert,omitempty"`     // 未处理 ALERT 数量
	CRIT     int    `json:"crit,omitempty"`      // 未处理 CRIT 数量
	ERR      int    `json:"err,omitempty"`       // 未处理 ERR 数量
	WARNING  int    `json:"warning,omitempty"`   // 未处理 WARNING 数量
	NOTICE   int    `json:"notice,omitempty"`    // 未处理 NOTICE 数量
	NowAlarm string `json:"now_alarm,omitempty"` // 当前最高级别报警，格式 "AlarmLevel.AlarmType"，如 "EMERG.Fall"
	// Timestamp 最后一次更新时间（用于防止旧数据覆盖新数据）
	Timestamp int64 `json:"timestamp,omitempty"`
}

type ActveAlarms struct {
	Timestamp     int64  `json:"timestamp,omitempty"`            // 最后一次更新时间（防止旧数据覆盖新数据）
	activeEMERG   int    `json:"emerg,omitempty"`                // 未处理数量
	activeALERT   int    `json:"alert,omitempty"`                // 未处理数量
	activeCRIT    int    `json:"crit,omitempty"`                 // 未处理数量
	activeERR     int    `json:"err,omitempty"`                  // 未处理数量
	activeWARNING int    `json:"warning,omitempty"`              // 未处理数量
	activeNOTICE  int    `json:"notice,omitempty"`               // 未处理数量
	NowAlarm      string `json:"alarmLevel.alarmType,omitempty"` // 当前未处理的最高级别报警，可选项,防止低级别报警刷掉高级别报警。
}

// RealtimeData 动态数据（写入 Redis）；仅 card_id 作为索引，不含其它静态数据。
type RealtimeData struct {
	CardID    string `json:"card_id"`
	Timestamp int64  `json:"timestamp"`

	//{[deviceID,deviceTimestamp,respiratoryRate,heartRate,sleepStatus,stability], ...}
	Vital []VitalSimplified `json:"vital,omitempty"` // 生命体征按设备

	// Postures: 设备姿态数组（值 0-11，与 consts.Pose* 一致）；长度即该设备人数
	// 数组格式，每个元素自带 device_id 和 timestamp
	// JSON: [{"device_id":"device_id_1","timestamp":1234567890,"postures":[0,5,9]},{"device_id":"device_id_2","timestamp":1234567890,"postures":[6]}]
	Postures []DevicePosture `json:"postures,omitempty"`
	// DeviceStatus: device_id -> 状态数组（使用 consts.DeviceStatus 整型，避免 string 大小写）
	// 值见 consts.StatusOnline/StatusOffline/StatusSignalPoor 等；JSON 为数字数组，如 {"device_id_1": [1,5], "device_id_2": [0]}
	DeviceStatus map[string][]consts.DeviceStatus `json:"device_status,omitempty"`
	BedState     *BedState                        `json:"bed_state,omitempty"`     // 床状态，可选项（UnitCard 可能没有）
	RoomState    *RoomState                       `json:"room_state,omitempty"`    // 房间状态，可选项（LocationCard 可能没有）
	ActiveAlarms *ActiveAlarmState                `json:"active_alarms,omitempty"` // 活跃报警：未处理数量 + 最高级别报警
}

// ----
// CardWithContent card with devices and residents JSONB content (for comparison)
type CardWithContent struct {
	CardID        string
	CardType      string
	BedID         *string
	UnitID        string
	CardName      string
	CardAddress   string
	Timezone      string
	ResidentID    *string
	DevicesJSON   []byte
	ResidentsJSON []byte
}

// ExpectedCard represents an expected card (for comparison, without card_id)
type ExpectedCard struct {
	CardType      string
	BedID         *string
	UnitID        string
	CardName      string
	CardAddress   string
	Timezone      string
	ResidentID    *string
	DevicesJSON   []byte
	ResidentsJSON []byte
}

// CardUpdateStats statistics for card updates
type CardUpdateStats struct {
	ExistingCount  int // Number of existing cards before update
	DeletedCount   int // Number of cards deleted
	CreatedCount   int // Number of cards created
	UpdatedCount   int // Number of cards updated (deleted + created)
	UnchangedCount int // Number of cards that remained unchanged
}

// CardIndexItem 卡片简化信息（用于列表页）
// 包含前端显示卡片列表所需的最少信息
type CardIndexItem struct {
	CardID            string   `json:"card_id"`
	CardName          string   `json:"card_name"`
	CardAddress       string   `json:"card_address"`
	BranchID          string   `json:"branch_id"`
	IconAlarmLevel    int      `json:"icon_alarm_level,omitempty"`
	PopAlarmEmerge    int      `json:"pop_alarm_emerge,omitempty"`
	DeviceIDs         []string `json:"device_ids,omitempty"`
	PrimaryResidentID *string  `json:"primary_resident_id,omitempty"`
}
