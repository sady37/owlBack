package repository

import (
	"context"
	"database/sql"
	"fmt"

	"wisefido-data/internal/domain"
)

// PostgresAlarmDeviceRepository 设备告警配置Repository实现（强类型版本）
type PostgresAlarmDeviceRepository struct {
	db *sql.DB
}

// NewPostgresAlarmDeviceRepository 创建设备告警配置Repository
func NewPostgresAlarmDeviceRepository(db *sql.DB) *PostgresAlarmDeviceRepository {
	return &PostgresAlarmDeviceRepository{db: db}
}

// 确保实现了接口
var _ AlarmDeviceRepository = (*PostgresAlarmDeviceRepository)(nil)

// GetAlarmDevice 获取设备的告警配置
func (r *PostgresAlarmDeviceRepository) GetAlarmDevice(ctx context.Context, tenantID, deviceID string) (*domain.AlarmDevice, error) {
	if tenantID == "" || deviceID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			device_id::text,
			tenant_id::text,
			monitor_config,
			vendor_config,
			metadata
		FROM alarm_device
		WHERE tenant_id = $1 AND device_id = $2
	`

	var alarmDevice domain.AlarmDevice
	var monitorConfig, vendorConfig, metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, deviceID).Scan(
		&alarmDevice.DeviceID,
		&alarmDevice.TenantID,
		&monitorConfig,
		&vendorConfig,
		&metadata,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alarm device not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get alarm device: %w", err)
	}

	if monitorConfig.Valid {
		alarmDevice.MonitorConfig = []byte(monitorConfig.String)
	}
	if vendorConfig.Valid {
		alarmDevice.VendorConfig = []byte(vendorConfig.String)
	}
	if metadata.Valid {
		alarmDevice.Metadata = []byte(metadata.String)
	}

	return &alarmDevice, nil
}

// UpsertAlarmDevice 创建或更新设备的告警配置
func (r *PostgresAlarmDeviceRepository) UpsertAlarmDevice(ctx context.Context, tenantID, deviceID string, alarmDevice *domain.AlarmDevice) error {
	if tenantID == "" || deviceID == "" {
		return fmt.Errorf("tenant_id and device_id are required")
	}

	query := `
		INSERT INTO alarm_device (
			device_id,
			tenant_id,
			monitor_config,
			vendor_config,
			metadata
		) VALUES ($1, $2, $3::jsonb, $4::jsonb, $5::jsonb)
		ON CONFLICT (device_id) DO UPDATE SET
			tenant_id = EXCLUDED.tenant_id,
			monitor_config = EXCLUDED.monitor_config,
			vendor_config = EXCLUDED.vendor_config,
			metadata = EXCLUDED.metadata
	`

	var monitorConfig, vendorConfig, metadata interface{}
	if len(alarmDevice.MonitorConfig) > 0 {
		monitorConfig = string(alarmDevice.MonitorConfig)
	} else {
		monitorConfig = "{}"
	}
	if len(alarmDevice.VendorConfig) > 0 {
		vendorConfig = string(alarmDevice.VendorConfig)
	}
	if len(alarmDevice.Metadata) > 0 {
		metadata = string(alarmDevice.Metadata)
	}

	_, err := r.db.ExecContext(ctx, query, deviceID, tenantID, monitorConfig, vendorConfig, metadata)
	if err != nil {
		return fmt.Errorf("failed to upsert alarm device: %w", err)
	}

	return nil
}

// DeleteAlarmDevice 删除设备的告警配置
func (r *PostgresAlarmDeviceRepository) DeleteAlarmDevice(ctx context.Context, tenantID, deviceID string) error {
	if tenantID == "" || deviceID == "" {
		return fmt.Errorf("tenant_id and device_id are required")
	}

	query := `
		DELETE FROM alarm_device
		WHERE tenant_id = $1 AND device_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete alarm device: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("alarm device not found")
	}

	return nil
}

// ListAlarmDevices 批量查询设备的告警配置（支持分页）
func (r *PostgresAlarmDeviceRepository) ListAlarmDevices(ctx context.Context, tenantID string, page, size int) ([]*domain.AlarmDevice, int, error) {
	if tenantID == "" {
		return []*domain.AlarmDevice{}, 0, nil
	}

	// 查询总数
	queryCount := `
		SELECT COUNT(*)
		FROM alarm_device
		WHERE tenant_id = $1
	`
	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count alarm devices: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	// 查询列表
	query := `
		SELECT 
			device_id::text,
			tenant_id::text,
			monitor_config,
			vendor_config,
			metadata
		FROM alarm_device
		WHERE tenant_id = $1
		ORDER BY device_id
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list alarm devices: %w", err)
	}
	defer rows.Close()

	var alarmDevices []*domain.AlarmDevice
	for rows.Next() {
		var alarmDevice domain.AlarmDevice
		var monitorConfig, vendorConfig, metadata sql.NullString

		if err := rows.Scan(
			&alarmDevice.DeviceID,
			&alarmDevice.TenantID,
			&monitorConfig,
			&vendorConfig,
			&metadata,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan alarm device: %w", err)
		}

		if monitorConfig.Valid {
			alarmDevice.MonitorConfig = []byte(monitorConfig.String)
		}
		if vendorConfig.Valid {
			alarmDevice.VendorConfig = []byte(vendorConfig.String)
		}
		if metadata.Valid {
			alarmDevice.Metadata = []byte(metadata.String)
		}

		alarmDevices = append(alarmDevices, &alarmDevice)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate alarm devices: %w", err)
	}

	return alarmDevices, total, nil
}

