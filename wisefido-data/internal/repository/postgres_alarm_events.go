package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"
)

// PostgresAlarmEventsRepository 报警事件Repository实现
// 注意：与 wisefido-sensor 的实现保持一致，但使用 wisefido-data 的 domain 模型
type PostgresAlarmEventsRepository struct {
	db *sql.DB
}

// NewPostgresAlarmEventsRepository 创建报警事件Repository
func NewPostgresAlarmEventsRepository(db *sql.DB) *PostgresAlarmEventsRepository {
	return &PostgresAlarmEventsRepository{db: db}
}

// 确保实现了接口
var _ AlarmEventsRepository = (*PostgresAlarmEventsRepository)(nil)

// buildWhereClause v2：构建 WHERE 子句。
//
// v2 重大变化：
//   - 删除 ae.tenant_id 列 → 改用 ae.device_addr <<= tenant_prefix（/48 INET CIDR）
//   - 删除 ae.metadata JSONB 列 → 无软删，alarm_events 是不可变事件日志
//   - 删除 ae.unit_id / ae.room_id 列 → 用 snapshot ae.unit_name / ae.room_name 过滤，或用 device_addr prefix-match
func (r *PostgresAlarmEventsRepository) buildWhereClause(tenantID string, filters AlarmEventFilters, args *[]interface{}, argN *int) []string {
	// v2: tenant prefix matching（tenantID 期望 v2 host 字符串 "fd00:0:T::" 或 CIDR "fd00:0:T::/48"）
	tenantPrefix := tenantID
	if !strings.Contains(tenantPrefix, "/") && strings.Contains(tenantPrefix, ":") {
		tenantPrefix = tenantPrefix + "/48"
	}
	where := []string{fmt.Sprintf("ae.device_addr <<= $%d::INET", *argN)}
	*args = append(*args, tenantPrefix)
	*argN++

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

	// 5W 过滤改用 snapshot 名字列（v2 ae.unit_name / ae.room_name 派生时锁住）
	if filters.EventUnitID != nil {
		// EventUnitID 入参语义已变：v2 期望 unit_name 而非 UUID
		where = append(where, fmt.Sprintf("ae.unit_name = $%d", *argN))
		*args = append(*args, *filters.EventUnitID)
		*argN++
	}
	if filters.EventRoomID != nil {
		where = append(where, fmt.Sprintf("ae.room_name = $%d", *argN))
		*args = append(*args, *filters.EventRoomID)
		*argN++
	}

	// 设备过滤（v2 仍存 ae.device_id snapshot 列）
	if filters.DeviceID != nil {
		where = append(where, fmt.Sprintf("ae.device_id = $%d::uuid", *argN))
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
	if len(filters.EventTypes) > 0 {
		placeholders := make([]string, len(filters.EventTypes))
		for i := range filters.EventTypes {
			placeholders[i] = fmt.Sprintf("$%d", *argN)
			*args = append(*args, filters.EventTypes[i])
			*argN++
		}
		where = append(where, fmt.Sprintf("ae.event_type IN (%s)", strings.Join(placeholders, ", ")))
	}
	if filters.Category != nil {
		where = append(where, fmt.Sprintf("ae.category = $%d", *argN))
		*args = append(*args, *filters.Category)
		*argN++
	}
	if len(filters.Categories) > 0 {
		placeholders := make([]string, len(filters.Categories))
		for i := range filters.Categories {
			placeholders[i] = fmt.Sprintf("$%d", *argN)
			*args = append(*args, filters.Categories[i])
			*argN++
		}
		where = append(where, fmt.Sprintf("ae.category IN (%s)", strings.Join(placeholders, ", ")))
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

	// v2：alarm_events 自带 snapshot 名字列（tenant_name/branch_name/unit_name/room_name/bed_name/resident_nickname/device_uid）
	// 不需要 JOIN devices/units/rooms/beds/residents 表过滤（FE filter 都改 snapshot 列）。
	joins := []string{}

	// 设备名称过滤 → v2 用 ae.device_uid snapshot
	if filters.DeviceName != nil {
		where = append(where, fmt.Sprintf("ae.device_uid ILIKE $%d", argN))
		args = append(args, "%"+*filters.DeviceName+"%")
		argN++
	}
	// 住户 ID → v2 ae.resident_id INET snapshot
	if filters.ResidentID != nil {
		where = append(where, fmt.Sprintf("ae.resident_id::text = $%d", argN))
		args = append(args, *filters.ResidentID)
		argN++
	}
	// 分支标签 → v2 ae.branch_name
	if filters.BranchTag != nil {
		where = append(where, fmt.Sprintf("ae.branch_name = $%d", argN))
		args = append(args, *filters.BranchTag)
		argN++
	}
	// 单元 → v2 ae.unit_name
	if filters.UnitID != nil {
		where = append(where, fmt.Sprintf("ae.unit_name = $%d", argN))
		args = append(args, *filters.UnitID)
		argN++
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

	// v2 SELECT bridge：
	//   - tenant_id 列已删 → 用 ae.device_addr 反推 /48 host
	//   - iot_timeseries_id / notified_users / metadata / updated_at 列已删 → NULL/默认占位
	//   - notes → ae.handler_notes (v2 列名)
	//   - trigger_data → ae.payload (v2 列名)
	//   - alarm_level SMALLINT → ::text 适配 domain.AlarmLevel string
	//   - handler UUID → ::text
	query := fmt.Sprintf(`
		SELECT
			ae.event_id::text,
			host(network(set_masklen(ae.device_addr, 48)))    AS tenant_id,
			COALESCE(ae.device_id::text, '')                  AS device_id,
			ae.event_type,
			COALESCE(ae.category, '')                          AS category,
			ae.alarm_level::text                               AS alarm_level,
			ae.alarm_status,
			ae.triggered_at,
			ae.hand_time,
			NULL::bigint                                       AS iot_timeseries_id,
			COALESCE(ae.payload, '{}'::jsonb)                  AS trigger_data,
			ae.handler::text,
			ae.operation,
			ae.handler_notes                                   AS notes,
			'[]'::jsonb                                        AS notified_users,
			'{}'::jsonb                                        AS metadata,
			ae.created_at,
			ae.created_at                                      AS updated_at
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

	// v2 GetAlarmEvent bridge — 与 ListAlarmEvents 同样列映射
	tenantPrefix := tenantID
	if !strings.Contains(tenantPrefix, "/") && strings.Contains(tenantPrefix, ":") {
		tenantPrefix = tenantPrefix + "/48"
	}
	query := `
		SELECT
			ae.event_id::text,
			host(network(set_masklen(ae.device_addr, 48)))    AS tenant_id,
			COALESCE(ae.device_id::text, '')                  AS device_id,
			ae.event_type,
			COALESCE(ae.category, '')                          AS category,
			ae.alarm_level::text                               AS alarm_level,
			ae.alarm_status,
			ae.triggered_at,
			ae.hand_time,
			NULL::bigint                                       AS iot_timeseries_id,
			COALESCE(ae.payload, '{}'::jsonb)                  AS trigger_data,
			ae.handler::text,
			ae.operation,
			ae.handler_notes                                   AS notes,
			'[]'::jsonb                                        AS notified_users,
			'{}'::jsonb                                        AS metadata,
			ae.created_at,
			ae.created_at                                      AS updated_at
		FROM alarm_events ae
		WHERE ae.event_id = $1::uuid
		  AND ae.device_addr <<= $2::INET
	`

	var event domain.AlarmEvent
	var handTimePtr sql.NullTime
	var iotTimeSeriesID sql.NullInt64
	var handler, operation, notes sql.NullString
	var triggerData, notifiedUsers, metadata []byte

	err := r.db.QueryRowContext(ctx, query, eventID, tenantPrefix).Scan(
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
		"alarm_status":      true,
		"hand_time":         true,
		"handler":           true,
		"operation":         true,
		"notes":             true,
		"notified_users":    true,
		"metadata":          true,
		"trigger_data":      true,
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

	// v2: 删 updated_at（v2 alarm_events 不可变 / 无 updated_at 列）

	// 添加 WHERE 条件（v2: tenant_id 列已删，用 device_addr <<= tenant_prefix）
	tenantPrefix := tenantID
	if !strings.Contains(tenantPrefix, "/") && strings.Contains(tenantPrefix, ":") {
		tenantPrefix = tenantPrefix + "/48"
	}
	args = append(args, eventID, tenantPrefix)
	whereClause := fmt.Sprintf("event_id = $%d::uuid AND device_addr <<= $%d::INET", argN, argN+1)

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

	// v2 GetRecentAlarmEvent bridge
	tenantPrefix := tenantID
	if !strings.Contains(tenantPrefix, "/") && strings.Contains(tenantPrefix, ":") {
		tenantPrefix = tenantPrefix + "/48"
	}
	query := `
		SELECT
			ae.event_id::text,
			host(network(set_masklen(ae.device_addr, 48)))    AS tenant_id,
			COALESCE(ae.device_id::text, '')                  AS device_id,
			COALESCE(ae.card_id::text, '')                    AS card_id,
			ae.event_type,
			COALESCE(ae.category, '')                          AS category,
			ae.alarm_level::text                               AS alarm_level,
			ae.alarm_status,
			ae.triggered_at,
			ae.hand_time,
			NULL::bigint                                       AS iot_timeseries_id,
			COALESCE(ae.payload, '{}'::jsonb)                  AS trigger_data,
			ae.handler::text,
			ae.operation,
			ae.handler_notes                                   AS notes,
			'[]'::jsonb                                        AS notified_users,
			'{}'::jsonb                                        AS metadata,
			ae.created_at,
			ae.created_at                                      AS updated_at
		FROM alarm_events ae
		WHERE ae.device_addr <<= $1::INET
		  AND ae.device_id = $2::uuid
		  AND ae.event_type = $3
		  AND ae.triggered_at > $4
		  AND ae.alarm_status = 'active'
		ORDER BY ae.triggered_at DESC
		LIMIT 1
	`

	var event domain.AlarmEvent
	var cardID sql.NullString
	var handTimePtr sql.NullTime
	var iotTimeSeriesID sql.NullInt64
	var handler, operation, notes sql.NullString
	var triggerData, notifiedUsers, metadata []byte

	err := r.db.QueryRowContext(ctx, query, tenantPrefix, deviceID, eventType, thresholdTime).Scan(
		&event.EventID,
		&event.TenantID,
		&event.DeviceID,
		&cardID,
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
	if cardID.Valid {
		event.CardID = &cardID.String
	}
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

	// v2：CountAlarmEvents 使用 snapshot 列，不 JOIN devices/rooms/units
	joins := []string{}
	if filters.DeviceName != nil {
		where = append(where, fmt.Sprintf("ae.device_uid ILIKE $%d", argN))
		args = append(args, "%"+*filters.DeviceName+"%")
		argN++
	}
	if filters.ResidentID != nil {
		where = append(where, fmt.Sprintf("ae.resident_id::text = $%d", argN))
		args = append(args, *filters.ResidentID)
		argN++
	}
	if filters.BranchTag != nil {
		where = append(where, fmt.Sprintf("ae.branch_name = $%d", argN))
		args = append(args, *filters.BranchTag)
		argN++
	}
	if filters.UnitID != nil {
		where = append(where, fmt.Sprintf("ae.unit_name = $%d", argN))
		args = append(args, *filters.UnitID)
		argN++
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
