// alarm_playback_repo.go — alarm_events 历史回放查询。
// 与 iot_playback_repo 平行：device_addr + 时间窗 + event_type 白名单。

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type AlarmPlaybackRow struct {
	EventID     string
	DeviceAddr  string
	TriggeredAt time.Time
	AlertedAt   *time.Time
	EventType   string
	AlarmLevel  int
	Payload     map[string]interface{}
}

type AlarmPlaybackRepository interface {
	GetAlarmRowsByAddr(ctx context.Context, deviceAddr string, kinds []string, start, end time.Time, limit int) ([]AlarmPlaybackRow, error)
}

type PostgresAlarmPlaybackRepository struct {
	db *sql.DB
}

func NewPostgresAlarmPlaybackRepository(db *sql.DB) *PostgresAlarmPlaybackRepository {
	return &PostgresAlarmPlaybackRepository{db: db}
}

var _ AlarmPlaybackRepository = (*PostgresAlarmPlaybackRepository)(nil)

func (r *PostgresAlarmPlaybackRepository) GetAlarmRowsByAddr(
	ctx context.Context,
	deviceAddr string,
	kinds []string,
	start, end time.Time,
	limit int,
) ([]AlarmPlaybackRow, error) {
	if limit <= 0 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}
	const q = `
SELECT event_id::text,
       host(device_addr) AS device_addr,
       triggered_at,
       alerted_at,
       event_type,
       alarm_level,
       payload
FROM alarm_events
WHERE device_addr = $1::INET
  AND triggered_at >= $2
  AND triggered_at <= $3
  AND event_type = ANY($4)
ORDER BY triggered_at ASC
LIMIT $5`
	rows, err := r.db.QueryContext(ctx, q, deviceAddr, start, end, pq.Array(kinds), limit)
	if err != nil {
		return nil, fmt.Errorf("alarm_events query: %w", err)
	}
	defer rows.Close()

	var out []AlarmPlaybackRow
	for rows.Next() {
		var (
			id          string
			addr        string
			triggeredAt time.Time
			alertedAt   sql.NullTime
			eventType   string
			alarmLevel  int
			payloadRaw  []byte
		)
		if err := rows.Scan(&id, &addr, &triggeredAt, &alertedAt, &eventType, &alarmLevel, &payloadRaw); err != nil {
			return nil, fmt.Errorf("alarm_events scan: %w", err)
		}
		var payload map[string]interface{}
		if len(payloadRaw) > 0 {
			_ = json.Unmarshal(payloadRaw, &payload)
		}
		row := AlarmPlaybackRow{
			EventID:     id,
			DeviceAddr:  addr,
			TriggeredAt: triggeredAt,
			EventType:   eventType,
			AlarmLevel:  alarmLevel,
			Payload:     payload,
		}
		if alertedAt.Valid {
			t := alertedAt.Time
			row.AlertedAt = &t
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
