package roomengine

// belief_cell_contract.go — P0:cell engine ↔ DBN 唯一耦合边(只读契约冻结)。
// 依据 signal_map §11.5 / belief_dbn_impl_plan §9 P0。
//
// 正向只读边(cell → DBN,P0.1):DBN 只读 cell 的 AreaType + Source 当 π(zone),
// 绝不写、不调 cell 学习/promote(R2 单时间尺度 / R3 cell engine 只读)。
// 本文件把"DBN 读什么"冻结成命名只读 accessor —— belief 侧 geom* 一律走这里,
// 不再裸取 c.Belief[0],防后续 P-task 误把读边扩成写边。
//
// 反向单源边(track still-box → cell,P0.2):still-box 仅在 track_manager.go
// updateContinuousIndicators(BoxRangeWithinMs(30s) → StillBoxRunStart)算一次,
// cell engine(MarkDwell/MarkLongStill)与 DBN(belief_adapter/belief_shadow)
// 双读同一个 StillBoxRunStart 字段,谁都不重算 —— 避免两套阈值 drift(§2.4 producer/maintainer)。

// CellPrior 是 DBN 侧对 cell engine 结果的唯一只读视图。
// 只暴露读、不返回 *Cell(不给写句柄)。*RoomGrid 实现它。
type CellPrior interface {
	AreaTypeAt(x, y int) (AreaType, bool)
	SourceAt(x, y int) (Source, bool)
	NearestEntryDistCm(x, y int) int
}

var _ CellPrior = (*RoomGrid)(nil)

// AreaTypeAt 只读 Belief[0].Type;cell 缺失返回 (AreaUnknown,false)。不触发学习、不改 cell。
func (g *RoomGrid) AreaTypeAt(x, y int) (AreaType, bool) {
	if g == nil {
		return AreaUnknown, false
	}
	c := g.CellAt(x, y)
	if c == nil || len(c.Belief) == 0 {
		return AreaUnknown, false
	}
	return c.Belief[0].Type, true
}

// SourceAt 只读 Belief[0].Source;cell 缺失返回 (SourceUnset,false)。
func (g *RoomGrid) SourceAt(x, y int) (Source, bool) {
	if g == nil {
		return SourceUnset, false
	}
	c := g.CellAt(x, y)
	if c == nil || len(c.Belief) == 0 {
		return SourceUnset, false
	}
	return c.Belief[0].Source, true
}

// NearestEntryDistCm 只读最近门距 cm(纯几何,不经 cell 学习)。
func (g *RoomGrid) NearestEntryDistCm(x, y int) int {
	if g == nil {
		return 1<<31 - 1
	}
	return g.NearestEntryDist(x, y)
}
