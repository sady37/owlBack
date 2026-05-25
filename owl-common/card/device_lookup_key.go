package card

import "strings"

// NormalizeDeviceLookupKey 规范化外部 device 标识（MAC / device_code）：
// 去空格 / 冒号 / 连字符 / 点并大写。
//
// （doc/device_ipv6_migration_checklist.md）：
//   - UUID device_addr 已不参与 lookup 路径（R-001 删除 UUID 入口）
//   - 仅外部 device gateway auth handshake (qinglan/sleepace 收 MQTT) 用 MAC/device_code 作 lookup key (R-002 边界)
//   - 内部消费链一律用 netip.Addr 走 LookupCardByDeviceAddr LPM
func NormalizeDeviceLookupKey(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToUpper(s) {
		if r == ':' || r == '-' || r == ' ' || r == '.' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
