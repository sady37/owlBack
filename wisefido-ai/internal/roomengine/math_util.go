package roomengine

import (
	"math"
	"time"
)

// 夜间时段默认 23:30 - 07:30 本地时间（可由 RiskTimeConfig 覆盖）。
// 设计动机：
//   - 夜间老人活动稀少，长时间静止 + 跌倒占总跌倒比例 50%+
//   - 白天家属/护工活动多，老人坐着聊天/看电视常有 8-15min 不动
//   - 监管要求所有 anomaly 必须人确认，所以无法用 cooldown 抑制——只能从源头降误报
// 因此：当前阈值（StillTimeoutSec）按夜间标准；白天放宽 20%。
const (
	defaultNightStartH = 23
	defaultNightStartM = 30
	defaultNightEndH   = 7
	defaultNightEndM   = 30
)

// RiskTimeConfig 风险时段配置（夜间），由 engine.Configure 注入。
// 0 值视为未设置，IsNightTime 用 defaultXxx。
type RiskTimeConfig struct {
	NightStartH int
	NightStartM int
	NightEndH   int
	NightEndM   int
}

// 默认配置（包级 var，由 engine.Configure 覆盖）
var nightCfg = RiskTimeConfig{
	NightStartH: defaultNightStartH, NightStartM: defaultNightStartM,
	NightEndH: defaultNightEndH, NightEndM: defaultNightEndM,
}

// SetRiskTimeConfig 由 engine.Configure 调用，注入 yaml 配置的夜间时段
func SetRiskTimeConfig(c RiskTimeConfig) {
	if c.NightStartH == 0 && c.NightStartM == 0 && c.NightEndH == 0 && c.NightEndM == 0 {
		return // 全 0 视为未设置
	}
	nightCfg = c
}

// IsNightTime 用毫秒时间戳判断本地时间是否在夜间时段。
// 默认 23:30-07:30，可由 SetRiskTimeConfig 覆盖。
// nowMs 0 时按系统当前时间。设备 timestamp 可直接传入。
func IsNightTime(nowMs int64) bool {
	var t time.Time
	if nowMs > 0 {
		t = time.UnixMilli(nowMs).Local()
	} else {
		t = time.Now().Local()
	}
	h, m := t.Hour(), t.Minute()
	if h > nightCfg.NightStartH || (h == nightCfg.NightStartH && m >= nightCfg.NightStartM) {
		return true
	}
	if h < nightCfg.NightEndH || (h == nightCfg.NightEndH && m < nightCfg.NightEndM) {
		return true
	}
	return false
}

// dist 二维欧几里得距离。
func dist(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}

// clamp 将 v 限制在 [lo, hi] 范围内。
func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// clampInt 将 v 限制在 [lo, hi] 范围内（int 版本）。
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// distInt 画布坐标（cm 整数）下两点欧氏距离，四舍五入取整。
func distInt(x1, y1, x2, y2 int) int {
	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	return int(math.Round(math.Sqrt(dx*dx + dy*dy)))
}
