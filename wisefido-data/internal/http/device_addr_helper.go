package httpapi

import (
	"net/netip"
	"strings"
)

// parseDeviceAddrFromPath 从 URL 路径段解析 device_addr，统一返回 canonical /128 CIDR
// （与 tenant /48、branch /56、unit /80 等其他 spatial endpoint 一致 — 所有 entity ID 走 CIDR
// 形态，层级信息内嵌前缀长度）。
//
// 接受两种 input form：
//   1) /128 CIDR: "fd00:0:3:411::1/128" — 直接 normalize 后返回
//   2) host text: "fd00:0:3:411::1"      — 补 /128 后返回
//
// 拒绝：空串 / "undefined" / "null" / 非 /128 prefix / 非 IPv6 / 多段 path 残留。
func parseDeviceAddrFromPath(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" || s == "undefined" || s == "null" {
		return "", false
	}
	if prefix, err := netip.ParsePrefix(s); err == nil {
		if prefix.IsValid() && prefix.Addr().Is6() && prefix.Bits() == 128 {
			return prefix.Masked().String(), true
		}
		return "", false
	}
	addr, err := netip.ParseAddr(s)
	if err != nil || !addr.IsValid() || !addr.Is6() {
		return "", false
	}
	return addr.String() + "/128", true
}
