package repository

import (
	"context"
	"database/sql"
	"fmt"

	"wisefido-data/internal/domain"
)

// PostgresAlarmCloudRepository 云端告警策略Repository实现（强类型版本）
type PostgresAlarmCloudRepository struct {
	db *sql.DB
}

// NewPostgresAlarmCloudRepository 创建云端告警策略Repository
func NewPostgresAlarmCloudRepository(db *sql.DB) *PostgresAlarmCloudRepository {
	return &PostgresAlarmCloudRepository{db: db}
}

// 确保实现了接口
var _ AlarmCloudRepository = (*PostgresAlarmCloudRepository)(nil)

// SystemTenantID 系统租户ID（用于系统默认模板）
const SystemTenantID = "00000000-0000-0000-0000-000000000001"

// GetAlarmCloud 获取租户的告警策略配置
func (r *PostgresAlarmCloudRepository) GetAlarmCloud(ctx context.Context, tenantID string) (*domain.AlarmCloud, error) {
	if tenantID == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT 
			tenant_id::text,
			OfflineAlarm,
			LowBattery,
			DeviceFailure,
			device_alarms,
			conditions,
			notification_rules,
			metadata
		FROM alarm_cloud
		WHERE tenant_id = $1
	`

	var alarmCloud domain.AlarmCloud
	var offlineAlarm, lowBattery, deviceFailure sql.NullString
	var deviceAlarms, conditions, notificationRules, metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&alarmCloud.TenantID,
		&offlineAlarm,
		&lowBattery,
		&deviceFailure,
		&deviceAlarms,
		&conditions,
		&notificationRules,
		&metadata,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alarm cloud not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get alarm cloud: %w", err)
	}

	if offlineAlarm.Valid {
		alarmCloud.OfflineAlarm = offlineAlarm.String
	}
	if lowBattery.Valid {
		alarmCloud.LowBattery = lowBattery.String
	}
	if deviceFailure.Valid {
		alarmCloud.DeviceFailure = deviceFailure.String
	}
	if deviceAlarms.Valid {
		alarmCloud.DeviceAlarms = []byte(deviceAlarms.String)
	}
	if conditions.Valid {
		alarmCloud.Conditions = []byte(conditions.String)
	}
	if notificationRules.Valid {
		alarmCloud.NotificationRules = []byte(notificationRules.String)
	}
	if metadata.Valid {
		alarmCloud.Metadata = []byte(metadata.String)
	}

	return &alarmCloud, nil
}

// UpsertAlarmCloud 创建或更新租户的告警策略配置
func (r *PostgresAlarmCloudRepository) UpsertAlarmCloud(ctx context.Context, tenantID string, alarmCloud *domain.AlarmCloud) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	query := `
		INSERT INTO alarm_cloud (
			tenant_id,
			OfflineAlarm,
			LowBattery,
			DeviceFailure,
			device_alarms,
			conditions,
			notification_rules,
			metadata
		) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, $8::jsonb)
		ON CONFLICT (tenant_id) DO UPDATE SET
			OfflineAlarm = EXCLUDED.OfflineAlarm,
			LowBattery = EXCLUDED.LowBattery,
			DeviceFailure = EXCLUDED.DeviceFailure,
			device_alarms = EXCLUDED.device_alarms,
			conditions = EXCLUDED.conditions,
			notification_rules = EXCLUDED.notification_rules,
			metadata = EXCLUDED.metadata
	`

	var offlineAlarm, lowBattery, deviceFailure interface{}
	if alarmCloud.OfflineAlarm != "" {
		offlineAlarm = alarmCloud.OfflineAlarm
	}
	if alarmCloud.LowBattery != "" {
		lowBattery = alarmCloud.LowBattery
	}
	if alarmCloud.DeviceFailure != "" {
		deviceFailure = alarmCloud.DeviceFailure
	}

	var deviceAlarms, conditions, notificationRules, metadata interface{}
	if len(alarmCloud.DeviceAlarms) > 0 {
		deviceAlarms = string(alarmCloud.DeviceAlarms)
	} else {
		deviceAlarms = "{}"
	}
	if len(alarmCloud.Conditions) > 0 {
		conditions = string(alarmCloud.Conditions)
	}
	if len(alarmCloud.NotificationRules) > 0 {
		notificationRules = string(alarmCloud.NotificationRules)
	}
	if len(alarmCloud.Metadata) > 0 {
		metadata = string(alarmCloud.Metadata)
	}

	_, err := r.db.ExecContext(ctx, query, tenantID, offlineAlarm, lowBattery, deviceFailure,
		deviceAlarms, conditions, notificationRules, metadata)
	if err != nil {
		return fmt.Errorf("failed to upsert alarm cloud: %w", err)
	}

	return nil
}

// GetSystemAlarmCloud 获取系统默认告警策略模板
func (r *PostgresAlarmCloudRepository) GetSystemAlarmCloud(ctx context.Context) (*domain.AlarmCloud, error) {
	return r.GetAlarmCloud(ctx, SystemTenantID)
}

// DeleteAlarmCloud 删除租户的告警策略配置（回退到系统默认）
func (r *PostgresAlarmCloudRepository) DeleteAlarmCloud(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if tenantID == SystemTenantID {
		return fmt.Errorf("cannot delete system alarm cloud")
	}

	query := `
		DELETE FROM alarm_cloud
		WHERE tenant_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete alarm cloud: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("alarm cloud not found")
	}

	return nil
}

