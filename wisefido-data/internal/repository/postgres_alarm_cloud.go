package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
			offlinealarm,
			lowbattery,
			devicefailure,
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
			offlinealarm,
			lowbattery,
			devicefailure,
			device_alarms,
			conditions,
			notification_rules,
			metadata
		) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, $8::jsonb)
		ON CONFLICT (tenant_id) DO UPDATE SET
			offlinealarm = EXCLUDED.offlinealarm,
			lowbattery = EXCLUDED.lowbattery,
			devicefailure = EXCLUDED.devicefailure,
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

// UpsertAlarmCloudFields 创建或更新租户的告警策略配置（使用更新模型）
func (r *PostgresAlarmCloudRepository) UpsertAlarmCloudFields(ctx context.Context, tenantID string, update *domain.AlarmCloudUpdate) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if update == nil {
		return fmt.Errorf("update is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 检查记录是否存在
	var exists bool
	err = tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM alarm_cloud WHERE tenant_id = $1)`,
		tenantID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check existing record: %w", err)
	}

	if exists {
		// UPDATE 模式：只更新提供的字段
		updates := []string{}
		args := []any{tenantID}
		argIdx := 2

		// 处理 UpdateString
		if update.OfflineAlarm != nil {
			switch update.OfflineAlarm.Action {
			case domain.UpdateActionUpdate:
				updates = append(updates, fmt.Sprintf("offlinealarm = $%d", argIdx))
				args = append(args, update.OfflineAlarm.Value)
				argIdx++
			case domain.UpdateActionDelete:
				updates = append(updates, "offlinealarm = NULL")
			case domain.UpdateActionKeep:
				// 不更新，跳过
			}
		}

		if update.LowBattery != nil {
			switch update.LowBattery.Action {
			case domain.UpdateActionUpdate:
				updates = append(updates, fmt.Sprintf("lowbattery = $%d", argIdx))
				args = append(args, update.LowBattery.Value)
				argIdx++
			case domain.UpdateActionDelete:
				updates = append(updates, "lowbattery = NULL")
			case domain.UpdateActionKeep:
				// 不更新，跳过
			}
		}

		if update.DeviceFailure != nil {
			switch update.DeviceFailure.Action {
			case domain.UpdateActionUpdate:
				updates = append(updates, fmt.Sprintf("devicefailure = $%d", argIdx))
				args = append(args, update.DeviceFailure.Value)
				argIdx++
			case domain.UpdateActionDelete:
				updates = append(updates, "devicefailure = NULL")
			case domain.UpdateActionKeep:
				// 不更新，跳过
			}
		}

		// 处理 UpdateJSON
		if update.DeviceAlarms != nil {
			switch update.DeviceAlarms.Action {
			case domain.UpdateActionUpdate:
				if len(update.DeviceAlarms.Value) > 0 {
					updates = append(updates, fmt.Sprintf("device_alarms = $%d::jsonb", argIdx))
					args = append(args, string(update.DeviceAlarms.Value))
					argIdx++
				} else {
					updates = append(updates, "device_alarms = '{}'::jsonb")
				}
			case domain.UpdateActionDelete:
				updates = append(updates, "device_alarms = NULL")
			case domain.UpdateActionKeep:
				// 不更新，跳过
			}
		}

		if update.Conditions != nil {
			switch update.Conditions.Action {
			case domain.UpdateActionUpdate:
				if len(update.Conditions.Value) > 0 {
					updates = append(updates, fmt.Sprintf("conditions = $%d::jsonb", argIdx))
					args = append(args, string(update.Conditions.Value))
					argIdx++
				} else {
					updates = append(updates, "conditions = NULL")
				}
			case domain.UpdateActionDelete:
				updates = append(updates, "conditions = NULL")
			case domain.UpdateActionKeep:
				// 不更新，跳过
			}
		}

		if update.NotificationRules != nil {
			switch update.NotificationRules.Action {
			case domain.UpdateActionUpdate:
				if len(update.NotificationRules.Value) > 0 {
					updates = append(updates, fmt.Sprintf("notification_rules = $%d::jsonb", argIdx))
					args = append(args, string(update.NotificationRules.Value))
					argIdx++
				} else {
					updates = append(updates, "notification_rules = NULL")
				}
			case domain.UpdateActionDelete:
				updates = append(updates, "notification_rules = NULL")
			case domain.UpdateActionKeep:
				// 不更新，跳过
			}
		}

		if update.Metadata != nil {
			switch update.Metadata.Action {
			case domain.UpdateActionUpdate:
				if len(update.Metadata.Value) > 0 {
					updates = append(updates, fmt.Sprintf("metadata = $%d::jsonb", argIdx))
					args = append(args, string(update.Metadata.Value))
					argIdx++
				} else {
					updates = append(updates, "metadata = NULL")
				}
			case domain.UpdateActionDelete:
				updates = append(updates, "metadata = NULL")
			case domain.UpdateActionKeep:
				// 不更新，跳过
			}
		}

		if len(updates) > 0 {
			query := fmt.Sprintf(`
				UPDATE alarm_cloud
				SET %s
				WHERE tenant_id = $1
			`, strings.Join(updates, ", "))

			_, err = tx.ExecContext(ctx, query, args...)
			if err != nil {
				return fmt.Errorf("failed to update alarm cloud: %w", err)
			}
		}
	} else {
		// INSERT 模式：必须提供所有 NOT NULL 字段
		// device_alarms 有默认值 '{}'::jsonb，所以可以为空
		query := `
			INSERT INTO alarm_cloud (
				tenant_id,
				offlinealarm,
				lowbattery,
				devicefailure,
				device_alarms,
				conditions,
				notification_rules,
				metadata
			) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, $8::jsonb)
		`

		var offlineAlarm, lowBattery, deviceFailure interface{}
		if update.OfflineAlarm != nil && update.OfflineAlarm.Action == domain.UpdateActionUpdate {
			offlineAlarm = update.OfflineAlarm.Value
		}
		if update.LowBattery != nil && update.LowBattery.Action == domain.UpdateActionUpdate {
			lowBattery = update.LowBattery.Value
		}
		if update.DeviceFailure != nil && update.DeviceFailure.Action == domain.UpdateActionUpdate {
			deviceFailure = update.DeviceFailure.Value
		}

		var deviceAlarms interface{} = "{}"
		if update.DeviceAlarms != nil && update.DeviceAlarms.Action == domain.UpdateActionUpdate {
			if len(update.DeviceAlarms.Value) > 0 {
				deviceAlarms = string(update.DeviceAlarms.Value)
			}
		}

		var conditions, notificationRules, metadata interface{}
		if update.Conditions != nil && update.Conditions.Action == domain.UpdateActionUpdate && len(update.Conditions.Value) > 0 {
			conditions = string(update.Conditions.Value)
		}
		if update.NotificationRules != nil && update.NotificationRules.Action == domain.UpdateActionUpdate && len(update.NotificationRules.Value) > 0 {
			notificationRules = string(update.NotificationRules.Value)
		}
		if update.Metadata != nil && update.Metadata.Action == domain.UpdateActionUpdate && len(update.Metadata.Value) > 0 {
			metadata = string(update.Metadata.Value)
		}

		_, err = tx.ExecContext(ctx, query, tenantID, offlineAlarm, lowBattery, deviceFailure,
			deviceAlarms, conditions, notificationRules, metadata)
		if err != nil {
			return fmt.Errorf("failed to insert alarm cloud: %w", err)
		}
	}

	return tx.Commit()
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

