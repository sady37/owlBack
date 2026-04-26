package roomengine

import (
	"context"
	"database/sql"
	"fmt"
)

// Persister 抽象 grid snapshot 字节存储。
// 实现：PostgresPersister（生产）；nil = 不持久化（测试/playback 模式）。
type Persister interface {
	// Save UPSERT 一条 snapshot。payload 是 GridSnapshot 的 JSON 字节。
	Save(ctx context.Context, roomID, layoutHash string, cellCount int, payload []byte) error
	// Load 读取一条 snapshot。found=false 表示无记录（首次启动）。
	Load(ctx context.Context, roomID string) (layoutHash string, payload []byte, found bool, err error)
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

// Save UPSERT (room_id) DO UPDATE
func (p *PostgresPersister) Save(ctx context.Context, roomID, layoutHash string, cellCount int, payload []byte) error {
	// 直接拼表名安全：table 由代码控制不来自用户输入；参数化 SQL 保护数据
	q := fmt.Sprintf(`
		INSERT INTO %s (room_id, layout_hash, schema_version, cell_count, payload, updated_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, NOW())
		ON CONFLICT (room_id) DO UPDATE
		SET layout_hash    = EXCLUDED.layout_hash,
		    schema_version = EXCLUDED.schema_version,
		    cell_count     = EXCLUDED.cell_count,
		    payload        = EXCLUDED.payload,
		    updated_at     = EXCLUDED.updated_at
	`, p.table)
	_, err := p.db.ExecContext(ctx, q, roomID, layoutHash, SnapshotSchemaVersion, cellCount, payload)
	return err
}

// Load 取回当前 snapshot；found=false 表示首次启动无记录
func (p *PostgresPersister) Load(ctx context.Context, roomID string) (string, []byte, bool, error) {
	q := fmt.Sprintf(`SELECT layout_hash, payload FROM %s WHERE room_id = $1`, p.table)
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
