package alarm

import (
	"fmt"
	"strconv"
	"strings"
)

// ========== Constants (常量) ==========

// Alarm Level constants (报警级别常量)
const (
	AlarmLevelEmerg  = "EMERG"
	AlarmLevelAlert  = "ALERT"
	AlarmLevelCrit   = "CRIT"
	AlarmLevelErr    = "ERR"
	AlarmLevelWarn   = "WARNING"
	AlarmLevelNotice = "NOTICE"
	AlarmLevelInfo   = "INFO"
	AlarmLevelDebug  = "DEBUG"
)

// Common Alarm default values (通用报警默认值)
// 这些是所有设备类型都支持的通用报警项，存储在 alarm_cloud 表的 OfflineAlarm, LowBattery, DeviceFailure 字段
// 使用统一的报警级别常量，保持一致性
const (
	DefaultOfflineAlarm  = AlarmLevelErr  // 设备离线报警默认级别
	DefaultLowBattery    = AlarmLevelWarn // 低电量报警默认级别
	DefaultDeviceFailure = AlarmLevelErr  // 设备故障报警默认级别
)

// Cloud Vital Alarm Threshold default values (生理指标阈值默认值)
const (
	DefaultRespiratoryRateMin = 8  // 呼吸率nomal最小值默认值
	DefaultRespiratoryRateMax = 24 // 呼吸率nomal最大值默认值
	DefaultHeartRateMin       = 50 // 心率正nomal小值默认值
	DefaultHeartRateMax       = 95 // 心率nomal最大值默认值
)

// ========== ProcessType 常量定义 ==========
const (
	ProcessTypeImmediate           = "immediate"            // 立即触发
	ProcessTypeTimeBased          = "time_based"          // 基于时间阈值
	ProcessTypeStateBased         = "state_based"          // 基于状态变化
	ProcessTypeActivityMonitoring = "activity_monitoring"  // 活动监控
	ProcessTypeConditionalTimeBased = "conditional_time_based" // 条件时间型
	ProcessTypeBedStateChange     = "bed_state_change"    // 床位状态变化
	ProcessTypeRoomStateChange    = "room_state_change"   // 房间状态变化
)

// ExampleDefaultAlarmSettingJSON 完整的 DefaultAlarmSetting 示例 JSON，便于检查 json.Marshal(DefaultAlarmSetting) 是否正常
// is_enabled: 0=关闭 1=开启 | alarm_level: nil=无报警级别 | alarm_params: nil=无参数 | display_setting: 显示设置
const ExampleDefaultAlarmSettingJSON = `{
	"sleepad": [
	  {
		"alarm_type": "ResetTime",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "InBedTime": "21:30",
		  "OutBedTime": "07:30"
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "NapTime",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "InBedTime": "13:00",
		  "OutBedTime": "14:00"
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "ApneaHypopnea",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "duration_sec": 60
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "AbnormalHeartRate",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "min": 50,
		  "max": 100,
		  "slow_duration_sec": 120,
		  "fast_duration_sec": 120
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "AbnormalRespiratoryRate",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "min": 8,
		  "max": 24,
		  "slow_duration_sec": 120,
		  "fast_duration_sec": 120
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "LeftBed",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		   "duration_sec": 8
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "LeftBedTooLong",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "leave_minutes": 45
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "InBed",
		"is_enabled": 0,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_min": 5
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "BedSitUp",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 3
	  },
	  {
		"alarm_type": "AbnormalBodyMovement",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_min": 10
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "NoBodyMove",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_min": 60
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "NoTurnOver",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_min": 120
		},
		"display_setting": 3
	  },
	{
		"alarm_type": "SensorDetached",
		"is_enabled": 1,
		"alarm_level": "ERR",
		"alarm_params": {
		  "duration_min": 120
		},
		"display_setting": 3
	  }	  
	],
	"radar": [
	  {
		"alarm_type": "ResetTime",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "InBedTime": "21:30",
		  "OutBedTime": "07:30"
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "NapTime",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "InBedTime": "13:00",
		  "OutBedTime": "14:00"
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "MonitoringMode",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "mode": 15
		},
		"display_setting": 2
	  },
	  {
		"alarm_type": "PostureDetection",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {},
		"display_setting": 2
	  },
	  {
		"alarm_type": "Fall",
		"is_enabled": 1,
		"alarm_level": "EMERG",
				"alarm_params": {
		  "duration_sec": 60
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SuspectedFall",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 0  
	  },
	  {
		"alarm_type": "SuspectedSittingOnGround",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_sec": 90
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SittingOnGround",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_sec": 90
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SuspectedBedSitUp",
		"is_enabled": 0,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 0
	  },
	  {
		"alarm_type": "BedSitUp",
		"is_enabled": 0,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 0
	  },	  
	  {
		"alarm_type": "ApneaHypopnea",
		"is_enabled": 0,
		"alarm_level": null,
		"alarm_params": {
		  "apnea_60s_min_events": 4,
		  "apnea_120min_min_events": 7
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "AbnormalHeartRate",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "min": 50,
		  "max": 95
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "AbnormalRespiratoryRate",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "min": 8,
		  "max": 24
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "VitalsWeak",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "duration_min": 10,
		  "sensitivity": 35
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "LeftBed",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_sec": 8
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "LeftBedTooLong",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "leave_minutes": 45
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Stay",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_min": 45
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "NoActivity24h",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SignalPoor",
		"is_enabled": 0,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 3
	  },
	  {
		"alarm_type": "AngleException",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 3
	  }
	]
  }`

// ========== 报警类型 ==========

const (
	AlarmTypeOfflineAlarm   = "OfflineAlarm"
	AlarmTypeDeviceFailure  = "DeviceFailure"
	AlarmTypeDeviceRecovery = "DeviceRecovery"
	AlarmTypeUnknown        = "Unknown"

	// 通用报警类型（不区分设备，设备类型通过 device_type 字段区分）
	ApneaHypopnea           = "ApneaHypopnea"
	AbnormalHeartRate       = "AbnormalHeartRate"
	AbnormalRespiratoryRate = "AbnormalRespiratoryRate"
	LeftBed                 = "LeftBed"
	LeftBedTooLong          = "LeftBedTooLong"
	AbnormalBodyMovement    = "AbnormalBodyMovement"
	NoBodyMove              = "NoBodyMove"
	NoTurnOver              = "NoTurnOver"
	ResetTime               = "ResetTime"
	NapTime                 = "NapTime"
	SensorDetached          = "SensorDetached"
	Fall                    = "Fall"
	SuspectedFall           = "SuspectedFall"
	SittingOnGround         = "SittingOnGround"
	SuspectedSittingOnGround = "SuspectedSittingOnGround"
	BedSitUp                = "BedSitUp"
	SuspectedBedSitUp       = "SuspectedBedSitUp"
	VitalsWeak              = "VitalsWeak"
	InBed                   = "InBed"
	Stay                    = "Stay"
	NoActivity24h           = "NoActivity24h"
	WarningArea             = "WarningArea"
	SignalPoor              = "SignalPoor"
	AngleException          = "AngleException"
	MonitoringMode          = "MonitoringMode"
	PostureDetection        = "PostureDetection"

)

// ========== 配置结构：alarm_type + is_enabled(0/1) + alarm_level + alarm_params + display_setting ==========

const (
	IsEnabledOff = 0 // 关闭报警
	IsEnabledOn  = 1 // 开启
)

const (
	DisplayNone                = 0 // 不显示
	DisplayAlarmCloud          = 1 // alarm_cloud
	DisplayAlarmDevice         = 2 // alarm_device
	DisplayAlarmCloudAndDevice = 3 // alarm_cloud + alarm_device
)

const (
	ParamDurationSec = "duration_sec"
	ParamDurationMin = "duration_min"
	ParamMin         = "min"
	ParamMax         = "max"
	ParamSensitivity = "sensitivity"
)

const (
	MonitoringModePeopleTracking  = 3  // People Tracking
	MonitoringModeFallMonitoring  = 7  // Fall Monitoring
	MonitoringModeSleepMonitoring = 11 // Sleep Monitoring
	MonitoringModeFullFunction    = 15 // Full Function
)

// 报警级别数值定义（参考 TDPv2-1122.md）
// 0=EMERG, 1=ALERT, 2=CRIT, 3=ERR, 4=WARNING, 5=NOTICE, 6=INFO, 7=DEBUG, 8=Cancel/恢复
const (
	AlarmLevelIntEmerg   = 0 // EMERG: 紧急：系统不可用
	AlarmLevelIntAlert   = 1 // ALERT: 警报：必须立即采取行动（如跌倒、心率/呼吸率严重异常持续≥1分钟）
	AlarmLevelIntCrit    = 2 // CRIT: 严重：严重情况（如心率/呼吸率严重异常持续≥1分钟）
	AlarmLevelIntErr     = 3 // ERR: 错误：错误条件（如设备故障、传感器断线、角度错误）
	AlarmLevelIntWarning = 4 // WARNING: 警告：警告信息（如可疑跌倒、心率/呼吸率中度异常持续≥5分钟）
	AlarmLevelIntNotice  = 5 // NOTICE: 通知：正常但重要的事件（如配置指令下发）
	AlarmLevelIntInfo    = 6 // INFO: 信息：一般信息性消息（如设备上线、状态变化）
	AlarmLevelIntDebug   = 7 // DEBUG: 调试：调试信息
	AlarmLevelIntCancel  = 8 // Cancel: 取消/恢复（如设备上线、报警恢复）
)

// ========== Types (类型) ==========

// CloudVitalAlarmThreshold 生理指标阈值配置
// 存储在 alarm_cloud 表的 conditions JSONB 字段中
// 用于定义生理指标类报警的阈值范围
// 包含完整的 EMERGENCY/WARNING/Normal ranges
type CloudVitalAlarmThreshold struct {
	// 完整的 conditions 结构（包含 EMERGENCY/WARNING/Normal ranges）
	// Heart Rate: EMERGENCY (0-44, 116+), WARNING (45-54, 96-109), Normal (55-95)
	// Respiratory Rate: EMERGENCY (0-7, 25+),  Normal (8-24)
	Conditions *VitalAlarmConditions `json:"conditions,omitempty"`
}

// VitalAlarmConditions 完整的生理指标报警条件（包含所有级别的 ranges）
type VitalAlarmConditions struct {
	HeartRate *struct {
		EMERGENCY *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"EMERGENCY,omitempty"`
		WARNING *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"WARNING,omitempty"`
		Normal *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"Normal,omitempty"`
	} `json:"heart_rate,omitempty"`
	RespiratoryRate *struct {
		EMERGENCY *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"EMERGENCY,omitempty"`
		// WARNING: 已简化，不包含 WARNING 级别
		Normal *struct {
			Ranges []struct {
				Min *int `json:"min,omitempty"`
				Max *int `json:"max,omitempty"`
			} `json:"ranges,omitempty"`
			DurationSec int `json:"duration_sec,omitempty"`
		} `json:"Normal,omitempty"`
	} `json:"respiratory_rate,omitempty"`
}

// TenantResetTime 租户级别的作息时间配置
// 存储在 alarm_cloud 表的 metadata JSONB 字段中，或者从 device_alarms 中的 ResetTime/NapTime 项提取
type TenantResetTime struct {
	ResetTime ResetTimeParams `json:"reset_time,omitempty"`
	NapTime   NapTimeParams   `json:"nap_time,omitempty"`
}

// ResetTimeParams 重置时间参数
type ResetTimeParams struct {
	InBedTime  string `json:"InBedTime"`  // 格式: "21:30"
	OutBedTime string `json:"OutBedTime"` // 格式: "07:30"
}

// NapTimeParams 午睡时间参数
type NapTimeParams struct {
	InBedTime  string `json:"InBedTime"`  // 格式: "13:00"
	OutBedTime string `json:"OutBedTime"` // 格式: "14:00"
}

// CommonAlarm 通用报警配置（所有设备类型都支持的报警项）
type CommonAlarm struct {
	OfflineAlarm  string `json:"OfflineAlarm,omitempty"`  // 设备离线报警级别
	LowBattery    string `json:"LowBattery,omitempty"`    // 低电量报警级别
	DeviceFailure string `json:"DeviceFailure,omitempty"` // 设备故障报警级别
}

// AlarmCloudConfig 完整的 alarm_cloud 配置
// 所有配置保存在 metadata 中，用于保存到 config_versions 表
type AlarmCloudConfig struct {
	TenantResetTime TenantResetTime `json:"TenantResetTime,omitempty"`
	CommonAlarm     CommonAlarm     `json:"common_alarm,omitempty"`
	AlarmSetting    struct {
		Sleepad []AlarmItem `json:"sleepad,omitempty"`
		Radar   []AlarmItem `json:"radar,omitempty"`
	} `json:"AlarmSetting,omitempty"`
	CloudVitalAlarmThreshold CloudVitalAlarmThreshold `json:"CloudVitalAlarmThreshold,omitempty"`
}

// AlarmItem 单条报警项
// alarm_type: 报警类型 | is_enabled: 0=关闭 1=开启 | alarm_level: nil=无报警级别 | alarm_params: nil=无参数 | display_setting: 显示设置
type AlarmItem struct {
	AlarmType      string                 `json:"alarm_type"`
	IsEnabled      *int                   `json:"is_enabled,omitempty"`   // 0=关闭报警 1=开启
	AlarmLevel     *string                `json:"alarm_level,omitempty"`  // nil=无报警级别
	AlarmParams    map[string]interface{} `json:"alarm_params,omitempty"` // nil=无参数
	DisplaySetting int                    `json:"display_setting"`
}

// AlarmEnablementMapRadar Radar 设备的报警使能配置表：map[alarm_type]0/1 (0=禁用 1=启用)
// 注意：此类型已废弃，建议使用 AlarmEnablementItem
type AlarmEnablementMapRadar map[string]int

// AlarmEnablementMapSleepad Sleepad 设备的报警使能配置表：map[alarm_type]0/1 (0=禁用 1=启用)
// 注意：此类型已废弃，建议使用 AlarmEnablementItem
type AlarmEnablementMapSleepad map[string]int

// AlarmEnablementItem 报警使能配置项
// 用于 GetAlarmEnablementMap 函数的返回值
type AlarmEnablementItem struct {
	AlarmType  string // 报警类型
	IsEnabled  int    // 是否启用：0=禁用 1=启用
	AlarmLevel string // 报警级别（不为空）
}

// ========== Variables (变量) ==========

// DefaultCloudVitalAlarmThreshold 默认的生理指标阈值配置
// 包含完整的 conditions 结构，包含所有级别的 ranges
var DefaultCloudVitalAlarmThreshold = CloudVitalAlarmThreshold{
	Conditions: &VitalAlarmConditions{
		HeartRate: &struct {
			EMERGENCY *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"EMERGENCY,omitempty"`
			WARNING *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"WARNING,omitempty"`
			Normal *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"Normal,omitempty"`
		}{
			EMERGENCY: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(0), Max: intPtr(44)},
					{Min: intPtr(116), Max: nil},
				},
				DurationSec: 60,
			},
			WARNING: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(45), Max: intPtr(54)},
					{Min: intPtr(96), Max: intPtr(109)},
				},
				DurationSec: 300,
			},
			Normal: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(55), Max: intPtr(95)},
				},
				DurationSec: 0,
			},
		},
		RespiratoryRate: &struct {
			EMERGENCY *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"EMERGENCY,omitempty"`
			// WARNING: 已简化，不包含 WARNING 级别
			Normal *struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			} `json:"Normal,omitempty"`
		}{
			EMERGENCY: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(0), Max: intPtr(7)},
					{Min: intPtr(25), Max: nil},
				},
				DurationSec: 60,
			},
			Normal: &struct {
				Ranges []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				} `json:"ranges,omitempty"`
				DurationSec int `json:"duration_sec,omitempty"`
			}{
				Ranges: []struct {
					Min *int `json:"min,omitempty"`
					Max *int `json:"max,omitempty"`
				}{
					{Min: intPtr(8), Max: intPtr(24)},
				},
				DurationSec: 0,
			},
		},
	},
}

// DefaultTenantResetTime 默认的租户作息时间配置
var DefaultTenantResetTime = TenantResetTime{
	ResetTime: ResetTimeParams{
		InBedTime:  "21:30",
		OutBedTime: "07:30",
	},
	NapTime: NapTimeParams{
		InBedTime:  "13:00",
		OutBedTime: "14:00",
	},
}

// DefaultAlarmSetting 各设备报警项默认值；直接改此变量后引用，json.Marshal(DefaultAlarmSetting) 可得 JSON
// 各报警项默认值（合并顺序：本包 < alarm_cloud < vendor_config < monitor_config）
var DefaultAlarmSetting = struct {
	Sleepad []AlarmItem `json:"sleepad"`
	Radar   []AlarmItem `json:"radar"`
}{
	Sleepad: []AlarmItem{
		{
			AlarmType:  ResetTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "21:30",
				"OutBedTime": "07:30",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  NapTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "13:00",
				"OutBedTime": "14:00",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SensorDetached,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelErr),
			AlarmParams: map[string]interface{}{
				"duration_min": 120,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  ApneaHypopnea,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 30,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  AbnormalHeartRate,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamMin:            50,
				ParamMax:            100,
				"slow_duration_sec": 120,
				"fast_duration_sec": 120,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  AbnormalRespiratoryRate,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamMin:            8,
				ParamMax:            24,
				"slow_duration_sec": 120,
				"fast_duration_sec": 120,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  LeftBed,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 8,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  LeftBedTooLong,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				"leave_minutes": 45,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  InBed,
			IsEnabled:  intPtr(IsEnabledOff),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 5,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      BedSitUp,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  AbnormalBodyMovement,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 10,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  NoBodyMove,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 45,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  NoTurnOver,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 120,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
	},
	Radar: []AlarmItem{
		{
			AlarmType:  ResetTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "21:30",
				"OutBedTime": "07:30",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  NapTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "13:00",
				"OutBedTime": "14:00",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  MonitoringMode,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"mode": 15,
			},
			DisplaySetting: DisplayAlarmDevice,
		},
		{
			AlarmType:      PostureDetection,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     nil,
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmDevice,
		},
		{
			AlarmType:  Fall,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 60,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      SuspectedFall,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayNone,
		},
		{
			AlarmType:  SittingOnGround,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 90,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  BedSitUp,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{},
			DisplaySetting: DisplayNone,
		},
		{
			AlarmType:  ApneaHypopnea,
			IsEnabled:  intPtr(IsEnabledOff),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				"apnea_60s_min_events":    4,
				"apnea_120min_min_events": 7,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  AbnormalHeartRate,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamMin: 50,
				ParamMax: 95,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  AbnormalRespiratoryRate,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamMin: 8,
				ParamMax: 24,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  VitalsWeak,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 10,
				ParamSensitivity: 35,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  LeftBed,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 8,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  LeftBedTooLong,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				"leave_minutes": 45,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  Stay,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 45,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      NoActivity24h,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      SignalPoor,
			IsEnabled:      intPtr(IsEnabledOff),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      AngleException,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
	},
}

// 注意：设备上报的报警类型到报警级别的映射应该从 alarm_cloud 配置中动态获取，而不是硬编码
// 报警级别应该由 alarm_cloud.device_alarms 中的配置决定，每个租户可以有不同的配置
// 因此不再提供 DeviceAlarmLevelMapSleepace 和 DeviceAlarmLevelMapRadar 硬编码映射表
// 如果需要获取报警级别，应该：
// 1. 从 alarm_cloud.device_alarms[device_type][alarm_type] 获取配置的 alarm_level
// 2. 如果配置不存在，使用 DefaultAlarmSetting 中的默认值
// 3. 将字符串级别的 alarm_level（如 "EMERG", "WARNING"）转换为整数级别（0-8）

// ========== Functions (函数) ==========

// strPtr 返回字符串指针
func strPtr(s string) *string { return &s }

// intPtr 返回整数指针
func intPtr(i int) *int { return &i }

// GetRadarAlarmTypes 获取 Radar 设备的报警类型列表
// 注意：报警类型已统一，不再区分设备前缀，设备类型通过 device_type 字段区分
func GetRadarAlarmTypes() []string {
	return []string{AlarmTypeOfflineAlarm, AlarmTypeDeviceFailure, Fall, SuspectedFall, SittingOnGround, SuspectedSittingOnGround, SuspectedBedSitUp, ApneaHypopnea, AbnormalHeartRate, AbnormalRespiratoryRate, VitalsWeak, LeftBed, InBed, LeftBedTooLong, BedSitUp, Stay, NoActivity24h, WarningArea, SignalPoor, AngleException, MonitoringMode, NapTime, ResetTime, PostureDetection}
}

// GetSleepPadAlarmTypes 获取 SleepPad 设备的报警类型列表
// 注意：报警类型已统一，不再区分设备前缀，设备类型通过 device_type 字段区分
func GetSleepPadAlarmTypes() []string {
	return []string{AlarmTypeOfflineAlarm, AlarmTypeDeviceFailure, SensorDetached, ApneaHypopnea, AbnormalHeartRate, AbnormalRespiratoryRate, LeftBed, LeftBedTooLong, InBed, BedSitUp, AbnormalBodyMovement, NoBodyMove, NoTurnOver}
}

// GetSupportedAlarmTypes 根据设备类型获取支持的报警类型列表
func GetSupportedAlarmTypes(deviceType string) []string {
	switch deviceType {
	case "radar":
		return GetRadarAlarmTypes()
	case "Sleepad", "sleepad", "sleepace":
		return GetSleepPadAlarmTypes()
	default:
		return []string{AlarmTypeOfflineAlarm, AlarmTypeDeviceFailure}
	}
}

// IsDisplayableAlarmType 判断报警类型是否可显示
func IsDisplayableAlarmType(alarmType string) bool { return alarmType != AlarmTypeUnknown }

// GetDefaultAlarmItemsSleepPad 获取 SleepPad 设备的默认报警项列表
func GetDefaultAlarmItemsSleepPad() []AlarmItem { return DefaultAlarmSetting.Sleepad }

// GetDefaultAlarmItemsRadar 获取 Radar 设备的默认报警项列表
func GetDefaultAlarmItemsRadar() []AlarmItem { return DefaultAlarmSetting.Radar }

// GetDefaultAlarmItems 根据设备类型获取默认报警项列表
func GetDefaultAlarmItems(deviceType string) []AlarmItem {
	switch deviceType {
	case "Sleepad", "sleepace", "sleepad":
		return GetDefaultAlarmItemsSleepPad()
	case "radar":
		return GetDefaultAlarmItemsRadar()
	default:
		return nil
	}
}

// ParseTimeString 解析时间字符串 "HH:MM" 为 hour 和 minute
// 例如: "06:20" -> (6, 20, nil), "21:30" -> (21, 30, nil)
func ParseTimeString(timeStr string) (hour, minute int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s, expected HH:MM", timeStr)
	}
	hour, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}
	minute, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}
	return hour, minute, nil
}

// GetAlarmEnablementMapRadar 获取 Radar 设备的报警使能配置表
// 只返回 IsEnabled 和 AlarmLevel 都不为空的项
// LeftBed 与 LeftBedTooLong 统一设为 LeftBed（LeftBedTooLong 仅限 reset 期间）
// 注意：此函数已被注释，因为使用硬编码的 DefaultAlarmSetting，而不是从 alarm_cloud 配置中获取
// 报警使能配置应该从 alarm_cloud.device_alarms 动态获取，而不是硬编码
/*
func GetAlarmEnablementMapRadar() AlarmEnablementMapRadar {
	enablement := make(AlarmEnablementMapRadar)
	var leftBedTooLongEnabled *int

	// 先遍历，收集 LeftBedTooLong 的配置（作为后备）
	// 同时处理其他所有项
	for _, item := range DefaultAlarmSetting.Radar {
		// 只处理 IsEnabled 和 AlarmLevel 都不为空的项
		if item.IsEnabled == nil || item.AlarmLevel == nil {
			continue
		}

		if item.AlarmType == LeftBedTooLong {
			// 保存 LeftBedTooLong 的配置作为后备
			leftBedTooLongEnabled = item.IsEnabled
			continue
		}

		// 处理其他项（包括 LeftBed）
		enablement[item.AlarmType] = *item.IsEnabled
	}

	// 如果 LeftBed 不存在但 LeftBedTooLong 存在，使用 LeftBedTooLong 的配置
	if leftBedTooLongEnabled != nil {
		if _, exists := enablement[RadarLeftBed]; !exists {
			enablement[RadarLeftBed] = *leftBedTooLongEnabled
		}
	}

	return enablement
}
*/

// ShouldEnableEnter2OutEventForAlarm 判断是否需要启用 event type=1（进出事件）监听以生成报警
// 参数：alarmType - 报警类型（如 RadarStay, NoActivity24h, RadarWarningArea, Radar_InBed, Radar_LeftBedTooLong）
// 返回：是否需要启用进出事件监听
// 简化逻辑：如果启用了以下任一报警类型，所有 event type=1（进出事件）都转为 alarm
// - RadarStay（滞留）
// - NoActivity24h（长时间无人活动）
// - RadarWarningArea（警告区域）
// - Radar_InBed/Radar_LeftBedTooLong（进床/离床）
// 注意：此函数已被注释，因为使用硬编码的逻辑，应该基于 alarm_cloud 配置动态判断
/*
func ShouldEnableEnter2OutEventForAlarm(alarmType string) bool {
	switch alarmType {
	case RadarStay, NoActivity24h, RadarWarningArea, RadarInBed, LeftBedTooLong:
		return true
	default:
		return false
	}
}
*/

// ShouldEnablePoseEventForAlarm 判断是否需要启用 event type=2（姿态变化事件）监听以生成报警
// 参数：alarmType - 报警类型（如 RadarFall, RadarSittingOnGround, RadarBedSitUp）
// 返回：是否需要启用姿态变化事件监听
// 简化逻辑：如果启用了以下任一报警类型，所有 event type=2（姿态变化事件）都转为 alarm
// - RadarFall（跌倒）
// - RadarSittingOnGround（坐地）
// - RadarBedSitUp（床上坐起）
// 注意：此函数已被注释，因为使用硬编码的逻辑，应该基于 alarm_cloud 配置动态判断
/*
func ShouldEnablePoseEventForAlarm(alarmType string) bool {
	switch alarmType {
	case RadarFall, SuspectedFall, RadarSittingOnGround, RadarSuspectedSittingOnGround, RadarBedSitUp, RadarSuspectedBedSitUp:
		return true
	default:
		return false
	}
}
*/

// GetAlarmEnablementMapSleepad 获取 Sleepad 设备的报警使能配置表
// 只返回 IsEnabled 和 AlarmLevel 都不为空的项
// LeftBed 与 LeftBedTooLong 统一设为 LeftBed（LeftBedTooLong 仅限 reset 期间）
// 注意：此函数已被注释，因为使用硬编码的 DefaultAlarmSetting，而不是从 alarm_cloud 配置中获取
// 报警使能配置应该从 alarm_cloud.device_alarms 动态获取，而不是硬编码
/*
func GetAlarmEnablementMapSleepad() AlarmEnablementMapSleepad {
	enablement := make(AlarmEnablementMapSleepad)
	var leftBedTooLongEnabled *int

	// 先遍历，收集 LeftBedTooLong 的配置（作为后备）
	// 同时处理其他所有项
	for _, item := range DefaultAlarmSetting.Sleepad {
		// 只处理 IsEnabled 和 AlarmLevel 都不为空的项
		if item.IsEnabled == nil || item.AlarmLevel == nil {
			continue
		}

		if item.AlarmType == LeftBedTooLong {
			// 保存 LeftBedTooLong 的配置作为后备
			leftBedTooLongEnabled = item.IsEnabled
			continue
		}

		// 处理其他项（包括 LeftBed）
		enablement[item.AlarmType] = *item.IsEnabled
	}

	// 如果 LeftBed 不存在但 LeftBedTooLong 存在，使用 LeftBedTooLong 的配置
	if leftBedTooLongEnabled != nil {
		if _, exists := enablement[LeftBed]; !exists {
			enablement[LeftBed] = *leftBedTooLongEnabled
		}
	}

	return enablement
}
*/

// GetAlarmEnablementMap 获取设备的报警使能配置表
// 参数：deviceType - 设备类型（"Sleepad"/"sleepace"/"sleepad" 或 "radar"）
// 参数：alarmItems - 报警项列表（从 alarm_cloud.device_alarms 获取，如果为 nil 则使用 DefaultAlarmSetting）
// 返回：[]AlarmEnablementItem，只包含 isEnabled=1 且 alarm_level 不为空的项
// 用于过滤决定是否将事件发布到 iot:alarm:stream
func GetAlarmEnablementMap(deviceType string, alarmItems []AlarmItem) []AlarmEnablementItem {
	// 如果 alarmItems 为 nil，使用默认配置
	if alarmItems == nil {
		switch deviceType {
		case "Sleepad", "sleepace", "sleepad":
			alarmItems = DefaultAlarmSetting.Sleepad
		case "radar":
			alarmItems = DefaultAlarmSetting.Radar
		default:
			return []AlarmEnablementItem{}
		}
	}

	result := make([]AlarmEnablementItem, 0)
	for _, item := range alarmItems {
		// 过滤条件：isEnabled=1 且 alarm_level 不为空
		if item.IsEnabled == nil || *item.IsEnabled != IsEnabledOn {
			continue
		}
		if item.AlarmLevel == nil || *item.AlarmLevel == "" {
			continue
		}

		result = append(result, AlarmEnablementItem{
			AlarmType:  item.AlarmType,
			IsEnabled:  *item.IsEnabled,
			AlarmLevel: *item.AlarmLevel,
		})
	}

	return result
}

// MQTTToAlarmTypeMapSleepace Sleepace MQTT 消息到 alarm_type 的映射表
// 基于文档 radar-Qinlan-code-v3.0.md (598-610)
// 注意：报警类型已统一，不再区分设备前缀，设备类型通过 device_type 字段区分
var MQTTToAlarmTypeMapSleepace = map[string]string{
	"alarmLeftBed":         LeftBed, // 或 LeftBedTooLong（根据上下文判断）
	"alarmHeartRateFast":   AbnormalHeartRate,
	"alarmHeartRateSlow":   AbnormalHeartRate,
	"alarmBreathRateFast":  AbnormalRespiratoryRate,
	"alarmBreathRateSlow":  AbnormalRespiratoryRate,
	"alarmBreathRatePause": ApneaHypopnea,
	"alarmBodymove":        AbnormalBodyMovement,
	"alarmNoBodymove":      NoBodyMove,
	"alarmNoTurnOver":      NoTurnOver,
	"alarmBedSitup":        BedSitUp,
	"alarmInBed":           InBed,   //En 在床多用InBed
	"alarmSensorFall":      SensorDetached,
	"offLine":              "", // 离线报警由通用报警处理，不映射到具体 alarm_type
}

// MQTTToAlarmTypeMapRadar Radar MQTT 消息到 alarm_type 的映射表
// 基于文档 radar-Qinlan-code-v3.0.md (612-636)
// 注意：需要根据 event_type, area_type, pose 等字段进行判断
// 注意：报警类型已统一，不再区分设备前缀，设备类型通过 device_type 字段区分
var MQTTToAlarmTypeMapRadar = map[string]string{
	// Event type=1 (进出事件)
	"event_type_1_room":              Stay,        // 进出房间
	"event_type_1_area_2_or_5":       InBed,       // 进出区域+Area_type={2||5}
	"event_type_1_area_6":            WarningArea, // 进入区域+Area_type=6
	"event_type_1_left_bed_too_long": LeftBedTooLong,
	// Event type=2 (姿态变化)
	"event_type_2_pose_5":  Fall,                    // 5-确认跌倒
	"event_type_2_pose_2":  SuspectedFall,           // 2-疑似跌倒
	"event_type_2_pose_7":  SuspectedSittingOnGround, // 7-疑似坐地
	"event_type_2_pose_8":  SittingOnGround,         // 8-确认坐地
	"event_type_2_pose_10": SuspectedBedSitUp,       // 10-疑似床上坐起
	"event_type_2_pose_11": BedSitUp,                // 11-确认床上坐起
	// Event type=7
	"event_type_7_signal_poor": SignalPoor, // 信号差事件
	// Event type=8
	"event_type_8_angle_abnormal": AngleException, // 倾角异常事件
	// Statistics (sleep)
	"stat_sleep_breath_01":   AbnormalRespiratoryRate, // 01: 呼吸过低
	"stat_sleep_breath_10":   AbnormalRespiratoryRate, // 10: 呼吸过高
	"stat_sleep_heart_01":    AbnormalHeartRate,       // 01: 心率过低
	"stat_sleep_heart_10":    AbnormalHeartRate,       // 10: 心率过高
	"stat_sleep_breath_11":   ApneaHypopnea,           // 11: 呼吸暂停
	"stat_sleep_vitals_11":   VitalsWeak,              // 11: 生命体征弱
	"stat_sleep_stay":        Stay,                    // 滞留
	"stat_sleep_no_activity": NoActivity24h,           // 长时间无人活动
}

// ConvertMQTTToAlarmTypeSleepace 将 Sleepace MQTT 消息转换为 alarm_type
// 参数：mqttKey - MQTT 消息中的键（如 "alarmLeftBed", "alarmHeartRateFast" 等）
// 返回：对应的 alarm_type，如果未找到则返回空字符串
func ConvertMQTTToAlarmTypeSleepace(mqttKey string) string {
	if alarmType, ok := MQTTToAlarmTypeMapSleepace[mqttKey]; ok {
		return alarmType
	}
	return ""
}

// ConvertMQTTToAlarmTypeRadar 将 Radar MQTT 消息转换为 alarm_type
// 参数：eventType - event_type 字段值（1=进出事件, 2=姿态变化, 7=信号差, 8=倾角异常）
// 参数：areaType - area_type 字段值（用于 event_type=1）
// 参数：pose - pose 字段值（用于 event_type=2）
// 参数：statType - statistics 类型（用于 sleep 统计）
// 参数：statAlarmType - statistics 中的 alarm_type 字段值
// 返回：对应的 alarm_type，如果未找到则返回空字符串
func ConvertMQTTToAlarmTypeRadar(eventType, areaType, pose, statType, statAlarmType string) string {
	// Event type=1 (进出事件)
	if eventType == "1" {
		if areaType == "6" {
			return WarningArea
		}
		if areaType == "2" || areaType == "5" {
			// 注意：RadarInBed 和 LeftBedTooLong 需要根据具体业务逻辑判断
			// 这里先返回 RadarInBed，实际使用时可能需要根据其他字段（如 duration）判断
			return InBed
		}
		// 进出房间
		return Stay
	}

	// Event type=2 (姿态变化)
	if eventType == "2" {
		switch pose {
		case "5":
			return Fall
		case "2":
			return SuspectedFall
		case "7":
			return SuspectedSittingOnGround
		case "8":
			return SittingOnGround
		case "10":
			return SuspectedBedSitUp
		case "11":
			return BedSitUp
		}
	}

	// Event type=7 (信号差)
	if eventType == "7" {
		return SignalPoor
	}

	// Event type=8 (倾角异常)
	if eventType == "8" {
		return AngleException
	}

	// Statistics (sleep)
	if statType == "sleep" {
		switch statAlarmType {
		case "01": // 呼吸过低或心率过低
			// 需要根据具体字段判断是呼吸还是心率
			// 这里假设有额外的字段来区分，实际使用时需要根据 MQTT 消息结构调整
			return AbnormalRespiratoryRate // 或 AbnormalHeartRate
		case "10": // 呼吸过高或心率过高
			return AbnormalRespiratoryRate // 或 AbnormalHeartRate
		case "11": // 呼吸暂停或生命体征弱
			return ApneaHypopnea // 或 VitalsWeak
		}
	}

	return ""
}

// ShouldEnableEventForAlarm 判断 MQTT 事件是否应该转换为报警
// 参数：deviceType - 设备类型（"Sleepad"/"sleepace"/"sleepad" 或 "radar"）
// 参数：alarmType - 报警类型（通过 ConvertMQTTToAlarmType* 函数转换得到）
// 参数：enablementMap - 报警使能配置表（通过 GetAlarmEnablementMap 函数获取）
// 返回：是否应该转换为报警（true=转换为 alarm topic，false=保持原 topic）
func ShouldEnableEventForAlarm(deviceType, alarmType string, enablementMap []AlarmEnablementItem) bool {
	if alarmType == "" {
		return false
	}

	// 在 enablementMap 中查找对应的 alarm_type
	for _, item := range enablementMap {
		if item.AlarmType == alarmType {
			// 找到匹配项，且 isEnabled=1，alarm_level 不为空，应该转换为报警
			return true
		}
	}

	return false
}

// AlarmTypeToNumericCode 报警类型到数字组合的映射
// 用于将标准报警项转换为数字组合，便于在 device_monitor 中查找配置
// 格式：event 类型使用 eventType+event+area_type，statistics 类型使用不同的编码
type AlarmNumericCode struct {
	EventType int    // 事件类型 (1=进出事件, 2=姿态变化, 7=信号差, 8=倾角异常)
	Event     int    // 事件值 (用于 event_type=1, 如 event=4 表示离床)
	AreaType  int    // 区域类型 (用于 event_type=1, 2=普通床, 5=监护床)
	Pose      int    // 姿态值 (用于 event_type=2)
	StatType  string // 统计类型 (用于 statistics, 如 "sleep")
	StatCode  string // 统计代码 (用于 statistics, 如 "01", "10", "11")
	Source    string // 数据源: "event" 或 "statistics" 或 "both"
}

// AlarmTypeToNumericCodeMap 报警类型到数字组合的映射表
// 注意：event 和 statistics 分别转换，因为有些在event，有些在statistics，有些同时在两个里（如fall）
// 注意：报警类型已统一，不再区分设备前缀，设备类型通过 device_type 字段区分
var AlarmTypeToNumericCodeMap = map[string][]AlarmNumericCode{
	// Event 类型报警
	LeftBed: {
		{EventType: 1, Event: 4, AreaType: 2, Source: "event"}, // 普通床离床
		{EventType: 1, Event: 4, AreaType: 5, Source: "event"}, // 监护床离床
	},
	LeftBedTooLong: {
		{EventType: 1, Event: 4, AreaType: 2, Source: "event"}, // 普通床离床过长
		{EventType: 1, Event: 4, AreaType: 5, Source: "event"}, // 监护床离床过长
	},
	InBed: {
		{EventType: 1, Event: 4, AreaType: 2, Source: "event"}, // 普通床在床
		{EventType: 1, Event: 4, AreaType: 5, Source: "event"}, // 监护床在床
	},
	WarningArea: {
		{EventType: 1, Event: 1, AreaType: 6, Source: "event"}, // 进入警告区域
	},
	Fall: {
		{EventType: 2, Pose: 5, Source: "event"}, // 确认跌倒（event）
	},
	SuspectedFall: {
		{EventType: 2, Pose: 2, Source: "event"}, // 疑似跌倒（event）
	},
	SittingOnGround: {
		{EventType: 2, Pose: 8, Source: "event"}, // 确认坐地
	},
	BedSitUp: {
		{EventType: 2, Pose: 11, Source: "event"}, // 确认床上坐起
	},
	SignalPoor: {
		{EventType: 7, Source: "event"}, // 信号差
	},
	AngleException: {
		{EventType: 8, Source: "event"}, // 倾角异常
	},
	// Statistics 类型报警（4组状态分别映射）
	// bit 1 & bit 0: 呼吸状态 (00=正常, 01=过低, 10=过高, 11=暂停)
	AbnormalRespiratoryRate: {
		{StatType: "sleep", StatCode: "breath_01", Source: "statistics"}, // 呼吸过低
		{StatType: "sleep", StatCode: "breath_10", Source: "statistics"}, // 呼吸过高
	},
	ApneaHypopnea: {
		{StatType: "sleep", StatCode: "breath_11", Source: "statistics"}, // 呼吸暂停
	},
	// bit 3 & bit 2: 心率状态 (00=正常, 01=过低, 10=过高, 11=未定义)
	AbnormalHeartRate: {
		{StatType: "sleep", StatCode: "heart_01", Source: "statistics"}, // 心率过低
		{StatType: "sleep", StatCode: "heart_10", Source: "statistics"}, // 心率过高
	},
	// bit 5 & bit 4: 生命体征情况 (00=正常, 01=未定义, 02=未定义, 11=生命体征弱)
	VitalsWeak: {
		{StatType: "sleep", StatCode: "vitals_11", Source: "statistics"}, // 生命体征弱
	},
	// bit 7 & bit 6: 睡眠状态 (00=未定义, 01=浅睡, 10=深睡, 11=清醒) - 通常不是报警
}

// GetNumericCodesForAlarmType 获取报警类型对应的所有数字组合
// 参数：alarmType - 报警类型（如 RadarLeftBed, RadarFall）
// 返回：数字组合列表，可能包含多个组合（如 LeftBed 对应普通床和监护床）
func GetNumericCodesForAlarmType(alarmType string) []AlarmNumericCode {
	if codes, ok := AlarmTypeToNumericCodeMap[alarmType]; ok {
		return codes
	}
	return []AlarmNumericCode{}
}

// NumericCodeToAlarmTypeMap 数字组合到报警类型的反向映射表
// 用于根据数字组合（如"142"）直接查找对应的报警类型列表
// 注意：同一个数字组合可能对应多个报警类型（如"142"对应 LeftBed/LeftBedTooLong/InBed），需要根据event值区分
// 格式：数字组合字符串 -> 报警类型列表
var NumericCodeToAlarmTypeMap = map[string][]string{
	// Event type=1 (进出事件)
	// "142" 和 "145" 可能对应多个报警类型，需要根据event值（4=离床, 3=在床等）进一步判断
	// 这里先列出所有可能的组合，实际使用时需要结合event字段判断
	"142": {LeftBed, LeftBedTooLong, InBed}, // eventType=1, event=4, area_type=2 (普通床：离床/离床过长/在床)
	"145": {LeftBed, LeftBedTooLong, InBed}, // eventType=1, event=4, area_type=5 (监护床：离床/离床过长/在床)
	"116": {WarningArea},                    // eventType=1, event=1, area_type=6 (进入警告区域)
	// Event type=2 (姿态变化)
	"25":  {Fall},            // eventType=2, pose=5 (确认跌倒)
	"22":  {Fall},   // eventType=2, pose=2 (疑似跌倒)
	"27":  {SittingOnGround}, // eventType=2, pose=7 (疑似坐地)
	"28":  {SittingOnGround}, // eventType=2, pose=8 (确认坐地)
	"210": {BedSitUp},        // eventType=2, pose=10 (疑似床上坐起)
	"211": {BedSitUp},        // eventType=2, pose=11 (确认床上坐起)
	// Event type=7, 8
	"7": {SignalPoor},     // eventType=7 (信号差)
	"8": {AngleException}, // eventType=8 (倾角异常)
	// Statistics (sleep) - 4组状态分别映射
	// bit 1 & bit 0: 呼吸状态
	"sleep_breath_01": {AbnormalRespiratoryRate}, // 呼吸过低
	"sleep_breath_10": {AbnormalRespiratoryRate}, // 呼吸过高
	"sleep_breath_11": {ApneaHypopnea},           // 呼吸暂停
	// bit 3 & bit 2: 心率状态
	"sleep_heart_01": {AbnormalHeartRate}, // 心率过低
	"sleep_heart_10": {AbnormalHeartRate}, // 心率过高
	// bit 5 & bit 4: 生命体征情况
	"sleep_vitals_11": {VitalsWeak}, // 生命体征弱
	// bit 7 & bit 6: 睡眠状态 (00/01/10/11) - 通常不是报警，不映射
}

// GetAlarmTypesFromNumericCode 根据数字组合获取所有可能的报警类型
// 参数：numericCode - 数字组合字符串（如"142", "25", "210"）
// 返回：对应的报警类型列表，如果未找到则返回空列表
func GetAlarmTypesFromNumericCode(numericCode string) []string {
	if alarmTypes, ok := NumericCodeToAlarmTypeMap[numericCode]; ok {
		return alarmTypes
	}
	return []string{}
}

// CheckAlarmEnabledByNumericCode 根据数字组合检查报警是否启用
// 参数：numericCode - 数字组合字符串（如"142"）
// 参数：enablementMap - 报警使能配置表
// 返回：是否启用（true=任一匹配的报警类型启用，false=全部未启用或未找到）
// 注意：如果数字组合对应多个报警类型，任一启用即返回true
func CheckAlarmEnabledByNumericCode(numericCode string, enablementMap []AlarmEnablementItem) bool {
	alarmTypes := GetAlarmTypesFromNumericCode(numericCode)
	if len(alarmTypes) == 0 {
		return false
	}
	// 检查任一报警类型是否启用
	for _, alarmType := range alarmTypes {
		for _, item := range enablementMap {
			if item.AlarmType == alarmType && item.IsEnabled == 1 {
				return true
			}
		}
	}
	return false
}

// GetAlarmEnabledMapByNumericCodes 根据数字组合列表获取启用状态映射表
// 参数：numericCodes - 数字组合列表（如["142", "145", "25"]）
// 参数：enablementMap - 报警使能配置表
// 返回：数字组合 -> 启用状态的映射表，例如：{"142": 0, "145": 1, "25": 0}
func GetAlarmEnabledMapByNumericCodes(numericCodes []string, enablementMap []AlarmEnablementItem) map[string]int {
	result := make(map[string]int)
	for _, code := range numericCodes {
		enabled := CheckAlarmEnabledByNumericCode(code, enablementMap)
		if enabled {
			result[code] = 1
		} else {
			result[code] = 0
		}
	}
	return result
}

// ExtractNumericCodesFromEvent 从 event 数据中提取数字组合
// 参数：eventData - 解码后的事件数据（map[string]interface{}）
// 返回：数字组合列表，例如：["142", "25", "7"]
func ExtractNumericCodesFromEvent(eventData map[string]interface{}) []string {
	var codes []string

	// 获取 event_type
	eventType, _ := eventData["event_type"].(string)
	if eventType == "" {
		// 如果没有 event_type，尝试从 type 字段获取
		if typeInt, ok := eventData["type"].(int); ok {
			eventType = fmt.Sprintf("%d", typeInt)
		}
	}

	// Event type=1 (进出事件)
	if eventType == "1" {
		// 获取 area_type
		areaType, _ := eventData["area_type_raw"].(string)
		if areaType == "" {
			// 如果没有 area_type_raw，尝试从 area_type 获取（可能是 int）
			if areaTypeInt, ok := eventData["area_type"].(int); ok {
				areaType = fmt.Sprintf("%d", areaTypeInt)
			}
		}

		// 获取 event 的原始值（优先从 event_raw 获取）
		var eventValue string
		if eventRaw, ok := eventData["event_raw"].(string); ok {
			eventValue = eventRaw
		} else if eventRawInt, ok := eventData["event_raw"].(int); ok {
			eventValue = fmt.Sprintf("%d", eventRawInt)
		} else if eventDisplay, ok := eventData["event"].(string); ok {
			// 如果没有 event_raw，从 display_en 反向查找原始值（与 radar_convert_table 全小写一致，不区分大小写）
			switch strings.ToLower(strings.TrimSpace(eventDisplay)) {
			case "enter room":
				eventValue = "1"
			case "leave room":
				eventValue = "2"
			case "enter area":
				eventValue = "3"
			case "leave area":
				eventValue = "4"
			}
		}

		// 生成数字组合：eventType + event + areaType
		if eventValue != "" && areaType != "" {
			code := fmt.Sprintf("%s%s%s", eventType, eventValue, areaType)
			codes = append(codes, code)
		} else if areaType == "6" {
			// 进入警告区域：eventType=1, event=1, area_type=6 -> "116"
			codes = append(codes, "116")
		}
	}

	// Event type=2 (姿态变化)
	if eventType == "2" {
		// 获取 pose 的原始值（decoder 应设置 pose_raw；JSON 解析后可能为 float64）
		pose, _ := eventData["pose_raw"].(string)
		if pose == "" {
			if poseInt, ok := eventData["pose"].(int); ok {
				pose = fmt.Sprintf("%d", poseInt)
			} else if poseFloat, ok := eventData["pose"].(float64); ok {
				pose = fmt.Sprintf("%.0f", poseFloat)
			}
		}
		if pose != "" {
			// 生成数字组合：eventType + pose -> "25", "210" 等
			code := fmt.Sprintf("%s%s", eventType, pose)
			codes = append(codes, code)
		}
	}

	// Event type=7 (信号差)
	if eventType == "7" {
		codes = append(codes, "7")
	}

	// Event type=8 (倾角异常)
	if eventType == "8" {
		codes = append(codes, "8")
	}

	return codes
}

// reverseMapPoseFromSNOMED 将 SNOMED display_en 字符串反向映射回数字 pose 值
// 用于 ExtractNumericCodesFromEvent 中处理 SNOMED 映射后的 pose 字符串
func reverseMapPoseFromSNOMED(displayEn string) string {
	// SNOMED 映射表（从文档 Reside_stream_stand.md 和 radar_convert_table.json）
	// 注意：这里使用 display_en 值进行反向映射
	poseMap := map[string]string{
		"Initialization":     "0",
		"Walking":            "1",
		"SuspectedFall":      "2",
		"Sitting":            "3",
		"Standing":           "4",
		"Fall":               "5",
		"Lying":              "6",
		"SuspectedSittingOnGround": "7",
		"SittingOnGround":           "8",
		"BedSitUp":           "9",  // pose 9 和 11 都是 BedSitUp，优先使用 pose 9
		"SuspectedBedSitUp":  "10",
		// 注意：pose 11 也是 "BedSitUp"，但由于无法区分，优先映射到 pose 9
	}
	
	// 精确匹配
	if pose, ok := poseMap[displayEn]; ok {
		return pose
	}
	
	// 不区分大小写的匹配
	for key, pose := range poseMap {
		if strings.EqualFold(key, displayEn) {
			return pose
		}
	}
	
	// 如果无法映射，返回空字符串（ExtractNumericCodesFromEvent 会跳过）
	return ""
}

// ExtractNumericCodesFromStat 从 stat 数据中提取数字组合
// 参数：statData - 解码后的统计数据（map[string]interface{}）
// 返回：数字组合列表，例如：["sleep_breath_01", "sleep_heart_10", "sleep_vitals_11"]
func ExtractNumericCodesFromStat(statData map[string]interface{}) []string {
	var codes []string

	// 检查是否有 stat_numeric_codes 字段（decoder 已生成）
	if numericCodes, ok := statData["stat_numeric_codes"].([]interface{}); ok {
		for _, code := range numericCodes {
			if codeStr, ok := code.(string); ok {
				codes = append(codes, codeStr)
			}
		}
		return codes
	}

	// 如果没有 stat_numeric_codes，从各个字段提取
	// sleep 类型
	if category, _ := statData["category"].(string); category == "sleep" {
		// 检查是否有 stat_numeric_code_breath
		if breathCode, ok := statData["stat_numeric_code_breath"].(string); ok && breathCode != "" {
			codes = append(codes, breathCode)
		}
		// 检查是否有 stat_numeric_code_heart
		if heartCode, ok := statData["stat_numeric_code_heart"].(string); ok && heartCode != "" {
			codes = append(codes, heartCode)
		}
		// 检查是否有 stat_numeric_code_vitals
		if vitalsCode, ok := statData["stat_numeric_code_vitals"].(string); ok && vitalsCode != "" {
			codes = append(codes, vitalsCode)
		}
	}

	return codes
}

// ShouldConvertToAlarm 根据使能表决定是否将 event/stat 转换为 alarm
// 参数：dataValue - 解码后的数据（event 或 stat 的单个对象）
// 参数：topicType - 主题类型（"event" 或 "stat"）
// 参数：enablementMap - 报警使能配置表
// 返回：是否应该转换为 alarm（true=转换为 alarm，false=保持原 topic）
func ShouldConvertToAlarm(dataValue map[string]interface{}, topicType string, enablementMap []AlarmEnablementItem) bool {
	var numericCodes []string

	// 根据 topicType 提取数字组合
	if topicType == "event" {
		numericCodes = ExtractNumericCodesFromEvent(dataValue)
	} else if topicType == "stat" {
		numericCodes = ExtractNumericCodesFromStat(dataValue)
	}

	if len(numericCodes) == 0 {
		return false
	}

	// 生成使能表
	enabledMap := GetAlarmEnabledMapByNumericCodes(numericCodes, enablementMap)

	// 如果任一数字组合对应的报警启用，则转换为 alarm
	for _, enabled := range enabledMap {
		if enabled == 1 {
			return true
		}
	}

	return false
}

// ========== Radar Event/Stat 到 Alarm_Type 映射表 ==========

// RadarEventStatToAlarmMapping Radar Event/Stat 到 Alarm_Type 的映射规则
// 用于 card-aggregator 处理所有 event/stat 到 alarm_type 的映射
type RadarEventStatToAlarmMapping struct {
	// 匹配条件
	TopicType string  // "event" | "stat" | "monitor" | "alarm"
	Category  string  // category 字段值（如 "enter2out", "track", "sleep", "Fall", "SuspectedFall"）
	EventType *int    // event_type 字段值（1=进出, 2=姿态, 7=信号差, 8=倾角）
	AreaType  *int    // area_type 字段值（2=普通床, 5=监护床, 6=警告区）
	Pose      *int    // pose 字段值（2=疑似跌倒, 5=确认跌倒, 7=疑似坐地, 8=确认坐地等）
	StatType  *string // stat 类型（"sleep"）
	StatCode  *string // stat 代码（"breath_01", "heart_10", "stay", "no_activity" 等）

	// 映射结果
	AlarmType   string // 对应的 alarm_type（可为空，如床位/房间状态变化仅更新状态不产生报警）
	ProcessType string // ProcessTypeImmediate | ProcessTypeTimeBased | ...

	// 时间相关参数（仅 ProcessType=ProcessTypeTimeBased 时使用）
	DurationSec *int    // 持续时间阈值（秒），如 60 表示持续60秒后触发
	UpgradeTo   *string // 升级目标 alarm_type（如 SuspectedFall 持续60秒升级为 Fall）

	Description string // 规则说明，便于维护与文档
}



// ========== 完整的 Radar Event/Stat 到 Alarm 映射表 ==========
var RadarEventStatToAlarmMap = []RadarEventStatToAlarmMapping{
	// ========== 即时触发型报警（已在 qinglan 转换为 alarm）==========
	{
		TopicType:    "alarm",
		Category:     "Fall",
		AlarmType:    Fall,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Confirmed fall alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "SuspectedFall",
		AlarmType:    SuspectedFall,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Suspected fall alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "SittingOnGround",
		AlarmType:    SittingOnGround,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Confirmed sitting on ground alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "SuspectedSittingOnGround",
		AlarmType:    SuspectedSittingOnGround,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Suspected sitting on ground alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "BedSitUp",
		AlarmType:    BedSitUp,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Confirmed bed sit-up alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "SuspectedBedSitUp",
		AlarmType:    SuspectedBedSitUp,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Suspected bed sit-up alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "WarningArea",
		AlarmType:    WarningArea,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Enter warning area alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "SignalPoor",
		AlarmType:    SignalPoor,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Radar signal poor alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "AngleException",
		AlarmType:    AngleException,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Radar angle exception alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "OfflineAlarm",
		AlarmType:    AlarmTypeOfflineAlarm,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Device offline alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "Stay",
		AlarmType:    Stay,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Stay alarm (evaluated in qinglan)",
	},
	{
		TopicType:    "alarm",
		Category:     "NoActivity24h",
		AlarmType:    NoActivity24h,
		ProcessType:  ProcessTypeImmediate,
		Description:  "24h no activity alarm (evaluated in qinglan)",
	},
	{
		TopicType:    "alarm",
		Category:     "VitalsWeak",
		AlarmType:    VitalsWeak,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Weak vitals alarm",
	},
	{
		TopicType:    "alarm",
		Category:     "LeftBed",
		AlarmType:    LeftBed,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Left bed alarm (evaluated in qinglan)",
	},
	{
		TopicType:    "alarm",
		Category:     "LeftBedTooLong",
		AlarmType:    LeftBedTooLong,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Left bed too long alarm (evaluated in qinglan)",
	},
	{
		TopicType:    "alarm",
		Category:     "InBed",
		AlarmType:    InBed,
		ProcessType:  ProcessTypeImmediate,
		Description:  "In bed alarm (evaluated in qinglan)",
	},

	// ========== 设备在线状态事件 ==========
	{
		TopicType:    "alarm",
		Category:     "isOnline",
		AlarmType:    AlarmTypeDeviceRecovery,
		ProcessType:  ProcessTypeStateBased,
		Description:  "Device recovered online, update device state",
	},

	// ========== Event Stream 中的事件 ==========
	// Event type=1 (进出事件)
	{
		TopicType:    "event",
		Category:     "enter2out",
		EventType:    intPtr(1),
		AreaType:     intPtr(6), // 警告区域
		AlarmType:    WarningArea,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Enter warning area event",
	},
	{
		TopicType:    "event",
		Category:     "enter2out",
		EventType:    intPtr(1),
		AreaType:     intPtr(2), // 普通床
		ProcessType:  ProcessTypeBedStateChange,
		AlarmType:    "",
		Description:  "Bed area enter/exit, update bed state",
	},
	{
		TopicType:    "event",
		Category:     "enter2out",
		EventType:    intPtr(1),
		AreaType:     intPtr(5), // 监护床
		ProcessType:  ProcessTypeBedStateChange,
		AlarmType:    "",
		Description:  "Monitor bed enter/exit, update bed state",
	},
	{
		TopicType:    "event",
		Category:     "enter2out",
		EventType:    intPtr(1),
		AreaType:     intPtr(4), // 房间
		ProcessType:  ProcessTypeRoomStateChange,
		AlarmType:    "",
		Description:  "Room enter/exit, update room state",
	},

	// Event type=2 (姿态变化) - 时间相关处理
	{
		TopicType:    "event",
		Category:     "pose",
		EventType:    intPtr(2),
		Pose:         intPtr(2), // 疑似跌倒
		AlarmType:    SuspectedFall,
		ProcessType:  ProcessTypeTimeBased,
		DurationSec:  intPtr(60),
		UpgradeTo:    strPtr(Fall),
		Description:  "Suspected fall, upgrade to confirmed fall after 60s",
	},
	{
		TopicType:    "event",
		Category:     "pose",
		EventType:    intPtr(2),
		Pose:         intPtr(7), // 疑似坐地
		AlarmType:    SuspectedSittingOnGround,
		ProcessType:  ProcessTypeTimeBased,
		DurationSec:  intPtr(60),
		UpgradeTo:    strPtr(SittingOnGround),
		Description:  "Suspected sitting on ground, upgrade after 60s",
	},
	{
		TopicType:    "event",
		Category:     "pose",
		EventType:    intPtr(2),
		Pose:         intPtr(10), // 疑似床上坐起
		AlarmType:    SuspectedBedSitUp,
		ProcessType:  ProcessTypeTimeBased,
		DurationSec:  intPtr(60),
		UpgradeTo:    strPtr(BedSitUp),
		Description:  "Suspected bed sit-up, upgrade after 60s",
	},
	// Event type=2 (姿态变化) - 即时触发
	{
		TopicType:    "event",
		Category:     "pose",
		EventType:    intPtr(2),
		Pose:         intPtr(5), // 确认跌倒
		AlarmType:    Fall,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Confirmed fall event",
	},
	{
		TopicType:    "event",
		Category:     "pose",
		EventType:    intPtr(2),
		Pose:         intPtr(8), // 确认坐地
		AlarmType:    SittingOnGround,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Confirmed sitting on ground event",
	},
	{
		TopicType:    "event",
		Category:     "pose",
		EventType:    intPtr(2),
		Pose:         intPtr(11), // 确认床上坐起
		AlarmType:    BedSitUp,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Confirmed bed sit-up event",
	},

	// Event type=3 (number-people) - 人数变化
	{
		TopicType:    "event",
		Category:     "number-people",
		EventType:    intPtr(3),
		ProcessType:  ProcessTypeActivityMonitoring,
		AlarmType:    "",
		Description:  "Number of people change, for activity monitoring",
	},

	// Event type=7 (信号差)
	{
		TopicType:    "event",
		Category:     "signal",
		EventType:    intPtr(7),
		AlarmType:    SignalPoor,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Signal poor event",
	},

	// Event type=8 (倾角异常)
	{
		TopicType:    "event",
		Category:     "angle",
		EventType:    intPtr(8),
		AlarmType:    AngleException,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Angle exception event",
	},

	// ========== Stat Stream 中的统计 ==========
	// 睡眠统计相关
	{
		TopicType:    "stat",
		Category:     "sleep",
		StatType:     strPtr("sleep"),
		StatCode:     strPtr("breath_01"),
		AlarmType:    AbnormalRespiratoryRate,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Low respiratory rate stat",
	},
	{
		TopicType:    "stat",
		Category:     "sleep",
		StatType:     strPtr("sleep"),
		StatCode:     strPtr("breath_10"),
		AlarmType:    AbnormalRespiratoryRate,
		ProcessType:  ProcessTypeImmediate,
		Description:  "High respiratory rate stat",
	},
	{
		TopicType:    "stat",
		Category:     "sleep",
		StatType:     strPtr("sleep"),
		StatCode:     strPtr("breath_11"),
		AlarmType:    ApneaHypopnea,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Apnea/hypopnea stat",
	},
	{
		TopicType:    "stat",
		Category:     "sleep",
		StatType:     strPtr("sleep"),
		StatCode:     strPtr("heart_01"),
		AlarmType:    AbnormalHeartRate,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Low heart rate stat",
	},
	{
		TopicType:    "stat",
		Category:     "sleep",
		StatType:     strPtr("sleep"),
		StatCode:     strPtr("heart_10"),
		AlarmType:    AbnormalHeartRate,
		ProcessType:  ProcessTypeImmediate,
		Description:  "High heart rate stat",
	},
	{
		TopicType:    "stat",
		Category:     "sleep",
		StatType:     strPtr("sleep"),
		StatCode:     strPtr("vitals_11"),
		AlarmType:    VitalsWeak,
		ProcessType:  ProcessTypeImmediate,
		Description:  "Weak vitals stat",
	},
	{
		TopicType:    "stat",
		Category:     "sleep",
		StatType:     strPtr("sleep"),
		StatCode:     strPtr("stay"),
		AlarmType:    Stay,
		ProcessType:  ProcessTypeTimeBased,
		DurationSec:  intPtr(45 * 60), // 45分钟
		Description:  "Stay stat, trigger after 45min",
	},
	{
		TopicType:    "stat",
		Category:     "sleep",
		StatType:     strPtr("sleep"),
		StatCode:     strPtr("no_activity"),
		AlarmType:    NoActivity24h,
		ProcessType:  ProcessTypeTimeBased,
		DurationSec:  intPtr(24 * 60 * 60), // 24小时
		Description:  "No activity stat, trigger after 24h",
	},

	// ========== Monitor Stream 中的监控数据 ==========
	// 呼吸率监控
	{
		TopicType:    "monitor",
		Category:     "vital",
		StatType:     strPtr("respiratory_rate"),
		AlarmType:    AbnormalRespiratoryRate,
		ProcessType:  ProcessTypeTimeBased,
		DurationSec:  intPtr(60), // 持续1分钟异常
		Description:  "Respiratory rate abnormal monitor",
	},
	// 心率监控
	{
		TopicType:    "monitor",
		Category:     "vital",
		StatType:     strPtr("heart_rate"),
		AlarmType:    AbnormalHeartRate,
		ProcessType:  ProcessTypeTimeBased,
		DurationSec:  intPtr(60), // 持续1分钟异常
		Description:  "Heart rate abnormal monitor",
	},
}

// MatchRadarEventStatToAlarm 匹配 Radar Event/Stat 到 Alarm_Type
// 参数：
//   - topicType: "event" | "stat" | "monitor" | "alarm"
//   - category: category 字段值
//   - eventType: event_type 字段值（可为 nil）
//   - areaType: area_type 字段值（可为 nil）
//   - pose: pose 字段值（可为 nil）
//   - statType: stat 类型（可为 nil）
//   - statCode: stat 代码（可为 nil）
// 返回：匹配的映射规则列表（可能有多个匹配，按优先级返回第一个）
func MatchRadarEventStatToAlarm(
	topicType, category string,
	eventType, areaType, pose *int,
	statType, statCode *string,
) []RadarEventStatToAlarmMapping {
	var matches []RadarEventStatToAlarmMapping

	for _, mapping := range RadarEventStatToAlarmMap {
		// 匹配 topic_type
		if mapping.TopicType != "" && mapping.TopicType != topicType {
			continue
		}
		// 匹配 category
		if mapping.Category != "" && mapping.Category != category {
			continue
		}
		// 匹配 event_type
		if mapping.EventType != nil && (eventType == nil || *mapping.EventType != *eventType) {
			continue
		}
		// 匹配 area_type
		if mapping.AreaType != nil && (areaType == nil || *mapping.AreaType != *areaType) {
			continue
		}
		// 匹配 pose
		if mapping.Pose != nil && (pose == nil || *mapping.Pose != *pose) {
			continue
		}
		// 匹配 stat_type
		if mapping.StatType != nil && (statType == nil || *mapping.StatType != *statType) {
			continue
		}
		// 匹配 stat_code
		if mapping.StatCode != nil && (statCode == nil || *mapping.StatCode != *statCode) {
			continue
		}

		matches = append(matches, mapping)
	}

	return matches
}

// GetAlarmTypeFromRadarEventStat 根据 Radar Event/Stat 获取对应的 Alarm_Type
// 返回第一个匹配的 alarm_type，如果未找到则返回空字符串
func GetAlarmTypeFromRadarEventStat(
	topicType, category string,
	eventType, areaType, pose *int,
	statType, statCode *string,
) string {
	matches := MatchRadarEventStatToAlarm(topicType, category, eventType, areaType, pose, statType, statCode)
	if len(matches) > 0 {
		return matches[0].AlarmType
	}
	return ""
}
