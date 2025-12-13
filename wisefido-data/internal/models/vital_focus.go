package models

// 注意：字段命名全部使用 snake_case（json tag），与 owlFront monitorModel.ts 对齐

type ServiceLevelInfo struct {
	LevelCode     string `json:"level_code"`
	DisplayName   string `json:"display_name"`
	DisplayNameCN string `json:"display_name_cn,omitempty"`
	ColorTag      string `json:"color_tag"`
	ColorHex      string `json:"color_hex"`
	Priority      int    `json:"priority"`
}

type CardResident struct {
	ResidentID       string           `json:"resident_id"`
	LastName         string           `json:"last_name,omitempty"`
	FirstName        string           `json:"first_name,omitempty"`
	Nickname         string           `json:"nickname,omitempty"`
	ServiceLevel     string           `json:"service_level,omitempty"`
	ServiceLevelInfo *ServiceLevelInfo `json:"service_level_info,omitempty"`
}

type CardDevice struct {
	DeviceID    string `json:"device_id"`
	DeviceName  string `json:"device_name"`
	DeviceType  any    `json:"device_type"` // 前端目前用 number；后端当前聚合里是 string，先允许 any 兼容
	DeviceModel string `json:"device_model,omitempty"`
	// binding_type 前端仍保留，但我们在 owlRD 里已去掉；这里先不强制输出，由 full cache 决定
	BindingType string `json:"binding_type,omitempty"`
}

type AlarmItem struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	Category     string `json:"category,omitempty"`
	AlarmLevel   any    `json:"alarm_level"` // string | number
	AlarmStatus  string `json:"alarm_status"`
	TriggeredAt  int64  `json:"triggered_at"`
	TriggeredBy  string `json:"triggered_by,omitempty"`
	TriggerData  any    `json:"trigger_data,omitempty"`
	IoTTimeSeriesID *int64 `json:"iot_timeseries_id,omitempty"`
}

// VitalFocusCard 对齐 owlFront/src/api/monitors/model/monitorModel.ts
type VitalFocusCard struct {
	CardID    string `json:"card_id"`
	TenantID  string `json:"tenant_id"`
	CardType  string `json:"card_type"` // 'ActiveBed' | 'Location'
	BedID     string `json:"bed_id,omitempty"`
	LocationID string `json:"location_id,omitempty"`
	CardName  string `json:"card_name"`
	CardAddress string `json:"card_address"`
	PrimaryResidentID string `json:"primary_resident_id,omitempty"`

	Residents []CardResident `json:"residents"`
	Devices   []CardDevice   `json:"devices"`

	DeviceCount   int `json:"device_count"`
	ResidentCount int `json:"resident_count"`

	UnhandledAlarm0 *int `json:"unhandled_alarm_0,omitempty"`
	UnhandledAlarm1 *int `json:"unhandled_alarm_1,omitempty"`
	UnhandledAlarm2 *int `json:"unhandled_alarm_2,omitempty"`
	UnhandledAlarm3 *int `json:"unhandled_alarm_3,omitempty"`
	UnhandledAlarm4 *int `json:"unhandled_alarm_4,omitempty"`
	TotalUnhandledAlarms *int `json:"total_unhandled_alarms,omitempty"`

	IconAlarmLevel *int `json:"icon_alarm_level,omitempty"`
	PopAlarmEmerge *int `json:"pop_alarm_emerge,omitempty"`

	RConnection *int `json:"r_connection,omitempty"`
	SConnection *int `json:"s_connection,omitempty"`

	Statuses map[string]any `json:"statuses,omitempty"`

	Breath *int `json:"breath,omitempty"`
	Heart  *int `json:"heart,omitempty"`
	BreathSource string `json:"breath_source,omitempty"` // 's' | 'r' | '-'
	HeartSource  string `json:"heart_source,omitempty"`

	SleepStage *int `json:"sleep_stage,omitempty"`
	SleepStateSNOMEDCode string `json:"sleep_state_snomed_code,omitempty"`
	SleepStateDisplay string `json:"sleep_state_display,omitempty"`

	BedStatus *int `json:"bed_status,omitempty"`

	PersonCount *int  `json:"person_count,omitempty"`
	Postures    []int `json:"postures,omitempty"`

	BedStatusTimestamp string `json:"bed_status_timestamp,omitempty"`
	StatusDuration     string `json:"status_duration,omitempty"`

	Alarms []AlarmItem `json:"alarms,omitempty"`
}

type GetVitalFocusCardsModel struct {
	Items      []VitalFocusCard   `json:"items"`
	Pagination BackendPagination `json:"pagination"`
}

type VitalFocusCardInfo struct {
	CardID    string `json:"card_id"`
	TenantID  string `json:"tenant_id"`
	CardType  string `json:"card_type"`
	BedID     string `json:"bed_id,omitempty"`
	LocationID string `json:"location_id,omitempty"`
	CardName  string `json:"card_name"`
	CardAddress string `json:"card_address"`
	PrimaryResidentID string `json:"primary_resident_id,omitempty"`
	Residents []CardResident `json:"residents"`
	Devices   []CardDevice   `json:"devices"`
}


