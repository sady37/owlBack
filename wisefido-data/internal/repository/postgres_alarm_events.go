package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"wisefido-data/internal/domain"
)

// PostgresAlarmEventsRepository 报警事件Repository实现
// 注意：与 wisefido-alarm 的实现保持一致，但使用 wisefido-data 的 domain 模型
type PostgresAlarmEventsRepository struct {
	db *sql.DB
}

// NewPostgresAlarmEventsRepository 创建报警事件Repository
func NewPostgresAlarmEventsRepository(db *sql.DB) *PostgresAlarmEventsRepository {
	return &PostgresAlarmEventsRepository{db: db}
}

// 确保实现了接口
var _ AlarmEventsRepository = (*PostgresAlarmEventsRepository)(nil)

// buildWhereClause 构建 WHERE 子句（用于 ListAlarmEvents 等查询方法）
func (r *PostgresAlarmEventsRepository) buildWhereClause(tenantID string, filters AlarmEventFilters, args *[]interface{}, argN *int) []string {
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
	if len(filters.DeviceIDs) > 0 {
		placeholders := make([]string, len(filters.DeviceIDs))
		for i := range filters.DeviceIDs {
			placeholders[i] = fmt.Sprintf("$%d", *argN)
			*args = append(*args, filters.DeviceIDs[i])
			*argN++
		}
		where = append(where, fmt.Sprintf("ae.device_id IN (%s)", strings.Join(placeholders, ", ")))
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
// 注意：此方法需要支持跨表 JOIN 查询关联数据（设备、卡片、住户、地址信息）
func (r *PostgresAlarmEventsRepository) ListAlarmEvents(ctx context.Context, tenantID string, filters AlarmEventFilters, page, size int) ([]*domain.AlarmEvent, int, error) {
	if tenantID == "" {
		return []*domain.AlarmEvent{}, 0, nil
	}

	// 构建基础 WHERE 子句
	args := []interface{}{}
	argN := 1
	where := r.buildWhereClause(tenantID, filters, &args, &argN)

	// 构建 JOIN 子句（如果需要）
	joins := []string{}

	// 需要 JOIN devices 表的情况
	needDevicesJoin := filters.DeviceName != nil ||
		filters.ResidentID != nil || filters.BranchTag != nil || filters.UnitID != nil

	if needDevicesJoin {
		joins = append(joins, "LEFT JOIN devices d ON ae.device_id = d.device_id")

		// 设备名称过滤
		if filters.DeviceName != nil {
			where = append(where, fmt.Sprintf("d.device_name ILIKE $%d", argN))
			args = append(args, "%"+*filters.DeviceName+"%")
			argN++
		}
	}

	// 需要 JOIN beds/rooms/units 的情况
	needLocationJoin := filters.ResidentID != nil || filters.BranchTag != nil || filters.UnitID != nil

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
		if filters.BranchTag != nil {
			where = append(where, fmt.Sprintf("u.branch_tag = $%d", argN))
			args = append(args, *filters.BranchTag)
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
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	// 查询数据
	query := fmt.Sprintf(`
		SELECT DISTINCT
			ae.event_id::text,
			ae.tenant_id::text,
			ae.device_id::text,
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

	events := []*domain.AlarmEvent{}
	for rows.Next() {
		var event domain.AlarmEvent
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

		// 处理 JSONB 字段（统一使用 json.RawMessage）
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

// GetAlarmEvent 根据 event_id 获取单个报警事件（需验证 tenant_id）
func (r *PostgresAlarmEventsRepository) GetAlarmEvent(ctx context.Context, tenantID, eventID string) (*domain.AlarmEvent, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return nil, fmt.Errorf("event_id is required")
	}

	query := `
		SELECT 
			event_id::text,
			tenant_id::text,
			device_id::text,
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

	var event domain.AlarmEvent
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

	// 处理 JSONB 字段（统一使用 json.RawMessage）
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
func (r *PostgresAlarmEventsRepository) CreateAlarmEvent(ctx context.Context, tenantID string, event *domain.AlarmEvent) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if event == nil {
		return fmt.Errorf("event is required")
	}
	if event.TenantID != tenantID {
		return fmt.Errorf("event.tenant_id must match tenant_id parameter")
	}

	// 设置默认值
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.AlarmStatus == "" {
		event.AlarmStatus = "active"
	}
	if event.TriggeredAt.IsZero() {
		event.TriggeredAt = time.Now()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = time.Now()
	}
	if len(event.TriggerData) == 0 {
		event.TriggerData = json.RawMessage("{}")
	}
	if len(event.NotifiedUsers) == 0 {
		event.NotifiedUsers = json.RawMessage("[]")
	}
	if len(event.Metadata) == 0 {
		event.Metadata = json.RawMessage("{}")
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

// AcknowledgeAlarmEvent 确认报警（更新状态为 acknowledged，设置 hand_time 和 handler）
func (r *PostgresAlarmEventsRepository) AcknowledgeAlarmEvent(ctx context.Context, tenantID, eventID, handlerID string) error {
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

// UpdateAlarmEventOperation 更新操作结果（verified_and_processed, false_alarm, test）
func (r *PostgresAlarmEventsRepository) UpdateAlarmEventOperation(ctx context.Context, tenantID, eventID, operation, handlerID string, notes *string) error {
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

// UpdateAlarmEvent 更新报警事件（需验证 tenant_id，支持部分更新）
func (r *PostgresAlarmEventsRepository) UpdateAlarmEvent(ctx context.Context, tenantID, eventID string, updates map[string]interface{}) error {
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
		"alarm_status":     true,
		"hand_time":        true,
		"handler":          true,
		"operation":        true,
		"notes":            true,
		"notified_users":   true,
		"metadata":         true,
		"trigger_data":     true,
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
func (r *PostgresAlarmEventsRepository) DeleteAlarmEvent(ctx context.Context, tenantID, eventID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if eventID == "" {
		return fmt.Errorf("event_id is required")
	}

	// 先获取当前的 metadata
	var currentMetadata []byte
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

// GetRecentAlarmEvent 获取最近的报警事件（用于去重检查，改进版）
// 检查最近 N 分钟内是否已有相同类型的报警
func (r *PostgresAlarmEventsRepository) GetRecentAlarmEvent(ctx context.Context, tenantID, deviceID, eventType string, withinMinutes int) (*domain.AlarmEvent, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	if eventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}

	// 计算时间阈值
	thresholdTime := time.Now().Add(-time.Duration(withinMinutes) * time.Minute)

	query := `
		SELECT 
			event_id::text,
			tenant_id::text,
			device_id::text,
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

	var event domain.AlarmEvent
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

	// 处理 JSONB 字段（统一使用 json.RawMessage）
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

// CountAlarmEvents 统计报警事件数量（按条件）
func (r *PostgresAlarmEventsRepository) CountAlarmEvents(ctx context.Context, tenantID string, filters AlarmEventFilters) (int, error) {
	if tenantID == "" {
		return 0, nil
	}

	// 构建基础 WHERE 子句
	args := []interface{}{}
	argN := 1
	where := r.buildWhereClause(tenantID, filters, &args, &argN)

	// 构建 JOIN 子句（如果需要）
	joins := []string{}
	needDevicesJoin := filters.DeviceName != nil ||
		filters.ResidentID != nil || filters.BranchTag != nil || filters.UnitID != nil

	if needDevicesJoin {
		joins = append(joins, "LEFT JOIN devices d ON ae.device_id = d.device_id")

		if filters.DeviceName != nil {
			where = append(where, fmt.Sprintf("d.device_name ILIKE $%d", argN))
			args = append(args, "%"+*filters.DeviceName+"%")
			argN++
		}
	}

	needLocationJoin := filters.ResidentID != nil || filters.BranchTag != nil || filters.UnitID != nil
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
		if filters.BranchTag != nil {
			where = append(where, fmt.Sprintf("u.branch_tag = $%d", argN))
			args = append(args, *filters.BranchTag)
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

