package models

import (
	"time"
)

// IoTTimeSeries IoT 时序数据（从 PostgreSQL 读取）
type IoTTimeSeries struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`

	// 生命体征
	HeartRate              *int    `json:"heart_rate"`
	HeartRateCode          *string `json:"heart_rate_code"`
	HeartRateDisplay       *string `json:"heart_rate_display"`
	RespiratoryRate        *int    `json:"respiratory_rate"`
	RespiratoryRateCode    *string `json:"respiratory_rate_code"`
	RespiratoryRateDisplay *string `json:"respiratory_rate_display"`

	// 姿态数据
	PostureSNOMEDCode *string `json:"posture_snomed_code"`
	PostureDisplay    *string `json:"posture_display"`
	TrackingID        *string `json:"tracking_id"` // Radar 设备的 tracking_id

	// 位置数据（Radar 设备，单位：厘米；原样透传，不转换）
	PositionX *int `json:"position_x,omitempty"` // 轨迹字节 2，分米×10→厘米
	PositionY *int `json:"position_y,omitempty"` // 轨迹字节 3，分米×10→厘米
	PositionZ *int `json:"position_z,omitempty"` // 轨迹字节 4，厘米 0-255

	// 区域ID（Radar设备）
	AreaID *int `json:"area_id,omitempty"` // area_id（床区域ID）

	// 床状态
	BedStatusSNOMEDCode *string `json:"bed_status_snomed_code"`
	BedStatusDisplay    *string `json:"bed_status_display"`

	// 睡眠状态
	SleepStateSNOMEDCode *string `json:"sleep_state_snomed_code"`
	SleepStateDisplay    *string `json:"sleep_state_display"`

	// 设备类型（从 devices 表查询）
	DeviceType string `json:"device_type"` // "Radar" 或 "Sleepace"
}

// VitalSource 按源存储的生命体征与睡眠状态（Radar / Sleepad 分源，display 由 card/vue 计算）
type VitalSource struct {
	Heart       *int    `json:"heart,omitempty"`        // 心率
	Breath      *int    `json:"breath,omitempty"`       // 呼吸率
	SleepStatus *string `json:"sleep_status,omitempty"`  // 睡眠状态 SNOMED
	BedStatus   *string `json:"bed_status,omitempty"`    // 床状态 SNOMED
}

// RealtimeData 按源存储的实时数据（写入 Redis）：radar / sleepad 分源，display 由 card/vue 或 mergeRealtimeData 计算
type RealtimeData struct {
	// 按源存储（不再融合 Heart/Breath/HeartSource/BreathSource）
	Radar  *VitalSource `json:"radar,omitempty"`  // Radar 源：heart, breath, sleep_status, bed_status
	Sleepad *VitalSource `json:"sleepad,omitempty"` // Sleepad 源：heart, breath, sleep_status, bed_status

	// 姿态数据（来自 Radar）
	PersonCount int       `json:"person_count"` // 人数（tracking_id 数量）
	Postures    []Posture `json:"postures"`     // 姿态列表

	// 时间戳
	Timestamp int64 `json:"timestamp"` // Unix 时间戳（使用数据中的最大时间戳）
}

// Posture 姿态数据
type Posture struct {
	TrackingID     string `json:"tracking_id"`     // Radar tracking_id
	PostureCode    string `json:"posture_code"`    // SNOMED 编码
	PostureDisplay string `json:"posture_display"` // 显示名称

	// 位置数据（来自 Radar，单位：厘米；原样透传 position_z）
	PositionX *int `json:"position_x,omitempty"`
	PositionY *int `json:"position_y,omitempty"`
	PositionZ *int `json:"position_z,omitempty"`

	// 区域ID（来自 Radar）
	AreaID *int `json:"area_id,omitempty"` // area_id（床区域ID）
}
