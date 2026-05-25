// alarm_enablement.go — cardagg 端 per-device alarm 使能缓存。
//
// **唯一数据源**：spatial_config (config_key='alarm.device_config', spatial_prefix=device /128)。
// JSONB schema：{alarm_items: []alarm.AlarmItem, updated_at: ts}。
//
// **不读 alarm.cloud_config**（tenant 级模板）—— 那是 wisefido-data 给 UI"reset to default"和
// 新设备 bind 时取 snapshot 用的，与运行时无关。device 一旦有 alarm.device_config row 即独立，
// tenant 模板改动不影响已配置 device（用户拍板：snapshot model，不级联）。
//
// 失效路径：wisefido-data UpdateDeviceMonitorSettings 写 spatial_config 后 publish
// config:alarmDevice:stream → cardagg AlarmDeviceConfigConsumer.Invalidate(deviceID) →
// 下次 IsEnabled lazy reload。

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

// AlarmEnablementCache lazy-loading + invalidation-aware per-device alarm 使能缓存。
//
// Key: deviceID (canonical IPv6 string) → map[alarmType]*EnabledAlarm（仅已启用的 alarm）。
type AlarmEnablementCache struct {
	mu     sync.RWMutex
	cache  map[string]*deviceEnablement
	db     *sql.DB
	logger *zap.Logger
}

type deviceEnablement struct {
	loaded  bool
	enabled map[string]*EnabledAlarm // alarm_type → item
}

func NewAlarmEnablementCache(db *sql.DB, logger *zap.Logger) *AlarmEnablementCache {
	return &AlarmEnablementCache{
		cache:  make(map[string]*deviceEnablement),
		db:     db,
		logger: logger,
	}
}

// IsEnabled 查 device 是否启用了某条 alarm。返 (enabledAlarm, ok)。
// tenantID 参数保留兼容旧 signature；本实现不用（device_config 按 /128 精确匹配）。
func (c *AlarmEnablementCache) IsEnabled(ctx context.Context, tenantID, deviceID, alarmType string) (*EnabledAlarm, bool) {
	_ = tenantID
	if deviceID == "" || alarmType == "" {
		return nil, false
	}
	c.mu.RLock()
	de := c.cache[deviceID]
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

	c.loadDevice(ctx, deviceID)

	c.mu.RLock()
	defer c.mu.RUnlock()
	de = c.cache[deviceID]
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
func (c *AlarmEnablementCache) Invalidate(deviceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, deviceID)
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
func (c *AlarmEnablementCache) InvalidateDevices(deviceIDs []string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, id := range deviceIDs {
		if id == "" {
			continue
		}
		delete(c.cache, id)
	}
}

// loadDevice 解析 device 的 alarm 使能配置：
//  1. loadFromDeviceConfig 查 spatial_config alarm.device_config /128 精确匹配
//  2. 无 row → alarm.GetDefaultAlarmItems(deviceType) 硬编码默认值兜底
//
// tenant 级 alarm.cloud_config 不读 — 那是 wisefido-data UI 的模板，与运行时无关。
func (c *AlarmEnablementCache) loadDevice(ctx context.Context, deviceID string) {
	if c.db == nil {
		c.setDefaults(deviceID, "")
		return
	}
	if items := c.loadFromDeviceConfig(ctx, deviceID); items != nil {
		c.setFromItems(deviceID, items)
		return
	}
	deviceType := c.resolveDeviceType(ctx, deviceID)
	c.setFromItems(deviceID, alarm.GetDefaultAlarmItems(deviceType))
}

// resolveDeviceType 用 device_addr (IPv6 text) 反查 device_type。
// 仅在 device_config 缺失时用（取 defaults 需知道设备类型）。
func (c *AlarmEnablementCache) resolveDeviceType(ctx context.Context, deviceID string) string {
	if c.db == nil || deviceID == "" {
		return ""
	}
	var dt sql.NullString
	_ = c.db.QueryRowContext(ctx,
		`SELECT dfm.device_type::text
		 FROM device_factory_meta dfm
		 JOIN devices d ON d.device_uid = dfm.device_uid
		 WHERE d.device_addr = $1::inet
		 LIMIT 1`, deviceID,
	).Scan(&dt)
	return dt.String
}

// loadFromDeviceConfig 读 spatial_config alarm.device_config /128 精确匹配。
// 配置由 wisefido-data UpdateDeviceMonitorSettings 写入；schema = {alarm_items: []AlarmItem}。
// 未配置 → nil（caller 用 GetDefaultAlarmItems 兜底）。
func (c *AlarmEnablementCache) loadFromDeviceConfig(ctx context.Context, deviceID string) []alarm.AlarmItem {
	if c.db == nil || deviceID == "" {
		return nil
	}
	var raw []byte
	err := c.db.QueryRowContext(ctx,
		`SELECT config_value
		 FROM spatial_config
		 WHERE config_key = 'alarm.device_config'
		   AND spatial_prefix = $1::inet
		 LIMIT 1`, deviceID,
	).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var packed struct {
		AlarmItems []alarm.AlarmItem `json:"alarm_items"`
	}
	if err := json.Unmarshal(raw, &packed); err != nil {
		c.logger.Warn("parse alarm.device_config", zap.String("device", deviceID), zap.Error(err))
		return nil
	}
	if len(packed.AlarmItems) == 0 {
		return nil
	}
	return packed.AlarmItems
}

func (c *AlarmEnablementCache) setDefaults(deviceID, deviceType string) {
	items := alarm.GetDefaultAlarmItems(deviceType)
	c.setFromItems(deviceID, items)
}

func (c *AlarmEnablementCache) setFromItems(deviceID string, items []alarm.AlarmItem) {
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
	c.cache[deviceID] = &deviceEnablement{loaded: true, enabled: enabled}
	c.mu.Unlock()
}
