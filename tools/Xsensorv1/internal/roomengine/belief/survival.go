package belief

import "math"

// survival.go — dwell 久静 → fall 生存函数（P4 移植自 Tsensor belief/survival.go，2026-06-16）。
// 硬阈悬崖 → 平滑 ramp：lnLR_fall(d) = −ln S_vol(d|zone)；S_vol = exp(−(d/scale)^shape)
//   ⟹ fallLR = 1 + (d/scale)^shape（封顶 dwellFallCap）。
// 取代 emission 原 `StillSec ≥ stillTau=60` 全局硬阈——60s 平阈让 DBN 在正常久坐/久站（马桶/沙发）
// 60s 即抬 SFallen 误火（d523-0611 静物伪迹 2.7min 即触发；Tsensor 早在 101/Kitchen/Hunzi/Ton 撞过同类）。
// per-zone 尾尺度按 cell areaType 分档（toilet/shower 20min·learned sit 90min·默认 12min·bed/deny 不报），
// 让 area 配置重新掌控久静阈值（治本：CellAreaType 已过 seam，本轮接进 emission）。
//
// 锚现有硬阈（铁律 [[fall_data_is_artificial_test]]：人为数据不可标定，留 oracle 收紧尾形）。

const (
	dwellShape         = 2.0   // Weibull 形状 k（沿用现行 k=2）
	dwellScaleOpenSec  = 720.0 // 开阔地/Unknown base 12min（≥ AreaSit 自学习阈，让自学习先跑+雷达极限保守）
	dwellEdgeDistCm    = 500.0 // ≥此（雷达远边缘 5m）→ still/pose/z 不可信 → 尾 ×dwellEdgeMult
	dwellEdgeMult      = 1.5   // 边缘风险系数：edge cell 阈值 = base × 1.5
	dwellFallCap       = 2.5   // fallLR 封顶（温和；真 dwell-fall 靠 Decider 窗累积）
	dwellNightTailMult = 0.7   // 夜间短尾（scale×此 = 更快 ramp；久静夜间更可疑）
	dwellTolFlipK      = 2.0   // tolerance 翻转权重：mult>1 时 (1−tolWeight)<0 → 久静反向压 SFallen
)

// dwellTail 一个 zone 的生存尾（scale=尾尺度秒，shape=Weibull 形状参数）。
type dwellTail struct {
	scaleSec float64
	shape    float64
}

// dwellTailFor zone → 生存尾。areaType 数值同 roomengine.AreaType 枚举
// （Bed=2 / Sit=3 / Deny=5 / Shower=6 / Toilet=7）；roomType=card.RoomType（Bathroom=1）。
// 不在表的 zone（bed/deny）= 久驻正常不报（返回 ok=false → LR=1）。
func dwellTailFor(roomType int, areaType int) (dwellTail, bool) {
	const roomTypeBathroom = 1
	// Bathroom toilet/shower：20min（constipation-safe，医学数据）。
	if roomType == roomTypeBathroom && (areaType == 7 || areaType == 6) {
		return dwellTail{scaleSec: 20 * 60, shape: dwellShape}, true
	}
	// Bathroom rest（站/坐被误判）：12min。
	if roomType == roomTypeBathroom {
		return dwellTail{scaleSec: 12 * 60, shape: dwellShape}, true
	}
	switch areaType {
	case 2, 5:
		return dwellTail{}, false // bed/deny：lying 区 / 家具，不报
	case 3:
		return dwellTail{scaleSec: 90 * 60, shape: dwellShape}, true // learned sitting：90min
	default:
		// Unknown/Active/Enter/未学习：12min（dwellScaleOpenSec）。
		return dwellTail{scaleSec: dwellScaleOpenSec, shape: dwellShape}, true
	}
}

// fallLRFromDwell 生存函数 fall 似然比（喂 SFallen）。
//   - toleranceMult ∈[1, MaxToleranceFactor]：cell 自适应容忍。mult>1 → 久静反向压 SFallen
//     （"此处久站越久=越正常=反 fall 证据"，破 SFallen 近吸收棘轮）。本轮先传 1.0（纯正向）。
//   - night：scale×dwellNightTailMult<1 → 更快 ramp（仅作用正向段）。本轮先传 false。
//   - radarDistCm ≥ dwellEdgeDistCm：雷达远边缘 → 尾 ×1.5。本轮先传 0（关）。
//   - zone 不在尾表 ∨ dwellSec≤0 → 1.0 中性。
func fallLRFromDwell(dwellSec, toleranceMult float64, roomType int, areaType int, night bool, radarDistCm int) float64 {
	tail, ok := dwellTailFor(roomType, areaType)
	if !ok || dwellSec <= 0 {
		return 1.0
	}
	if toleranceMult < 1.0 {
		toleranceMult = 1.0
	}
	scale := tail.scaleSec
	if night {
		scale *= dwellNightTailMult
	}
	if float64(radarDistCm) >= dwellEdgeDistCm {
		scale *= dwellEdgeMult
	}
	ramp := math.Pow(dwellSec/scale, tail.shape)
	tolWeight := (toleranceMult - 1) * dwellTolFlipK
	lr := 1 + (1-tolWeight)*ramp
	if lr < likelihoodFloor {
		lr = likelihoodFloor
	}
	if lr > dwellFallCap {
		lr = dwellFallCap
	}
	return lr
}
