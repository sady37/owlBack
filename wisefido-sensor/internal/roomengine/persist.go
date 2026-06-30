package roomengine

// persist.go
//
// RoomGrid 学习状态持久化（snapshot save/load）。
//
// 设计要点（与 owlRD/db/36_roomengine_grid_snapshot.sql 同步）：
//   1. 几何/物理字段（InRoom/InFOV/EdgeDist/MaxZ/MinZ）永不持久化——RegisterRoom 重建
//   2. 人标 Belief 也不需特别处理——RegisterRoom 的 SetPrior 会重灌
//   3. 持久化目标：cell 的累积 counter + 学习出的 Belief（Source != Unset）
//   4. 稀疏存储：跳过"无任何信念 + 全 0 counter"的 cell（典型房间 50%+ 是空地）
//   5. layout_hash 改变时丢弃整个快照（geometry 变了，cell 索引语义变了，不试图合并）
//
// 序列化层 vs 存储层：
//   - EncodeSnapshot / DecodeSnapshot — 纯函数，做 grid ↔ JSON 转换
//   - Persister 接口 — 字节存储抽象（PostgresPersister 实现见 persist_postgres.go）
//   两者解耦：测试可用 NoopPersister + 直接断言 EncodeSnapshot 输出。

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// SnapshotSchemaVersion payload JSON 结构版本。
// 开发阶段：不考虑向后兼容，DecodeSnapshot 严格匹配当前版本，schema 不一致 → 丢弃旧快照重学。
//
// v2 (2026-04-25): Counters 加 NearTraverseCount（NTC）— 用于 auto-Deny 推断
// v3 (2026-04-27): Counters 加 FakeAlarmCount（FA）+ ToleratedStillCount（TS）— cell history integral 自适应阈值
// v4 (2026-04-27): Counters 加 BlindSpotRecoveryCount（BSR）— 盲区返回端点累计（lost-fall recovery 反馈）
// v5 (2026-04-29): Counters 加 GhostCount（GC）+ RestZoneConfirmed（RZC）+ RealFallCount（RFC）—
//
//	PR-6 结构化人类反馈（False Alarm Reason / Observed Conditions checkbox）
//
// v6 (2026-04-29): Counters 加 AutoDenyQualifiedSinceMs（ADS）— PR-15.3 Auto-Deny 15 天时间门控状态
// v7 (2026-05-18): sensor_v2 决定 15+20 — Cell 加 EnterTarget 字符串 + InsideEnterEvidenceN + InsideEnterLearned
// v8 (2026-05-19): L1 mirror pair 自学习 — Cell 加 MirrorBounceCount + LastMirrorMs
// v9 (2026-06-04): sticky 否决 — Cell 加 LearnBlocked（verified 真摔人勾"永不在此自动学抑制"，跨重启保留）
// v11 (2026-06-27): Counters 加 SS（SitScore log-odds）。
// v12 (2026-06-27): Chair 久坐学习由 EMA(DM/DSQ/DN) 改 per-chair 14 日滚动窗——Counters 删 DM/DSQ,
//                   加 DWin（anchor 14 日桶：n/sum/sumSq+5min直方图）+ DMu/DSig/DN（缓存 μ/σ/样本数，floor 热路径读）。
//                   floor 连续 tFloor=clamp(μ+1.5σ,[12,90]) 只对 chair anchor 算,需跨重启保留;否则重启塌回 active(12min)。
//                   非破坏:旧 v11 快照无 DWin → 冷启回退 90min,14 天内边学边收敛(FN-safe)。
// v13 (2026-06-30): chair floor 改 tFloor=clamp(max(μ+1.5σ+10min,maxSit),≤90min)——Counters 加 DMS（DwellMaxSit:
//                   false_alarm+"Sit on Chair" 反馈棘轮,人工确认安全久坐下限,跨重启保留;否则重启丢棘轮→长坐人群复发误报)。
//                   非破坏:旧快照无 DMS → maxSit=0,只走 μ+1.5σ+10min 自然收敛。
// v14 (2026-06-30): maxSit 棘轮加 30 天半衰期慢衰减(治"只升不降→FN 永久退化")——Counters 加 DMSM（DwellMaxSitMs:
//                   当前 maxSit 值的 as-of 时刻,跨重启保衰减基准)。非破坏:旧 v13 快照无 DMSM → 从加载时刻起算。
const SnapshotSchemaVersion = 14 // v10: AreaType 重编号(deny/reflector/monitorbed/interfer/lying，删 shower/toilet)，旧快照 area 值失效须冷启

// CellSnapshot 单 cell 的可持久化字段（紧凑 JSON，short keys 节省空间）
type CellSnapshot struct {
	I int       `json:"i"`           // cell index = col*W + row
	B [3][3]int `json:"b"`           // Belief[i] = [type, confidence, source]
	C *Counters `json:"c,omitempty"` // counter 全 0 时整个 c 段省略
}

// Counters 单 cell 的累积字段（零值字段省略）
type Counters struct {
	RD  int       `json:"rd,omitempty"`  // RealDecay
	GD  int       `json:"gd,omitempty"`  // GhostDecay
	AT  [4]uint16 `json:"at,omitempty"`  // ActiveType[Move,Stand,Sit,Lie] — 数组永不省略，但全 0 时也只 8 字节
	TC  uint16    `json:"tc,omitempty"`  // TraverseCount
	NTC uint16    `json:"ntc,omitempty"` // NearTraverseCount (schema_v ≥ 2)
	LR  int       `json:"lr,omitempty"`  // LieRetract
	FE  int       `json:"fe,omitempty"`  // FallEventCount
	LA  int       `json:"la,omitempty"`  // LieAnomalyCount
	LS  int       `json:"ls,omitempty"`  // LongStillCount
	SIB int       `json:"sib,omitempty"` // SleepadInBedCount
	SLB int       `json:"slb,omitempty"` // SleepadLeftBedCount
	DE  int       `json:"de,omitempty"`  // DoorEventCount
	FX  int       `json:"fx,omitempty"`  // FlowX
	FY  int       `json:"fy,omitempty"`  // FlowY
	DW  int       `json:"dw,omitempty"`  // DwellEMA
	FA  int       `json:"fa,omitempty"`  // FakeAlarmCount (schema_v ≥ 3)
	TS  int       `json:"ts,omitempty"`  // ToleratedStillCount (schema_v ≥ 3)
	BSR int       `json:"bsr,omitempty"` // BlindSpotRecoveryCount (schema_v ≥ 4)
	GC  int       `json:"gc,omitempty"`  // GhostCount (schema_v ≥ 5)
	RZC int       `json:"rzc,omitempty"` // RestZoneConfirmed (schema_v ≥ 5)
	RFC int       `json:"rfc,omitempty"` // RealFallCount (schema_v ≥ 5)
	SS  float64   `json:"ss,omitempty"`  // SitScore log-odds (schema_v ≥ 11) — Belief→AreaSit 学习层
	DWin []DwellBucket `json:"dwin,omitempty"` // Chair anchor 14 日久坐窗 (schema_v ≥ 12)
	DMu  float64       `json:"dmu,omitempty"`  // 缓存窗口 μ 秒 (schema_v ≥ 12) — floor 热路径读
	DSig float64       `json:"dsg,omitempty"`  // 缓存窗口 σ 秒 (schema_v ≥ 12)
	DN   int           `json:"dn,omitempty"`   // 缓存窗口样本数 (schema_v ≥ 12)
	DMS  float64       `json:"dms,omitempty"`  // DwellMaxSit 秒 (schema_v ≥ 13) — false_alarm 反馈棘轮
	DMSM int64         `json:"dmsm,omitempty"` // DwellMaxSitMs as-of 时刻 (schema_v ≥ 14) — 衰减基准
	ADS int64     `json:"ads,omitempty"` // AutoDenyQualifiedSinceMs (schema_v ≥ 6)

	// sensor_v2 v7 (决定 15+20):
	ET  string `json:"et,omitempty"`  // EnterTarget ""/"outside"/"bathroom" (schema_v ≥ 7)
	IEN int    `json:"ien,omitempty"` // InsideEnterEvidenceN (schema_v ≥ 7)
	IEL bool   `json:"iel,omitempty"` // InsideEnterLearned (schema_v ≥ 7)

	// L1 mirror pair 自学习 (schema_v ≥ 8):
	MBC int   `json:"mbc,omitempty"` // MirrorBounceCount
	LMM int64 `json:"lmm,omitempty"` // LastMirrorMs
	LB  bool  `json:"lb,omitempty"`  // LearnBlocked (schema_v ≥ 9) — sticky 否决
}

// GridSnapshot 顶层 payload 结构（写入 JSONB 字段）
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

// LayoutHash 计算 RoomConfig 的几何 hash。
// 用于判定 layout 是否有结构性变更（任意矩形/多边形/雷达安装/物体高度变了即丢弃旧 snapshot）。
//
// 故意不包含 RoomID、RoomW、RoomH、Origin——这些是元信息不影响 cell 索引语义。
// 实际影响 cell→几何映射的字段：WallPolygon + 各类矩形 + 各物体高度 + Radar 安装。
// （未来 RoomEngine 用 height 决定"吊灯不阻挡通行"等，所以 height 也是 cell 语义的一部分）
func LayoutHash(cfg RoomConfig) string {
	h := sha256.New()
	enc := json.NewEncoder(h)
	// 顺序固定（json 编码按字段顺序），保证同输入同输出
	_ = enc.Encode(cfg.WallPolygon)
	_ = enc.Encode(cfg.Enters)
	_ = enc.Encode(cfg.EnterHeights)
	_ = enc.Encode(cfg.EnterTargets) // sensor_v2 决定 15：EnterTarget 变更视作 layout 变更，invalidate grid
	_ = enc.Encode(cfg.RoomType)     // sensor_v2 决定 16：RoomType 变更（default↔bathroom）触发完整重学
	_ = enc.Encode(cfg.Beds)
	_ = enc.Encode(cfg.BedHeights)
	_ = enc.Encode(cfg.Toilets)
	_ = enc.Encode(cfg.ToiletHeights)
	_ = enc.Encode(cfg.Showers)
	_ = enc.Encode(cfg.ShowerHeights)
	_ = enc.Encode(cfg.Chairs)
	_ = enc.Encode(cfg.ChairHeights)
	_ = enc.Encode(cfg.Furnitures)
	_ = enc.Encode(cfg.FurnitureHeights)
	_ = enc.Encode(cfg.Interferes)
	_ = enc.Encode(cfg.InterfereHeights)
	_ = enc.Encode(cfg.Radar)
	return hex.EncodeToString(h.Sum(nil))
}

// EncodeSnapshot 把 grid 序列化为 GridSnapshot（稀疏：跳过 trivial cell）。
// trivial = Belief[0].Source == SourceUnset 且所有 counter 都为 0。
func EncodeSnapshot(g *RoomGrid) GridSnapshot {
	snap := GridSnapshot{SchemaVer: SnapshotSchemaVersion}
	snap.Grid.W = g.Width
	snap.Grid.H = g.Height
	snap.Grid.OX = g.OriginX
	snap.Grid.OY = g.OriginY

	for i := range g.Cells {
		c := &g.Cells[i]
		ctrs := buildCounters(c)
		hasBelief := c.Belief[0].Source != SourceUnset ||
			c.Belief[1].Source != SourceUnset ||
			c.Belief[2].Source != SourceUnset
		if ctrs == nil && !hasBelief {
			continue // trivial cell
		}
		snap.Cells = append(snap.Cells, CellSnapshot{
			I: i,
			B: [3][3]int{
				{int(c.Belief[0].Type), c.Belief[0].Confidence, int(c.Belief[0].Source)},
				{int(c.Belief[1].Type), c.Belief[1].Confidence, int(c.Belief[1].Source)},
				{int(c.Belief[2].Type), c.Belief[2].Confidence, int(c.Belief[2].Source)},
			},
			C: ctrs,
		})
	}
	return snap
}

// buildCounters 把 cell 的累积字段抽到 Counters；全 0 返回 nil（让 c 段省略）
func buildCounters(c *Cell) *Counters {
	ct := Counters{
		RD:  c.RealDecay,
		GD:  c.GhostDecay,
		AT:  c.ActiveType,
		TC:  c.TraverseCount,
		NTC: c.NearTraverseCount,
		LR:  c.LieRetract,
		FE:  c.FallEventCount,
		LA:  c.LieAnomalyCount,
		LS:  c.LongStillCount,
		SIB: c.SleepadInBedCount,
		SLB: c.SleepadLeftBedCount,
		DE:  c.DoorEventCount,
		FX:  c.FlowX,
		FY:  c.FlowY,
		DW:  c.DwellEMA,
		FA:  c.FakeAlarmCount,
		TS:  c.ToleratedStillCount,
		BSR: c.BlindSpotRecoveryCount,
		GC:  c.GhostCount,
		RZC: c.RestZoneConfirmed,
		RFC: c.RealFallCount,
		SS:  c.SitScore,
		DWin: c.DwellWindow,
		DMu:  c.DwellMu,
		DSig: c.DwellSig,
		DN:   c.DwellN,
		DMS:  c.DwellMaxSit,
		DMSM: c.DwellMaxSitMs,
		ADS: c.AutoDenyQualifiedSinceMs,
		ET:  c.EnterTarget,
		IEN: c.InsideEnterEvidenceN,
		IEL: c.InsideEnterLearned,
		MBC: c.MirrorBounceCount,
		LMM: c.LastMirrorMs,
		LB:  c.LearnBlocked,
	}
	if ct.RD == 0 && ct.GD == 0 &&
		ct.AT == [4]uint16{} &&
		ct.TC == 0 && ct.NTC == 0 && ct.LR == 0 && ct.FE == 0 && ct.LA == 0 && ct.LS == 0 &&
		ct.SIB == 0 && ct.SLB == 0 && ct.DE == 0 && ct.FX == 0 && ct.FY == 0 && ct.DW == 0 &&
		ct.FA == 0 && ct.TS == 0 && ct.BSR == 0 && ct.ADS == 0 &&
		ct.ET == "" && ct.IEN == 0 && !ct.IEL &&
		ct.MBC == 0 && ct.LMM == 0 && !ct.LB && ct.SS == 0 &&
		len(ct.DWin) == 0 && ct.DMu == 0 && ct.DSig == 0 && ct.DN == 0 {
		return nil
	}
	return &ct
}

// DecodeSnapshot 把 GridSnapshot 灌回 grid。
//
// 调用前提：grid 已由 NewRoomGrid + StampRoomPolygon + StampRadar + StampEnters + SetPrior 建好；
// 本函数仅覆盖 Belief + counter（geometry 不动）。
//
// 返回错误的场景：
//   - schema_version 不支持
//   - grid 尺寸（W/H/OX/OY）与 snapshot 不匹配（也是 layout 变更的一种）
//
// 单 cell 越界（i >= len(grid.Cells)）会跳过并不报错——可能是 grid 裁剪过（边界 ±1 cell）。
func DecodeSnapshot(snap GridSnapshot, g *RoomGrid) error {
	// 向后兼容：低版本 snapshot 也接受（仅新字段视为 0）；高于当前版本拒绝
	if snap.SchemaVer > SnapshotSchemaVersion {
		return fmt.Errorf("snapshot schema_v=%d > current %d (downgrade unsupported)", snap.SchemaVer, SnapshotSchemaVersion)
	}
	if snap.SchemaVer < 10 {
		return fmt.Errorf("snapshot schema_v=%d too old (min supported = 10; AreaType 重编号前的快照 area 值失效)", snap.SchemaVer)
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
		// Belief：直接覆盖；SourceHuman 在 RegisterRoom→SetPrior 时已写入，
		// 这里如果 snapshot 里也是 SourceHuman 则冗余但等价；
		// 如果 snapshot 是 SourceLearned 但 baseline 是 SourceHuman——以 SnapShot 为准
		// 是错误的，因为人标神圣。所以加保护：仅当 baseline 不是 Human 时才覆盖。
		for bi := 0; bi < 3; bi++ {
			if c.Belief[bi].Source == SourceHuman {
				continue // 当前人标不可被 snapshot 覆盖
			}
			if Source(cs.B[bi][2]) == SourceHuman {
				continue // snapshot 里的人标不灌回：人标权威只活在 layout(SetPrior)，防 layout 改后旧人标区从 28 复活
			}
			c.Belief[bi].Type = AreaType(cs.B[bi][0])
			c.Belief[bi].Confidence = cs.B[bi][1]
			c.Belief[bi].Source = Source(cs.B[bi][2])
		}
		// AreaType mirror（同 promoteCell 行为）
		if c.Belief[0].Source != SourceHuman {
			c.AreaType = c.Belief[0].Type
		}
		// Counters：snapshot 中 c 段缺失即视为全 0（保持 grid 当前值——通常也是 0）
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
			c.DwellWindow = cs.C.DWin
			c.DwellMu = cs.C.DMu
			c.DwellSig = cs.C.DSig
			c.DwellN = cs.C.DN
			c.DwellMaxSit = cs.C.DMS
			c.DwellMaxSitMs = cs.C.DMSM
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

// MarshalSnapshot 序列化 + 压缩（暂不压缩，留 hook）。
// 返回 (payload []byte, cellCount int)；cellCount 写入 PG 列便于运维监控。
func MarshalSnapshot(snap GridSnapshot) ([]byte, int, error) {
	data, err := json.Marshal(snap)
	if err != nil {
		return nil, 0, err
	}
	return data, len(snap.Cells), nil
}

// UnmarshalSnapshot 反序列化
func UnmarshalSnapshot(data []byte) (GridSnapshot, error) {
	var snap GridSnapshot
	err := json.Unmarshal(data, &snap)
	return snap, err
}
