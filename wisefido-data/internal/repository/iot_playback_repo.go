// iot_playback_repo.go — wisefido-data 仅查询 monitor_stream 回放原始行。
// 写入路径见 wisefido-iot internal/repository/stream_repo.go。

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type MonitorPlaybackRepository interface {
	GetMonitorRowsByAddr(ctx context.Context, deviceAddr string, start, end time.Time, limit int) ([]map[string]interface{}, error)
}

type PostgresMonitorPlaybackRepository struct {
	db *sql.DB
}

func NewPostgresMonitorPlaybackRepository(db *sql.DB) *PostgresMonitorPlaybackRepository {
	return &PostgresMonitorPlaybackRepository{db: db}
}

var _ MonitorPlaybackRepository = (*PostgresMonitorPlaybackRepository)(nil)

func (r *PostgresMonitorPlaybackRepository) GetMonitorRowsByAddr(ctx context.Context, deviceAddr string, start, end time.Time, limit int) ([]map[string]interface{}, error) {
	if deviceAddr == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 3000
	}
	if limit > 30000 {
		limit = 30000
	}
	const q = `
SELECT ts, host(device_addr) AS device_addr, device_type, stream_type, payload
FROM monitor_stream
WHERE device_addr = $1::INET
  AND ts >= $2
  AND ts <= $3
ORDER BY ts ASC
LIMIT $4`
	rows, err := r.db.QueryContext(ctx, q, deviceAddr, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("monitor_stream query: %w", err)
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var (
			ts         time.Time
			addr       string
			deviceType sql.NullString
			streamType string
			payload    []byte
		)
		if err := rows.Scan(&ts, &addr, &deviceType, &streamType, &payload); err != nil {
			return nil, fmt.Errorf("monitor_stream scan: %w", err)
		}
		var dv interface{}
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &dv)
		}
		row := map[string]interface{}{
			"timestamp":   ts.UnixMilli(),
			"device_addr": addr,
			"stream_type": streamType,
			"data_value":  dv,
		}
		if deviceType.Valid {
			row["device_type"] = deviceType.String
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
