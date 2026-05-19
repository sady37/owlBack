// alarm_enablement.go — sensor 端 per-device alarm 使能缓存。
//
// 数据源：spatial_config 表，config_key='alarm.cloud_config'，按 device_addr LPM 查（v2 IPv6 单程票）。
// JSONB 结构：{"device_alarms": {"Radar": {"Fall": {"is_enabled":1, "alarm_level":"CRITICAL"}, ...}, "SleepPad": {...}}}
//
// 命名 quirk：device_factory_meta.device_type 用 "Sleepad"，但 JSONB key 是 "SleepPad"（首字母大写 P），需兼容映射。
//
// 用途：sensor 内部所有 alarm 出口（zonealarm.Supervisor / fall firers / aiPublisher）
// 在 publish 前 IsEnabled 检查；未启用直接丢弃，不发往 cardagg。
//
// 注：sensor 不查 cards 表 / 不引 CardMeta；按"sensor 不知 card"原则，
// enablement 完全按物理实体地址（device /128 通过 spatial_config LPM 上溯）。

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

// Invalidate 失效单个 device 的使能缓存（spatial_config 改了 / device 解绑时调）。
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
//  1. resolveDeviceType 查 device_type（device_factory_meta JOIN devices via device_ipv6）
//  2. loadFromSpatialConfig 按 device_addr LPM 取 spatial_config 'alarm.cloud_config'，解析 device_type 节
//  3. 与 alarm.GetDefaultAlarmItems 合并（spatial_config 覆盖 default）
//  4. spatial_config 缺省 → 用 defaults
func (c *AlarmEnablementCache) loadDevice(ctx context.Context, deviceAddr string) {
	if c.db == nil {
		c.setDefaults(deviceAddr, "")
		return
	}
	deviceType := c.resolveDeviceType(ctx, deviceAddr)
	items := c.loadFromSpatialConfig(ctx, deviceAddr, deviceType)
	if items == nil {
		items = alarm.GetDefaultAlarmItems(deviceType)
	}
	c.setFromItems(deviceAddr, items)
}

// resolveDeviceType 按 device_ipv6 反查 device_type（v2 IPv6 单程票）。
func (c *AlarmEnablementCache) resolveDeviceType(ctx context.Context, deviceAddr string) string {
	if c.db == nil || deviceAddr == "" {
		return ""
	}
	var dt sql.NullString
	_ = c.db.QueryRowContext(ctx,
		`SELECT dfm.device_type::text
		 FROM device_factory_meta dfm
		 JOIN devices d ON d.device_id = dfm.device_id
		 WHERE d.device_ipv6 = $1::inet
		 LIMIT 1`, deviceAddr,
	).Scan(&dt)
	return dt.String
}

// loadFromSpatialConfig 按 device_addr LPM 取 alarm.cloud_config，解析对应 device_type 节，
// 与 defaults 合并返回。无配置返回 nil（caller fallback 到 defaults）。
func (c *AlarmEnablementCache) loadFromSpatialConfig(ctx context.Context, deviceAddr, deviceType string) []alarm.AlarmItem {
	if c.db == nil || deviceAddr == "" {
		return nil
	}
	var raw []byte
	err := c.db.QueryRowContext(ctx,
		`SELECT config_value
		 FROM spatial_config
		 WHERE config_key = 'alarm.cloud_config'
		   AND spatial_prefix >>= $1::inet
		 ORDER BY masklen(spatial_prefix) DESC
		 LIMIT 1`, deviceAddr,
	).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var cfg struct {
		DeviceAlarms map[string]map[string]struct {
			IsEnabled   *int                   `json:"is_enabled,omitempty"`
			AlarmLevel  *string                `json:"alarm_level,omitempty"`
			AlarmParams map[string]interface{} `json:"alarm_params,omitempty"`
		} `json:"device_alarms"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		c.logger.Warn("parse alarm.cloud_config", zap.String("device", deviceAddr), zap.Error(err))
		return nil
	}
	if len(cfg.DeviceAlarms) == 0 {
		return nil
	}
	candidateKeys := alarmCloudConfigKeysForDeviceType(deviceType)
	var override map[string]struct {
		IsEnabled   *int                   `json:"is_enabled,omitempty"`
		AlarmLevel  *string                `json:"alarm_level,omitempty"`
		AlarmParams map[string]interface{} `json:"alarm_params,omitempty"`
	}
	for _, k := range candidateKeys {
		if v, ok := cfg.DeviceAlarms[k]; ok {
			override = v
			break
		}
	}
	if override == nil {
		return nil
	}
	defaults := alarm.GetDefaultAlarmItems(deviceType)
	defaultTypes := make(map[string]bool, len(defaults))
	result := make([]alarm.AlarmItem, 0, len(defaults)+len(override))
	for _, d := range defaults {
		defaultTypes[d.AlarmType] = true
		item := d
		if o, ok := override[d.AlarmType]; ok {
			if o.IsEnabled != nil {
				item.IsEnabled = o.IsEnabled
			}
			if o.AlarmLevel != nil && *o.AlarmLevel != "" {
				lvl := *o.AlarmLevel
				item.AlarmLevel = &lvl
			}
			if len(o.AlarmParams) > 0 {
				item.AlarmParams = o.AlarmParams
			}
		}
		result = append(result, item)
	}
	// spatial_config 里有但 defaults 没有的项也保留（设备健康类 Offline/SignalPoor 等）
	for alarmType, o := range override {
		if defaultTypes[alarmType] {
			continue
		}
		if o.IsEnabled == nil || *o.IsEnabled != 1 {
			continue
		}
		var lvl *string
		if o.AlarmLevel != nil && *o.AlarmLevel != "" {
			s := *o.AlarmLevel
			lvl = &s
		}
		result = append(result, alarm.AlarmItem{
			AlarmType:   alarmType,
			IsEnabled:   o.IsEnabled,
			AlarmLevel:  lvl,
			AlarmParams: o.AlarmParams,
		})
	}
	return result
}

// alarmCloudConfigKeysForDeviceType device_factory_meta.device_type → JSONB key 映射。
// quirk：DB device_type="Sleepad"，JSONB key="SleepPad"（首字母大写 P），同时兼容 lowercase。
func alarmCloudConfigKeysForDeviceType(deviceType string) []string {
	switch deviceType {
	case "Radar", "radar":
		return []string{"Radar", "radar"}
	case "Sleepad", "sleepad", "SleepPad", "sleeppad":
		return []string{"SleepPad", "Sleepad", "sleeppad", "sleepad"}
	}
	return nil
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
