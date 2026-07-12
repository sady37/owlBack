package roomengine

import "os"

// dbn_mode.go — DBN cutover 门控（S0.c-4，迁自已删 belief_shadow.go 的 dbnMode，B3.2/A·R3.4 单源）。
//   0 = shadow：DBN 不自发，固件 floor 照转发，无否决腿（dbn_xray 对账）。
//   1 = DBN 自发，固件 floor 照转发，无否决腿。
//   2 = DBN 自发 + 否决腿：躺卧面固件即时 fall/坐地不转发（vetoFirmwareFallLying，读全局 dbnMode）；floor 90min 兜底不受影响。
// 自发门控 dbnSelfFireEnabled=dbnMode≥1；否决门控独立读全局 dbnMode==2。
// 新单元冷启安全由 tFloor 默认兜底承担（chair 90min / deny 12min / bath 20min），不单设 DBN 冷启 cap。
// 运维 .env 翻 DBN_MODE。门控放 engine 内 publish 处（A·R12.3），=0 仍走完整裁决可 log diff。

var dbnMode = parseDBNMode(os.Getenv("DBN_MODE"))

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

// dbnSelfFireEnabled DBN 自发 fire 是否启用（engine 内 publish 门控，A·R12.3）。dbnMode=0 静默。
func dbnSelfFireEnabled() bool {
	return dbnMode >= 1
}
