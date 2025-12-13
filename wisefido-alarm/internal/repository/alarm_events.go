package repository

import (
	"database/sql"
	"fmt"

	"wisefido-alarm/internal/models"

	"go.uber.org/zap"
)

// AlarmEventsRepository 报警事件仓库
type AlarmEventsRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewAlarmEventsRepository 创建报警事件仓库
func NewAlarmEventsRepository(db *sql.DB, logger *zap.Logger) *AlarmEventsRepository {
	return &AlarmEventsRepository{
		db:     db,
		logger: logger,
	}
}

// CreateAlarmEvent 创建报警事件
func (r *AlarmEventsRepository) CreateAlarmEvent(event *models.AlarmEvent) error {
	query := `
		INSERT INTO alarm_events (
			event_id,
			tenant_id,
			device_id,
			event_type,
			category,
			alarm_level,
			alarm_status,
			triggered_at,
			hand_time,
			iot_timeseries_id,
			trigger_data,
			handler,
			operation,
			notes,
			notified_users,
			metadata,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	_, err := r.db.Exec(
		query,
		event.EventID,
		event.TenantID,
		event.DeviceID,
		event.EventType,
		event.Category,
		event.AlarmLevel,
		event.AlarmStatus,
		event.TriggeredAt,
		event.HandTime,
		event.IoTTimeSeriesID,
		event.TriggerData,
		event.Handler,
		event.Operation,
		event.Notes,
		event.NotifiedUsers,
		event.Metadata,
		event.CreatedAt,
		event.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create alarm event: %w", err)
	}

	return nil
}

// GetRecentAlarmEvent 获取最近的报警事件（用于去重检查）
// 检查最近 N 分钟内是否已有相同类型的报警
func (r *AlarmEventsRepository) GetRecentAlarmEvent(
	tenantID, deviceID, eventType string,
	withinMinutes int,
) (*models.AlarmEvent, error) {
	query := `
		SELECT 
			event_id,
			tenant_id,
			device_id,
			event_type,
			category,
			alarm_level,
			alarm_status,
			triggered_at,
			hand_time,
			iot_timeseries_id,
			trigger_data,
			handler,
			operation,
			notes,
			notified_users,
			metadata,
			created_at,
			updated_at
		FROM alarm_events
		WHERE tenant_id = $1
		  AND device_id = $2
		  AND event_type = $3
		  AND triggered_at > NOW() - INTERVAL '%d minutes'
		  AND alarm_status = 'active'
		ORDER BY triggered_at DESC
		LIMIT 1
	`

	query = fmt.Sprintf(query, withinMinutes)

	var event models.AlarmEvent
	err := r.db.QueryRow(query, tenantID, deviceID, eventType).Scan(
		&event.EventID,
		&event.TenantID,
		&event.DeviceID,
		&event.EventType,
		&event.Category,
		&event.AlarmLevel,
		&event.AlarmStatus,
		&event.TriggeredAt,
		&event.HandTime,
		&event.IoTTimeSeriesID,
		&event.TriggerData,
		&event.Handler,
		&event.Operation,
		&event.Notes,
		&event.NotifiedUsers,
		&event.Metadata,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到最近的报警事件
		}
		return nil, fmt.Errorf("failed to query recent alarm event: %w", err)
	}

	return &event, nil
}

// UpdateAlarmEvent 更新报警事件（用于延长持续时间等）
// 目前简化实现，只更新 updated_at
func (r *AlarmEventsRepository) UpdateAlarmEvent(eventID string) error {
	query := `
		UPDATE alarm_events
		SET updated_at = CURRENT_TIMESTAMP
		WHERE event_id = $1
	`

	_, err := r.db.Exec(query, eventID)
	if err != nil {
		return fmt.Errorf("failed to update alarm event: %w", err)
	}

	return nil
}
