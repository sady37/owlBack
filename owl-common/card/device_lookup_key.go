package card

import "strings"

// IsUUIDString 与 cardagg service.IsUUID 一致：36 字符带四段连字符。
func IsUUIDString(s string) bool {
	if len(s) != 36 {
		return false
	}
	return s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-'
}

// NormalizeDeviceLookupKey 解析 device_uid / MAC 时用：去空格、冒号、连字符、点并大写。
// UUID（36 字符标准形）原样返回，避免破坏 device_id 查库。
func NormalizeDeviceLookupKey(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if IsUUIDString(s) {
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
