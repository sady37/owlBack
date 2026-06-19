package engine

import "math"

// repeat_fall.go — 重复跌倒提前报（§J，用户 2026-06-18 重头设计）。
// 定位：**第一次/孤立摔归 firmware**（pose=2→5@60s / pose=7→8@90s 确认报）；Xsensor 只在
//   "人刚摔过、又摔"时抢在 firmware 前提前报。per-logicID leaky 残余 R 跨"摔→起→再摔"存活。
//
// 机制（每 present 帧、PReal≥0.5 真人）：fall episode = 连续处于 fall 族 pose{2,5,7,8}；起身离族 →
//   episode 结束、credit 烘进 R。effective = R_prior(前科，已衰减) + min(1, FallSec/Throu)；
//   **R_prior>0（有前科）且 effective≥1 → 提前 fire**；R_prior=0（第一次）→ 永不触发（天然交 firmware）。
//   R 随时间 e^{-Δt/τ} 漏衰减，隔久没摔→0→又算第一次。
// 与「无动态阈值调制」铁律相容：不报单次/孤立摔（R_prior>0 闸）；调制源="刚摔过"=同一跌倒风险轴（内禀，
//   非 WeakBio 式正交外来信号）；不改 firmware 阈、不否决 firmware。
// 常数（Throu/τ）留 oracle（[[fall_data_is_artificial_test]]：无真实重复跌倒数据）。

const (
	poseSuspectFall = 2 // SuspectedFall
	poseFall        = 5 // Fall（firmware confirmed）
	poseSuspectSit  = 7 // SuspectedSittingOnGround
	poseSitOnGround = 8 // SittingOnGround（firmware confirmed）

	throuFallSec   = 60.0  // pose 2/5 族单次 confirm 秒（=firmware 保守端，用户 2026-06-18 拍）
	throuSitSec    = 90.0  // pose 7/8 族
	repeatFallTauS = 866.0 // 残余漏衰减时间常数 τ 秒（半衰期 ~10min = 急性聚集窗）
)

func isFallPose(pose int) bool {
	return pose == poseSuspectFall || pose == poseFall || pose == poseSuspectSit || pose == poseSitOnGround
}

func throuFor(pose int) float64 {
	if pose == poseSuspectSit || pose == poseSitOnGround {
		return throuSitSec
	}
	return throuFallSec
}

// RepeatFallEscalator per-logicID 重复摔残余器（跨"摔→起→再摔"存活；dropTrack 时随 logicID 销毁）。
type RepeatFallEscalator struct {
	r            float64 // 已完成 episode 的累积残余（衰减锚在 lastEndMs）
	lastEndMs    int64   // 上个 episode 结束 ms（算 R 衰减；0=从未有过 episode）
	episodeStart int64   // 当前 fall episode 起 ms（0=不在 episode）
	episodeLast  int64   // 当前 episode 最后一帧 ms（episode 结束时算 FallSec 用）
	priorR       float64 // 本 episode 起点的前科残余（=episode 开始衰减后的 r，决定能否提前 fire）
	throu        float64 // 本 episode confirm 阈（按起始 pose 族）
}

func NewRepeatFallEscalator() *RepeatFallEscalator { return &RepeatFallEscalator{} }

// Residual 当前累积残余（forensic 暴露，xray 看"前科"强度）。
func (e *RepeatFallEscalator) Residual() float64 { return e.r }

// Step 一帧推进。pose=firmware 原始姿态，real=本 track PReal≥0.5。
// 返回 (提前 fire, 本帧 episode 刚结束=记录线 self_recovered)。
func (e *RepeatFallEscalator) Step(nowMs int64, pose int, real bool) (earlyFire, recovered bool) {
	if real && isFallPose(pose) {
		if e.episodeStart == 0 { // episode 起：把残余衰减到现在、记前科
			if e.lastEndMs > 0 && nowMs > e.lastEndMs {
				e.r *= math.Exp(-float64(nowMs-e.lastEndMs) / 1000.0 / repeatFallTauS)
			}
			e.episodeStart = nowMs
			e.priorR = e.r
			e.throu = throuFor(pose)
		}
		e.episodeLast = nowMs
		credit := float64(nowMs-e.episodeStart) / 1000.0 / e.throu
		if credit > 1 {
			credit = 1
		}
		// 第一次（priorR=0）永不触发（交 firmware）；有前科且累计过阈 → 提前 fire。
		return e.priorR > 0 && e.priorR+credit >= 1, false
	}
	if e.episodeStart != 0 { // episode 结束：烘 credit 进 R 攒给下次 + 记录线
		credit := float64(e.episodeLast-e.episodeStart) / 1000.0 / e.throu
		if credit > 1 {
			credit = 1
		}
		e.r = e.priorR + credit
		e.episodeStart = 0
		e.lastEndMs = nowMs
		return false, true
	}
	return false, false
}
