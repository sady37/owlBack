package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// AlarmDeviceRepository 设备报警配置仓库
type AlarmDeviceRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewAlarmDeviceRepository 创建设备报警配置仓库
func NewAlarmDeviceRepository(db *sql.DB, logger *zap.Logger) *AlarmDeviceRepository {
	return &AlarmDeviceRepository{
		db:     db,
		logger: logger,
	}
}

// AlarmDeviceConfig 设备报警配置
type AlarmDeviceConfig struct {
	DeviceID      string          `json:"device_id"`
	TenantID      string          `json:"tenant_id"`
	MonitorConfig json.RawMessage `json:"monitor_config"` // 完整监控配置（JSONB）
	VendorConfig  json.RawMessage `json:"vendor_config"`  // 厂家参考配置（JSONB）
	Metadata      json.RawMessage `json:"metadata"`       // 元数据（JSONB）
}

// GetAlarmDeviceConfig 获取设备的报警配置（需验证 tenant_id）
func (r *AlarmDeviceRepository) GetAlarmDeviceConfig(ctx context.Context, tenantID, deviceID string) (*AlarmDeviceConfig, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	query := `
		SELECT 
			device_id,
			tenant_id,
			monitor_config,
			vendor_config,
			metadata
		FROM alarm_device
		WHERE device_id = $1 AND tenant_id = $2
	`

	var config AlarmDeviceConfig
	err := r.db.QueryRowContext(ctx, query, deviceID, tenantID).Scan(
		&config.DeviceID,
		&config.TenantID,
		&config.MonitorConfig,
		&config.VendorConfig,
		&config.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 设备没有配置，返回 nil
		}
		return nil, fmt.Errorf("failed to query alarm_device: %w", err)
	}

	return &config, nil
}

// GetDeviceMonitorConfig 获取设备的完整监控配置（使用数据库函数，需验证 tenant_id）
func (r *AlarmDeviceRepository) GetDeviceMonitorConfig(ctx context.Context, tenantID, deviceID string) (json.RawMessage, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}

	query := `SELECT get_iot_monitor_config($1, $2)`

	var config json.RawMessage
	err := r.db.QueryRowContext(ctx, query, tenantID, deviceID).Scan(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to get device monitor config: %w", err)
	}

	return config, nil
}

// GetDeviceDefaultMonitorConfig 获取设备类型的默认配置（用于初次配置，需验证 tenant_id）
func (r *AlarmDeviceRepository) GetDeviceDefaultMonitorConfig(ctx context.Context, tenantID, deviceType string) (json.RawMessage, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceType == "" {
		return nil, fmt.Errorf("device_type is required")
	}

	query := `SELECT get_device_default_monitor_config($1, $2, NULL)`

	var config json.RawMessage
	err := r.db.QueryRowContext(ctx, query, tenantID, deviceType).Scan(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to get device default monitor config: %w", err)
	}

	return config, nil
}
