package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// AlarmCloudRepository 报警策略仓库
type AlarmCloudRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewAlarmCloudRepository 创建报警策略仓库
func NewAlarmCloudRepository(db *sql.DB, logger *zap.Logger) *AlarmCloudRepository {
	return &AlarmCloudRepository{
		db:     db,
		logger: logger,
	}
}

// AlarmCloudConfig 报警策略配置
type AlarmCloudConfig struct {
	TenantID          *string         `json:"tenant_id"`          // NULL 表示系统默认策略
	OfflineAlarm      *string         `json:"offline_alarm"`      // 通用报警：设备离线
	LowBattery        *string         `json:"low_battery"`        // 通用报警：低电量
	DeviceFailure     *string         `json:"device_failure"`     // 通用报警：设备故障
	DeviceAlarms      json.RawMessage `json:"device_alarms"`      // 设备特定报警配置（JSONB）
	Conditions        json.RawMessage `json:"conditions"`         // 阈值配置（JSONB）
	NotificationRules json.RawMessage `json:"notification_rules"` // 通知规则（JSONB）
	Metadata          json.RawMessage `json:"metadata"`           // 元数据（JSONB）
}

// GetAlarmCloudConfig 获取租户的报警策略配置
// 匹配优先级：1) 租户特定配置，2) 系统默认配置（tenant_id = NULL）
func (r *AlarmCloudRepository) GetAlarmCloudConfig(tenantID string) (*AlarmCloudConfig, error) {
	// 1. 优先查询租户特定配置
	var config AlarmCloudConfig
	query := `
		SELECT 
			tenant_id,
			"OfflineAlarm",
			"LowBattery",
			"DeviceFailure",
			device_alarms,
			conditions,
			notification_rules,
			metadata
		FROM alarm_cloud
		WHERE tenant_id = $1
	`

	err := r.db.QueryRow(query, tenantID).Scan(
		&config.TenantID,
		&config.OfflineAlarm,
		&config.LowBattery,
		&config.DeviceFailure,
		&config.DeviceAlarms,
		&config.Conditions,
		&config.NotificationRules,
		&config.Metadata,
	)

	if err == nil {
		// 找到租户特定配置
		return &config, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query alarm_cloud: %w", err)
	}

	// 2. 如果租户没有配置，查询系统默认配置（tenant_id = NULL）
	query = `
		SELECT 
			tenant_id,
			"OfflineAlarm",
			"LowBattery",
			"DeviceFailure",
			device_alarms,
			conditions,
			notification_rules,
			metadata
		FROM alarm_cloud
		WHERE tenant_id IS NULL
	`

	err = r.db.QueryRow(query).Scan(
		&config.TenantID,
		&config.OfflineAlarm,
		&config.LowBattery,
		&config.DeviceFailure,
		&config.DeviceAlarms,
		&config.Conditions,
		&config.NotificationRules,
		&config.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 没有系统默认配置，返回空配置
			return &AlarmCloudConfig{
				DeviceAlarms: json.RawMessage("{}"),
			}, nil
		}
		return nil, fmt.Errorf("failed to query system default alarm_cloud: %w", err)
	}

	return &config, nil
}

// GetDeviceTypeAlarmConfig 获取设备类型的报警配置（使用数据库函数）
// 返回包含通用报警和设备特定报警的配置
func (r *AlarmCloudRepository) GetDeviceTypeAlarmConfig(tenantID, deviceType string) (json.RawMessage, error) {
	query := `SELECT get_device_type_alarm_config($1, $2)`

	var config json.RawMessage
	err := r.db.QueryRow(query, tenantID, deviceType).Scan(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to get device type alarm config: %w", err)
	}

	return config, nil
}
