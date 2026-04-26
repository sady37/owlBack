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

// loadDevice queries alarm_device from DB and builds the enablement map.
func (c *AlarmEnablementCache) loadDevice(ctx context.Context, tenantID, deviceID string) {
	if c.db == nil {
		c.setDefaults(deviceID, "")
		return
	}

	var monitorConfigRaw sql.NullString
	err := c.db.QueryRowContext(ctx,
		`SELECT monitor_config FROM alarm_device WHERE tenant_id = $1 AND device_id = $2`,
		tenantID, deviceID,
	).Scan(&monitorConfigRaw)

	if err != nil {
		c.logger.Debug("alarm_device not found, using defaults",
			zap.String("did", deviceID), zap.Error(err))
		c.setDefaults(deviceID, c.resolveDeviceType(ctx, deviceID))
		return
	}

	items := c.parseMonitorConfig(monitorConfigRaw.String, c.resolveDeviceType(ctx, deviceID))
	c.setFromItems(deviceID, items)
}

func (c *AlarmEnablementCache) resolveDeviceType(ctx context.Context, deviceID string) string {
	if c.metaCache == nil {
		return ""
	}
	// DeviceMetaCache stores per-card, but we can scan all cards for this deviceID.
	// For simplicity, query DB directly.
	if c.db == nil {
		return ""
	}
	var dt sql.NullString
	_ = c.db.QueryRowContext(ctx,
		`SELECT device_type FROM device_store WHERE device_id = $1 LIMIT 1`, deviceID,
	).Scan(&dt)
	return dt.String
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

	result := make([]alarm.AlarmItem, 0, len(defaults))
	for _, d := range defaults {
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
