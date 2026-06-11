package belief

import "math"

// survival.go — P4 dwell HSMM 生存函数 S_vol(d|zone)。硬阈悬崖 → 平滑 ramp：
// lnLR_fall(d) = −ln S_vol(d|zone)；S_vol = exp(−(d/scale)^shape) ⟹ fallLR = 1 + (d/scale)^shape（封顶）。
// per-zone 尾尺度锚现有硬阈（样本不足先粗档，待 P9 oracle 收紧尾形）；P4.3 夜间短尾（久静更可疑）。
// 单源（#1.3）：likelihood.go ObsDwellStill 唯一调此，不再 inline ramp。

// dwellTail 一个 zone 的生存尾（scale=尾尺度秒，shape=Weibull 形状参数）。
type dwellTail struct {
	scaleSec float64
	shape    float64
}

// dwellTailFor zone → 生存尾。不在表的 zone（bed/rest/enter/unknown）= 久驻正常不报（返回 ok=false → LR=1）。
// scale 锚现有硬阈（P9.6 待 oracle 收紧；shape=2 沿用现行 Weibull k=2）。
func dwellTailFor(roomType int, areaType int) (dwellTail, bool) {
	const (
		RoomTypeBathroom = 1
	)
	// Bathroom toilet/shower: 20min (constipation-safe, medical data)
	if roomType == RoomTypeBathroom && (areaType == 7 || areaType == 6) {
		return dwellTail{scaleSec: 20 * 60, shape: dwellShape}, true
	}
	// Bathroom rest (standing/sitting misclassified): 12min
	if roomType == RoomTypeBathroom {
		return dwellTail{scaleSec: 12 * 60, shape: dwellShape}, true
	}
	switch areaType {
	case 2, 5:
		return dwellTail{}, false // bed/deny: lying area / furniture, 不报
	case 3:
		return dwellTail{scaleSec: 90 * 60, shape: dwellShape}, true // learned sitting area: 90min
	default:
		return dwellTail{scaleSec: 60, shape: dwellShape}, true // Unknown/Active/Enter/unlearned: 60s unit tick
	}
}

// fallLRFromDwell P4.1 生存函数 fall 似然比 = 1 + (d/scale)^shape（封顶 dwellFallCap）。
//   - toleranceMult：开阔地 cell tolerance 拉长尾（被容忍久站→尾更长→久站真人不报，R3 只读；<1 视为 1）。
//   - night：P4.3 风险时段夜间短尾（scale×dwellNightTailMult<1 → 更快 ramp，夜间久静更可疑）。
//   - zone 不在尾表 ∨ dwell≤0 → 1.0 中性（不抬 fall）。
func fallLRFromDwell(dwellSec, toleranceMult float64, roomType int, areaType int, night bool) float64 {
	tail, ok := dwellTailFor(roomType, areaType)
	if !ok || dwellSec <= 0 {
		return 1.0
	}
	if toleranceMult < 1.0 {
		toleranceMult = 1.0 // 默认/未设=1.0（无容忍证据→不放宽）
	}
	scale := tail.scaleSec * toleranceMult
	if night {
		scale *= dwellNightTailMult // 短尾
	}
	r := dwellSec / scale
	lr := 1 + math.Pow(r, tail.shape)
	if lr > dwellFallCap {
		lr = dwellFallCap
	}
	return lr
}
