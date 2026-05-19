package zonealarm

import (
	"fmt"
	"os"
	"strings"

	"owl-common/alarm"
	"wisefido-sensor-v1/internal/zoneengine"

	"gopkg.in/yaml.v3"
)

// Config zone_alarm.yaml 顶层结构。
type Config struct {
	Rules []Rule `yaml:"rules"`
}

// LoadFromFile 从 yaml 加载 + 字符串字段（zone/status）→ enum 解析。
//
// 不存在文件 / 加载失败时返回 (DefaultRules(), nil) — 让 wiring 层日志后兜底。
func LoadFromFile(path string) ([]Rule, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}
	for i := range cfg.Rules {
		if err := resolveRule(&cfg.Rules[i]); err != nil {
			return nil, fmt.Errorf("rule[%d] %s: %w", i, cfg.Rules[i].AlarmType, err)
		}
	}
	return cfg.Rules, nil
}

// resolveRule 把 yaml 字符串字段（ArmZoneStr / ArmStatusStr / CancelTrigger.ZoneStr / StatusStr）
// 转换为 enum 类型，便于 supervisor 主路径用 == 判断。
func resolveRule(r *Rule) error {
	zt, err := parseZoneType(r.ArmZoneStr)
	if err != nil {
		return fmt.Errorf("arm_zone: %w", err)
	}
	r.ArmZone = zt

	st, err := parseZoneStatus(r.ArmStatusStr)
	if err != nil {
		return fmt.Errorf("arm_status: %w", err)
	}
	r.ArmStatus = st

	for j := range r.Cancels {
		c := &r.Cancels[j]
		ct, err := parseZoneType(c.ZoneStr)
		if err != nil {
			return fmt.Errorf("cancels[%d].zone: %w", j, err)
		}
		c.Zone = ct
		if strings.TrimSpace(c.StatusStr) == "" {
			c.StatusAny = true
		} else {
			cs, err := parseZoneStatus(c.StatusStr)
			if err != nil {
				return fmt.Errorf("cancels[%d].status: %w", j, err)
			}
			c.Status = cs
		}
	}
	return nil
}

func parseZoneType(s string) (zoneengine.ZoneType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "bed":
		return zoneengine.ZoneTypeBed, nil
	case "room":
		return zoneengine.ZoneTypeRoom, nil
	case "bathroom":
		return zoneengine.ZoneTypeBathroom, nil
	default:
		return 0, fmt.Errorf("unknown zone type %q (expect bed/room/bathroom)", s)
	}
}

func parseZoneStatus(s string) (zoneengine.ZoneStatus, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "vacant":
		return zoneengine.StatusVacant, nil
	case "occupied":
		return zoneengine.StatusOccupied, nil
	case "leaving":
		return zoneengine.StatusLeaving, nil
	default:
		return 0, fmt.Errorf("unknown zone status %q (expect vacant/occupied/leaving)", s)
	}
}

// DefaultRules — 4 条 provisional alarm rules，硬约束来自用户拍板（[[zoneengine_adapters_done]] memory）。
//
// 数值（duration / time window）是 placeholder，最终由用户对照真实信号 trace + tenant
// 配置 calibrate；yaml hot reload 经 Supervisor.ReloadRules 即可。
//
// AlarmType 取自 owl-common/alarm 常量，与 cardagg alarm_handler 接收侧严格对齐。
func DefaultRules() []Rule {
	rules := []Rule{
		// 1. Stay — bathroom 持续占用 N min →  fire
		{
			AlarmType:    alarm.Stay,
			Level:        alarm.AlarmLevelWarn,
			ArmZoneStr:   "bathroom",
			ArmStatusStr: "occupied",
			DurationSec:  600, // 10 min
			Cancels: []CancelTrigger{
				{ZoneStr: "bathroom", StatusStr: "vacant"},
				{ZoneStr: "room", StatusStr: "occupied"}, // 进其它房间
				{ZoneStr: "bed", StatusStr: "occupied"},  // 去躺床
			},
		},
		// 2. LeftBed — bed 离床持续 N min →  fire（rest window 内）
		{
			AlarmType:    alarm.LeftBed,
			Level:        alarm.AlarmLevelWarn,
			ArmZoneStr:   "bed",
			ArmStatusStr: "vacant",
			DurationSec:  1800, // 30 min
			Cancels: []CancelTrigger{
				{ZoneStr: "bed", StatusStr: "occupied"},
				{ZoneStr: "room", StatusStr: "occupied"}, // 人回房算回床区域（用户拍板）
			},
			// LeftBed 默认 24h；rest window 由 tenant 级 yaml 覆盖
		},
		// 3. NightAbsence (Night Out-of-Room) — room→vacant 持续 N min（21-7 时段）
		{
			AlarmType:    alarm.NightAbsence,
			Level:        alarm.AlarmLevelWarn,
			ArmZoneStr:   "room",
			ArmStatusStr: "vacant",
			DurationSec:  1800, // 30 min
			Cancels: []CancelTrigger{
				{ZoneStr: "room", StatusStr: "occupied"}, // EnterRoom
				{ZoneStr: "bed", StatusStr: "occupied"},  // InBed
				{ZoneStr: "room", CountGtZero: true},     // NumberPeople>0
			},
			TimeWindow: &TimeWindow{StartH: 21, StartM: 0, EndH: 7, EndM: 0},
		},
		// 4. BedNightAbsence (Night Out-of-Bed) — bed→vacant 持续 N min（21-7 时段）
		{
			AlarmType:    alarm.BedNightAbsence,
			Level:        alarm.AlarmLevelWarn,
			ArmZoneStr:   "bed",
			ArmStatusStr: "vacant",
			DurationSec:  1800, // 30 min
			Cancels: []CancelTrigger{
				{ZoneStr: "bed", StatusStr: "occupied"}, // InBed
			},
			TimeWindow: &TimeWindow{StartH: 21, StartM: 0, EndH: 7, EndM: 0},
		},
	}
	for i := range rules {
		_ = resolveRule(&rules[i]) // 默认值都合法，错忽略
	}
	return rules
}
