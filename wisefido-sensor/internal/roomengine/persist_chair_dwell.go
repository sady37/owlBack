package roomengine

// persist_chair_dwell.go — per-chair 久坐学习态字节存储（chair_dwell_state 表，object_id 锚定）。
//
// 与 grid snapshot（persist_postgres.go）并列的 sibling 存储：grid 按 cell index 存（layout 编辑重建
// 冷启清零），本表按 canvas object_id 存（跨编辑保号）。sensor 单一写入口，layout 不碰（rule 1.3）。

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// ChairDwellRow 一把椅子的可持久化久坐学习态（= chair_dwell_state 一行，去掉 spatial_prefix）。
type ChairDwellRow struct {
	ObjectID string
	Mu, Sig  float64
	Window   []DwellBucket
	MaxSit   float64
	MaxSitMs int64
	CX, CY   int
}

// ChairDwellPersister 抽象 chair_dwell_state 读写。实现：PostgresChairDwellPersister；nil = 不持久化。
type ChairDwellPersister interface {
	// LoadRoom 读一个 room（spatial_prefix=roomID）下所有 chair 的久坐学习态。
	LoadRoom(ctx context.Context, roomID string) ([]ChairDwellRow, error)
	// SaveRoom UPSERT 一个 room 下各 chair 的久坐学习态（按 object_id）。
	SaveRoom(ctx context.Context, roomID string, rows []ChairDwellRow) error
}

// PostgresChairDwellPersister 写到 owlRD/dbv2/28_room_cell_state.sql 定义的 chair_dwell_state 表。
type PostgresChairDwellPersister struct {
	db    *sql.DB
	table string // 默认 "chair_dwell_state"
}

// NewPostgresChairDwellPersister table 默认 "chair_dwell_state"。
func NewPostgresChairDwellPersister(db *sql.DB, table string) *PostgresChairDwellPersister {
	if table == "" {
		table = "chair_dwell_state"
	}
	return &PostgresChairDwellPersister{db: db, table: table}
}

// LoadRoom 取回该 room 所有 chair 的久坐学习态；无记录返回空切片。
func (p *PostgresChairDwellPersister) LoadRoom(ctx context.Context, roomID string) ([]ChairDwellRow, error) {
	q := fmt.Sprintf(`SELECT object_id, dmu, dsg, dwin, dms, dmsm, cx, cy FROM %s WHERE spatial_prefix = $1::INET`, p.table)
	rows, err := p.db.QueryContext(ctx, q, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChairDwellRow
	for rows.Next() {
		var r ChairDwellRow
		var win []byte
		if err := rows.Scan(&r.ObjectID, &r.Mu, &r.Sig, &win, &r.MaxSit, &r.MaxSitMs, &r.CX, &r.CY); err != nil {
			return nil, err
		}
		if len(win) > 0 {
			if err := json.Unmarshal(win, &r.Window); err != nil {
				return nil, fmt.Errorf("unmarshal dwin (object %s): %w", r.ObjectID, err)
			}
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// SaveRoom UPSERT (spatial_prefix, object_id) DO UPDATE。逐 chair upsert；stale object（layout 删除的椅子）
// 不主动清（object_id 是 uuid 不复用，读侧按当前 chairs 过滤，残行无害）。
func (p *PostgresChairDwellPersister) SaveRoom(ctx context.Context, roomID string, rows []ChairDwellRow) error {
	if len(rows) == 0 {
		return nil
	}
	q := fmt.Sprintf(`
		INSERT INTO %s (spatial_prefix, object_id, dmu, dsg, dwin, dms, dmsm, cx, cy, updated_at)
		VALUES ($1::INET, $2, $3, $4, $5::jsonb, $6, $7, $8, $9, NOW())
		ON CONFLICT (spatial_prefix, object_id) DO UPDATE
		SET dmu = EXCLUDED.dmu, dsg = EXCLUDED.dsg, dwin = EXCLUDED.dwin,
		    dms = EXCLUDED.dms, dmsm = EXCLUDED.dmsm, cx = EXCLUDED.cx, cy = EXCLUDED.cy,
		    updated_at = EXCLUDED.updated_at
	`, p.table)
	for _, r := range rows {
		win, err := json.Marshal(r.Window)
		if err != nil {
			return err
		}
		if _, err := p.db.ExecContext(ctx, q, roomID, r.ObjectID, r.Mu, r.Sig, win, r.MaxSit, r.MaxSitMs, r.CX, r.CY); err != nil {
			return err
		}
	}
	return nil
}
