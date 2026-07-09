package roomengine

// chair_dwell.go — per-chair 久坐学习态（object_id 锚定）的 **只读** 镜像。
//
// 与 wisefido-sensor 正本 1:1：正本学习+写 chair_dwell_state；Xsensorv1 replay 只 hydrate 读入缓存 μ/σ/N/maxSit
// 喂 floor 连续 tFloor，不学不重算不衰减（衰减在 sensor 侧已落值）。学习态从 grid cell 迁到 per-object 后，
// 存活载体 = sibling 表 chair_dwell_state，跨 layout 编辑保号。

import "owl-common/radarutils"

// ChairRect layout 解析出的 chair pin（rect + object_id）。object_id = per-chair 学习态锚定 key。
type ChairRect struct {
	Rect     radarutils.Rect
	ObjectID string
}

// DwellColdMinN dwell 分布冷启门控：样本 < 此值 → floor 回退默认阈（chair 90min / bathroom 20min，与正本一致）。
const DwellColdMinN = 3

// bathDwellObjectID per-room 浴室停留学习态在 chair_dwell_state 表里的哨兵 object_id（浴室无 pin，一房一个）。
// hydrate 侧按此哨兵拣出灌 tm.bathDwell（与 wisefido-sensor 正本 1:1）。
const bathDwellObjectID = "__bathroom__"

// ChairDwell per-chair 久坐学习态只读镜像（hydrate 自 chair_dwell_state；Xsensorv1 不含 Window/不学）。
type ChairDwell struct {
	Mu       float64 // 缓存窗口 μ（秒）；0=冷启
	Sig      float64 // 缓存窗口 σ（秒）
	N        int     // 窗口样本数（由 dwin Σ bucket.n 派生；冷启门控 <DwellColdMinN）
	MaxSit   float64 // false_alarm 反馈棘轮（秒，只读）
	MaxSitMs int64   // 棘轮值 as-of 时刻（只读，1:1 保字段）
}

// ChairObjectAt 命中解析（与正本同判据，电子云 halo=sitSpreadCm）：取 field 最大椅子的 object_id。无命中 → ("", false)。
func ChairObjectAt(chairs []ChairRect, sitSpreadCm, x, y int) (string, bool) {
	halo := sitSpreadCm
	bestW := 0.0
	bestID := ""
	for _, cr := range chairs {
		w := 0.0
		if d := cr.Rect.DistTo(x, y); d == 0 {
			w = 1.0
		} else if halo > 0 && d < halo {
			w = 1.0 - float64(d)/float64(halo)
		}
		if w > bestW {
			bestW, bestID = w, cr.ObjectID
		}
	}
	if bestW <= 0 {
		return "", false
	}
	return bestID, true
}
