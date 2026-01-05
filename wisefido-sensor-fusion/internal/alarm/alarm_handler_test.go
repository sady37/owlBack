package alarm

import (
	"testing"
)

func TestIsDeviceDirectAlarm(t *testing.T) {
	tests := []struct {
		name     string
		eventType string
		expected bool
	}{
		// 设备直接报警
		{"Fall", "Fall", true},
		{"SuspectedFall", "SuspectedFall", true},
		{"OfflineAlarm", "OfflineAlarm", true},
		{"LowBattery", "LowBattery", true},
		{"DeviceFailure", "DeviceFailure", true},
		{"Stay", "Stay", true},
		{"NoActivity24h", "NoActivity24h", true},
		{"AngleException", "AngleException", true},
		{"LeftBed", "LeftBed", true},
		{"SitUp", "SitUp", true},

		// 云端事件报警（不在本函数处理）
		{"HeartRateHigh", "HeartRateHigh", false},
		{"BreathRateLow", "BreathRateLow", false},
		{"Apnea", "Apnea", false},
		{"NoTurning2H", "NoTurning2H", false},
		{"NoBodyMovement2H", "NoBodyMovement2H", false},
		{"Wandering", "Wandering", false},
		{"ProlongedStay", "ProlongedStay", false},

		// 边界情况
		{"Empty", "", false},
		{"Unknown", "UnknownEvent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDeviceDirectAlarm(tt.eventType)
			if result != tt.expected {
				t.Errorf("IsDeviceDirectAlarm(%q) = %v, want %v", tt.eventType, result, tt.expected)
			}
		})
	}
}

func TestGetAlarmCategory(t *testing.T) {
	tests := []struct {
		name     string
		eventType string
		expected string
	}{
		// 安全报警
		{"Fall", "Fall", "safety"},
		{"SuspectedFall", "SuspectedFall", "safety"},
		{"Stay", "Stay", "safety"},
		{"NoActivity24h", "NoActivity24h", "safety"},

		// 设备报警
		{"OfflineAlarm", "OfflineAlarm", "device"},
		{"LowBattery", "LowBattery", "device"},
		{"DeviceFailure", "DeviceFailure", "device"},

		// 行为报警
		{"LeftBed", "LeftBed", "behavioral"},
		{"SitUp", "SitUp", "behavioral"},
		{"AngleException", "AngleException", "device"}, // 根据实际代码，AngleException 是 device 类型

		// 默认
		{"Unknown", "UnknownEvent", "safety"},
		{"Empty", "", "safety"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAlarmCategory(tt.eventType)
			if result != tt.expected {
				t.Errorf("GetAlarmCategory(%q) = %q, want %q", tt.eventType, result, tt.expected)
			}
		})
	}
}

func TestGetAlarmLevel(t *testing.T) {
	tests := []struct {
		name     string
		eventType string
		expected string
	}{
		// EMERGENCY 级别
		{"Fall", "Fall", "EMERGENCY"},
		{"NoActivity24h", "NoActivity24h", "EMERGENCY"},
		{"DeviceFailure", "DeviceFailure", "EMERGENCY"},

		// WARNING 级别
		{"SuspectedFall", "SuspectedFall", "WARNING"},
		{"Stay", "Stay", "WARNING"},
		{"OfflineAlarm", "OfflineAlarm", "WARNING"},
		{"LowBattery", "LowBattery", "WARNING"},
		{"AngleException", "AngleException", "WARNING"},
		{"LeftBed", "LeftBed", "WARNING"},
		{"SitUp", "SitUp", "WARNING"},

		// 默认
		{"Unknown", "UnknownEvent", "WARNING"},
		{"Empty", "", "WARNING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAlarmLevel(tt.eventType)
			if result != tt.expected {
				t.Errorf("GetAlarmLevel(%q) = %q, want %q", tt.eventType, result, tt.expected)
			}
		})
	}
}
