package models

// VitalFocusCard 完整的卡片对象（聚合后的数据）
// 用于 API 服务返回给前端
type VitalFocusCard struct {
	// 基础信息（来自 cards 表）
	CardID          string   `json:"card_id"`
	TenantID        string   `json:"tenant_id"`
	CardType        string   `json:"card_type"` // "ActiveBed" 或 "Location"
	BedID           *string  `json:"bed_id,omitempty"`           // ActiveBed 卡片
	LocationID      *string  `json:"location_id,omitempty"`      // Location 卡片（unit_id）
	CardName        string   `json:"card_name"`
	CardAddress     string   `json:"card_address"`
	PrimaryResidentID *string `json:"primary_resident_id,omitempty"` // ActiveBed 卡片的主住户

	// 住户和设备（来自 cards.residents 和 cards.devices JSONB）
	Residents       []CardResident `json:"residents"`
	Devices         []CardDevice   `json:"devices"`
	DeviceCount     int            `json:"device_count"`
	ResidentCount   int            `json:"resident_count"`

	// 报警统计（来自 cards 表）
	UnhandledAlarm0 *int `json:"unhandled_alarm_0,omitempty"` // EMERG(0)
	UnhandledAlarm1 *int `json:"unhandled_alarm_1,omitempty"` // ALERT(1)
	UnhandledAlarm2 *int `json:"unhandled_alarm_2,omitempty"` // CRIT(2)
	UnhandledAlarm3 *int `json:"unhandled_alarm_3,omitempty"` // ERR(3)
	UnhandledAlarm4 *int `json:"unhandled_alarm_4,omitempty"` // WARNING(4)
	TotalUnhandledAlarms *int `json:"total_unhandled_alarms,omitempty"` // 总计

	// 报警显示控制（来自 cards 表）
	IconAlarmLevel  *int `json:"icon_alarm_level,omitempty"`  // 图标报警级别阈值（默认 3）
	PopAlarmEmerge  *int `json:"pop_alarm_emerge,omitempty"`   // 弹出报警级别阈值（默认 0）

	// 设备连接状态（待实现，需要设备状态 API）
	RConnection     *int `json:"r_connection,omitempty"` // Radar 连接：0=offline, 1=online
	SConnection     *int `json:"s_connection,omitempty"` // Sleepace 连接：0=offline, 1=online

	// 实时数据（来自 Redis: vital-focus:card:{card_id}:realtime）
	// 生命体征
	Heart           *int    `json:"heart,omitempty"`           // 心率 (bpm)
	Breath          *int    `json:"breath,omitempty"`         // 呼吸频率 (次/分钟)
	HeartSource     *string `json:"heart_source,omitempty"`    // 's'=sleepace, 'r'=radar, '-'=无数据
	BreathSource    *string `json:"breath_source,omitempty"`  // 's'=sleepace, 'r'=radar, '-'=无数据

	// 睡眠状态
	SleepStage      *int    `json:"sleep_stage,omitempty"`    // 1=awake, 2=light sleep, 4=deep sleep
	SleepStateSNOMEDCode *string `json:"sleep_state_snomed_code,omitempty"`
	SleepStateDisplay    *string `json:"sleep_state_display,omitempty"`

	// 床状态
	BedStatus       *int    `json:"bed_status,omitempty"`     // 0=in bed, 1=out of bed

	// 姿态数据（Location 卡片）
	PersonCount     *int    `json:"person_count,omitempty"`   // 人数
	Postures        []int   `json:"postures,omitempty"`       // 姿态数组

	// 时间信息（ActiveBed 卡片）
	BedStatusTimestamp *string `json:"bed_status_timestamp,omitempty"` // 床状态变化时间（格式化）
	StatusDuration     *string `json:"status_duration,omitempty"`     // 持续时间（格式化）

	// 报警列表（来自 Redis: vital-focus:card:{card_id}:alarms）
	Alarms          []AlarmItem `json:"alarms,omitempty"`
}

// CardResident 卡片关联的住户
type CardResident struct {
	ResidentID string `json:"resident_id"`
	Nickname   string `json:"nickname"`
	UnitID     *string `json:"unit_id,omitempty"`
	BedID      *string `json:"bed_id,omitempty"`
}

// CardDevice 卡片关联的设备
type CardDevice struct {
	DeviceID    string  `json:"device_id"`
	DeviceName  string  `json:"device_name"`
	DeviceType  string  `json:"device_type"`
	DeviceModel string  `json:"device_model"`
	BedID       *string `json:"bed_id,omitempty"`
	BedName     *string `json:"bed_name,omitempty"`
	RoomID      *string `json:"room_id,omitempty"`
	RoomName    *string `json:"room_name,omitempty"`
	UnitID      string  `json:"unit_id"`
}

// AlarmItem 报警项（来自 alarm_events 表）
type AlarmItem struct {
	EventID          string                 `json:"event_id"`
	EventType        string                 `json:"event_type"`
	Category         *string                `json:"category,omitempty"` // safety, clinical, behavioral, device
	AlarmLevel       string                 `json:"alarm_level"`       // '0'/'EMERG', '1'/'ALERT', etc.
	AlarmStatus      string                 `json:"alarm_status"`       // active, acknowledged
	TriggeredAt      int64                  `json:"triggered_at"`       // timestamp
	TriggeredBy      *string                `json:"triggered_by,omitempty"` // 设备名称或 'Cloud'
	TriggerData      map[string]interface{} `json:"trigger_data,omitempty"`
	IoTTimeSeriesID  *int64                 `json:"iot_timeseries_id,omitempty"`
}

