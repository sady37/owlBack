// router_adapters.go — cache 满足 consumer 层接口（CLAUDE.md 规则 #1.3
// 单源：cache 是数据 maintainer，consumer 通过窄接口取数据）。

package service

import (
	"context"

	"owl-common/alarm"
	"owl-common/observation"
)

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
