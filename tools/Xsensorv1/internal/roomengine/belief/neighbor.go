package belief

import "math"

// neighbor.go — 跨房 neighbor 隐轴（§A ρ_xroom 有向门控；方程 DBN-Zone-Room §A.1–§A.3；
// 五领域规则补全 §38：核形态改 band-pass + track 守恒 + W/D 随 unit 自适应）。
//
// 本房 real-present track 消失（lost-track，进 Blind*）→ 查同 unit 兄弟房 **fresh 有向 hand-off**
// → ρ_xroom 驱动 T_S 把 Blind 行的 →Fallen 整流入 →Left（人挪去邻房 vs 本房真摔，联合推断）。
// ρ=0（无 hand-off/stale/反向）→ 行不变 → 保 lost-fall（安全默认，铁律 [[partial_monitoring_fall_suppression_law]]）。
//
// 五领域规则（用户 2026-06-16）：
//  ① 越近越强：时间相距越近相关越强 → wDir 新鲜度。
//  ② track 数守恒（P6.5）：短时 unit track 数恒定；本房 −1 ↔ 兄弟房 +1 = 同一人挪过去。**hand-off 权威信号
//     = 丢的 track 在兄弟房重现（守恒），非裸"兄弟房有人"**。载体复用 SuiteCensus/P_id 跨区 track 账。
//  ③ 同 tick 不相关 / 1-5s 峰：Δ≈0（同 tick 两房）= 人没法瞬移 = 可疑（**压低但非零**）；1-5s = 正常走过去 =
//     峰；之后衰减 → wDir **band-pass（先升后降）**，非 exp 单调衰减（旧 §A.1 Δ=0 峰是错的，本轮纠正）。
//  ④ 公共场所窗小：unit 越公共/陌生人流大 → **W（hand-off 检测窗）越小**（防陌生人偶合误判 hand-off）。
//  ⑤ 资源不足窗放长：设备不足/覆盖差 → 二义只能靠 neighbor → **D（延迟裁决窗）放长**（锚 静止消失门限 + 余量）。
//
// 有向性（区别 ghost 对称）：源房先丢、宿房后现，wDir 对 sign(Δ) 不对称。
// 与 §10 接口：q 吃兄弟房**去 ghost** 占用后验。neighbor 是房间 T_S 转移耦合，不加本房 J 隐维（不爆）。

// NeighborParams §A 形态/标定锚（曲线留 oracle；"形状"对即可——band-pass 峰 1-5s）。
type NeighborParams struct {
	HandoffWindowMs int64   // W：radar 穿堂 hand-off 检测窗上界（规则④随公共度收小）
	JitterMs        int64   // J：反向余量（时钟抖动容差）
	TauPeakMs       float64 // radar band-pass 峰时（~1-5s，老人穿堂步速）
	OffsetMs        float64 // 小偏移：令 wDir(Δ=0)>0（同 tick 压低但非零，规则③）
	TauJMs          float64 // Δ<0 抖动段衰减
	Beta            float64 // η sole-resident 连续衰减率
	// sleepad InBed 接力 = 走去另一房 + 躺上床（数十秒，≠ radar 秒级穿堂）→ 慢核（峰后移、窗放长）。
	//   实测 lost→InBed 71s；τpeak=60s 令 wDir(71s)≈0.98、wDir(150s)≈0.56，低端 wDir(5s)≈0.22 保"太快上不了床"。
	SleepadTauPeakMs float64 // sleepad 慢核峰时（~60s）
	SleepadWindowMs  int64   // sleepad 慢核硬窗（~150s）
}

func DefaultNeighborParams() NeighborParams {
	return NeighborParams{HandoffWindowMs: 60_000, JitterMs: 5_000, TauPeakMs: 3_000, OffsetMs: 250, TauJMs: 2_000, Beta: 0.7,
		SleepadTauPeakMs: 60_000, SleepadWindowMs: 150_000}
}

// gammaPeak1 x·e^(1-x)：x=0→0，峰=1 at x=1，x>1 衰减。band-pass 基形。
func gammaPeak1(x float64) float64 { return x * math.Exp(1-x) }

// wDir §A.1 有向 band-pass 核（规则③：同 tick 压低非零、峰后衰减；对 sign(Δ) 不对称）。
//   slow=true（sleepad InBed 接力）→ 用慢核（峰 ~60s、窗 ~150s）；false（radar 穿堂）→ 快核（峰 ~3s、窗 W）。
func wDir(p NeighborParams, deltaMs int64, slow bool) float64 {
	tauPeak, window := p.TauPeakMs, p.HandoffWindowMs
	if slow {
		tauPeak, window = p.SleepadTauPeakMs, p.SleepadWindowMs
	}
	if deltaMs < -p.JitterMs || deltaMs > window {
		return 0 // 窗外 / 真反向 → 非 hand-off（stale 证不了此刻在哪）
	}
	if deltaMs >= 0 {
		// 先升后降：x=(Δ+offset)/τpeak，Δ=0→x 小（压低非零），峰在 Δ≈τpeak。
		return gammaPeak1((float64(deltaMs) + p.OffsetMs) / tauPeak)
	}
	// −J ≤ Δ < 0：从 Δ=0 的压低值按 τj 再衰减（抖动反向余量）。
	return gammaPeak1(p.OffsetMs/tauPeak) * math.Exp(float64(deltaMs)/p.TauJMs)
}

// SiblingHandoff 一个兄弟房的 hand-off 证据（规则② track 守恒：丢的 track 在此重现）。
type SiblingHandoff struct {
	ArrivalDeltaMs int64   // Δ = t_recapture(兄弟房 +1 track) − t_lost(本房)；>0 = 先走后到
	CAttr          float64 // 源型可信度（sleepad InBed 0.9 / room-enter 0.8 / radar-only 0.2）
	GainedReal     float64 // 兄弟房**新增 real track**（守恒重现）的去 ghost 后验（§10 接口）= P(丢的人在此重现)
	Slow           bool    // 慢核：兄弟房现身=sleepad InBed（走+躺，数十秒）→ wDir 用 sleepad τpeak/窗；false=radar 穿堂快核
}

// RhoXroom §A.1：ρ_xroom = η(rc) · max_r'[ w^dir(Δ)·c_attr·GainedReal ]。
// max（非加总）：sole-resident「人只在一处」+ 守恒「丢的 track 在一个兄弟房重现」→ 取最强单房命中
// （加总会两弱信号凑成强 ρ → 过度抑制 → 漏真摔，危险；max 偏低 → 少抑制 → 倾向当摔，安全，C §38 认）。
func RhoXroom(p NeighborParams, residentCount int, sibs []SiblingHandoff) float64 {
	if residentCount < 1 {
		residentCount = 1
	}
	eta := math.Exp(-p.Beta * float64(residentCount-1)) // 单住户 η=1；多住户弱不归零
	best := 0.0
	for _, s := range sibs {
		if v := wDir(p, s.ArrivalDeltaMs, s.Slow) * s.CAttr * s.GainedReal; v > best {
			best = v
		}
	}
	rho := eta * best
	if rho > 1 {
		rho = 1
	}
	return rho
}

// GateBlindRow §A.2：Blind 行 →Fallen 按 ρ 整流入 →Left（F↔L 行和守恒）。
// ρ=0 → 行不变（保 lost-fall 安全默认）；ρ→1 → Fallen 倾向整流入 Left。
func GateBlindRow(row [numStates]float64, rho float64) [numStates]float64 {
	moved := rho * row[SFallen]
	row[SFallen] -= moved
	row[SLeft] += moved
	return row
}

// UnitPublicness §38 规则④ unit 公共度（值同 v2 units.unit_type，DB 整数直转）。
type UnitPublicness int

const (
	UnitPrivate UnitPublicness = 1 // single：私有 suite（单住户）→ 找人窗最长（同住户巧合少，老人走得慢需久等）
	UnitShared  UnitPublicness = 2 // share：共享单元 → 中
	UnitPublic  UnitPublicness = 3 // public：Facility 公共区 → 窗最短（陌生人流大，收紧防偶合误判 hand-off）
)

// HandoffWindowFor 规则④：找人窗 W 按 unit 公共度 3 档（form-anchor 45/60/90s，留 oracle）。
//   越公共 → 窗越短（陌生人偶合越多 → 收紧才不把巧合误判成 hand-off 而误抑制真摔）。
//   未知/缺省 → 私有 90s（residential 单住户为主场景；该档巧合少，久窗低风险）。
func HandoffWindowFor(pub UnitPublicness) int64 {
	switch pub {
	case UnitPublic:
		return 45_000
	case UnitShared:
		return 60_000
	default: // UnitPrivate / 未知
		return 90_000
	}
}

// DelayWindowFor 规则⑤：D（延迟裁决窗）随覆盖差放长，**锚静止消失门限 + 余量**（同一物理时钟，用户裁定）。
//   stillVanishMs = 静止消失门限（~8min，雷达降功率滤掉静止人/物的物理时标）。
//   coverage ∈[0,1]（1=全覆盖，0=大量盲区）。覆盖差 → 二义只能靠 neighbor → 放长（上界 静止门限+余量）。
// D 由 decide/lost-fall 层消费（消失后等过此窗仍无 hand-off/重现 → fire），非 ρ_xroom 本身。
func DelayWindowFor(stillVanishMs int64, marginMs int64, coverage float64) int64 {
	if coverage < 0 {
		coverage = 0
	} else if coverage > 1 {
		coverage = 1
	}
	// 覆盖好 → 接近 静止门限（无需久等）；覆盖差 → 静止门限 + 满余量（neighbor 是唯一裁决，等过门限+缓冲）。
	return stillVanishMs + int64(float64(marginMs)*(1-coverage))
}
