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

// fallLRFromDwell P4.1 生存函数 fall 似然比。**A(P5,委员会选 A):tolerance 翻转 dwell 证据方向**——
// 非容忍 cell(mult=1)正向 ramp `1+(d/scale)^shape`;容忍 cell(mult>1)反向 `1−(tolWeight−1?)...`<1 **主动压**
// SFallen——"此处久站越久=越正常=反 fall 证据",给新拓扑(SFallen 近吸收)缺的反向力,破棘轮(否则任何持续
// LR>1 都饱和,tolerance 只延时不治本)。**不伤 TP**:真摔在非容忍点(床边/浴室地,mult=1)正常 ramp。
//   - toleranceMult ∈[1, MaxToleranceFactor]：R3 只读 cell 自适应容忍(FakeAlarmCount+ToleratedStillCount 学得)。
//   - night：scale×dwellNightTailMult<1 → 更快 ramp(夜间久静更可疑;仅作用正向段)。
//   - zone 不在尾表 ∨ dwell≤0 → 1.0 中性。
func fallLRFromDwell(dwellSec, toleranceMult float64, roomType int, areaType int, night bool) float64 {
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
	ramp := math.Pow(dwellSec/scale, tail.shape)
	// tolWeight=0 当 mult=1（纯正向 ramp，TP 不受影响）；mult 越大权重越大，(1−tolWeight)<0 → 随 dwell **单调下压**。
	tolWeight := (toleranceMult - 1) * dwellTolFlipK
	lr := 1 + (1-tolWeight)*ramp
	if lr < likelihoodFloor {
		lr = likelihoodFloor // 容忍久站强压封底（不湮灭，留可竞争）
	}
	if lr > dwellFallCap {
		lr = dwellFallCap
	}
	return lr
}
