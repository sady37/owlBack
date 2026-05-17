// router_adapters.go — 让 v1 cache 实现 consumer 层接口（CLAUDE.md 规则 #1.3
// 单源：cache 是数据 maintainer，consumer 通过窄接口取数据）。

package service

import (
	"context"
	"net/netip"

	"owl-common/alarm"
	"owl-common/observation"
)

// ResolveDeviceAddr 给 alarm_router 的 MetaResolver 接口用。
// hint 已是 canonical /128 字符串时直接返回；空 hint 时返回空（caller 早退）。
// device_ipv6 单程票后不再做 UUID/UID 反查（R-001）。
func (c *DeviceMetaCache) ResolveDeviceAddr(ctx context.Context, cardID, hint string) string {
	return c.ResolveDeviceID(ctx, cardID, hint, hint)
}

// LookupCardByDevice device_addr (/128 IPv6 字符串) → cardID。
// sensor 发 alarm 时 SubjectEntity 留空（不做 device→card 反查），由 cardagg LPM 兜底。
func (c *DeviceMetaCache) LookupCardByDevice(ctx context.Context, deviceAddr string) string {
	addr, err := netip.ParseAddr(deviceAddr)
	if err != nil {
		return ""
	}
	return c.LookupCardByDeviceAddr(ctx, addr)
}

// Resolve 给 alarm_router 的 EnablementResolver 接口用。
// 缺省 level 取 alarm.Registry.DefaultLevel；DefaultLevel 仍空视为未配置。
func (c *AlarmEnablementCache) Resolve(ctx context.Context, tenantPref, deviceAddr, alarmType string) (level string, enabled bool) {
	ea, ok := c.IsEnabled(ctx, tenantPref, deviceAddr, alarmType)
	if !ok {
		return "", false
	}
	level = ea.AlarmLevel
	if level == "" {
		if def := alarm.LookupAlarm(alarmType); def != nil {
			level = def.DefaultLevel
		}
	}
	if level != "" {
		level = observation.NormalizeEventLevel(level)
	}
	return level, level != ""
}
