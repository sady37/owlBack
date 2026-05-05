package models

// AlarmConfig 报警配置（从 alarm_cloud 和 alarm_device 合并）
type AlarmConfig struct {
	TenantID    string
	DeviceID    string
	DeviceType  string // "Radar" 或 "Sleepace"
	
	// 通用报警级别
	OfflineAlarm    string // DISABLE, EMERGENCY, WARNING, ERROR, INFORMATIONAL
	LowBattery      string
	DeviceFailure   string
	
	// 设备特定报警配置（从 alarm_device.monitor_config.alarms 读取）
	DeviceAlarms map[string]DeviceAlarmConfig
	
	// 阈值配置（从 alarm_cloud.conditions 读取）
	Conditions *VitalAlarmConditions
	
	// 睡眠时间配置（从 alarm_device.monitor_config.sleep_period 读取）
	SleepPeriod *SleepPeriodConfig
}

// DeviceAlarmConfig 设备报警配置
type DeviceAlarmConfig struct {
	Level     string                 `json:"level"`     // EMERGENCY, WARNING, etc.
	Enabled   bool                   `json:"enabled"`   // 是否启用
	Threshold map[string]interface{} `json:"threshold,omitempty"` // 阈值配置
}

// VitalAlarmConditions 生命体征报警阈值配置
type VitalAlarmConditions struct {
	HeartRate       *VitalThreshold `json:"heart_rate,omitempty"`
	RespiratoryRate *VitalThreshold `json:"respiratory_rate,omitempty"`
}

// VitalThreshold 生命体征阈值
type VitalThreshold struct {
	EMERGENCY *ThresholdRange `json:"EMERGENCY,omitempty"`
	WARNING   *ThresholdRange `json:"WARNING,omitempty"`
	Normal    *ThresholdRange `json:"Normal,omitempty"`
}

// ThresholdRange 阈值范围
type ThresholdRange struct {
	Ranges      []Range `json:"ranges"`      // 范围列表
	DurationSec int     `json:"duration_sec"` // 持续时间（秒）
}

// Range 范围
type Range struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

// SleepPeriodConfig 睡眠时间配置
type SleepPeriodConfig struct {
	StartTime string `json:"start_time"` // "22:00"
	EndTime   string `json:"end_time"`   // "06:30"
	Timezone  string `json:"timezone"`   // "Asia/Shanghai"
}

