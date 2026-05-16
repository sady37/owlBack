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

// AlarmEnablementCache is a lazy-loading, invalidation-aware in-memory cache
// of per-device alarm enablement.
//
// Key: deviceID → map[alarmType]*EnabledAlarm (only enabled alarms).
// Load path: alarm_device.monitor_config → parse → filter enabled.
// Fallback: alarm.DefaultAlarmSetting by device type.
type AlarmEnablementCache struct {
	mu        sync.RWMutex
	cache     map[string]*deviceEnablement
	db        *sql.DB
	metaCache *DeviceMetaCache
	logger    *zap.Logger
}

type deviceEnablement struct {
	loaded  bool
	enabled map[string]*EnabledAlarm // alarm_type → item
}

func NewAlarmEnablementCache(db *sql.DB, metaCache *DeviceMetaCache, logger *zap.Logger) *AlarmEnablementCache {
	return &AlarmEnablementCache{
		cache:     make(map[string]*deviceEnablement),
		db:        db,
		metaCache: metaCache,
		logger:    logger,
	}
}

// IsEnabled checks whether the given alarm type is enabled for a device.
// Returns (enabledAlarm, ok). When ok=false, the alarm is not enabled or not found.
func (c *AlarmEnablementCache) IsEnabled(ctx context.Context, tenantID, deviceID, alarmType string) (*EnabledAlarm, bool) {
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

	c.loadDevice(ctx, tenantID, deviceID)

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

// Invalidate marks a device's enablement as stale so it reloads on next access.
func (c *AlarmEnablementCache) Invalidate(deviceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, deviceID)
}

// InvalidateAll clears the entire enablement cache (used on reset).
func (c *AlarmEnablementCache) InvalidateAll() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.cache = make(map[string]*deviceEnablement)
	c.mu.Unlock()
}

// InvalidateDevices 批量失效告警使能缓存（同 unit 下多设备）。
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

// loadDevice v2：从 spatial_config longest-prefix-match 解析 device 的 alarm 使能配置。
//
// 数据源：spatial_config 表，config_key='alarm.cloud_config'，spatial_prefix LPM 反查。
// JSONB 结构：{"device_alarms": {"Radar": {"Fall": {"is_enabled":1, "alarm_level":"CRITICAL"}, ...}, "SleepPad": {...}}}
// 命名 quirk：device_factory_meta.device_type 用 "Sleepad"，但 JSONB 键是 "SleepPad"（首字母大写 P），需兼容映射。
//
// 路径：
//   1. resolveDeviceType 用 device_ipv6 LPM 反查 device_type（v2 IPv6 单程票后 deviceID 是 IPv6 字符串）
//   2. 读 spatial_config LPM；解析 device_alarms.<deviceTypeKey>.<alarmType>.{is_enabled,alarm_level}
//   3. 与 alarm.GetDefaultAlarmItems 合并（spatial_config 项覆盖 defaults，未配置项保留 default）
//   4. spatial_config 完全未配 → 直接用 defaults
//
// 缓存：调用方通过 IsEnabled 触发，命中即缓存；spatial_config 改了需 Invalidate 才生效（admin 更新 alarm 配置后应触发，目前不自动）。
func (c *AlarmEnablementCache) loadDevice(ctx context.Context, tenantID, deviceID string) {
	if c.db == nil {
		c.setDefaults(deviceID, "")
		return
	}
	deviceType := c.resolveDeviceType(ctx, deviceID)
	items := c.loadFromSpatialConfig(ctx, deviceID, deviceType)
	if items == nil {
		items = alarm.GetDefaultAlarmItems(deviceType)
	}
	c.setFromItems(deviceID, items)
}

// resolveDeviceType v2 IPv6 单程票：cardagg 调用方传入的 deviceID 是 IPv6 字符串（device_ipv6），
// 用 devices JOIN device_factory_meta 反查 device_type。
// 旧实现 `device_id = $1::uuid` 在 v2 cutover 后 cast 失败被静默吞，导致 device_type="" → defaults nil → 全 alarm 静默 drop。
func (c *AlarmEnablementCache) resolveDeviceType(ctx context.Context, deviceID string) string {
	if c.db == nil || deviceID == "" {
		return ""
	}
	var dt sql.NullString
	_ = c.db.QueryRowContext(ctx,
		`SELECT dfm.device_type::text
		 FROM device_factory_meta dfm
		 JOIN devices d ON d.device_id = dfm.device_id
		 WHERE d.device_ipv6 = $1::inet
		 LIMIT 1`, deviceID,
	).Scan(&dt)
	return dt.String
}

// loadFromSpatialConfig 按 device_addr LPM 取 alarm.cloud_config，解析对应 device_type 节，
// 与 defaults 合并返回。无配置或解析失败返回 nil（调用方 fallback 到 defaults）。
func (c *AlarmEnablementCache) loadFromSpatialConfig(ctx context.Context, deviceID, deviceType string) []alarm.AlarmItem {
	if c.db == nil || deviceID == "" {
		return nil
	}
	var raw []byte
	err := c.db.QueryRowContext(ctx,
		`SELECT config_value
		 FROM spatial_config
		 WHERE config_key = 'alarm.cloud_config'
		   AND spatial_prefix >>= $1::inet
		 ORDER BY masklen(spatial_prefix) DESC
		 LIMIT 1`, deviceID,
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
		c.logger.Warn("parse alarm.cloud_config", zap.String("device", deviceID), zap.Error(err))
		return nil
	}
	if len(cfg.DeviceAlarms) == 0 {
		return nil
	}
	// 命名映射：device_factory_meta="Sleepad" ↔ JSONB key="SleepPad"
	// 同时容错 lowercase（未来若改 lowercase 不破）
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
	// spatial_config 里有但 defaults 没有的项也保留（设备健康类 Offline / SignalPoor 等通常不在 defaults，由系统强制）
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

// alarmCloudConfigKeysForDeviceType 把 device_factory_meta.device_type 映射到 alarm.cloud_config JSONB key。
// 已知 quirk：DB device_type="Sleepad"，但 JSONB key="SleepPad"（首字母大写 P）；同时兼容 lowercase 防未来归一。
func alarmCloudConfigKeysForDeviceType(deviceType string) []string {
	switch deviceType {
	case "Radar", "radar":
		return []string{"Radar", "radar"}
	case "Sleepad", "sleepad", "SleepPad", "sleeppad":
		return []string{"SleepPad", "Sleepad", "sleeppad", "sleepad"}
	}
	return nil
}

// parseMonitorConfig extracts AlarmItems from monitor_config JSON.
// monitor_config 结构：{"items": [{alarm_type, is_enabled, alarm_level, alarm_params, display_setting}, ...]}
// 由 wisefido-data 的 device_monitor_settings_service 写入，必须保持一致。
func (c *AlarmEnablementCache) parseMonitorConfig(raw string, deviceType string) []alarm.AlarmItem {
	if raw == "" {
		return alarm.GetDefaultAlarmItems(deviceType)
	}

	var mc struct {
		Items []alarm.AlarmItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(raw), &mc); err != nil {
		c.logger.Warn("parse monitor_config", zap.Error(err))
		return alarm.GetDefaultAlarmItems(deviceType)
	}

	defaults := alarm.GetDefaultAlarmItems(deviceType)
	if defaults == nil {
		return nil
	}

	overrides := make(map[string]alarm.AlarmItem, len(mc.Items))
	for _, it := range mc.Items {
		if it.AlarmType == "" {
			continue
		}
		overrides[it.AlarmType] = it
	}

	defaultTypes := make(map[string]bool, len(defaults))
	result := make([]alarm.AlarmItem, 0, len(defaults)+len(overrides))
	for _, d := range defaults {
		defaultTypes[d.AlarmType] = true
		item := d
		if cfg, ok := overrides[d.AlarmType]; ok {
			if cfg.IsEnabled != nil {
				item.IsEnabled = cfg.IsEnabled
			}
			if cfg.AlarmLevel != nil {
				item.AlarmLevel = cfg.AlarmLevel
			}
			if len(cfg.AlarmParams) > 0 {
				item.AlarmParams = cfg.AlarmParams
			}
			if cfg.DisplaySetting != 0 {
				item.DisplaySetting = cfg.DisplaySetting
			}
		}
		result = append(result, item)
	}
	// monitor_config 里有但 defaults 没有的项也保留（设备健康类 Offline / SignalPoor /
	// AngleException / DeviceFailure 通常不在 defaults 里——它们由系统强制审计，
	// is_enabled / alarm_level 来自 monitor_config 即可）。
	for _, cfg := range overrides {
		if defaultTypes[cfg.AlarmType] {
			continue
		}
		result = append(result, cfg)
	}
	return result
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
