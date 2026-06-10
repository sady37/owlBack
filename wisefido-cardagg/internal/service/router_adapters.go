// router_adapters.go — cache 满足 consumer 层接口（CLAUDE.md 规则 #1.3
// 单源：cache 是数据 maintainer，consumer 通过窄接口取数据）。

package service

import (
	"context"

	"owl-common/observation"
)

// Resolve 给 alarm_router 的 EnablementResolver 接口用。
// level 唯一来自 device_config（生成时已填 default 作保底）；未配置（!ok）返回空 + false。
func (c *AlarmEnablementCache) Resolve(ctx context.Context, tenantPref, deviceAddr, alarmType string) (level string, enabled bool) {
	ea, ok := c.IsEnabled(ctx, tenantPref, deviceAddr, alarmType)
	if !ok {
		return "", false
	}
	level = ea.AlarmLevel
	if level != "" {
		level = observation.NormalizeEventLevel(level)
	}
	return level, level != ""
}
