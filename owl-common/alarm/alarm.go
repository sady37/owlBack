package alarm

import (
	"fmt"
	"strconv"
	"strings"
)

// ========== Constants (常量) ==========

// Common Alarm default values (通用报警默认值)
// 这些是所有设备类型都支持的通用报警项，存储在 alarm_cloud 表的 OfflineAlarm, LowBattery, DeviceFailure 字段
const (
	DefaultOfflineAlarm  = "ERROR"   // 设备离线报警默认级别
	DefaultLowBattery    = "WARNING" // 低电量报警默认级别
	DefaultDeviceFailure = "ERROR"   // 设备故障报警默认级别
)

// Cloud Vital Alarm Threshold default values (生理指标阈值默认值)
const (
	DefaultRespiratoryRateMin = 8  // 呼吸率nomal最小值默认值
	DefaultRespiratoryRateMax = 24 // 呼吸率nomal最大值默认值
	DefaultHeartRateMin       = 50 // 心率正nomal小值默认值
	DefaultHeartRateMax       = 95 // 心率nomal最大值默认值
)

// ExampleDefaultAlarmSettingJSON 完整的 DefaultAlarmSetting 示例 JSON，便于检查 json.Marshal(DefaultAlarmSetting) 是否正常
// is_enabled: 0=关闭 1=开启 | alarm_level: nil=无报警级别 | alarm_params: nil=无参数 | display_setting: 显示设置
const ExampleDefaultAlarmSettingJSON = `{
	"sleepad": [
	  {
		"alarm_type": "SleepPad_ResetTime",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "InBedTime": "21:30",
		  "OutBedTime": "07:30"
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SleepPad_NapTime",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "InBedTime": "13:00",
		  "OutBedTime": "14:00"
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SleepPad_ApneaHypopnea",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "duration_sec": 60
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SleepPad_AbnormalHeartRate",
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
		"alarm_type": "SleepPad_AbnormalRespiratoryRate",
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
		"alarm_type": "SleepPad_LeftBed",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		   "duration_sec": 8
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SleepPad_LeftBedTooLong",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "leave_minutes": 45
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SleepPad_InBed",
		"is_enabled": 0,
		"alarm_level": null,
		"alarm_params": {
		  "duration_sec": 300
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SleepPad_BedSitUp",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SleepPad_AbnormalBodyMovement",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_min": 10
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SleepPad_NoBodyMove",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_min": 60
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "SleepPad_NoTurnOver",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_min": 120
		},
		"display_setting": 3
	  },
	{
		"alarm_type": "SleepPad_SensorDetached",
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
		"alarm_type": "Radar_ResetTime",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "InBedTime": "21:30",
		  "OutBedTime": "07:30"
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_NapTime",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "InBedTime": "13:00",
		  "OutBedTime": "14:00"
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_MonitoringMode",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {
		  "mode": 15
		},
		"display_setting": 2
	  },
	  {
		"alarm_type": "Radar_PostureDetection",
		"is_enabled": 1,
		"alarm_level": null,
		"alarm_params": {},
		"display_setting": 2
	  },
	  {
		"alarm_type": "Radar_Fall",
		"is_enabled": 1,
		"alarm_level": "EMERG",
				"alarm_params": {
		  "duration_sec": 60
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_SuspectedFall",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_SittingOnGround",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_sec": 90
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_BedSitUp",
		"is_enabled": 0,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 0
	  },	  
	  {
		"alarm_type": "Radar_ApneaHypopnea",
		"is_enabled": 0,
		"alarm_level": null,
		"alarm_params": {
		  "apnea_60s_min_events": 4,
		  "apnea_120min_min_events": 7
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_AbnormalHeartRate",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "min": 50,
		  "max": 95
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_AbnormalRespiratoryRate",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "min": 8,
		  "max": 24
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_VitalsWeak",
		"is_enabled": 1,
		"alarm_level": "EMERG",
		"alarm_params": {
		  "duration_min": 10,
		  "sensitivity": 35
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_LeftBed",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_sec": 8
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_LeftBedTooLong",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "leave_minutes": 45
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_Stay",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {
		  "duration_min": 45
		},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_NoActivity24h",
		"is_enabled": 1,
		"alarm_level": "WARNING",
		"alarm_params": {},
		"display_setting": 3
	  },
	  {
		"alarm_type": "Radar_SignalPoor",
		"is_enabled": 0,
		"alarm_level": "ERR",
		"alarm_params": {},
		"display_setting": 1
	  },
	  {
		"alarm_type": "Radar_AngleException",
		"is_enabled": 1,
		"alarm_level": "ERR",
		"alarm_params": {},
		"display_setting": 1
	  }
	]
  }`

// ========== 报警类型 ==========

const (
	AlarmTypeOfflineAlarm  = "OfflineAlarm"
	AlarmTypeDeviceFailure = "DeviceFailure"
	AlarmTypeUnknown       = "Unknown"

	SleepPadApneaHypopnea           = "SleepPad_ApneaHypopnea"
	SleepPadAbnormalHeartRate       = "SleepPad_AbnormalHeartRate"
	SleepPadAbnormalRespiratoryRate = "SleepPad_AbnormalRespiratoryRate"
	SleepPadLeftBed                 = "SleepPad_LeftBed"
	SleepPadLeftBedTooLong          = "SleepPad_LeftBedTooLong"
	SleepPadOnBed                   = "SleepPad_OnBed"
	SleepPadBedSitUp                = "SleepPad_BedSitUp"
	SleepPadAbnormalBodyMovement    = "SleepPad_AbnormalBodyMovement"
	SleepPadNoBodyMove              = "SleepPad_NoBodyMove"
	SleepPadNoTurnOver              = "SleepPad_NoTurnOver"
	SleepPadResetTime               = "SleepPad_ResetTime"
	SleepPadNapTime                 = "SleepPad_NapTime"
	SleepPadSensorDetached          = "SleepPad_SensorDetached"

	RadarResetTime               = "Radar_ResetTime"
	RadarNapTime                 = "Radar_NapTime"
	RadarFall                    = "Radar_Fall"
	RadarSuspectedFall           = "Radar_SuspectedFall"
	RadarSittingOnGround         = "Radar_SittingOnGround"
	RadarApneaHypopnea           = "Radar_ApneaHypopnea"
	RadarAbnormalHeartRate       = "Radar_AbnormalHeartRate"
	RadarAbnormalRespiratoryRate = "Radar_AbnormalRespiratoryRate"
	RadarVitalsWeak              = "Radar_VitalsWeak"
	RadarLeftBed                 = "Radar_LeftBed"
	RadarInBed                   = "Radar_InBed"
	RadarLeftBedTooLong          = "Radar_LeftBedTooLong"
	RadarBedSitUp                = "Radar_BedSitUp"
	RadarStay                    = "Radar_Stay"
	RadarNoActivity24h           = "Radar_NoActivity24h"
	RadarWarningArea             = "Radar_WarningArea"
	RadarSignalPoor              = "Radar_SignalPoor"
	RadarAngleException          = "Radar_AngleException"
	RadarMonitoringMode          = "Radar_MonitoringMode"
	RadarPostureDetection        = "Radar_PostureDetection"
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
	AlarmLevelEmerg  = "EMERG"
	AlarmLevelAlert  = "ALERT"
	AlarmLevelCrit   = "CRIT"
	AlarmLevelErr    = "ERR"
	AlarmLevelWarn   = "WARNING"
	AlarmLevelNotice = "NOTICE"
	AlarmLevelInfo   = "INFO"
	AlarmLevelDebug  = "DEBUG"
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

// ResetTime 作息时间（用于 REST API）
type ResetTime struct {
	InBedTime  string `json:"In-bed time"`  // 格式: "21:30"
	OutBedTime string `json:"Out-bed time"` // 格式: "07:30"
}

// NapTime 午睡时间（用于 REST API）
type NapTime struct {
	InBedTime  string `json:"In-bed time"`  // 格式: "21:30"
	OutBedTime string `json:"Out-bed time"` // 格式: "07:30"
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
			AlarmType:  SleepPadResetTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "21:30",
				"OutBedTime": "07:30",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepPadNapTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "13:00",
				"OutBedTime": "14:00",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepPadSensorDetached,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelErr),
			AlarmParams: map[string]interface{}{
				"duration_min": 120,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepPadApneaHypopnea,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 30,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepPadAbnormalHeartRate,
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
			AlarmType:  SleepPadAbnormalRespiratoryRate,
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
			AlarmType:  SleepPadLeftBed,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 8,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepPadLeftBedTooLong,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				"leave_minutes": 45,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepPadOnBed,
			IsEnabled:  intPtr(IsEnabledOff),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 300,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      SleepPadBedSitUp,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepPadAbnormalBodyMovement,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 10,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepPadNoBodyMove,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 45,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  SleepPadNoTurnOver,
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
			AlarmType:  RadarResetTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "21:30",
				"OutBedTime": "07:30",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarNapTime,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"InBedTime":  "13:00",
				"OutBedTime": "14:00",
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarMonitoringMode,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"mode": 15,
			},
			DisplaySetting: DisplayAlarmDevice,
		},
		{
			AlarmType:      RadarPostureDetection,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     nil,
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmDevice,
		},
		{
			AlarmType:  RadarFall,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 60,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      RadarSuspectedFall,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarSittingOnGround,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 90,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarBedSitUp,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{},
			DisplaySetting: 0,
		},
		{
			AlarmType:  RadarApneaHypopnea,
			IsEnabled:  intPtr(IsEnabledOff),
			AlarmLevel: nil,
			AlarmParams: map[string]interface{}{
				"apnea_60s_min_events":    4,
				"apnea_120min_min_events": 7,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarAbnormalHeartRate,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamMin: 50,
				ParamMax: 95,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarAbnormalRespiratoryRate,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamMin: 8,
				ParamMax: 24,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarVitalsWeak,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelEmerg),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 10,
				ParamSensitivity: 35,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarLeftBed,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationSec: 8,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarLeftBedTooLong,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				"leave_minutes": 45,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:  RadarStay,
			IsEnabled:  intPtr(IsEnabledOn),
			AlarmLevel: strPtr(AlarmLevelWarn),
			AlarmParams: map[string]interface{}{
				ParamDurationMin: 45,
			},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      RadarNoActivity24h,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelWarn),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      RadarSignalPoor,
			IsEnabled:      intPtr(IsEnabledOff),
			AlarmLevel:     strPtr(AlarmLevelErr),
			AlarmParams:    map[string]interface{}{},
			DisplaySetting: DisplayAlarmCloudAndDevice,
		},
		{
			AlarmType:      RadarAngleException,
			IsEnabled:      intPtr(IsEnabledOn),
			AlarmLevel:     strPtr(AlarmLevelErr),
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
func GetRadarAlarmTypes() []string {
	return []string{AlarmTypeOfflineAlarm, AlarmTypeDeviceFailure, RadarFall, RadarSuspectedFall, RadarSittingOnGround, RadarApneaHypopnea, RadarAbnormalHeartRate, RadarAbnormalRespiratoryRate, RadarVitalsWeak, RadarLeftBed, RadarInBed, RadarLeftBedTooLong, RadarBedSitUp, RadarStay, RadarNoActivity24h, RadarWarningArea, RadarSignalPoor, RadarAngleException, RadarMonitoringMode, RadarNapTime, RadarResetTime, RadarPostureDetection}
}

// GetSleepPadAlarmTypes 获取 SleepPad 设备的报警类型列表
func GetSleepPadAlarmTypes() []string {
	return []string{AlarmTypeOfflineAlarm, AlarmTypeDeviceFailure, SleepPadSensorDetached, SleepPadApneaHypopnea, SleepPadAbnormalHeartRate, SleepPadAbnormalRespiratoryRate, SleepPadLeftBed, SleepPadLeftBedTooLong, SleepPadOnBed, SleepPadBedSitUp, SleepPadAbnormalBodyMovement, SleepPadNoBodyMove, SleepPadNoTurnOver}
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

		if item.AlarmType == RadarLeftBedTooLong {
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
// 参数：alarmType - 报警类型（如 RadarStay, RadarNoActivity24h, RadarWarningArea, Radar_InBed, Radar_LeftBedTooLong）
// 返回：是否需要启用进出事件监听
// 简化逻辑：如果启用了以下任一报警类型，所有 event type=1（进出事件）都转为 alarm
// - RadarStay（滞留）
// - RadarNoActivity24h（长时间无人活动）
// - RadarWarningArea（警告区域）
// - Radar_InBed/Radar_LeftBedTooLong（进床/离床）
// 注意：此函数已被注释，因为使用硬编码的逻辑，应该基于 alarm_cloud 配置动态判断
/*
func ShouldEnableEnter2OutEventForAlarm(alarmType string) bool {
	switch alarmType {
	case RadarStay, RadarNoActivity24h, RadarWarningArea, RadarInBed, RadarLeftBedTooLong:
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
	case RadarFall, RadarSittingOnGround, RadarBedSitUp:
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

		if item.AlarmType == SleepPadLeftBedTooLong {
			// 保存 LeftBedTooLong 的配置作为后备
			leftBedTooLongEnabled = item.IsEnabled
			continue
		}

		// 处理其他项（包括 LeftBed）
		enablement[item.AlarmType] = *item.IsEnabled
	}

	// 如果 LeftBed 不存在但 LeftBedTooLong 存在，使用 LeftBedTooLong 的配置
	if leftBedTooLongEnabled != nil {
		if _, exists := enablement[SleepPadLeftBed]; !exists {
			enablement[SleepPadLeftBed] = *leftBedTooLongEnabled
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
var MQTTToAlarmTypeMapSleepace = map[string]string{
	"alarmLeftBed":         SleepPadLeftBed, // 或 SleepPadLeftBedTooLong（根据上下文判断）
	"alarmHeartRateFast":   SleepPadAbnormalHeartRate,
	"alarmHeartRateSlow":   SleepPadAbnormalHeartRate,
	"alarmBreathRateFast":  SleepPadAbnormalRespiratoryRate,
	"alarmBreathRateSlow":  SleepPadAbnormalRespiratoryRate,
	"alarmBreathRatePause": SleepPadApneaHypopnea,
	"alarmBodymove":        SleepPadAbnormalBodyMovement,
	"alarmNoBodymove":      SleepPadNoBodyMove,
	"alarmNoTurnOver":      SleepPadNoTurnOver,
	"alarmSitup":           SleepPadBedSitUp,
	"alarmOnBed":           SleepPadOnBed,
	"alarmSensorFall":      SleepPadSensorDetached,
	"offLine":              "", // 离线报警由通用报警处理，不映射到具体 alarm_type
}

// MQTTToAlarmTypeMapRadar Radar MQTT 消息到 alarm_type 的映射表
// 基于文档 radar-Qinlan-code-v3.0.md (612-636)
// 注意：需要根据 event_type, area_type, pose 等字段进行判断
var MQTTToAlarmTypeMapRadar = map[string]string{
	// Event type=1 (进出事件)
	"event_type_1_room":              RadarStay,        // 进出房间
	"event_type_1_area_2_or_5":       RadarInBed,       // 进出区域+Area_type={2||5}
	"event_type_1_area_6":            RadarWarningArea, // 进入区域+Area_type=6
	"event_type_1_left_bed_too_long": RadarLeftBedTooLong,
	// Event type=2 (姿态变化)
	"event_type_2_pose_5":  RadarFall,            // 5-确认跌倒
	"event_type_2_pose_2":  RadarSuspectedFall,   // 2-疑似跌倒
	"event_type_2_pose_7":  RadarSittingOnGround, // 7-疑似坐地
	"event_type_2_pose_8":  RadarSittingOnGround, // 8-确认坐地
	"event_type_2_pose_10": RadarBedSitUp,        // 10-疑似床上坐起
	"event_type_2_pose_11": RadarBedSitUp,        // 11-确认床上坐起
	// Event type=7
	"event_type_7_signal_poor": RadarSignalPoor, // 信号差事件
	// Event type=8
	"event_type_8_angle_abnormal": RadarAngleException, // 倾角异常事件
	// Statistics (sleep)
	"stat_sleep_breath_01":   RadarAbnormalRespiratoryRate, // 01: 呼吸过低
	"stat_sleep_breath_10":   RadarAbnormalRespiratoryRate, // 10: 呼吸过高
	"stat_sleep_heart_01":    RadarAbnormalHeartRate,       // 01: 心率过低
	"stat_sleep_heart_10":    RadarAbnormalHeartRate,       // 10: 心率过高
	"stat_sleep_breath_11":   RadarApneaHypopnea,           // 11: 呼吸暂停
	"stat_sleep_vitals_11":   RadarVitalsWeak,              // 11: 生命体征弱
	"stat_sleep_stay":        RadarStay,                    // 滞留
	"stat_sleep_no_activity": RadarNoActivity24h,           // 长时间无人活动
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
			return RadarWarningArea
		}
		if areaType == "2" || areaType == "5" {
			// 注意：RadarInBed 和 RadarLeftBedTooLong 需要根据具体业务逻辑判断
			// 这里先返回 RadarInBed，实际使用时可能需要根据其他字段（如 duration）判断
			return RadarInBed
		}
		// 进出房间
		return RadarStay
	}

	// Event type=2 (姿态变化)
	if eventType == "2" {
		switch pose {
		case "5":
			return RadarFall
		case "2":
			return RadarSuspectedFall
		case "7", "8":
			return RadarSittingOnGround
		case "10", "11":
			return RadarBedSitUp
		}
	}

	// Event type=7 (信号差)
	if eventType == "7" {
		return RadarSignalPoor
	}

	// Event type=8 (倾角异常)
	if eventType == "8" {
		return RadarAngleException
	}

	// Statistics (sleep)
	if statType == "sleep" {
		switch statAlarmType {
		case "01": // 呼吸过低或心率过低
			// 需要根据具体字段判断是呼吸还是心率
			// 这里假设有额外的字段来区分，实际使用时需要根据 MQTT 消息结构调整
			return RadarAbnormalRespiratoryRate // 或 RadarAbnormalHeartRate
		case "10": // 呼吸过高或心率过高
			return RadarAbnormalRespiratoryRate // 或 RadarAbnormalHeartRate
		case "11": // 呼吸暂停或生命体征弱
			return RadarApneaHypopnea // 或 RadarVitalsWeak
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
var AlarmTypeToNumericCodeMap = map[string][]AlarmNumericCode{
	// Event 类型报警
	RadarLeftBed: {
		{EventType: 1, Event: 4, AreaType: 2, Source: "event"}, // 普通床离床
		{EventType: 1, Event: 4, AreaType: 5, Source: "event"}, // 监护床离床
	},
	RadarLeftBedTooLong: {
		{EventType: 1, Event: 4, AreaType: 2, Source: "event"}, // 普通床离床过长
		{EventType: 1, Event: 4, AreaType: 5, Source: "event"}, // 监护床离床过长
	},
	RadarInBed: {
		{EventType: 1, Event: 4, AreaType: 2, Source: "event"}, // 普通床在床
		{EventType: 1, Event: 4, AreaType: 5, Source: "event"}, // 监护床在床
	},
	RadarWarningArea: {
		{EventType: 1, Event: 1, AreaType: 6, Source: "event"}, // 进入警告区域
	},
	RadarFall: {
		{EventType: 2, Pose: 5, Source: "event"},                  // 确认跌倒（event）
		{StatType: "sleep", StatCode: "11", Source: "statistics"}, // 生命体征弱（statistics，可能触发跌倒）
	},
	RadarSuspectedFall: {
		{EventType: 2, Pose: 2, Source: "event"}, // 疑似跌倒
	},
	RadarSittingOnGround: {
		{EventType: 2, Pose: 7, Source: "event"}, // 疑似坐地
		{EventType: 2, Pose: 8, Source: "event"}, // 确认坐地
	},
	RadarBedSitUp: {
		{EventType: 2, Pose: 10, Source: "event"}, // 疑似床上坐起
		{EventType: 2, Pose: 11, Source: "event"}, // 确认床上坐起
	},
	RadarSignalPoor: {
		{EventType: 7, Source: "event"}, // 信号差
	},
	RadarAngleException: {
		{EventType: 8, Source: "event"}, // 倾角异常
	},
	// Statistics 类型报警（4组状态分别映射）
	// bit 1 & bit 0: 呼吸状态 (00=正常, 01=过低, 10=过高, 11=暂停)
	RadarAbnormalRespiratoryRate: {
		{StatType: "sleep", StatCode: "breath_01", Source: "statistics"}, // 呼吸过低
		{StatType: "sleep", StatCode: "breath_10", Source: "statistics"}, // 呼吸过高
	},
	RadarApneaHypopnea: {
		{StatType: "sleep", StatCode: "breath_11", Source: "statistics"}, // 呼吸暂停
	},
	// bit 3 & bit 2: 心率状态 (00=正常, 01=过低, 10=过高, 11=未定义)
	RadarAbnormalHeartRate: {
		{StatType: "sleep", StatCode: "heart_01", Source: "statistics"}, // 心率过低
		{StatType: "sleep", StatCode: "heart_10", Source: "statistics"}, // 心率过高
	},
	// bit 5 & bit 4: 生命体征情况 (00=正常, 01=未定义, 02=未定义, 11=生命体征弱)
	RadarVitalsWeak: {
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
	"142": {RadarLeftBed, RadarLeftBedTooLong, RadarInBed}, // eventType=1, event=4, area_type=2 (普通床：离床/离床过长/在床)
	"145": {RadarLeftBed, RadarLeftBedTooLong, RadarInBed}, // eventType=1, event=4, area_type=5 (监护床：离床/离床过长/在床)
	"116": {RadarWarningArea},                              // eventType=1, event=1, area_type=6 (进入警告区域)
	// Event type=2 (姿态变化)
	"25":  {RadarFall},            // eventType=2, pose=5 (确认跌倒)
	"22":  {RadarSuspectedFall},   // eventType=2, pose=2 (疑似跌倒)
	"27":  {RadarSittingOnGround}, // eventType=2, pose=7 (疑似坐地)
	"28":  {RadarSittingOnGround}, // eventType=2, pose=8 (确认坐地)
	"210": {RadarBedSitUp},        // eventType=2, pose=10 (疑似床上坐起)
	"211": {RadarBedSitUp},        // eventType=2, pose=11 (确认床上坐起)
	// Event type=7, 8
	"7": {RadarSignalPoor},     // eventType=7 (信号差)
	"8": {RadarAngleException}, // eventType=8 (倾角异常)
	// Statistics (sleep) - 4组状态分别映射
	// bit 1 & bit 0: 呼吸状态
	"sleep_breath_01": {RadarAbnormalRespiratoryRate}, // 呼吸过低
	"sleep_breath_10": {RadarAbnormalRespiratoryRate}, // 呼吸过高
	"sleep_breath_11": {RadarApneaHypopnea},           // 呼吸暂停
	// bit 3 & bit 2: 心率状态
	"sleep_heart_01": {RadarAbnormalHeartRate}, // 心率过低
	"sleep_heart_10": {RadarAbnormalHeartRate}, // 心率过高
	// bit 5 & bit 4: 生命体征情况
	"sleep_vitals_11": {RadarVitalsWeak}, // 生命体征弱
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
			// 如果没有 event_raw，从 display_en 反向查找原始值
			// event 值：1=Enter room, 2=Leave room, 3=Enter area, 4=Leave area
			switch eventDisplay {
			case "Enter room":
				eventValue = "1"
			case "Leave room":
				eventValue = "2"
			case "Enter area":
				eventValue = "3"
			case "Leave area":
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
		// 获取 pose 的原始值
		pose, _ := eventData["pose_raw"].(string)
		if pose == "" {
			// 如果没有 pose_raw，尝试从 pose 获取（可能是 int）
			if poseInt, ok := eventData["pose"].(int); ok {
				pose = fmt.Sprintf("%d", poseInt)
			} else if poseStr, ok := eventData["pose"].(string); ok {
				// 可能是 SNOMED 映射后的值，需要反向查找
				// 这里先尝试直接使用，如果不行再反向查找
				pose = poseStr
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
