package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/domain"
)

// PostgresIoTTimeSeriesRepository IoT时序数据Repository实现（强类型版本）
type PostgresIoTTimeSeriesRepository struct {
	db *sql.DB
}

// NewPostgresIoTTimeSeriesRepository 创建IoT时序数据Repository
func NewPostgresIoTTimeSeriesRepository(db *sql.DB) *PostgresIoTTimeSeriesRepository {
	return &PostgresIoTTimeSeriesRepository{db: db}
}

// 确保实现了接口
var _ IoTTimeSeriesRepository = (*PostgresIoTTimeSeriesRepository)(nil)

// buildBaseQuery 构建基础查询（包含必需的 JOIN）
func (r *PostgresIoTTimeSeriesRepository) buildBaseQuery(includeAlarmEvent bool) string {
	baseSelect := `
		SELECT 
			its.id,
			its.tenant_id::text,
			its.device_id::text,
			its.timestamp,
			its.data_type,
			its.category,
			its.tracking_id,
			its.radar_pos_x,
			its.radar_pos_y,
			its.radar_pos_z,
			its.posture_snomed_code,
			its.posture_display,
			its.event_type,
			its.event_snomed_code,
			its.event_display,
			its.area_id,
			its.heart_rate_code,
			its.heart_rate_display,
			its.heart_rate,
			its.respiratory_rate_code,
			its.respiratory_rate_display,
			its.respiratory_rate,
			its.sleep_state_snomed_code,
			its.sleep_state_display,
			its.unit_id::text,
			its.room_id::text,
			its.alarm_event_id::text,
			its.confidence,
			its.remaining_time,
			its.raw_original,
			its.raw_format,
			its.raw_compression,
			its.metadata,
			its.created_at,
			COALESCE(d.serial_number, '') as device_sn,
			COALESCE(d.uid, '') as device_uid,
			COALESCE(ds.firmware_version, '') as firmware_version,
			COALESCE(ds.device_type, '') as device_type
		FROM iot_timeseries its
		LEFT JOIN devices d ON its.device_id = d.device_id
		LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
	`

	if includeAlarmEvent {
		baseSelect += `
		LEFT JOIN alarm_events ae ON its.alarm_event_id = ae.event_id
		`
	}

	return baseSelect
}

// buildWhereClause 构建 WHERE 子句
func (r *PostgresIoTTimeSeriesRepository) buildWhereClause(tenantID string, filters *IoTTimeSeriesFilters, args *[]interface{}, argN *int) string {
	var where []string
	
	if tenantID != "" {
		where = append(where, "its.tenant_id = $"+fmt.Sprintf("%d", *argN))
		*args = append(*args, tenantID)
		*argN++
	}

	if filters != nil {
		if filters.DeviceID != "" {
			where = append(where, "its.device_id = $"+fmt.Sprintf("%d", *argN))
			*args = append(*args, filters.DeviceID)
			*argN++
		}
		if filters.DataType != "" {
			where = append(where, "its.data_type = $"+fmt.Sprintf("%d", *argN))
			*args = append(*args, filters.DataType)
			*argN++
		}
		if filters.Category != "" {
			where = append(where, "its.category = $"+fmt.Sprintf("%d", *argN))
			*args = append(*args, filters.Category)
			*argN++
		}
		if filters.EventType != "" {
			where = append(where, "its.event_type = $"+fmt.Sprintf("%d", *argN))
			*args = append(*args, filters.EventType)
			*argN++
		}
		if filters.UnitID != "" {
			where = append(where, "its.unit_id = $"+fmt.Sprintf("%d", *argN))
			*args = append(*args, filters.UnitID)
			*argN++
		}
		if filters.RoomID != "" {
			where = append(where, "its.room_id = $"+fmt.Sprintf("%d", *argN))
			*args = append(*args, filters.RoomID)
			*argN++
		}
		if filters.StartTime != nil {
			where = append(where, "its.timestamp >= $"+fmt.Sprintf("%d", *argN))
			*args = append(*args, *filters.StartTime)
			*argN++
		}
		if filters.EndTime != nil {
			where = append(where, "its.timestamp <= $"+fmt.Sprintf("%d", *argN))
			*args = append(*args, *filters.EndTime)
			*argN++
		}
	}

	return strings.Join(where, " AND ")
}

// scanIoTTimeSeries 扫描单条记录
func (r *PostgresIoTTimeSeriesRepository) scanIoTTimeSeries(row *sql.Row) (*domain.IoTTimeSeries, error) {
	var ts domain.IoTTimeSeries
	var trackingID, radarPosX, radarPosY, radarPosZ sql.NullInt64
	var postureCode, postureDisplay sql.NullString
	var eventType, eventCode, eventDisplay sql.NullString
	var areaID sql.NullInt64
	var heartRateCode, heartRateDisplay sql.NullString
	var heartRate sql.NullInt64
	var respRateCode, respRateDisplay sql.NullString
	var respRate sql.NullInt64
	var sleepCode, sleepDisplay sql.NullString
	var unitID, roomID, alarmEventID sql.NullString
	var confidence, remainingTime sql.NullInt64
	var rawCompression sql.NullString
	var metadata []byte

	err := row.Scan(
		&ts.ID,
		&ts.TenantID,
		&ts.DeviceID,
		&ts.Timestamp,
		&ts.DataType,
		&ts.Category,
		&trackingID,
		&radarPosX,
		&radarPosY,
		&radarPosZ,
		&postureCode,
		&postureDisplay,
		&eventType,
		&eventCode,
		&eventDisplay,
		&areaID,
		&heartRateCode,
		&heartRateDisplay,
		&heartRate,
		&respRateCode,
		&respRateDisplay,
		&respRate,
		&sleepCode,
		&sleepDisplay,
		&unitID,
		&roomID,
		&alarmEventID,
		&confidence,
		&remainingTime,
		&ts.RawOriginal,
		&ts.RawFormat,
		&rawCompression,
		&metadata,
		&ts.CreatedAt,
		&ts.DeviceSN,
		&ts.DeviceUID,
		&ts.FirmwareVersion,
		&ts.DeviceType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan iot_timeseries: %w", err)
	}

	// 处理 nullable 字段
	if trackingID.Valid {
		tid := int(trackingID.Int64)
		ts.TrackingID = &tid
	}
	if radarPosX.Valid {
		x := int(radarPosX.Int64)
		ts.RadarPosX = &x
	}
	if radarPosY.Valid {
		y := int(radarPosY.Int64)
		ts.RadarPosY = &y
	}
	if radarPosZ.Valid {
		z := int(radarPosZ.Int64)
		ts.RadarPosZ = &z
	}
	if postureCode.Valid {
		ts.PostureSNOMEDCode = postureCode.String
	}
	if postureDisplay.Valid {
		ts.PostureDisplay = postureDisplay.String
	}
	if eventType.Valid {
		ts.EventType = eventType.String
	}
	if eventCode.Valid {
		ts.EventSNOMEDCode = eventCode.String
	}
	if eventDisplay.Valid {
		ts.EventDisplay = eventDisplay.String
	}
	if areaID.Valid {
		aid := int(areaID.Int64)
		ts.AreaID = &aid
	}
	if heartRateCode.Valid {
		ts.HeartRateCode = heartRateCode.String
	}
	if heartRateDisplay.Valid {
		ts.HeartRateDisplay = heartRateDisplay.String
	}
	if heartRate.Valid {
		hr := int(heartRate.Int64)
		ts.HeartRate = &hr
	}
	if respRateCode.Valid {
		ts.RespiratoryRateCode = respRateCode.String
	}
	if respRateDisplay.Valid {
		ts.RespiratoryRateDisplay = respRateDisplay.String
	}
	if respRate.Valid {
		rr := int(respRate.Int64)
		ts.RespiratoryRate = &rr
	}
	if sleepCode.Valid {
		ts.SleepStateSNOMEDCode = sleepCode.String
	}
	if sleepDisplay.Valid {
		ts.SleepStateDisplay = sleepDisplay.String
	}
	if unitID.Valid {
		ts.UnitID = unitID.String
	}
	if roomID.Valid {
		ts.RoomID = roomID.String
	}
	if alarmEventID.Valid {
		ts.AlarmEventID = alarmEventID.String
	}
	if confidence.Valid {
		c := int(confidence.Int64)
		ts.Confidence = &c
	}
	if remainingTime.Valid {
		rt := int(remainingTime.Int64)
		ts.RemainingTime = &rt
	}
	if rawCompression.Valid {
		ts.RawCompression = rawCompression.String
	}

	// 解析 metadata JSONB
	if len(metadata) > 0 {
		// 这里需要解析 JSONB，暂时先留空
		// 可以使用 encoding/json 或 pgx 的 JSONB 处理
		ts.Metadata = make(map[string]interface{})
	}

	return &ts, nil
}

// GetTimeSeriesData 获取时序数据（按ID）
func (r *PostgresIoTTimeSeriesRepository) GetTimeSeriesData(ctx context.Context, id int64) (*domain.IoTTimeSeries, error) {
	query := r.buildBaseQuery(false) + `
		WHERE its.id = $1
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanIoTTimeSeries(row)
}

// GetLatestData 获取最新数据（按设备）
func (r *PostgresIoTTimeSeriesRepository) GetLatestData(ctx context.Context, tenantID, deviceID string, limit int) ([]*domain.IoTTimeSeries, error) {
	if limit <= 0 {
		limit = 20
	}

	query := r.buildBaseQuery(false) + `
		WHERE its.tenant_id = $1 AND its.device_id = $2
		ORDER BY its.timestamp DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest data: %w", err)
	}
	defer rows.Close()

	var results []*domain.IoTTimeSeries
	for rows.Next() {
		// 复用 scanIoTTimeSeries 逻辑，但需要适配 rows.Scan
		var ts domain.IoTTimeSeries
		var trackingID, radarPosX, radarPosY, radarPosZ sql.NullInt64
		var postureCode, postureDisplay sql.NullString
		var eventType, eventCode, eventDisplay sql.NullString
		var areaID sql.NullInt64
		var heartRateCode, heartRateDisplay sql.NullString
		var heartRate sql.NullInt64
		var respRateCode, respRateDisplay sql.NullString
		var respRate sql.NullInt64
		var sleepCode, sleepDisplay sql.NullString
		var unitID, roomID, alarmEventID sql.NullString
		var confidence, remainingTime sql.NullInt64
		var rawCompression sql.NullString
		var metadata []byte

		if err := rows.Scan(
			&ts.ID,
			&ts.TenantID,
			&ts.DeviceID,
			&ts.Timestamp,
			&ts.DataType,
			&ts.Category,
			&trackingID,
			&radarPosX,
			&radarPosY,
			&radarPosZ,
			&postureCode,
			&postureDisplay,
			&eventType,
			&eventCode,
			&eventDisplay,
			&areaID,
			&heartRateCode,
			&heartRateDisplay,
			&heartRate,
			&respRateCode,
			&respRateDisplay,
			&respRate,
			&sleepCode,
			&sleepDisplay,
			&unitID,
			&roomID,
			&alarmEventID,
			&confidence,
			&remainingTime,
			&ts.RawOriginal,
			&ts.RawFormat,
			&rawCompression,
			&metadata,
			&ts.CreatedAt,
			&ts.DeviceSN,
			&ts.DeviceUID,
			&ts.FirmwareVersion,
			&ts.DeviceType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 处理 nullable 字段（复用逻辑）
		if trackingID.Valid {
			tid := int(trackingID.Int64)
			ts.TrackingID = &tid
		}
		if radarPosX.Valid {
			x := int(radarPosX.Int64)
			ts.RadarPosX = &x
		}
		if radarPosY.Valid {
			y := int(radarPosY.Int64)
			ts.RadarPosY = &y
		}
		if radarPosZ.Valid {
			z := int(radarPosZ.Int64)
			ts.RadarPosZ = &z
		}
		if postureCode.Valid {
			ts.PostureSNOMEDCode = postureCode.String
		}
		if postureDisplay.Valid {
			ts.PostureDisplay = postureDisplay.String
		}
		if eventType.Valid {
			ts.EventType = eventType.String
		}
		if eventCode.Valid {
			ts.EventSNOMEDCode = eventCode.String
		}
		if eventDisplay.Valid {
			ts.EventDisplay = eventDisplay.String
		}
		if areaID.Valid {
			aid := int(areaID.Int64)
			ts.AreaID = &aid
		}
		if heartRateCode.Valid {
			ts.HeartRateCode = heartRateCode.String
		}
		if heartRateDisplay.Valid {
			ts.HeartRateDisplay = heartRateDisplay.String
		}
		if heartRate.Valid {
			hr := int(heartRate.Int64)
			ts.HeartRate = &hr
		}
		if respRateCode.Valid {
			ts.RespiratoryRateCode = respRateCode.String
		}
		if respRateDisplay.Valid {
			ts.RespiratoryRateDisplay = respRateDisplay.String
		}
		if respRate.Valid {
			rr := int(respRate.Int64)
			ts.RespiratoryRate = &rr
		}
		if sleepCode.Valid {
			ts.SleepStateSNOMEDCode = sleepCode.String
		}
		if sleepDisplay.Valid {
			ts.SleepStateDisplay = sleepDisplay.String
		}
		if unitID.Valid {
			ts.UnitID = unitID.String
		}
		if roomID.Valid {
			ts.RoomID = roomID.String
		}
		if alarmEventID.Valid {
			ts.AlarmEventID = alarmEventID.String
		}
		if confidence.Valid {
			c := int(confidence.Int64)
			ts.Confidence = &c
		}
		if remainingTime.Valid {
			rt := int(remainingTime.Int64)
			ts.RemainingTime = &rt
		}
		if rawCompression.Valid {
			ts.RawCompression = rawCompression.String
		}
		if len(metadata) > 0 {
			ts.Metadata = make(map[string]interface{})
		}

		results = append(results, &ts)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return results, nil
}

// GetDataByDevice 按设备查询（支持过滤）
func (r *PostgresIoTTimeSeriesRepository) GetDataByDevice(ctx context.Context, tenantID, deviceID string, filters *IoTTimeSeriesFilters, page, size int) ([]*domain.IoTTimeSeries, int, error) {
	if deviceID == "" {
		return []*domain.IoTTimeSeries{}, 0, nil
	}

	// 设置 deviceID 到 filters
	if filters == nil {
		filters = &IoTTimeSeriesFilters{}
	}
	filters.DeviceID = deviceID

	return r.getDataWithFilters(ctx, tenantID, filters, false, page, size)
}

// GetDataByResident 按住户查询（通过device关联）
func (r *PostgresIoTTimeSeriesRepository) GetDataByResident(ctx context.Context, tenantID, residentID string, filters *IoTTimeSeriesFilters, page, size int) ([]*domain.IoTTimeSeries, int, error) {
	if residentID == "" {
		return []*domain.IoTTimeSeries{}, 0, nil
	}

	// 需要 JOIN beds 和 residents 表
	baseSelect := r.buildBaseQuery(filters != nil && filters.IncludeAlarmEvent)
	baseSelect += `
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id
		WHERE its.tenant_id = $1 AND b.resident_id = $2
	`

	args := []interface{}{tenantID, residentID}
	argN := 3

	// 添加其他过滤条件
	var additionalWhere []string
	if filters != nil {
		if filters.DataType != "" {
			additionalWhere = append(additionalWhere, fmt.Sprintf("its.data_type = $%d", argN))
			args = append(args, filters.DataType)
			argN++
		}
		if filters.Category != "" {
			additionalWhere = append(additionalWhere, fmt.Sprintf("its.category = $%d", argN))
			args = append(args, filters.Category)
			argN++
		}
		if filters.EventType != "" {
			additionalWhere = append(additionalWhere, fmt.Sprintf("its.event_type = $%d", argN))
			args = append(args, filters.EventType)
			argN++
		}
		if filters.UnitID != "" {
			additionalWhere = append(additionalWhere, fmt.Sprintf("its.unit_id = $%d", argN))
			args = append(args, filters.UnitID)
			argN++
		}
		if filters.RoomID != "" {
			additionalWhere = append(additionalWhere, fmt.Sprintf("its.room_id = $%d", argN))
			args = append(args, filters.RoomID)
			argN++
		}
		if filters.StartTime != nil {
			additionalWhere = append(additionalWhere, fmt.Sprintf("its.timestamp >= $%d", argN))
			args = append(args, *filters.StartTime)
			argN++
		}
		if filters.EndTime != nil {
			additionalWhere = append(additionalWhere, fmt.Sprintf("its.timestamp <= $%d", argN))
			args = append(args, *filters.EndTime)
			argN++
		}
		if len(additionalWhere) > 0 {
			baseSelect += " AND " + strings.Join(additionalWhere, " AND ")
		}
	}

	baseSelect += " ORDER BY its.timestamp DESC"

	// 查询总数（只查询 COUNT，不包含所有字段）
	queryCount := "SELECT COUNT(*) FROM iot_timeseries its"
	if filters != nil && filters.IncludeAlarmEvent {
		queryCount += " LEFT JOIN alarm_events ae ON its.alarm_event_id = ae.event_id"
	}
	queryCount += " LEFT JOIN devices d ON its.device_id = d.device_id"
	queryCount += " LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id"
	queryCount += " LEFT JOIN beds b ON d.bound_bed_id = b.bed_id"
	queryCount += " WHERE its.tenant_id = $1 AND b.resident_id = $2"
	
	// 添加其他过滤条件到 COUNT 查询（复用 additionalWhere）
	if len(additionalWhere) > 0 {
		queryCount += " AND " + strings.Join(additionalWhere, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count: %w", err)
	}

	// 分页查询
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	args = append(args, size, offset)
	query := baseSelect + fmt.Sprintf(" LIMIT $%d OFFSET $%d", argN, argN+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	results, err := r.scanRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetDataByTimeRange 时间范围查询
func (r *PostgresIoTTimeSeriesRepository) GetDataByTimeRange(ctx context.Context, tenantID string, startTime, endTime time.Time, filters *IoTTimeSeriesFilters, page, size int) ([]*domain.IoTTimeSeries, int, error) {
	if filters == nil {
		filters = &IoTTimeSeriesFilters{}
	}
	filters.StartTime = &startTime
	filters.EndTime = &endTime

	return r.getDataWithFilters(ctx, tenantID, filters, filters.IncludeAlarmEvent, page, size)
}

// GetDataByLocation 按位置查询（unit_id/room_id）
func (r *PostgresIoTTimeSeriesRepository) GetDataByLocation(ctx context.Context, tenantID string, unitID, roomID *string, filters *IoTTimeSeriesFilters, page, size int) ([]*domain.IoTTimeSeries, int, error) {
	if filters == nil {
		filters = &IoTTimeSeriesFilters{}
	}
	if unitID != nil {
		filters.UnitID = *unitID
	}
	if roomID != nil {
		filters.RoomID = *roomID
	}

	return r.getDataWithFilters(ctx, tenantID, filters, filters.IncludeAlarmEvent, page, size)
}

// getDataWithFilters 通用查询方法（支持过滤和分页）
func (r *PostgresIoTTimeSeriesRepository) getDataWithFilters(ctx context.Context, tenantID string, filters *IoTTimeSeriesFilters, includeAlarmEvent bool, page, size int) ([]*domain.IoTTimeSeries, int, error) {
	baseSelect := r.buildBaseQuery(includeAlarmEvent)
	args := []interface{}{}
	argN := 1

	whereClause := r.buildWhereClause(tenantID, filters, &args, &argN)
	baseSelect += " WHERE " + whereClause + " ORDER BY its.timestamp DESC"

	// 查询总数（只查询 COUNT，不包含所有字段）
	queryCount := "SELECT COUNT(*) FROM iot_timeseries its"
	if includeAlarmEvent {
		queryCount += " LEFT JOIN alarm_events ae ON its.alarm_event_id = ae.event_id"
	}
	queryCount += " LEFT JOIN devices d ON its.device_id = d.device_id"
	queryCount += " LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id"
	queryCount += " WHERE " + whereClause

	var total int
	if err := r.db.QueryRowContext(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count: %w", err)
	}

	// 分页查询
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	args = append(args, size, offset)
	query := baseSelect + fmt.Sprintf(" LIMIT $%d OFFSET $%d", argN, argN+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	results, err := r.scanRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// scanRows 扫描多行记录
func (r *PostgresIoTTimeSeriesRepository) scanRows(rows *sql.Rows) ([]*domain.IoTTimeSeries, error) {
	var results []*domain.IoTTimeSeries

	for rows.Next() {
		var ts domain.IoTTimeSeries
		var trackingID, radarPosX, radarPosY, radarPosZ sql.NullInt64
		var postureCode, postureDisplay sql.NullString
		var eventType, eventCode, eventDisplay sql.NullString
		var areaID sql.NullInt64
		var heartRateCode, heartRateDisplay sql.NullString
		var heartRate sql.NullInt64
		var respRateCode, respRateDisplay sql.NullString
		var respRate sql.NullInt64
		var sleepCode, sleepDisplay sql.NullString
		var unitID, roomID, alarmEventID sql.NullString
		var confidence, remainingTime sql.NullInt64
		var rawCompression sql.NullString
		var metadata []byte

		if err := rows.Scan(
			&ts.ID,
			&ts.TenantID,
			&ts.DeviceID,
			&ts.Timestamp,
			&ts.DataType,
			&ts.Category,
			&trackingID,
			&radarPosX,
			&radarPosY,
			&radarPosZ,
			&postureCode,
			&postureDisplay,
			&eventType,
			&eventCode,
			&eventDisplay,
			&areaID,
			&heartRateCode,
			&heartRateDisplay,
			&heartRate,
			&respRateCode,
			&respRateDisplay,
			&respRate,
			&sleepCode,
			&sleepDisplay,
			&unitID,
			&roomID,
			&alarmEventID,
			&confidence,
			&remainingTime,
			&ts.RawOriginal,
			&ts.RawFormat,
			&rawCompression,
			&metadata,
			&ts.CreatedAt,
			&ts.DeviceSN,
			&ts.DeviceUID,
			&ts.FirmwareVersion,
			&ts.DeviceType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 处理 nullable 字段
		if trackingID.Valid {
			tid := int(trackingID.Int64)
			ts.TrackingID = &tid
		}
		if radarPosX.Valid {
			x := int(radarPosX.Int64)
			ts.RadarPosX = &x
		}
		if radarPosY.Valid {
			y := int(radarPosY.Int64)
			ts.RadarPosY = &y
		}
		if radarPosZ.Valid {
			z := int(radarPosZ.Int64)
			ts.RadarPosZ = &z
		}
		if postureCode.Valid {
			ts.PostureSNOMEDCode = postureCode.String
		}
		if postureDisplay.Valid {
			ts.PostureDisplay = postureDisplay.String
		}
		if eventType.Valid {
			ts.EventType = eventType.String
		}
		if eventCode.Valid {
			ts.EventSNOMEDCode = eventCode.String
		}
		if eventDisplay.Valid {
			ts.EventDisplay = eventDisplay.String
		}
		if areaID.Valid {
			aid := int(areaID.Int64)
			ts.AreaID = &aid
		}
		if heartRateCode.Valid {
			ts.HeartRateCode = heartRateCode.String
		}
		if heartRateDisplay.Valid {
			ts.HeartRateDisplay = heartRateDisplay.String
		}
		if heartRate.Valid {
			hr := int(heartRate.Int64)
			ts.HeartRate = &hr
		}
		if respRateCode.Valid {
			ts.RespiratoryRateCode = respRateCode.String
		}
		if respRateDisplay.Valid {
			ts.RespiratoryRateDisplay = respRateDisplay.String
		}
		if respRate.Valid {
			rr := int(respRate.Int64)
			ts.RespiratoryRate = &rr
		}
		if sleepCode.Valid {
			ts.SleepStateSNOMEDCode = sleepCode.String
		}
		if sleepDisplay.Valid {
			ts.SleepStateDisplay = sleepDisplay.String
		}
		if unitID.Valid {
			ts.UnitID = unitID.String
		}
		if roomID.Valid {
			ts.RoomID = roomID.String
		}
		if alarmEventID.Valid {
			ts.AlarmEventID = alarmEventID.String
		}
		if confidence.Valid {
			c := int(confidence.Int64)
			ts.Confidence = &c
		}
		if remainingTime.Valid {
			rt := int(remainingTime.Int64)
			ts.RemainingTime = &rt
		}
		if rawCompression.Valid {
			ts.RawCompression = rawCompression.String
		}
		if len(metadata) > 0 {
			ts.Metadata = make(map[string]interface{})
		}

		results = append(results, &ts)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return results, nil
}

