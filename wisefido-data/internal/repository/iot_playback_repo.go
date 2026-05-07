// iot_playback_repo.go — wisefido-data 仅查询 iot_timeseries 回放原始行；写入在 wisefido-iot internal/repository/iot_timeseries_repo.go
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// IoTTimeSeriesRepository 本模块只用于轨迹回放读库
type IoTTimeSeriesRepository interface {
	GetPlaybackRawMonitorRowsForDevice(ctx context.Context, tenantID, deviceUID string, start, end time.Time, limit int) ([]map[string]interface{}, error)
}

// PostgresIoTTimeSeriesRepository 读 iot_timeseries
type PostgresIoTTimeSeriesRepository struct {
	db *sql.DB
}

func NewPostgresIoTTimeSeriesRepository(db *sql.DB) *PostgresIoTTimeSeriesRepository {
	return &PostgresIoTTimeSeriesRepository{db: db}
}

var _ IoTTimeSeriesRepository = (*PostgresIoTTimeSeriesRepository)(nil)

func (r *PostgresIoTTimeSeriesRepository) GetPlaybackRawMonitorRowsForDevice(ctx context.Context, tenantID, deviceUID string, start, end time.Time, limit int) ([]map[string]interface{}, error) {
	out, err := r.queryPlaybackRawRows(ctx, tenantID, deviceUID, start, end, limit, true)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		out, err = r.queryPlaybackRawRows(ctx, tenantID, deviceUID, start, end, limit, false)
	}
	return out, err
}

func (r *PostgresIoTTimeSeriesRepository) queryPlaybackRawRows(ctx context.Context, tenantID, deviceUID string, start, end time.Time, limit int, monitorOnly bool) ([]map[string]interface{}, error) {
	if deviceUID == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 3000
	}
	if limit > 30000 {
		limit = 30000
	}
	startMs, endMs := start.UnixMilli(), end.UnixMilli()
	topicFilter := ""
	if monitorOnly {
		topicFilter = "  AND its.topic_type = 'monitor'\n"
	}
	q := `
SELECT its.id, its.tenant_id::text, its.device_id::text, its.device_uid,
its.device_type, its.room_id::text, its.bed_id::text, its."timestamp", its.topic_type, its.category,
its.branch_name, its.building_name, its.unit_name, its.room_name, its.bed_name,
its.data_value
FROM iot_timeseries its
INNER JOIN devices d ON d.device_uid = its.device_uid AND d.tenant_id::text = $1
WHERE its.tenant_id::text = $1
  AND its.device_uid = $2
` + topicFilter + `  AND its."timestamp" >= $3
  AND its."timestamp" <= $4
ORDER BY its."timestamp" ASC
LIMIT $5`
	rows, err := r.db.QueryContext(ctx, q, tenantID, deviceUID, startMs, endMs, limit)
	if err != nil {
		return nil, fmt.Errorf("playback raw rows: %w", err)
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var (
			id                                                       int64
			tid, did, duid                                           sql.NullString
			deviceType, roomID, bedID                                sql.NullString
			tsMs                                                     int64
			topicType, category                                      sql.NullString
			branchName, buildingName, unitName, roomName, bedName sql.NullString
			dataVal                                                  []byte
		)
		if err := rows.Scan(
			&id, &tid, &did, &duid,
			&deviceType, &roomID, &bedID, &tsMs, &topicType, &category,
			&branchName, &buildingName, &unitName, &roomName, &bedName,
			&dataVal,
		); err != nil {
			return nil, fmt.Errorf("playback raw scan: %w", err)
		}
		var dv interface{}
		if len(dataVal) > 0 {
			_ = json.Unmarshal(dataVal, &dv)
		}
		row := map[string]interface{}{
			"id":            id,
			"tenant_id":     iotNullStr(tid),
			"device_uid":    iotNullStr(duid),
			"timestamp":     tsMs,
			"topic_type":    iotNullStr(topicType),
			"category":      iotNullStr(category),
			"data_value":    dv,
			"branch_name":   iotNullStr(branchName),
			"building_name": iotNullStr(buildingName),
			"unit_name":     iotNullStr(unitName),
			"room_name":     iotNullStr(roomName),
			"bed_name":      iotNullStr(bedName),
		}
		if did.Valid {
			row["device_id"] = did.String
		}
		if deviceType.Valid {
			row["device_type"] = deviceType.String
		}
		if roomID.Valid {
			row["room_id"] = roomID.String
		}
		if bedID.Valid {
			row["bed_id"] = bedID.String
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func iotNullStr(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}
