package roomengine

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Persister 抽象 grid snapshot 字节存储。
// 实现：PostgresPersister（生产）；nil = 不持久化（测试/playback 模式）。
type Persister interface {
	// Save UPSERT 一条 snapshot。payload 是 GridSnapshot 的 JSON 字节。
	Save(ctx context.Context, roomID, layoutHash string, cellCount int, payload []byte) error
	// Load 读取一条 snapshot。found=false 表示无记录（首次启动）。
	Load(ctx context.Context, roomID string) (layoutHash string, payload []byte, found bool, err error)
}

// HistoryPersister 抽象按日期归档的 grid snapshot 存储（每天一份，保留 365 天）。
// 与 Persister 并列：Persister 是 live state（room_id 唯一），HistoryPersister 是历史档案。
// 实现：PostgresHistoryPersister；nil = 不归档。
type HistoryPersister interface {
	// SaveDaily UPSERT 当天的 snapshot 并清理超期记录。
	// retainDays > 0 时，本次写入完成后 DELETE snapshot_date < NOW()-retainDays。
	SaveDaily(ctx context.Context, roomID, layoutHash string, snapshotDate time.Time, cellCount int, payload []byte, retainDays int) error
}

// PostgresPersister 写到 owlRD/db/36_roomengine_grid_snapshot.sql 定义的表
type PostgresPersister struct {
	db    *sql.DB
	table string // 默认 "roomengine_grid_snapshot"
}

// NewPostgresPersister table 默认 "roomengine_grid_snapshot"（与 36_*.sql 对齐）
func NewPostgresPersister(db *sql.DB, table string) *PostgresPersister {
	if table == "" {
		table = "roomengine_grid_snapshot"
	}
	return &PostgresPersister{db: db, table: table}
}

// Save UPSERT (spatial_prefix) DO UPDATE。
// v2 schema: roomengine_grid_snapshot.room_id (UUID) 已退役 → spatial_prefix INET (room /88 CIDR)。
func (p *PostgresPersister) Save(ctx context.Context, roomID, layoutHash string, cellCount int, payload []byte) error {
	// 直接拼表名安全：table 由代码控制不来自用户输入；参数化 SQL 保护数据
	q := fmt.Sprintf(`
		INSERT INTO %s (spatial_prefix, layout_hash, schema_version, cell_count, payload, snapshot_at)
		VALUES ($1::INET, $2, $3, $4, $5::jsonb, NOW())
		ON CONFLICT (spatial_prefix) DO UPDATE
		SET layout_hash    = EXCLUDED.layout_hash,
		    schema_version = EXCLUDED.schema_version,
		    cell_count     = EXCLUDED.cell_count,
		    payload        = EXCLUDED.payload,
		    snapshot_at    = EXCLUDED.snapshot_at
	`, p.table)
	_, err := p.db.ExecContext(ctx, q, roomID, layoutHash, SnapshotSchemaVersion, cellCount, payload)
	return err
}

// Load 取回当前 snapshot；found=false 表示首次启动无记录
func (p *PostgresPersister) Load(ctx context.Context, roomID string) (string, []byte, bool, error) {
	q := fmt.Sprintf(`SELECT layout_hash, payload FROM %s WHERE spatial_prefix = $1::INET`, p.table)
	var hash string
	var payload []byte
	err := p.db.QueryRowContext(ctx, q, roomID).Scan(&hash, &payload)
	if err == sql.ErrNoRows {
		return "", nil, false, nil
	}
	if err != nil {
		return "", nil, false, err
	}
	return hash, payload, true, nil
}

// PostgresHistoryPersister 写到 owlRD/db/37_roomengine_grid_snapshot_history.sql 定义的表
type PostgresHistoryPersister struct {
	db    *sql.DB
	table string // 默认 "roomengine_grid_snapshot_history"
}

// NewPostgresHistoryPersister table 默认 "roomengine_grid_snapshot_history"（与 37_*.sql 对齐）
func NewPostgresHistoryPersister(db *sql.DB, table string) *PostgresHistoryPersister {
	if table == "" {
		table = "roomengine_grid_snapshot_history"
	}
	return &PostgresHistoryPersister{db: db, table: table}
}

// SaveDaily UPSERT (spatial_prefix, archive_date) DO UPDATE，并按 retainDays 清理超期记录。
// snapshotDate 通常用 nowLocal.Truncate(24h)（仅取日期部分，时区由调用方决定）。
//
// v2 schema: roomengine_grid_snapshot_history 列名变化:
//
//	room_id UUID → spatial_prefix INET; snapshot_date → archive_date; taken_at → snapshot_at;
//	PK 仍是 (snapshot_id UUID auto)；UNIQUE/UPSERT 改 (spatial_prefix, archive_date)
func (p *PostgresHistoryPersister) SaveDaily(ctx context.Context, roomID, layoutHash string,
	snapshotDate time.Time, cellCount int, payload []byte, retainDays int) error {
	q := fmt.Sprintf(`
		INSERT INTO %s (spatial_prefix, archive_date, layout_hash, schema_version, cell_count, payload, snapshot_at)
		VALUES ($1::INET, $2::date, $3, $4, $5, $6::jsonb, NOW())
		ON CONFLICT (spatial_prefix, archive_date) DO UPDATE
		SET layout_hash    = EXCLUDED.layout_hash,
		    schema_version = EXCLUDED.schema_version,
		    cell_count     = EXCLUDED.cell_count,
		    payload        = EXCLUDED.payload,
		    snapshot_at    = EXCLUDED.snapshot_at
	`, p.table)
	if _, err := p.db.ExecContext(ctx, q, roomID, snapshotDate.Format("2006-01-02"),
		layoutHash, SnapshotSchemaVersion, cellCount, payload); err != nil {
		return err
	}
	if retainDays > 0 {
		// 顺手清理过期记录；失败仅返回，不阻挡主写入（已 INSERT 成功）
		dq := fmt.Sprintf(`DELETE FROM %s WHERE archive_date < (NOW() - ($1 || ' days')::interval)::date`, p.table)
		if _, err := p.db.ExecContext(ctx, dq, fmt.Sprintf("%d", retainDays)); err != nil {
			return fmt.Errorf("retain prune failed: %w", err)
		}
	}
	return nil
}
