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
//
// effective≥1 的"1.0"阈 = 维持(2026-06-22 文献核对)。真实跌倒数据支撑,不为"秒起"测试调低：
//   - 真摔倒地时长 = 平均 14min(范围 2–59min);frail 老人 26 例真摔**无一能自起**,50%+ 摔后起不来
//     (PMC3850536 / PMC2590903 / PMC7905119)。即**真摔每次 ≥2min ≫ throu 60s → credit 60s 内即封顶 1.0**。
//   - 故真摔场景:第1次(priorR=0)交 firmware(≥60s firmware 本就报);第2次 priorR≈0.5–0.9 → 约 15–30s 触发,
//     **阈 1.0 已能兜住"刚摔又摔"**,无需调低。
//   - **不调低的理由**:测试里"摔后秒起(11–14s)"→credit 仅 0.18 凑不够,但那是人为不真实;现实中"秒起"的是
//     绊一下/快速坐下(stumble/sit)=非跌倒 → 调低 1.0 会把这些误报(FP 暴涨,老人日均坐站几十次)。
//   - 文献"反复跌倒"=6 月内 >2 次(跨天/周),非几分钟内连摔;急性连摔且每次秒起临床罕见(癫痫/晕厥)。
//   - ⚠️ 验证 §J 须用"摔后躺 ≥30–60s 再起再摔"的真节奏 case,不能用"秒起"测试(credit 凑不够=测试假象非 bug)。

const (
	poseSuspectFall = 2 // SuspectedFall
	poseFall        = 5 // Fall（firmware confirmed）
	poseSuspectSit  = 7 // SuspectedSittingOnGround
	poseSitOnGround = 8 // SittingOnGround（firmware confirmed）

	areaLying = 8 // observation.AreaLying（沙发/躺椅）：躺区 fall-family pose = 正常斜躺，不计 repeat-fall

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

// Step 一帧推进。pose=firmware 原始姿态，real=本 track PReal≥0.5，areaType=当前 cell 区域。
// AreaLying（沙发/躺椅）里 fall-family pose 是正常斜躺不算摔 → 不起 episode（既不攒 R 也不提前 fire，
// 且不让躺区攒的前科泄漏到邻 Sit 区误火）；publish 腿 lying veto 只拦当场发布，堵不住此泄漏。
// 返回 (提前 fire, 本帧 episode 刚结束=记录线 self_recovered)。
func (e *RepeatFallEscalator) Step(nowMs int64, pose int, real bool, areaType int) (earlyFire, recovered bool) {
	if real && isFallPose(pose) && areaType != areaLying {
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
