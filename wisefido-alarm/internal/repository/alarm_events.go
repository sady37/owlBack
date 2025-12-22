package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wisefido-alarm/internal/models"

	"go.uber.org/zap"
)

// AlarmEventsRepository 报警事件仓库
// 遵循"bottom-up"设计原则，使用强规则实现
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

// AlarmEventFilters 报警事件过滤条件
type AlarmEventFilters struct {
	// 时间段过滤
	StartTime *time.Time // 开始时间（triggered_at >= StartTime）
	EndTime   *time.Time // 结束时间（triggered_at <= EndTime）

	// 住户过滤
	ResidentID *string // 住户ID（通过 device_id JOIN devices → beds → residents 获取）

	// 位置过滤
	BranchName *string // 分支名称（通过 device_id JOIN devices → beds/rooms → units → units.branch_name 获取）
	UnitID     *string // 单元ID（通过 device_id JOIN devices → beds/rooms → units 获取）

	// 设备过滤
	DeviceID     *string // 设备ID（直接过滤）
	DeviceName   *string // 设备名称（通过 device_id JOIN devices.device_name 获取，支持模糊匹配）
	DeviceSerial *string // 设备序列号（通过 device_id JOIN devices.serial_number 或 device_store.serial_number 获取，支持模糊匹配）

	// 事件类型和级别过滤
	EventType  *string   // 事件类型
	Category   *string   // 分类（safety, clinical, behavioral, device）
	AlarmLevel *string   // 报警级别
	AlarmLevels []string // 报警级别列表（IN 查询）

	// 状态过滤
	AlarmStatus *string   // 报警状态（active, acknowledged）
	AlarmStatuses []string // 报警状态列表（IN 查询）

	// 操作结果过滤
	Operation *string   // 操作结果
	Operations []string // 操作结果列表（IN 查询）

	// 处理人过滤
	HandlerID *string // 处理人ID
}

// ============================================
// 基础 CRUD 操作
// ============================================

// GetAlarmEvent 根据 event_id 获取单个报警事件（需验证 tenant_id）
func (r *AlarmEventsRepository) GetAlarmEvent(ctx context.Context, tenantID, eventID string) (*models.AlarmEvent, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return nil, fmt.Errorf("event_id is required")
	}

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
		WHERE event_id = $1
		  AND tenant_id = $2
		  AND (metadata->>'deleted_at' IS NULL)
	`

	var event models.AlarmEvent
	var handTimePtr sql.NullTime
	var iotTimeSeriesID sql.NullInt64
	var handler, operation, notes sql.NullString
	var triggerData, notifiedUsers, metadata []byte

	err := r.db.QueryRowContext(ctx, query, eventID, tenantID).Scan(
		&event.EventID,
		&event.TenantID,
		&event.DeviceID,
		&event.EventType,
		&event.Category,
		&event.AlarmLevel,
		&event.AlarmStatus,
		&event.TriggeredAt,
		&handTimePtr,
		&iotTimeSeriesID,
		&triggerData,
		&handler,
		&operation,
		&notes,
		&notifiedUsers,
		&metadata,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alarm event not found: event_id=%s, tenant_id=%s", eventID, tenantID)
		}
		return nil, fmt.Errorf("failed to get alarm event: %w", err)
	}

	// 处理可空字段
	if handTimePtr.Valid {
		event.HandTime = &handTimePtr.Time
	}
	if iotTimeSeriesID.Valid {
		event.IoTTimeSeriesID = &iotTimeSeriesID.Int64
	}
	if handler.Valid {
		event.Handler = &handler.String
	}
	if operation.Valid {
		event.Operation = &operation.String
	}
	if notes.Valid {
		event.Notes = &notes.String
	}

	// 处理 JSONB 字段
	if len(triggerData) > 0 {
		event.TriggerData = triggerData
	} else {
		event.TriggerData = json.RawMessage("{}")
	}
	if len(notifiedUsers) > 0 {
		event.NotifiedUsers = notifiedUsers
	} else {
		event.NotifiedUsers = json.RawMessage("[]")
	}
	if len(metadata) > 0 {
		event.Metadata = metadata
	} else {
		event.Metadata = json.RawMessage("{}")
	}

	return &event, nil
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

// UpdateAlarmEvent 更新报警事件（需验证 tenant_id，支持部分更新）
// updates 是一个 map，包含要更新的字段
func (r *AlarmEventsRepository) UpdateAlarmEvent(ctx context.Context, tenantID, eventID string, updates map[string]interface{}) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if len(updates) == 0 {
		return fmt.Errorf("updates cannot be empty")
	}

	// 构建 SET 子句
	setParts := []string{}
	args := []interface{}{}
	argN := 1

	// 允许更新的字段
	allowedFields := map[string]bool{
		"alarm_status":    true,
		"hand_time":       true,
		"handler":         true,
		"operation":       true,
		"notes":           true,
		"notified_users":  true,
		"metadata":        true,
		"trigger_data":    true,
		"iot_timeseries_id": true,
	}

	for field, value := range updates {
		if !allowedFields[field] {
			return fmt.Errorf("field '%s' is not allowed to update", field)
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argN))
		args = append(args, value)
		argN++
	}

	// 自动更新 updated_at
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	// 添加 WHERE 条件
	args = append(args, eventID, tenantID)
	whereClause := fmt.Sprintf("event_id = $%d AND tenant_id = $%d AND (metadata->>'deleted_at' IS NULL)", argN, argN+1)

	query := fmt.Sprintf(`
		UPDATE alarm_events
		SET %s
		WHERE %s
	`, strings.Join(setParts, ", "), whereClause)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update alarm event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("alarm event not found or already deleted: event_id=%s, tenant_id=%s", eventID, tenantID)
	}

	return nil
}

// DeleteAlarmEvent 软删除报警事件（需验证 tenant_id）
// 使用 metadata 字段标记删除时间
func (r *AlarmEventsRepository) DeleteAlarmEvent(ctx context.Context, tenantID, eventID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return fmt.Errorf("event_id is required")
	}

	// 先获取当前的 metadata
	var currentMetadata json.RawMessage
	err := r.db.QueryRowContext(ctx,
		`SELECT metadata FROM alarm_events WHERE event_id = $1 AND tenant_id = $2`,
		eventID, tenantID,
	).Scan(&currentMetadata)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("alarm event not found: event_id=%s, tenant_id=%s", eventID, tenantID)
		}
		return fmt.Errorf("failed to get alarm event metadata: %w", err)
	}

	// 解析 metadata
	var metadata map[string]interface{}
	if len(currentMetadata) > 0 {
		if err := json.Unmarshal(currentMetadata, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// 设置 deleted_at
	metadata["deleted_at"] = time.Now().Format(time.RFC3339)

	// 序列化 metadata
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// 更新记录
	query := `
		UPDATE alarm_events
		SET metadata = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE event_id = $2
		  AND tenant_id = $3
		  AND (metadata->>'deleted_at' IS NULL)
	`

	result, err := r.db.ExecContext(ctx, query, metadataJSON, eventID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete alarm event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("alarm event not found or already deleted: event_id=%s, tenant_id=%s", eventID, tenantID)
	}

	return nil
}
// ============================================
// 查询操作
// ============================================

// buildWhereClause 构建 WHERE 子句（用于 ListAlarmEvents 等查询方法）
func (r *AlarmEventsRepository) buildWhereClause(tenantID string, filters AlarmEventFilters, args *[]interface{}, argN *int) []string {
	where := []string{fmt.Sprintf("ae.tenant_id = $%d", *argN)}
	*args = append(*args, tenantID)
	*argN++

	// 软删除过滤
	where = append(where, "(ae.metadata->>'deleted_at' IS NULL)")

	// 时间段过滤
	if filters.StartTime != nil {
		where = append(where, fmt.Sprintf("ae.triggered_at >= $%d", *argN))
		*args = append(*args, *filters.StartTime)
		*argN++
	}
	if filters.EndTime != nil {
		where = append(where, fmt.Sprintf("ae.triggered_at <= $%d", *argN))
		*args = append(*args, *filters.EndTime)
		*argN++
	}

	// 设备过滤
	if filters.DeviceID != nil {
		where = append(where, fmt.Sprintf("ae.device_id = $%d", *argN))
		*args = append(*args, *filters.DeviceID)
		*argN++
	}

	// 事件类型和级别过滤
	if filters.EventType != nil {
		where = append(where, fmt.Sprintf("ae.event_type = $%d", *argN))
		*args = append(*args, *filters.EventType)
		*argN++
	}
	if filters.Category != nil {
		where = append(where, fmt.Sprintf("ae.category = $%d", *argN))
		*args = append(*args, *filters.Category)
		*argN++
	}
	if filters.AlarmLevel != nil {
		where = append(where, fmt.Sprintf("ae.alarm_level = $%d", *argN))
		*args = append(*args, *filters.AlarmLevel)
		*argN++
	}
	if len(filters.AlarmLevels) > 0 {
		placeholders := make([]string, len(filters.AlarmLevels))
		for i := range filters.AlarmLevels {
			placeholders[i] = fmt.Sprintf("$%d", *argN)
			*args = append(*args, filters.AlarmLevels[i])
			*argN++
		}
		where = append(where, fmt.Sprintf("ae.alarm_level IN (%s)", strings.Join(placeholders, ", ")))
	}

	// 状态过滤
	if filters.AlarmStatus != nil {
		where = append(where, fmt.Sprintf("ae.alarm_status = $%d", *argN))
		*args = append(*args, *filters.AlarmStatus)
		*argN++
	}
	if len(filters.AlarmStatuses) > 0 {
		placeholders := make([]string, len(filters.AlarmStatuses))
		for i := range filters.AlarmStatuses {
			placeholders[i] = fmt.Sprintf("$%d", *argN)
			*args = append(*args, filters.AlarmStatuses[i])
			*argN++
		}
		where = append(where, fmt.Sprintf("ae.alarm_status IN (%s)", strings.Join(placeholders, ", ")))
	}

	// 操作结果过滤
	if filters.Operation != nil {
		where = append(where, fmt.Sprintf("ae.operation = $%d", *argN))
		*args = append(*args, *filters.Operation)
		*argN++
	}
	if len(filters.Operations) > 0 {
		placeholders := make([]string, len(filters.Operations))
		for i := range filters.Operations {
			placeholders[i] = fmt.Sprintf("$%d", *argN)
			*args = append(*args, filters.Operations[i])
			*argN++
		}
		where = append(where, fmt.Sprintf("ae.operation IN (%s)", strings.Join(placeholders, ", ")))
	}

	// 处理人过滤
	if filters.HandlerID != nil {
		where = append(where, fmt.Sprintf("ae.handler = $%d", *argN))
		*args = append(*args, *filters.HandlerID)
		*argN++
	}

	return where
}

// ListAlarmEvents 列表查询（支持多条件过滤、分页）
func (r *AlarmEventsRepository) ListAlarmEvents(ctx context.Context, tenantID string, filters AlarmEventFilters, page, size int) ([]*models.AlarmEvent, int, error) {
	if tenantID == "" {
		return []*models.AlarmEvent{}, 0, nil
	}

	// 构建基础 WHERE 子句
	args := []interface{}{}
	argN := 1
	where := r.buildWhereClause(tenantID, filters, &args, &argN)

	// 构建 JOIN 子句（如果需要）
	joins := []string{}
	
	// 需要 JOIN devices 表的情况
	needDevicesJoin := filters.DeviceName != nil || filters.DeviceSerial != nil || 
		filters.ResidentID != nil || filters.BranchName != nil || filters.UnitID != nil

	if needDevicesJoin {
		joins = append(joins, "LEFT JOIN devices d ON ae.device_id = d.device_id")
		joins = append(joins, "LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id")

		// 设备名称过滤
		if filters.DeviceName != nil {
			where = append(where, fmt.Sprintf("d.device_name ILIKE $%d", argN))
			args = append(args, "%"+*filters.DeviceName+"%")
			argN++
		}

		// 设备序列号过滤
		if filters.DeviceSerial != nil {
			where = append(where, fmt.Sprintf("(d.serial_number ILIKE $%d OR ds.serial_number ILIKE $%d)", argN, argN))
			args = append(args, "%"+*filters.DeviceSerial+"%", "%"+*filters.DeviceSerial+"%")
			argN += 2
		}
	}

	// 需要 JOIN beds/rooms/units 的情况
	needLocationJoin := filters.ResidentID != nil || filters.BranchName != nil || filters.UnitID != nil

	if needLocationJoin {
		joins = append(joins, `
			LEFT JOIN beds b ON d.bound_bed_id = b.bed_id
			LEFT JOIN rooms r ON (d.bound_room_id = r.room_id OR b.room_id = r.room_id)
			LEFT JOIN units u ON r.unit_id = u.unit_id
		`)

		// 住户过滤
		if filters.ResidentID != nil {
			joins = append(joins, "LEFT JOIN residents res ON (b.bed_id = res.bed_id OR r.room_id = res.room_id OR u.unit_id = res.unit_id)")
			where = append(where, fmt.Sprintf("res.resident_id = $%d", argN))
			args = append(args, *filters.ResidentID)
			argN++
		}

		// 分支标签过滤
		if filters.BranchName != nil {
			where = append(where, fmt.Sprintf("u.branch_name = $%d", argN))
			args = append(args, *filters.BranchName)
			argN++
		}

		// 单元过滤
		if filters.UnitID != nil {
			where = append(where, fmt.Sprintf("u.unit_id = $%d", argN))
			args = append(args, *filters.UnitID)
			argN++
		}
	}

	joinClause := strings.Join(joins, " ")
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// 计算总数
	queryCount := fmt.Sprintf(`
		SELECT COUNT(DISTINCT ae.event_id)
		FROM alarm_events ae
		%s
		%s
	`, joinClause, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count alarm events: %w", err)
	}

	// 分页处理
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	// 查询数据
	query := fmt.Sprintf(`
		SELECT DISTINCT
			ae.event_id,
			ae.tenant_id,
			ae.device_id,
			ae.event_type,
			ae.category,
			ae.alarm_level,
			ae.alarm_status,
			ae.triggered_at,
			ae.hand_time,
			ae.iot_timeseries_id,
			ae.trigger_data,
			ae.handler,
			ae.operation,
			ae.notes,
			ae.notified_users,
			ae.metadata,
			ae.created_at,
			ae.updated_at
		FROM alarm_events ae
		%s
		%s
		ORDER BY ae.triggered_at DESC
		LIMIT $%d OFFSET $%d
	`, joinClause, whereClause, len(args)+1, len(args)+2)

	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query alarm events: %w", err)
	}
	defer rows.Close()

	events := []*models.AlarmEvent{}
	for rows.Next() {
		var event models.AlarmEvent
		var handTimePtr sql.NullTime
		var iotTimeSeriesID sql.NullInt64
		var handler, operation, notes sql.NullString
		var triggerData, notifiedUsers, metadata []byte

		err := rows.Scan(
			&event.EventID,
			&event.TenantID,
			&event.DeviceID,
			&event.EventType,
			&event.Category,
			&event.AlarmLevel,
			&event.AlarmStatus,
			&event.TriggeredAt,
			&handTimePtr,
			&iotTimeSeriesID,
			&triggerData,
			&handler,
			&operation,
			&notes,
			&notifiedUsers,
			&metadata,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan alarm event: %w", err)
		}

		// 处理可空字段
		if handTimePtr.Valid {
			event.HandTime = &handTimePtr.Time
		}
		if iotTimeSeriesID.Valid {
			event.IoTTimeSeriesID = &iotTimeSeriesID.Int64
		}
		if handler.Valid {
			event.Handler = &handler.String
		}
		if operation.Valid {
			event.Operation = &operation.String
		}
		if notes.Valid {
			event.Notes = &notes.String
		}

		// 处理 JSONB 字段
		if len(triggerData) > 0 {
			event.TriggerData = triggerData
		} else {
			event.TriggerData = json.RawMessage("{}")
		}
		if len(notifiedUsers) > 0 {
			event.NotifiedUsers = notifiedUsers
		} else {
			event.NotifiedUsers = json.RawMessage("[]")
		}
		if len(metadata) > 0 {
			event.Metadata = metadata
		} else {
			event.Metadata = json.RawMessage("{}")
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate alarm events: %w", err)
	}

	return events, total, nil
}

// GetRecentAlarmEvent 获取最近的报警事件（用于去重检查，改进版）
// 检查最近 N 分钟内是否已有相同类型的报警
func (r *AlarmEventsRepository) GetRecentAlarmEvent(ctx context.Context, tenantID, deviceID, eventType string, withinMinutes int) (*models.AlarmEvent, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	if eventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}

	// 计算时间阈值（使用参数化查询避免 SQL 注入）
	thresholdTime := time.Now().Add(-time.Duration(withinMinutes) * time.Minute)

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
		  AND triggered_at > $4
		  AND alarm_status = 'active'
		  AND (metadata->>'deleted_at' IS NULL)
		ORDER BY triggered_at DESC
		LIMIT 1
	`

	var event models.AlarmEvent
	var handTimePtr sql.NullTime
	var iotTimeSeriesID sql.NullInt64
	var handler, operation, notes sql.NullString
	var triggerData, notifiedUsers, metadata []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, deviceID, eventType, thresholdTime).Scan(
		&event.EventID,
		&event.TenantID,
		&event.DeviceID,
		&event.EventType,
		&event.Category,
		&event.AlarmLevel,
		&event.AlarmStatus,
		&event.TriggeredAt,
		&handTimePtr,
		&iotTimeSeriesID,
		&triggerData,
		&handler,
		&operation,
		&notes,
		&notifiedUsers,
		&metadata,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到最近的报警事件
		}
		return nil, fmt.Errorf("failed to query recent alarm event: %w", err)
	}

	// 处理可空字段
	if handTimePtr.Valid {
		event.HandTime = &handTimePtr.Time
	}
	if iotTimeSeriesID.Valid {
		event.IoTTimeSeriesID = &iotTimeSeriesID.Int64
	}
	if handler.Valid {
		event.Handler = &handler.String
	}
	if operation.Valid {
		event.Operation = &operation.String
	}
	if notes.Valid {
		event.Notes = &notes.String
	}

	// 处理 JSONB 字段
	if len(triggerData) > 0 {
		event.TriggerData = triggerData
	} else {
		event.TriggerData = json.RawMessage("{}")
	}
	if len(notifiedUsers) > 0 {
		event.NotifiedUsers = notifiedUsers
	} else {
		event.NotifiedUsers = json.RawMessage("[]")
	}
	if len(metadata) > 0 {
		event.Metadata = metadata
	} else {
		event.Metadata = json.RawMessage("{}")
	}

	return &event, nil
}

// ============================================
// 状态管理
// ============================================

// AcknowledgeAlarmEvent 确认报警（更新状态为 acknowledged，设置 hand_time 和 handler）
func (r *AlarmEventsRepository) AcknowledgeAlarmEvent(ctx context.Context, tenantID, eventID, handlerID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if handlerID == "" {
		return fmt.Errorf("handler_id is required")
	}

	updates := map[string]interface{}{
		"alarm_status": "acknowledged",
		"hand_time":    time.Now(),
		"handler":      handlerID,
	}

	return r.UpdateAlarmEvent(ctx, tenantID, eventID, updates)
}

// UpdateAlarmEventOperation 更新操作结果（verified_and_processed, false_alarm, test, auto_relieved）
func (r *AlarmEventsRepository) UpdateAlarmEventOperation(ctx context.Context, tenantID, eventID, operation, handlerID string, notes *string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if operation == "" {
		return fmt.Errorf("operation is required")
	}

	// 验证 operation 值
	validOperations := map[string]bool{
		"verified_and_processed": true,
		"false_alarm":            true,
		"test":                   true,
		"auto_relieved":          true,
	}
	if !validOperations[operation] {
		return fmt.Errorf("invalid operation: %s", operation)
	}

	updates := map[string]interface{}{
		"operation": operation,
	}

	if handlerID != "" {
		updates["handler"] = handlerID
		updates["hand_time"] = time.Now()
	}

	if notes != nil {
		updates["notes"] = *notes
	}

	return r.UpdateAlarmEvent(ctx, tenantID, eventID, updates)
}

// ============================================
// 统计查询
// ============================================

// CountAlarmEvents 统计报警事件数量（按条件）
func (r *AlarmEventsRepository) CountAlarmEvents(ctx context.Context, tenantID string, filters AlarmEventFilters) (int, error) {
	if tenantID == "" {
		return 0, nil
	}

	// 构建基础 WHERE 子句
	args := []interface{}{}
	argN := 1
	where := r.buildWhereClause(tenantID, filters, &args, &argN)

	// 构建 JOIN 子句（如果需要）
	joins := []string{}
	needDevicesJoin := filters.DeviceName != nil || filters.DeviceSerial != nil || 
		filters.ResidentID != nil || filters.BranchName != nil || filters.UnitID != nil

	if needDevicesJoin {
		joins = append(joins, "LEFT JOIN devices d ON ae.device_id = d.device_id")
		joins = append(joins, "LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id")

		if filters.DeviceName != nil {
			where = append(where, fmt.Sprintf("d.device_name ILIKE $%d", argN))
			args = append(args, "%"+*filters.DeviceName+"%")
			argN++
		}
		if filters.DeviceSerial != nil {
			where = append(where, fmt.Sprintf("(d.serial_number ILIKE $%d OR ds.serial_number ILIKE $%d)", argN, argN))
			args = append(args, "%"+*filters.DeviceSerial+"%", "%"+*filters.DeviceSerial+"%")
			argN += 2
		}
	}

	needLocationJoin := filters.ResidentID != nil || filters.BranchName != nil || filters.UnitID != nil
	if needLocationJoin {
		joins = append(joins, `
			LEFT JOIN beds b ON d.bound_bed_id = b.bed_id
			LEFT JOIN rooms r ON (d.bound_room_id = r.room_id OR b.room_id = r.room_id)
			LEFT JOIN units u ON r.unit_id = u.unit_id
		`)

		if filters.ResidentID != nil {
			joins = append(joins, "LEFT JOIN residents res ON (b.bed_id = res.bed_id OR r.room_id = res.room_id OR u.unit_id = res.unit_id)")
			where = append(where, fmt.Sprintf("res.resident_id = $%d", argN))
			args = append(args, *filters.ResidentID)
			argN++
		}
		if filters.BranchName != nil {
			where = append(where, fmt.Sprintf("u.branch_name = $%d", argN))
			args = append(args, *filters.BranchName)
			argN++
		}
		if filters.UnitID != nil {
			where = append(where, fmt.Sprintf("u.unit_id = $%d", argN))
			args = append(args, *filters.UnitID)
			argN++
		}
	}

	joinClause := strings.Join(joins, " ")
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT COUNT(DISTINCT ae.event_id)
		FROM alarm_events ae
		%s
		%s
	`, joinClause, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count alarm events: %w", err)
	}

	return total, nil
}

// GetAlarmEventsByDevice 获取设备的报警事件列表
func (r *AlarmEventsRepository) GetAlarmEventsByDevice(ctx context.Context, tenantID, deviceID string, filters AlarmEventFilters, page, size int) ([]*models.AlarmEvent, int, error) {
	filters.DeviceID = &deviceID
	return r.ListAlarmEvents(ctx, tenantID, filters, page, size)
}

// GetAlarmEventsByCategory 按分类查询
func (r *AlarmEventsRepository) GetAlarmEventsByCategory(ctx context.Context, tenantID, category string, filters AlarmEventFilters, page, size int) ([]*models.AlarmEvent, int, error) {
	filters.Category = &category
	return r.ListAlarmEvents(ctx, tenantID, filters, page, size)
}

// GetAlarmEventsByLevel 按级别查询
func (r *AlarmEventsRepository) GetAlarmEventsByLevel(ctx context.Context, tenantID, alarmLevel string, filters AlarmEventFilters, page, size int) ([]*models.AlarmEvent, int, error) {
	filters.AlarmLevel = &alarmLevel
	return r.ListAlarmEvents(ctx, tenantID, filters, page, size)
}

// ============================================
// 高级功能
// ============================================

// GetActiveAlarmEvents 获取待处理的报警（alarm_status = 'active'，排除 INFO/DEBUG 级别）
func (r *AlarmEventsRepository) GetActiveAlarmEvents(ctx context.Context, tenantID string, filters AlarmEventFilters, page, size int) ([]*models.AlarmEvent, int, error) {
	activeStatus := "active"
	filters.AlarmStatus = &activeStatus
	filters.AlarmLevels = []string{"0", "1", "2", "3", "4", "5", "EMERG", "ALERT", "CRIT", "ERR", "WARNING", "NOTICE"}
	return r.ListAlarmEvents(ctx, tenantID, filters, page, size)
}

// GetInformationalEvents 获取信息性事件（INFO(6) 和 DEBUG(7) 级别）
func (r *AlarmEventsRepository) GetInformationalEvents(ctx context.Context, tenantID string, filters AlarmEventFilters, page, size int) ([]*models.AlarmEvent, int, error) {
	filters.AlarmLevels = []string{"6", "7", "INFO", "DEBUG"}
	return r.ListAlarmEvents(ctx, tenantID, filters, page, size)
}

