package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"wisefido-data/internal/domain"
)

// PostgresAlarmCloudRepository 云端告警策略 Repository — owl_v2 spatial_config 版。
//
// v2 schema 中 alarm_cloud / alarm_device 两张表已合并进 spatial_config(longest-prefix-match KV)。
// 本 repo 保留 interface 不变（GetAlarmCloud/UpsertAlarmCloud/...），实现改为读写
// spatial_config 单行 JSONB blob：
//
//	spatial_prefix = tenant_id (INET CIDR, e.g. 'fd00:0:3::/48')
//	config_key     = alarmCloudConfigKey
//	config_value   = packed JSONB (含 offlinealarm/lowbattery/.../metadata)
//
// 单 JSONB 方案理由：该 UI（AlarmCloud.vue）定位为 tenant 首次批量设置，配置体小、整体读写、
// 无字段级粒度需求；将来若需 branch/device 层覆盖，可在更细 prefix 上再插行（spatial_config
// longest-prefix-match 自带覆盖语义），但本期不实现。
type PostgresAlarmCloudRepository struct {
	db *sql.DB
}

const alarmCloudConfigKey = "alarm.cloud_config"

// alarmCloudPacked 是 domain.AlarmCloud 的 JSONB 序列化壳（显式 json tag 防字段重命名漂移）。
type alarmCloudPacked struct {
	OfflineAlarm      string          `json:"offlinealarm,omitempty"`
	LowBattery        string          `json:"lowbattery,omitempty"`
	DeviceFailure     string          `json:"devicefailure,omitempty"`
	DeviceAlarms      json.RawMessage `json:"device_alarms,omitempty"`
	Conditions        json.RawMessage `json:"conditions,omitempty"`
	NotificationRules json.RawMessage `json:"notification_rules,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

func packAlarmCloud(a *domain.AlarmCloud) ([]byte, error) {
	return json.Marshal(alarmCloudPacked{
		OfflineAlarm:      a.OfflineAlarm,
		LowBattery:        a.LowBattery,
		DeviceFailure:     a.DeviceFailure,
		DeviceAlarms:      a.DeviceAlarms,
		Conditions:        a.Conditions,
		NotificationRules: a.NotificationRules,
		Metadata:          a.Metadata,
	})
}

func unpackAlarmCloud(jsonBytes []byte, tenantID string) (*domain.AlarmCloud, error) {
	var p alarmCloudPacked
	if err := json.Unmarshal(jsonBytes, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal alarm_cloud JSONB: %w", err)
	}
	return &domain.AlarmCloud{
		TenantID:          tenantID,
		OfflineAlarm:      p.OfflineAlarm,
		LowBattery:        p.LowBattery,
		DeviceFailure:     p.DeviceFailure,
		DeviceAlarms:      p.DeviceAlarms,
		Conditions:        p.Conditions,
		NotificationRules: p.NotificationRules,
		Metadata:          p.Metadata,
	}, nil
}

func NewPostgresAlarmCloudRepository(db *sql.DB) *PostgresAlarmCloudRepository {
	return &PostgresAlarmCloudRepository{db: db}
}

var _ AlarmCloudRepository = (*PostgresAlarmCloudRepository)(nil)

// GetAlarmCloud 读 spatial_config 中 tenant 级 alarm 配置。
func (r *PostgresAlarmCloudRepository) GetAlarmCloud(ctx context.Context, tenantID string) (*domain.AlarmCloud, error) {
	if tenantID == "" {
		return nil, sql.ErrNoRows
	}

	var configValue []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT config_value
		  FROM spatial_config
		 WHERE spatial_prefix = $1::inet
		   AND config_key = $2
	`, tenantID, alarmCloudConfigKey).Scan(&configValue)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alarm cloud not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get alarm cloud: %w", err)
	}

	return unpackAlarmCloud(configValue, tenantID)
}

// UpsertAlarmCloud 写 spatial_config 中 tenant 级 alarm 配置。
// source='manual_ui' 表示来自 admin UI；updated_by 由 service 层带；本 repo 接口无 userID 参数。
func (r *PostgresAlarmCloudRepository) UpsertAlarmCloud(ctx context.Context, tenantID string, alarmCloud *domain.AlarmCloud) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	payload, err := packAlarmCloud(alarmCloud)
	if err != nil {
		return fmt.Errorf("failed to pack alarm_cloud: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO spatial_config (
			spatial_prefix, config_key, config_value,
			source, state_db, last_synced_at, updated_at
		) VALUES (
			$1::inet, $2, $3::jsonb,
			'manual_ui', 3, now(), now()
		)
		ON CONFLICT (spatial_prefix, config_key) DO UPDATE SET
			config_value   = EXCLUDED.config_value,
			source         = EXCLUDED.source,
			state_db       = EXCLUDED.state_db,
			last_synced_at = EXCLUDED.last_synced_at,
			updated_at     = EXCLUDED.updated_at
	`, tenantID, alarmCloudConfigKey, string(payload))
	if err != nil {
		return fmt.Errorf("failed to upsert alarm cloud: %w", err)
	}
	return nil
}

// GetSystemAlarmCloud 系统默认模板（v2 不再维护此模板）。
// 返回 ErrNoRows，让上层走 owl-common/alarm 硬编码默认值。
func (r *PostgresAlarmCloudRepository) GetSystemAlarmCloud(ctx context.Context) (*domain.AlarmCloud, error) {
	return nil, sql.ErrNoRows
}

// DeleteAlarmCloud 删除租户 alarm 配置（删 spatial_config 单行）。
func (r *PostgresAlarmCloudRepository) DeleteAlarmCloud(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM spatial_config
		 WHERE spatial_prefix = $1::inet
		   AND config_key = $2
	`, tenantID, alarmCloudConfigKey)
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
