package repository

import (
	"context"
	"database/sql"
	"fmt"
	"wisefido-card-aggregator/internal/models"

	"go.uber.org/zap"
)

// AlarmEventsRepository 报警事件仓库（简化版，只用于创建报警事件）
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

// CreateAlarmEvent 创建报警事件（需验证 tenant_id）
func (r *AlarmEventsRepository) CreateAlarmEvent(ctx context.Context, tenantID string, event *models.AlarmEvent) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if event == nil {
		return fmt.Errorf("event is required")
	}
	if event.TenantID != tenantID {
		return fmt.Errorf("event.tenant_id must match tenant_id parameter")
	}

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

	_, err := r.db.ExecContext(ctx,
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

