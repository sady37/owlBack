package playback

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"wisefido-sensor/internal/roomengine"
)

const streamRadarTrack = "radar.track"

// OpenDB 用环境变量打开 PG 连接（DB_HOST/PORT/USER/PASSWORD/NAME/SSLMODE）
func OpenDB() (*sql.DB, error) {
	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "postgres")
	pass := getenv("DB_PASSWORD", "postgres")
	name := getenv("DB_NAME", "owl_v2")
	sslMode := getenv("DB_SSLMODE", "disable")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, sslMode)
	return sql.Open("postgres", dsn)
}

// LookupDeviceAddr device_uid(hex MAC) → device_addr /128 canonical IPv6 字符串（无掩码，netip 可直接 parse）
func LookupDeviceAddr(ctx context.Context, db *sql.DB, deviceUID string) (string, error) {
	var addr string
	row := db.QueryRowContext(ctx,
		`SELECT host(device_addr) FROM devices WHERE device_uid = $1 LIMIT 1`, deviceUID)
	if err := row.Scan(&addr); err != nil {
		return "", err
	}
	return addr, nil
}

// LoadDeviceRoomConfig 解析 device_uid 所属 /88 room 下全部 /128 device canvas（room_visual_layout）合并成 RoomConfig。
// 复用 live engine 同款 LoadRoomCanvases + BuildRoomConfigFromCanvases，避免加载逻辑双写漂移。
func LoadDeviceRoomConfig(ctx context.Context, db *sql.DB, deviceUID string, logger *zap.Logger) (string, roomengine.RoomConfig, error) {
	var roomID string
	row := db.QueryRowContext(ctx,
		`SELECT network(set_masklen(device_addr, 88))::text FROM devices WHERE device_uid = $1 LIMIT 1`, deviceUID)
	if err := row.Scan(&roomID); err != nil {
		return "", roomengine.RoomConfig{}, err
	}
	byRoom, err := roomengine.LoadRoomCanvases(ctx, db, logger)
	if err != nil {
		return "", roomengine.RoomConfig{}, err
	}
	cfg, ok := roomengine.BuildRoomConfigFromCanvases(roomID, byRoom[roomID], logger)
	if !ok {
		return roomID, roomengine.RoomConfig{}, fmt.Errorf("room %s has no usable layout canvas", roomID)
	}
	return roomID, cfg, nil
}

// LookupHistorySnapshot 从 roomengine_grid_snapshot_history 拉指定 (spatial_prefix /88, archive_date) 的 payload。
// 用于 playback FromDate：把那一天的 grid 状态作为演化起点。roomID = /88 room network 文本。
func LookupHistorySnapshot(ctx context.Context, db *sql.DB, roomID string, date time.Time) (payload []byte, layoutHash string, found bool, err error) {
	q := `
		SELECT layout_hash, payload
		FROM roomengine_grid_snapshot_history
		WHERE spatial_prefix = $1::inet AND archive_date = $2::date
		ORDER BY snapshot_at DESC
		LIMIT 1`
	err = db.QueryRowContext(ctx, q, roomID, date.Format("2006-01-02")).Scan(&layoutHash, &payload)
	if err == sql.ErrNoRows {
		return nil, "", false, nil
	}
	if err != nil {
		return nil, "", false, err
	}
	return payload, layoutHash, true, nil
}

// Row 一条回放数据（monitor_stream radar.track 帧 或 event_log 离散事件），Run 内部消费。
type Row struct {
	DeviceAddr  string // /128 canonical IPv6（喂 ParseRadarTracks → TrackFrame.DeviceAddr）
	DeviceUID   string
	TimestampMs int64
	TopicType   string // "monitor" / "event"（merge 后区分用）
	Category    string // event_log.event_kind（event 行）
	DataValue   interface{}
}

// QueryMonitorTracks 拉一段时间内 device_addr 的 radar.track 高频帧（monitor_stream 61_）。
func QueryMonitorTracks(ctx context.Context, db *sql.DB, deviceAddr, deviceUID string,
	start, end time.Time, limit int) ([]Row, error) {
	q := `
SELECT ts, payload
FROM monitor_stream
WHERE device_addr = $1
  AND stream_type = $2
  AND ts >= $3 AND ts <= $4
ORDER BY ts ASC
LIMIT $5`
	rows, err := db.QueryContext(ctx, q, deviceAddr, streamRadarTrack, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Row
	for rows.Next() {
		var ts time.Time
		var raw []byte
		if err := rows.Scan(&ts, &raw); err != nil {
			return nil, err
		}
		out = append(out, Row{
			DeviceAddr:  deviceAddr,
			DeviceUID:   deviceUID,
			TimestampMs: ts.UnixMilli(),
			TopicType:   "monitor",
			DataValue:   decodeJSON(raw),
		})
	}
	return out, rows.Err()
}

// QueryEvents 拉一段时间内 device_addr 的离散事件（event_log 62_）。
// 全 event_kind 拉回；ParseRadarTrackEvents 自行只取 Enter/Exit/InBed/LeftBed/NumberPeople，其余返回 nil。
func QueryEvents(ctx context.Context, db *sql.DB, deviceAddr, deviceUID string,
	start, end time.Time, limit int) ([]Row, error) {
	q := `
SELECT ts, event_kind, payload
FROM event_log
WHERE device_addr = $1
  AND ts >= $2 AND ts <= $3
ORDER BY ts ASC
LIMIT $4`
	rows, err := db.QueryContext(ctx, q, deviceAddr, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Row
	for rows.Next() {
		var ts time.Time
		var kind string
		var raw []byte
		if err := rows.Scan(&ts, &kind, &raw); err != nil {
			return nil, err
		}
		out = append(out, Row{
			DeviceAddr:  deviceAddr,
			DeviceUID:   deviceUID,
			TimestampMs: ts.UnixMilli(),
			TopicType:   "event",
			Category:    kind,
			DataValue:   decodeJSON(raw),
		})
	}
	return out, rows.Err()
}

func decodeJSON(raw []byte) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var v interface{}
	_ = json.Unmarshal(raw, &v)
	return v
}

// mergeRowsByTime 把两个已按 TMs 排序的 Row 列表 merge 为一个有序列表（merge-sort step）
func mergeRowsByTime(a, b []Row) []Row {
	out := make([]Row, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i].TimestampMs <= b[j].TimestampMs {
			out = append(out, a[i])
			i++
		} else {
			out = append(out, b[j])
			j++
		}
	}
	if i < len(a) {
		out = append(out, a[i:]...)
	}
	if j < len(b) {
		out = append(out, b[j:]...)
	}
	return out
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
