// Package testkit 提供乐高块加载工具——从 window.json 读取生产 StreamMessage 记录。
// 测试工具包，生产码不依赖。
package testkit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"owl-common/alarm"
)

// LegoV2Record 是 v2 window.json 单条记录的 Go 表示。
// 字段名与生产 StreamMessage 对齐（约束#1.1: 格式=生产 StreamMessage 原字段名）。
type LegoV2Record struct {
	Category  string                   `json:"category"`
	DeviceUID string                   `json:"device_uid"`
	Timestamp int64                    `json:"timestamp"`
	DataValue []map[string]interface{} `json:"data_value"`
}

// LoadWindow 读单个 window.json 文件，按 timestamp 升序返回记录。
// dir 是 case fixture 目录的完整路径（如 "<casesDir>/unit201-handoff-0609-bathroom-333B"）。
func LoadWindow(dir string) ([]LegoV2Record, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "window.json"))
	if err != nil {
		return nil, err
	}
	var recs []LegoV2Record
	if err := json.Unmarshal(raw, &recs); err != nil {
		return nil, err
	}
	sort.SliceStable(recs, func(i, j int) bool { return recs[i].Timestamp < recs[j].Timestamp })
	return recs, nil
}

// EventCategory 判断 category 是否属于事件流（走 topic_type="event"），
// 其余 track/heart 等走 topic_type="monitor"。
func EventCategory(cat string) bool {
	switch cat {
	case alarm.Fall, alarm.EnterRoom, alarm.ExitRoom, alarm.InBed, alarm.LeftBed, alarm.NumberPeople,
		alarm.Activity, alarm.Walking, alarm.Sitting, alarm.Standing:
		return true
	}
	return false
}
