package httpapi

// CIDR-aware URL helpers.
//
// 背景：v2 owl_v2 cutover 后，admin/v1 endpoint 的 :id 参数从 UUID 改为 IPv6 prefix CIDR
// （如 'fd00:0:3:100::/56'）。CIDR 字符串里有 '/' 与 URL path segment 分隔符冲突。
// 前端为零改动直接拼到 URL，导致 path 变成 'tenants/fd00:0:3:100::/56' 被 split 成
// ['tenants', 'fd00:0:3:100::', '56']。这两个 helper 在 handler 解析 URL 时把 CIDR 重新拼回。

import (
	"strconv"
	"strings"
)

// stitchCIDRSegments 把 strings.Split(path, "/") 的结果中"IPv6 + 数字尾巴"重新拼回 CIDR。
//
//	["fd00:0:3:100::", "56"] → ["fd00:0:3:100::/56"]
//	["uuid-style-id", "subaction"] → ["uuid-style-id", "subaction"]   (不含 ':' 时不动)
//	["fd00:0:6::", "48", "bootstrap-accounts", "reset"] → ["fd00:0:6::/48", "bootstrap-accounts", "reset"]
func stitchCIDRSegments(parts []string) []string {
	if len(parts) < 2 {
		return parts
	}
	if !strings.Contains(parts[0], ":") {
		return parts
	}
	// 第一段是 IPv6（含 ':'），第二段是 0-128 的纯数字 → 拼回成 CIDR
	if n, err := strconv.Atoi(parts[1]); err == nil && n >= 0 && n <= 128 {
		merged := append([]string{parts[0] + "/" + parts[1]}, parts[2:]...)
		return merged
	}
	return parts
}

// isMultiSegmentPath 判断 'rest'（TrimPrefix 后剩余的 path）是否是多段（含子 action）。
// CIDR 形式 'fd00:0:3:100::/56' 视作单段；'fd00:0:6::/48/bootstrap-accounts' 视作多段。
//
// 算法：把字符串按 '/' split，stitch CIDR 后看是否仍 > 1 段。
func isMultiSegmentPath(rest string) bool {
	parts := stitchCIDRSegments(strings.Split(rest, "/"))
	return len(parts) > 1
}
