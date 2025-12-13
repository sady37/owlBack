package models

import (
	"time"
)

// AlarmEvent 报警事件（对应 alarm_events 表）
type AlarmEvent struct {
	EventID          string    `json:"event_id" db:"event_id"`
	TenantID         string    `json:"tenant_id" db:"tenant_id"`
	DeviceID         string    `json:"device_id" db:"device_id"`
	EventType        string    `json:"event_type" db:"event_type"`
	Category         string    `json:"category" db:"category"` // safety, clinical, behavioral, device
	AlarmLevel       string    `json:"alarm_level" db:"alarm_level"` // ALERT, CRIT, WARNING, etc.
	AlarmStatus      string    `json:"alarm_status" db:"alarm_status"` // active, acknowledged
	TriggeredAt      time.Time `json:"triggered_at" db:"triggered_at"`
	HandTime         *time.Time `json:"hand_time,omitempty" db:"hand_time"`
	IoTTimeSeriesID  *int64     `json:"iot_timeseries_id,omitempty" db:"iot_timeseries_id"`
	TriggerData      string    `json:"trigger_data" db:"trigger_data"` // JSONB
	Handler          *string   `json:"handler,omitempty" db:"handler"`
	Operation        *string   `json:"operation,omitempty" db:"operation"`
	Notes            *string   `json:"notes,omitempty" db:"notes"`
	NotifiedUsers    string    `json:"notified_users" db:"notified_users"` // JSONB
	Metadata         string    `json:"metadata" db:"metadata"` // JSONB
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// TriggerData 触发数据快照（JSONB 结构）
type TriggerData struct {
	HeartRate         *int    `json:"heart_rate,omitempty"`
	RespiratoryRate   *int    `json:"respiratory_rate,omitempty"`
	Posture           *string `json:"posture,omitempty"`
	PostureDisplay    *string `json:"posture_display,omitempty"`
	EventType         string  `json:"event_type"`
	Confidence        *int    `json:"confidence,omitempty"`
	DurationSec       *int    `json:"duration_sec,omitempty"`
	Threshold         *ThresholdData `json:"threshold,omitempty"`
	SNOMEDCode        *string `json:"snomed_code,omitempty"`
	SNOMEDDisplay     *string `json:"snomed_display,omitempty"`
	Source            string  `json:"source"` // "Sleepace" 或 "Radar"
}

// ThresholdData 阈值数据
type ThresholdData struct {
	Min      *int `json:"min,omitempty"`
	Max      *int `json:"max,omitempty"`
	DurationSec *int `json:"duration_sec,omitempty"`
}

