package roomengine

// persist_hydrate.go — Xsensorv1 replay **只读** 灌入 28 快照（roomengine_grid_snapshot）。
//
// 与 wisefido-sensor 正本的区别：Xsensorv1 是馈送/replay 引擎，**不 Save、不学习**（无 EncodeSnapshot /
// Persister.Save / LearnCellAreas / UpdateBelief）。本文件只 port 读半边：结构 + UnmarshalSnapshot +
// DecodeSnapshot + 一个只读 Load。RegisterRoom(SetPrior 钉 pin) 之后 hydrate，replay 即拿到生产
// 学到的 cell areaType + dwell 分布(μ/σ) + SitScore，floor 连续 tFloor 与生产一致。
// 跳过 SourceHuman（人标恒在上，pin 已由 SetPrior 写好）。

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// snapshotSchemaVersionMax 当前可读的最高 schema（与 wisefido-sensor 正本 SnapshotSchemaVersion 对齐）。
// v15：chair 久坐学习态迁出 grid snapshot → chair_dwell_state（本表 dmu/dsg/dn/dms/dmsm 已删）。
const snapshotSchemaVersionMax = 15

// CellSnapshot 单 cell 可持久化字段（紧凑 JSON，short keys）。与正本一致。
type CellSnapshot struct {
	I int       `json:"i"`
	B [3][3]int `json:"b"`
	C *Counters `json:"c,omitempty"`
}

// Counters 单 cell 累积字段。与 wisefido-sensor 正本 persist.go 一致（含 v11 的 ss/dm/dsq/dn）。
type Counters struct {
	RD  int       `json:"rd,omitempty"`
	GD  int       `json:"gd,omitempty"`
	AT  [4]uint16 `json:"at,omitempty"`
	TC  uint16    `json:"tc,omitempty"`
	NTC uint16    `json:"ntc,omitempty"`
	LR  int       `json:"lr,omitempty"`
	FE  int       `json:"fe,omitempty"`
	LA  int       `json:"la,omitempty"`
	LS  int       `json:"ls,omitempty"`
	SIB int       `json:"sib,omitempty"`
	SLB int       `json:"slb,omitempty"`
	DE  int       `json:"de,omitempty"`
	FX  int       `json:"fx,omitempty"`
	FY  int       `json:"fy,omitempty"`
	DW  int       `json:"dw,omitempty"`
	FA  int       `json:"fa,omitempty"`
	TS  int       `json:"ts,omitempty"`
	BSR int       `json:"bsr,omitempty"`
	GC  int       `json:"gc,omitempty"`
	RZC int       `json:"rzc,omitempty"`
	RFC int       `json:"rfc,omitempty"`
	SS  float64   `json:"ss,omitempty"`  // SitScore log-odds
	// chair 久坐学习态（dmu/dsg/dn/dms/dmsm）已迁出 grid snapshot → chair_dwell_state（object_id 锚定），此处不再读。
	ADS int64     `json:"ads,omitempty"`
	ET  string    `json:"et,omitempty"`
	IEN int       `json:"ien,omitempty"`
	IEL bool      `json:"iel,omitempty"`
	MBC int       `json:"mbc,omitempty"`
	LMM int64     `json:"lmm,omitempty"`
	LB  bool      `json:"lb,omitempty"`
}

// GridSnapshot 顶层 payload 结构。与正本一致。
type GridSnapshot struct {
	SchemaVer int `json:"schema_v"`
	Grid      struct {
		W  int `json:"w"`
		H  int `json:"h"`
		OX int `json:"ox"`
		OY int `json:"oy"`
	} `json:"grid"`
	Cells []CellSnapshot `json:"cells"`
}

// UnmarshalSnapshot 反序列化。
func UnmarshalSnapshot(data []byte) (GridSnapshot, error) {
	var snap GridSnapshot
	err := json.Unmarshal(data, &snap)
	return snap, err
}

// DecodeSnapshot 把 GridSnapshot 灌回 grid（只覆 Belief + counter，几何不动）。镜像正本：
// SourceHuman 跳过（人标神圣，pin 已由 SetPrior 写）；grid 尺寸不匹配 = layout 变 → 报错冷启。
func DecodeSnapshot(snap GridSnapshot, g *RoomGrid) error {
	if snap.SchemaVer > snapshotSchemaVersionMax {
		return fmt.Errorf("snapshot schema_v=%d > current %d (downgrade unsupported)", snap.SchemaVer, snapshotSchemaVersionMax)
	}
	if snap.SchemaVer < 10 {
		return fmt.Errorf("snapshot schema_v=%d too old (min 10; AreaType 重编号前 area 值失效)", snap.SchemaVer)
	}
	if snap.Grid.W != g.Width || snap.Grid.H != g.Height ||
		snap.Grid.OX != g.OriginX || snap.Grid.OY != g.OriginY {
		return fmt.Errorf("grid dimension mismatch: snap=(%d,%d,%d,%d) grid=(%d,%d,%d,%d)",
			snap.Grid.W, snap.Grid.H, snap.Grid.OX, snap.Grid.OY,
			g.Width, g.Height, g.OriginX, g.OriginY)
	}
	for _, cs := range snap.Cells {
		if cs.I < 0 || cs.I >= len(g.Cells) {
			continue
		}
		c := &g.Cells[cs.I]
		for bi := 0; bi < 3; bi++ {
			if c.Belief[bi].Source == SourceHuman {
				continue // 当前人标不可被 snapshot 覆盖
			}
			if Source(cs.B[bi][2]) == SourceHuman {
				continue // snapshot 里的人标不灌回（人标权威只活在 layout SetPrior）
			}
			c.Belief[bi].Type = AreaType(cs.B[bi][0])
			c.Belief[bi].Confidence = cs.B[bi][1]
			c.Belief[bi].Source = Source(cs.B[bi][2])
		}
		if c.Belief[0].Source != SourceHuman {
			c.AreaType = c.Belief[0].Type
		}
		if cs.C != nil {
			c.RealDecay = cs.C.RD
			c.GhostDecay = cs.C.GD
			c.ActiveType = cs.C.AT
			c.TraverseCount = cs.C.TC
			c.NearTraverseCount = cs.C.NTC
			c.LieRetract = cs.C.LR
			c.FallEventCount = cs.C.FE
			c.LieAnomalyCount = cs.C.LA
			c.LongStillCount = cs.C.LS
			c.SleepadInBedCount = cs.C.SIB
			c.SleepadLeftBedCount = cs.C.SLB
			c.DoorEventCount = cs.C.DE
			c.FlowX = cs.C.FX
			c.FlowY = cs.C.FY
			c.DwellEMA = cs.C.DW
			c.FakeAlarmCount = cs.C.FA
			c.ToleratedStillCount = cs.C.TS
			c.BlindSpotRecoveryCount = cs.C.BSR
			c.GhostCount = cs.C.GC
			c.RestZoneConfirmed = cs.C.RZC
			c.RealFallCount = cs.C.RFC
			c.SitScore = cs.C.SS
			c.AutoDenyQualifiedSinceMs = cs.C.ADS
			c.EnterTarget = cs.C.ET
			c.InsideEnterEvidenceN = cs.C.IEN
			c.InsideEnterLearned = cs.C.IEL
			c.MirrorBounceCount = cs.C.MBC
			c.LastMirrorMs = cs.C.LMM
			c.LearnBlocked = cs.C.LB
		}
	}
	return nil
}

// HydrateRoomFromDB 只读灌入一房的 28 快照到已建好的 grid（RegisterRoom 之后调）。
// 失败/无记录/尺寸不匹配 = 静默冷启（replay 退回 pin-only + chair 冷启 90min），不致命。
func (e *Engine) HydrateRoomFromDB(ctx context.Context, db *sql.DB, table, roomID string) {
	if db == nil {
		return
	}
	if table == "" {
		table = "roomengine_grid_snapshot"
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	grid := e.grids[roomID]
	if grid == nil {
		return // sleepad-only / 无 layout 房：无 grid，跳过
	}
	q := fmt.Sprintf(`SELECT payload FROM %s WHERE spatial_prefix = $1::INET`, table)
	var payload []byte
	switch err := db.QueryRowContext(ctx, q, roomID).Scan(&payload); err {
	case nil:
		// fallthrough to decode
	case sql.ErrNoRows:
		e.logger.Info("xsensor hydrate: no snapshot (cold start)", zap.String("room_id", roomID))
		return
	default:
		e.logger.Warn("xsensor hydrate: load failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	snap, err := UnmarshalSnapshot(payload)
	if err != nil {
		e.logger.Warn("xsensor hydrate: unmarshal failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	if err := DecodeSnapshot(snap, grid); err != nil {
		e.logger.Info("xsensor hydrate: not reusable (grid resized by layout edit?), cold start", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	e.logger.Info("xsensor hydrate ok", zap.String("room_id", roomID), zap.Int("cells", len(snap.Cells)))
}

// ChairDwellRead per-chair 久坐学习态只读行（hydrate 用；N 由 dwin Σ bucket.n 派生）。
type ChairDwellRead struct {
	ObjectID string
	Mu, Sig  float64
	N        int
	MaxSit   float64
	MaxSitMs int64
	CX, CY   int // §12 影子 chair 重建 rect 用（=cx/cy ±20cm）
	Source   int // ChairRect.Source：1=Human / 2=Learned(影子 chair，按 CX/CY 重建 rect)
}

// HydrateChairDwellFromDB 只读灌入一房各 chair 的久坐学习态到 tm.chairDwell（RegisterRoom + SetChairs 之后调）。
// object_id 锚定跨 layout 编辑存活（替掉旧的从 grid-snapshot Counters 读 dwell 路径）。失败/无记录 = 冷启，不致命。
func (e *Engine) HydrateChairDwellFromDB(ctx context.Context, db *sql.DB, table, roomID string) {
	if db == nil {
		return
	}
	if table == "" {
		table = "chair_dwell_state"
	}
	e.mu.Lock()
	tm := e.rooms[roomID]
	e.mu.Unlock()
	if tm == nil {
		return
	}
	q := fmt.Sprintf(`SELECT object_id, dmu, dsg, dwin, dms, dmsm, cx, cy, src FROM %s WHERE spatial_prefix = $1::INET`, table)
	rows, err := db.QueryContext(ctx, q, roomID)
	if err != nil {
		e.logger.Warn("xsensor chair dwell hydrate: load failed", zap.String("room_id", roomID), zap.Error(err))
		return
	}
	defer rows.Close()
	var out []ChairDwellRead
	for rows.Next() {
		var r ChairDwellRead
		var win []byte
		if err := rows.Scan(&r.ObjectID, &r.Mu, &r.Sig, &win, &r.MaxSit, &r.MaxSitMs, &r.CX, &r.CY, &r.Source); err != nil {
			e.logger.Warn("xsensor chair dwell hydrate: scan failed", zap.String("room_id", roomID), zap.Error(err))
			return
		}
		r.N = dwellSampleN(win) // N 由 dwin 派生（无独立列，避免 drift）
		out = append(out, r)
	}
	if len(out) == 0 {
		return
	}
	tm.HydrateChairDwell(out)
	e.logger.Info("xsensor chair dwell hydrated", zap.String("room_id", roomID), zap.Int("chairs", len(out)))
}

// dwellSampleN 从 dwin jsonb 派生窗口样本数 N = Σ bucket.n。解析失败/空 → 0（冷启门控回退 90min）。
func dwellSampleN(dwin []byte) int {
	if len(dwin) == 0 {
		return 0
	}
	var buckets []struct {
		N int `json:"n"`
	}
	if err := json.Unmarshal(dwin, &buckets); err != nil {
		return 0
	}
	n := 0
	for _, b := range buckets {
		n += b.N
	}
	return n
}
