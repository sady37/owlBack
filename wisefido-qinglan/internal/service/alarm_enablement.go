// alarm_enablement.go — qinglan 端 per-device alarm 使能缓存。
//
// 复用 sensor (wisefido-sensor/internal/service/alarm_enablement.go) 同形态实现：
// **唯一数据源**：spatial_config (config_key='alarm.device_config', spatial_prefix=device /128)。
// JSONB schema：{alarm_items: []alarm.AlarmItem, updated_at: ts}。
//
// 不读 alarm.cloud_config（tenant 模板）—— 那是 wisefido-data UI 用的，与运行时无关。
// device 一旦有 alarm.device_config row 即独立，tenant 模板改动不影响已配置 device。
//
// 用途：qinglan StreamPublisher.PublishAlarm 在源头按 device /128 + alarmType 查使能；
// 未启用直接 drop（不发 iot:alarm:stream，cardagg 端不再收到）。把 enablement gate 收到
// producer 端，符合 [[platform_agent_addressing]] trust-platform-agent 原则。
//
// 阈值 (duration_sec / duration_min) 也存在 EnabledAlarm.AlarmParams 里，
// 当前 qinglan 只用 on/off（IsEnabled 二值）；threshold 主要给 sensor zonealarm 用。
//
// 注：qinglan 不查 cards 表 / 不引 CardMeta；enablement 完全按物理实体地址（device /128
// spatial_config 精确匹配）。无 row → 用 alarm.GetDefaultAlarmItems(deviceType) 兜底。

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"

	"owl-common/alarm"
	"owl-common/observation"

	"go.uber.org/zap"
)

// EnabledAlarm holds the resolved enablement info for one alarm type on one device.
type EnabledAlarm struct {
	AlarmType   string
	AlarmLevel  string
	AlarmParams map[string]interface{} // e.g. duration_sec for LeftBed
}

// AlarmEnablementCache lazy-load + invalidation-aware per-device alarm 使能缓存。
//
// Key: deviceAddr (canonical IPv6 string) → map[alarmType]*EnabledAlarm (only enabled alarms)
type AlarmEnablementCache struct {
	mu     sync.RWMutex
	cache  map[string]*deviceEnablement
	db     *sql.DB
	logger *zap.Logger
}

type deviceEnablement struct {
	loaded  bool
	enabled map[string]*EnabledAlarm
}

func NewAlarmEnablementCache(db *sql.DB, logger *zap.Logger) *AlarmEnablementCache {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AlarmEnablementCache{
		cache:  make(map[string]*deviceEnablement),
		db:     db,
		logger: logger,
	}
}

// IsEnabled 检查 device 是否启用了某条 alarm。返 (enabledAlarm, ok)；ok=false 表示未启用或未配置。
func (c *AlarmEnablementCache) IsEnabled(ctx context.Context, deviceAddr, alarmType string) (*EnabledAlarm, bool) {
	if deviceAddr == "" || alarmType == "" {
		return nil, false
	}
	c.mu.RLock()
	de := c.cache[deviceAddr]
	if de != nil && de.loaded {
		c.mu.RUnlock()
		ea := de.enabled[alarmType]
		if ea == nil {
			baseType, _ := alarm.SplitAlarmQualifier(alarmType)
			ea = de.enabled[baseType]
		}
		return ea, ea != nil
	}
	c.mu.RUnlock()

	c.loadDevice(ctx, deviceAddr)

	c.mu.RLock()
	defer c.mu.RUnlock()
	de = c.cache[deviceAddr]
	if de == nil {
		return nil, false
	}
	ea := de.enabled[alarmType]
	if ea == nil {
		baseType, _ := alarm.SplitAlarmQualifier(alarmType)
		ea = de.enabled[baseType]
	}
	return ea, ea != nil
}

// Invalidate 失效单 device cache（spatial_config 改了 → consumer 调）。
func (c *AlarmEnablementCache) Invalidate(deviceAddr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, deviceAddr)
}

// InvalidateAll 清空整个缓存。
func (c *AlarmEnablementCache) InvalidateAll() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.cache = make(map[string]*deviceEnablement)
	c.mu.Unlock()
}

// InvalidateDevices 批量失效。
func (c *AlarmEnablementCache) InvalidateDevices(addrs []string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, a := range addrs {
		if a == "" {
			continue
		}
		delete(c.cache, a)
	}
}

// loadDevice 解析 device 的 alarm 使能配置：
//  1. loadFromDeviceConfig 查 spatial_config alarm.device_config /128 精确匹配
//  2. 无 row → alarm.GetDefaultAlarmItems(deviceType) 硬编码默认值兜底
func (c *AlarmEnablementCache) loadDevice(ctx context.Context, deviceAddr string) {
	if c.db == nil {
		c.setDefaults(deviceAddr, "")
		return
	}
	if items := c.loadFromDeviceConfig(ctx, deviceAddr); items != nil {
		c.setFromItems(deviceAddr, items)
		return
	}
	deviceType := c.resolveDeviceType(ctx, deviceAddr)
	c.setFromItems(deviceAddr, alarm.GetDefaultAlarmItems(deviceType))
}

// loadFromDeviceConfig 读 spatial_config alarm.device_config 精确匹配 /128。
func (c *AlarmEnablementCache) loadFromDeviceConfig(ctx context.Context, deviceAddr string) []alarm.AlarmItem {
	if c.db == nil || deviceAddr == "" {
		return nil
	}
	var raw []byte
	err := c.db.QueryRowContext(ctx,
		`SELECT config_value
		 FROM spatial_config
		 WHERE config_key = 'alarm.device_config'
		   AND spatial_prefix = $1::inet
		 LIMIT 1`, deviceAddr,
	).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var packed struct {
		AlarmItems []alarm.AlarmItem `json:"alarm_items"`
	}
	if err := json.Unmarshal(raw, &packed); err != nil {
		c.logger.Warn("parse alarm.device_config", zap.String("device", deviceAddr), zap.Error(err))
		return nil
	}
	if len(packed.AlarmItems) == 0 {
		return nil
	}
	return packed.AlarmItems
}

// resolveDeviceType 按 device_addr 反查 device_type；仅在 device_config 缺失时用。
func (c *AlarmEnablementCache) resolveDeviceType(ctx context.Context, deviceAddr string) string {
	if c.db == nil || deviceAddr == "" {
		return ""
	}
	var dt sql.NullString
	_ = c.db.QueryRowContext(ctx,
		`SELECT dfm.device_type::text
		 FROM device_factory_meta dfm
		 JOIN devices d ON d.device_uid = dfm.device_uid
		 WHERE d.device_addr = $1::inet
		 LIMIT 1`, deviceAddr,
	).Scan(&dt)
	return dt.String
}

func (c *AlarmEnablementCache) setDefaults(deviceAddr, deviceType string) {
	items := alarm.GetDefaultAlarmItems(deviceType)
	c.setFromItems(deviceAddr, items)
}

func (c *AlarmEnablementCache) setFromItems(deviceAddr string, items []alarm.AlarmItem) {
	enabled := make(map[string]*EnabledAlarm)
	for _, item := range items {
		if item.IsEnabled == nil || *item.IsEnabled != 1 {
			continue
		}
		level := ""
		if item.AlarmLevel != nil {
			level = observation.NormalizeEventLevel(*item.AlarmLevel)
		}
		if level == "" {
			continue
		}
		ea := &EnabledAlarm{AlarmType: item.AlarmType, AlarmLevel: level}
		if len(item.AlarmParams) > 0 {
			ea.AlarmParams = item.AlarmParams
		}
		enabled[item.AlarmType] = ea
	}

	c.mu.Lock()
	c.cache[deviceAddr] = &deviceEnablement{loaded: true, enabled: enabled}
	c.mu.Unlock()
}
