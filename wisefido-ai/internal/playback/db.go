package playback

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// OpenDB 用环境变量打开 PG 连接（DB_HOST/PORT/USER/PASSWORD/NAME/SSLMODE）
func OpenDB() (*sql.DB, error) {
	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "postgres")
	pass := getenv("DB_PASSWORD", "postgres")
	name := getenv("DB_NAME", "owlrd")
	sslMode := getenv("DB_SSLMODE", "disable")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, sslMode)
	return sql.Open("postgres", dsn)
}

// LookupTenantID 从 devices 表查 device_uid 对应的 tenant
func LookupTenantID(ctx context.Context, db *sql.DB, deviceUID string) (string, error) {
	var tid string
	row := db.QueryRowContext(ctx,
		`SELECT tenant_id::text FROM devices WHERE device_uid = $1 LIMIT 1`, deviceUID)
	if err := row.Scan(&tid); err != nil {
		return "", err
	}
	return tid, nil
}

// LookupRoomLayout 从 rooms 表查 device_uid 绑定房间的 layout_config（JSONB raw bytes）
// 仅 API 模式用（CLI 模式直接读 layout 文件）
func LookupRoomLayout(ctx context.Context, db *sql.DB, deviceUID string) (roomID string, layoutRaw []byte, err error) {
	q := `
		SELECT r.room_id::text, r.layout_config::text
		FROM devices d
		JOIN rooms r ON r.room_id = d.bound_room_id
		WHERE d.device_uid = $1
		LIMIT 1`
	var layoutStr sql.NullString
	if err := db.QueryRowContext(ctx, q, deviceUID).Scan(&roomID, &layoutStr); err != nil {
		return "", nil, err
	}
	if !layoutStr.Valid || layoutStr.String == "" {
		return roomID, nil, fmt.Errorf("device %s bound room has no layout_config", deviceUID)
	}
	return roomID, []byte(layoutStr.String), nil
}

// Row iot_timeseries 单行（Run 内部消费）
type Row struct {
	ID          int64
	DeviceID    string
	DeviceUID   string
	TimestampMs int64
	DataValue   interface{}
}

// QueryRows 拉一段时间内 device_uid 的 monitor 数据
func QueryRows(ctx context.Context, db *sql.DB, tenantID, deviceUID string,
	start, end time.Time, limit int) ([]Row, error) {

	q := `
SELECT its.id, its.device_id::text, its.device_uid, its."timestamp", its.data_value
FROM iot_timeseries its
WHERE its.tenant_id::text = $1
  AND its.device_uid = $2
  AND its.topic_type = 'monitor'
  AND its."timestamp" >= $3
  AND its."timestamp" <= $4
ORDER BY its."timestamp" ASC
LIMIT $5`
	rows, err := db.QueryContext(ctx, q, tenantID, deviceUID,
		start.UnixMilli(), end.UnixMilli(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Row
	for rows.Next() {
		var (
			id    int64
			did   sql.NullString
			duid  sql.NullString
			tsMs  int64
			dvRaw []byte
		)
		if err := rows.Scan(&id, &did, &duid, &tsMs, &dvRaw); err != nil {
			return nil, err
		}
		var dv interface{}
		if len(dvRaw) > 0 {
			_ = json.Unmarshal(dvRaw, &dv)
		}
		out = append(out, Row{
			ID:          id,
			DeviceID:    did.String,
			DeviceUID:   duid.String,
			TimestampMs: tsMs,
			DataValue:   dv,
		})
	}
	return out, rows.Err()
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
