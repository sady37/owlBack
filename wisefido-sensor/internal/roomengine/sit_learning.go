package roomengine

import "math"

// sit_learning.go — AreaSit 4 通道 log-odds 学习评分核心（设计 doc/cell_area_learning_scoring.md §11；DBN-Zone-Room §N）。
//
// 纯评分函数 + oracle 常量。episode 锚-停留累积 / break 分类 / emit 接线在 track_manager.go。
// 4 通道（log-odds 相加）：
//   resolution  自走收场硬判据（walk-away→+；lost→veto 不计，fall 域）—— 姿态无关
//   dwell       静止时长平滑 bump（StillBox 驱动，<5min→Active）—— 姿态无关
//   g_fwSit     pose=Sitting 增益（base 0.6 > z 权 0.4，故 pose>z）—— 姿态相关
//   z           坐姿高度高斯隶属度（纯正向，z=0 缺失中性，无负分支）—— 姿态相关、与 g_fwSit 各算各的
// 治本：dwell/resolution 姿态无关，固件把坐误读 Standing/z=0（D523 角装）也能学成 Sit。

const (
	// episode / 升格（SitScore 衰减半衰期已移 DecayParams.SitScoreSec，config 可调）
	sitActiveCutoffMin = 5.0 // dwell <5min → Active，不计 Sit
	sitPromoteTau      = 6.0 // cell.SitScore ≥ τ → 升 AreaSit

	// resolution（硬判据；lost → veto 不调本函数）
	sitLLRWalkAway = 2.0

	// z（纯正向高斯隶属度，缺失中性，权重 < pose）
	sitZCenter = 80.0 // 坐姿 z 中心（实测 2026-06-23 monitor_stream：pose=Sitting z>0 中位 80；可调，175cm 人群→~83–85）
	sitZSigma  = 12.0
	sitZWeight = 0.4

	// dwell（升降双 sigmoid 平滑 bump，分钟）
	sitDwellRiseMid    = 8.0
	sitDwellRiseSlope  = 3.0
	sitDwellFallMid    = 55.0
	sitDwellFallSlope  = 10.0
	sitDwellBedLeanMin = 30.0 // dwell>此 ∧ Lying 主导 → bedLean（建议 Bed，不自升）

	// firmware-sit（pose=Sitting；base>z 故 pose>z；安全靠 episode walk-away 门控，非高 Tmin）
	sitFwBase         = 0.6
	sitFwSlopePerMin  = 0.04
	sitFwMinMin       = 1.0
	sitFwCap          = 1.5
	sitFwContigGateMs = 3000 // 3s 软门：连续 Sitting ≥3s 才计入 fwMax
)

func sitSigmoid(x float64) float64 { return 1.0 / (1.0 + math.Exp(-x)) }

// sitDwellLLR 平滑 bump；<cutoff 返回 0（调用方据此判 Active，不 emit Sit）。
func sitDwellLLR(dwellMin float64) float64 {
	if dwellMin < sitActiveCutoffMin {
		return 0
	}
	return sitSigmoid((dwellMin-sitDwellRiseMid)/sitDwellRiseSlope) *
		sitSigmoid((sitDwellFallMid-dwellMin)/sitDwellFallSlope)
}

// sitZMembership 坐姿 z 高斯隶属度 [0,1]（仅 z>0 调用；z=0 缺失由调用方不更新 = 中性）。
func sitZMembership(z float64) float64 {
	d := z - sitZCenter
	return math.Exp(-(d * d) / (2 * sitZSigma * sitZSigma))
}

// sitGFw firmware pose=Sitting 增益（fwMaxMin = episode 内最长连续 Sitting 分钟）。
// base 0.6 > sitZWeight 0.4 → pose 通道恒重于 z。
func sitGFw(fwMaxMin float64) float64 {
	if fwMaxMin < sitFwMinMin {
		return 0
	}
	g := sitFwBase + sitFwSlopePerMin*(fwMaxMin-sitFwMinMin)
	if g > sitFwCap {
		g = sitFwCap
	}
	return g
}

// sitEpisodeLLR 一条 walk-away episode 的 ΔL_sit（resolution=walk-away 已确认才调用；lost 不调本函数）。
//
//	dwellMin = StillBox 时长分钟；zBest = episode 内 max sitZMembership（z>0 帧，无则 0）；fwMaxMin = 最长连续 Sitting 分钟。
func sitEpisodeLLR(dwellMin, zBest, fwMaxMin float64) float64 {
	return sitLLRWalkAway + sitZWeight*zBest + sitDwellLLR(dwellMin) + sitGFw(fwMaxMin)
}

// sitSpreadWeight episode 加分按到锚的切比雪夫距离分层衰减（±10→1.0 / ±20→0.5 / ±30→0.3 / 更远→0）。
// 让锚漂 30cm 内的 episode 合并，边缘减权不把一大片误学满额 Sit。
func sitSpreadWeight(ringCm int) float64 {
	switch {
	case ringCm <= 10:
		return 1.0
	case ringCm <= 20:
		return 0.5
	case ringCm <= 30:
		return 0.3
	}
	return 0
}
