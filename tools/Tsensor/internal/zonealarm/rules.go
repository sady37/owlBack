package zonealarm

import (
	"fmt"
	"os"
	"strings"

	"owl-common/alarm"
	"owl-common/card"

	"gopkg.in/yaml.v3"
)

// Config zone_alarm.yaml 顶层结构。
type Config struct {
	Rules []Rule `yaml:"rules"`
}

// LoadFromFile 从 yaml 加载 + 字符串字段（anchor/state/entity）→ enum 解析。
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

// resolveRule 把 yaml 字符串字段（AnchorStr / KeepCounting.EntityStr/StateStr）转换为 enum。
func resolveRule(r *Rule) error {
	if r.AlarmType == "" {
		return fmt.Errorf("alarm_type required")
	}
	if r.AnchorField == "" {
		return fmt.Errorf("anchor_field required")
	}
	if r.ThresholdDaySec <= 0 {
		return fmt.Errorf("threshold_day_sec must be > 0")
	}
	ek, err := parseEntityKind(r.AnchorStr)
	if err != nil {
		return fmt.Errorf("anchor: %w", err)
	}
	r.Anchor = ek

	for j := range r.KeepCounting {
		c := &r.KeepCounting[j]
		if strings.TrimSpace(c.EntityStr) == "" {
			c.UsesSelf = true
			c.Entity = r.Anchor
		} else {
			e, err := parseEntityKind(c.EntityStr)
			if err != nil {
				return fmt.Errorf("keep_counting[%d].entity: %w", j, err)
			}
			c.Entity = e
		}
		st, err := parseEntityState(c.StateStr)
		if err != nil {
			return fmt.Errorf("keep_counting[%d].state: %w", j, err)
		}
		c.State = st
	}
	return nil
}

func parseEntityKind(s string) (EntityKind, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "bed":
		return EntityBed, nil
	case "room":
		return EntityRoom, nil
	case "bathroom":
		return EntityBathroom, nil
	default:
		return 0, fmt.Errorf("unknown entity %q (expect bed/room/bathroom)", s)
	}
}

func parseEntityState(s string) (EntityState, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "vacant":
		return StateVacant, nil
	case "occupied":
		return StateOccupied, nil
	case "alone":
		return StateAlone, nil
	default:
		return 0, fmt.Errorf("unknown state %q (expect vacant/occupied/alone)", s)
	}
}

// DefaultRules — 3 条 state-anchored 规则（[[zoneengine_adapters_done]]+ Stay alone-only 修正）。
//
// 阈值取自 owl-common/card/risk_thresholds.go BathroomAlone（Stay 与 cardagg display RiskLevel
// 共用同一份阈值，避免 sensor/cardagg 漂移）；LeftBed/NightAbsence 沿用原 30min。
//
// yaml hot reload 经 Supervisor.ReloadRules 即可（运行时 fired map 不动）。
func DefaultRules() []Rule {
	rules := []Rule{
		// 1. Stay — bathroom 独居超阈 → fire（仅 1 人才算滞留）
		{
			AlarmType:         alarm.Stay,
			AnchorStr:         "bathroom",
			AnchorField:       AnchorAloneContinuousMin,
			KeepCounting:      []Condition{{StateStr: "alone"}}, // Self=bathroom Alone
			ThresholdDaySec:   card.BathroomAlone.DayRiskMin * 60,       // 45*60 = 2700
			ThresholdNightSec: card.BathroomAlone.RiskTimeRiskMin * 60,  // 30*60 = 1800
		},
		// 2. LeftBed — bed 离床 30min + 同 /80 unit 内 room 也得空（人回房算回床区域）
		{
			AlarmType:       alarm.LeftBed,
			AnchorStr:       "bed",
			AnchorField:     AnchorBedStatusTs,
			KeepCounting: []Condition{
				{StateStr: "vacant"},                       // Self=bed Vacant
				{EntityStr: "room", StateStr: "vacant"},    // 同 unit room 也空
			},
			ThresholdDaySec: 30 * 60,
		},
		// 3. NightAbsence — room 空 30min + 21-7 时段 + bed 也得空
		{
			AlarmType:       alarm.NightAbsence,
			AnchorStr:       "room",
			AnchorField:     AnchorLastExitTs,
			KeepCounting: []Condition{
				{StateStr: "vacant"},                       // Self=room Vacant
				{EntityStr: "bed", StateStr: "vacant"},     // 同 unit bed 也空
			},
			ThresholdDaySec: 30 * 60,
			NightOnly:       &TimeWindow{StartH: 21, StartM: 0, EndH: 7, EndM: 0},
		},
	}
	for i := range rules {
		_ = resolveRule(&rules[i]) // 默认值都合法，错忽略
	}
	return rules
}
