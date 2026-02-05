package httpapi

import (
	"net"
	"net/http"
)

// IsRequestFromLocalhost 判断请求是否来自本机（用于 BackendCall 限制）
// 仅当客户端 IP 为 127.0.0.1 或 ::1 时返回 true
func IsRequestFromLocalhost(r *http.Request) bool {
	ip := getClientIP(r)
	if ip == "" {
		return false
	}
	// 支持 IPv4 和 IPv6 环回
	if ip == "127.0.0.1" || ip == "::1" {
		return true
	}
	// 支持带端口的格式，如 127.0.0.1:12345
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	return ip == "127.0.0.1" || ip == "::1"
}
