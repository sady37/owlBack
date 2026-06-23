package roomengine

import (
	"os"
	"strconv"
	"time"
)

// dbn_mode.go — DBN cutover 开关（S0.c-4 重建，迁自已删 belief_shadow.go 的 dbnMode 门控，B3.2/A·R3.4 单源）。
// 两正交轴=否决固件 × DBN 自发，三档；每-unit 冷启成熟度 cap=min(全局 DBN_MODE, unitCap)：
//   0 静默（engine 内跑裁决但不 publish，shadow 对账）；1 DBN 自发 fire + 固件 floor；2 满冷启 cap 后可否决固件。
// 运维 .env 翻 DBN_MODE / DBN_COLD_HOURS。门控放 engine 内 publish 处（A·R12.3），使 =0 仍走完整裁决可 log diff。

var dbnMode = parseDBNMode(os.Getenv("DBN_MODE"))
var coldGraduateMs = parseColdHours(os.Getenv("DBN_COLD_HOURS"))

func parseDBNMode(s string) int {
	switch s {
	case "1":
		return 1
	case "2":
		return 2
	default:
		return 0
	}
}

func parseColdHours(s string) int64 {
	const def = int64(7) * 24 * 3600 * 1000 // 默认 7d（委员会 §6 冷启毕业）
	if s == "" {
		return def
	}
	h, err := strconv.Atoi(s)
	if err != nil || h <= 0 {
		return def
	}
	return int64(h) * int64(time.Hour/time.Millisecond)
}

// unitCap 每-unit 冷启成熟度 cap：新 unit 首见 track 起 < coldGraduateMs 返 1（自发但不否决固件），
// 毕业后返 2（可否决固件）。首次见即登记 firstTrack。
func (e *Engine) unitCap(suiteID string, nowMs int64) int {
	e.coldStartMu.Lock()
	first, ok := e.unitFirstTrackMs[suiteID]
	if !ok {
		e.unitFirstTrackMs[suiteID] = nowMs
		first = nowMs
	}
	e.coldStartMu.Unlock()
	if nowMs-first >= coldGraduateMs {
		return 2
	}
	return 1
}

// dbnSelfFireEnabledFor DBN 自发 fire 是否启用（engine 内 publish 门控，A·R12.3）。
// effectiveMode = min(全局 dbnMode, 每-unit 冷启 cap)；≥1 = 可自发 publish。dbnMode=0 恒 false（静默）。
func (e *Engine) dbnSelfFireEnabledFor(suiteID string, nowMs int64) bool {
	return min(dbnMode, e.unitCap(suiteID, nowMs)) >= 1
}
