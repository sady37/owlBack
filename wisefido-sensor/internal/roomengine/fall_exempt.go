// fall_exempt.go — Fall 总闸：人工标定躺区豁免。
//
// 用户拍板：人工标定（SourceHuman）的 AreaBed（床 / long-sofa 躺区）= 100% 置信度，
// 老人在该区域静止 = normal use case，所有 Fall 规则不报。
// 仅 AreaBed —— Sit/Toilet/Shower 静止仍是真跌倒高发场景，不享豁免。
//
// 跨 Bedroom/silent_fall/Bathroom 各 fall 路径共享此判定，避免散在多处 drift。

package roomengine

// humanBedExemptMinConfidence 区分"真人工 layout 标定"vs"radar 自学习 lock"。
// 真人工配置 (engine.SetPrior) 写 Confidence=99；MarkRestZoneByFeedback 等自学习路径
// 同样写 SourceHuman 但 Confidence=95（命名 quirk，全局 Source 分级 audit 见 backlog）。
// 用 Confidence≥99 切，避免把 radar 自学习 cell 当人工豁免。
const humanBedExemptMinConfidence = 99

// isHumanBedAt 返回 (x,y) 所在 cell 是否人工 layout 标定的躺区
// （Source=Human + Type=Bed + Confidence ≥ 99，layout-config 专属）。
// grid 或 cell 缺失时返回 false（默认不豁免）。
func isHumanBedAt(grid *RoomGrid, x, y int) bool {
	if grid == nil {
		return false
	}
	cell := grid.CellAt(x, y)
	if cell == nil {
		return false
	}
	b := cell.Belief[0]
	return b.Source == SourceHuman && b.Type == AreaBed && b.Confidence >= humanBedExemptMinConfidence
}
