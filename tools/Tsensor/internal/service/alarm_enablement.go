// alarm_enablement.go — sensor 端 per-device alarm 使能缓存。
//
// **唯一数据源**：device_config (config_key='alarm.device_config', key=device_uid；经 devices 把 device_addr→uid)。
// JSONB schema：{alarm_items: []alarm.AlarmItem, updated_at: ts}。
//
// **不读 alarm.cloud_config**（tenant 级模板，在 spatial_config）—— 那是 wisefido-data 给 UI"reset to default"和
// 新设备 bind 时取 snapshot 用的，与运行时无关。device 一旦有 alarm.device_config row 即独立，
// tenant 模板改动不影响已配置 device（用户拍板：snapshot model，不级联）。
//
// 用途：sensor 内部所有 alarm 出口（zonealarm.Supervisor / fall firers / aiPublisher）
// 在 publish 前 IsEnabled 检查；未启用直接丢弃，不发往 cardagg。
// 阈值 (duration_sec/duration_min) 也从此处取（zonealarm.ThresholdResolver 包装 IsEnabled）。
//
// 注：sensor 不查 cards 表 / 不引 CardMeta；按"sensor 不知 card"原则，
// enablement 完全按物理实体地址（device_config 按 device_uid 精确匹配）。
// 无 row → 用 alarm.GetDefaultAlarmItems(deviceType) 兜底（新设备未经 UI 配置场景）。

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

// IsEnabled 检查 device 是否启用了某条 alarm。
// 返回 (enabledAlarm, ok)；ok=false 表示未启用或未配置。
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

// Invalidate 失效单个 device 的使能缓存（device_config 改了 / device 解绑时调）。
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
//  1. loadFromDeviceConfig 查 device_config alarm.device_config（key=device_uid）
//  2. 无 row → alarm.GetDefaultAlarmItems(deviceType) 硬编码默认值兜底
//
// tenant 级 alarm.cloud_config 不读 — 那是 wisefido-data UI 的模板，与运行时无关。
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

// loadFromDeviceConfig 读 device_config alarm.device_config（key=device_uid，经 devices 把 device_addr→uid）。
// 配置由 wisefido-data UpdateDeviceMonitorSettings 写入；schema = {alarm_items: []AlarmItem}。
// 未配置 → nil（caller 用 GetDefaultAlarmItems 兜底）。
func (c *AlarmEnablementCache) loadFromDeviceConfig(ctx context.Context, deviceAddr string) []alarm.AlarmItem {
	if c.db == nil || deviceAddr == "" {
		return nil
	}
	var raw []byte
	err := c.db.QueryRowContext(ctx,
		`SELECT dc.config_value
		 FROM device_config dc
		 JOIN devices d ON d.device_uid = dc.device_uid
		 WHERE dc.config_key = 'alarm.device_config'
		   AND d.device_addr = $1::inet
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

// resolveDeviceType 按 device_addr 反查 device_type。
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
