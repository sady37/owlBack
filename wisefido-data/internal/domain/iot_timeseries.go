package domain

import "time"

// IoTTimeSeries IoT时序数据领域模型（对应 iot_timeseries 表）
// 存储 Radar、SleepPad 等多种 IoT 设备的实时时间序列数据
type IoTTimeSeries struct {
	// 主键
	ID int64 `db:"id"` // BIGSERIAL

	// 设备索引
	TenantID string `db:"tenant_id"` // UUID, NOT NULL
	DeviceID string `db:"device_id"` // UUID, NOT NULL

	// 时间戳
	Timestamp time.Time `db:"timestamp"` // TIMESTAMPTZ, NOT NULL

	// 数据类型标识
	DataType string `db:"data_type"` // VARCHAR(20), 'observation'/'alarm'
	Category string `db:"category"`  // VARCHAR(50), FHIR Category

	// 轨迹数据
	TrackingID *int `db:"tracking_id"` // INTEGER, nullable
	RadarPosX  *int `db:"radar_pos_x"` // INTEGER, nullable
	RadarPosY  *int `db:"radar_pos_y"` // INTEGER, nullable
	RadarPosZ  *int `db:"radar_pos_z"` // INTEGER, nullable

	// 姿态/运动状态
	PostureSNOMEDCode string `db:"posture_snomed_code"` // VARCHAR(50), nullable
	PostureDisplay    string `db:"posture_display"`     // VARCHAR(100), nullable

	// 事件
	EventType       string `db:"event_type"`        // VARCHAR(50), nullable
	EventSNOMEDCode string `db:"event_snomed_code"` // VARCHAR(50), nullable
	EventDisplay    string `db:"event_display"`     // VARCHAR(100), nullable
	AreaID          *int   `db:"area_id"`           // INTEGER, nullable

	// 生命体征
	HeartRateCode        string `db:"heart_rate_code"`         // VARCHAR(50), nullable
	HeartRateDisplay     string `db:"heart_rate_display"`      // VARCHAR(100), nullable
	HeartRate            *int   `db:"heart_rate"`              // INTEGER, nullable
	RespiratoryRateCode  string `db:"respiratory_rate_code"`   // VARCHAR(50), nullable
	RespiratoryRateDisplay string `db:"respiratory_rate_display"` // VARCHAR(100), nullable
	RespiratoryRate      *int   `db:"respiratory_rate"`        // INTEGER, nullable

	// 睡眠状态
	SleepStateSNOMEDCode string `db:"sleep_state_snomed_code"` // VARCHAR(50), nullable
	SleepStateDisplay    string `db:"sleep_state_display"`    // VARCHAR(100), nullable

	// 位置信息（冗余字段）
	UnitID string `db:"unit_id"` // UUID, nullable
	RoomID string `db:"room_id"` // UUID, nullable

	// 告警关联
	AlarmEventID string `db:"alarm_event_id"` // UUID, nullable

	// 其他字段
	Confidence    *int `db:"confidence"`     // INTEGER, nullable
	RemainingTime *int `db:"remaining_time"` // INTEGER, nullable

	// 原始记录
	RawOriginal     []byte `db:"raw_original"`      // BYTEA, NOT NULL
	RawFormat       string `db:"raw_format"`        // VARCHAR(50), NOT NULL
	RawCompression  string `db:"raw_compression"`   // VARCHAR(50), nullable

	// 元数据
	Metadata map[string]interface{} `db:"metadata"` // JSONB

	// 时间戳
	CreatedAt time.Time `db:"created_at"` // TIMESTAMPTZ

	// 关联数据（通过 JOIN 获取）
	DeviceSN    string `db:"device_sn"`    // 从 devices 表获取
	DeviceUID   string `db:"device_uid"`    // 从 devices 表获取
	DeviceType  string `db:"device_type"`   // 从 device_store 表获取
	FirmwareVersion string `db:"firmware_version"` // 从 devices 表获取
}

