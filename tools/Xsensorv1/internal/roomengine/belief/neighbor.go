package belief

import "math"

// neighbor.go — 跨房 hand-off 注入（§7.7 v2 矩形核定稿）。
//
// 本房 real-present track 消失（lost-track）→ 查同 unit 兄弟房 **fresh 有向 hand-off**（丢的人在兄弟房重现）
// → 注入 SLeft 对数似然（"软 ExitRoom"）→ 正常滤波抬 SLeft → durable purge（人挪去邻房 vs 本房真摔）。
// 无 hand-off → 注入 0 → 保 lost-fall 兜底（安全默认，铁律 [[partial_monitoring_fall_suppression_law]]）。
//
// **核形状 = 矩形窗（平、不衰减），不是带通**（§7.6⑥/§7.7 v2 收敛）：
//   timing-shaping（带通/三角高端衰减）本质在判"是不是同一个人"——越晚越疑巧合。但**单住户只有一人，
//   80s 与 15s 重现的同人确定性完全一样 → 衰减=折损满分证据，逻辑错**。只保两条硬边界：
//   - 低端 Δ≤0：拒"反向/非先离后现"（gain 不晚于 loss = 共存或镜像 bleed，非接力）。**Δ 用 gain − A轨
//     last-observed**（非 coast 判定的 LostReal 帧时刻——那个迟到且抖、把 1-3s 真过门压成 Δ≈0 误拒）→ Δ 回到
//     (0, W] 正区间，1-3s 过门（最常见接力）正常放行；ghost 防线在 **GainedReal 去-ghost**（承重判别器），不在时间门。
//   - 高端 Δ>W：超最大合理步行、任何房没现身 → **当可能摔进盲区，不注入 = FN-safe**。
//   中间全满强度恒定（Δ 固定到达延迟，注入逐 tick 不变 → 几 tick 必 purge，远早于 tFloor）。
//
// **FN-safety = 结构性，不靠 kernel 衰减**（A 审 v2，承重红线）：
//   ① 多住户 hard-gate（rc!=1→0）：durable purge 无时序保护，多人时巧合 gain 会 drop 掉真摔轨=FN；
//   ② GainedReal≥阈 去-ghost（sleepad 强制 vital-gate，活 HR/RR=ghost 仿不出的真人重现铁证）；
//   ③ latch-first（已证摔免疫）：engine 层在注入最前判 fallLatched，绝不注入已确认摔的轨。

// NeighborParams hand-off 窗（radar/sleepad 各一上界；矩形窗只需上界 + 低端瞬移门，无 timing-shaping 参数）。
type NeighborParams struct {
	HandoffWindowMs int64 // W_radar：radar 穿堂 hand-off 窗上界（超此=超最大合理步行，不注入）
	SleepadWindowMs int64 // W_sleepad：sleepad InBed（走+躺）hand-off 窗上界
}

func DefaultNeighborParams() NeighborParams {
	// W_radar=90s / W_sleepad=150s：高端截断点 = 各自"最大合理重现时延"，超之当可能摔进盲区不抑制。
	//   不随公共度缩放（elder-care 套房无陌生人、多住户已 hard-gate、单住户无巧合 → 见 §7.6⑥）。
	return NeighborParams{HandoffWindowMs: 90_000, SleepadWindowMs: 150_000}
}

// handoffGainThresh GainedReal≥此（去-ghost real 阈，非裸 >0）；sleepad 另在 engine 层强制 vital-gate。
const handoffGainThresh = 0.5

// handoffLConst = logit(0.9) ≈ 2.197：矩形窗内满强度恒定注入。
//
//	≥ exitFlipLogOdds(=logit 0.85) → 注入即压 floor 兜底（engine floor 闸）+ 逐 tick 抬 SLeft 驱动 absorbed-drop。
//	封顶在此（非更高）= 纵深防御：单 tick 压不垮 confirm 级 SFallen，latch 才是硬底。
var handoffLConst = math.Log(0.9 / 0.1)

// SiblingHandoff 一个兄弟房的 hand-off 证据（规则② track 守恒：丢的 track 在此重现）。
type SiblingHandoff struct {
	ArrivalDeltaMs int64   // Δ = t_recapture(兄弟房 +1 track) − t_lost(本房)；>0 = 先走后到（有向，区别 ghost 对称）
	GainedReal     float64 // 兄弟房新增 real track 去 ghost 后验 = P(丢的人在此重现)
	Slow           bool    // 慢窗：兄弟房现身=sleepad InBed（走+躺）→ 用 W_sleepad；false=radar 穿堂用 W_radar
}

// HandoffLogOdds §7.7 v2：矩形窗 hand-off 注入强度，返回注入进 lost-track exitL 的 SLeft 对数似然（0=不注入）。
//
//	单住户=那个人，窗内任一兄弟房命中即满强度恒定（不随 Δ 衰减）。latch（已证摔免疫）由 engine 层在注入最前判。
func HandoffLogOdds(p NeighborParams, residentCount int, sibs []SiblingHandoff) float64 {
	if residentCount != 1 {
		return 0 // ① 多住户 hard-gate（承重）：durable purge 无时序保护，多人巧合 gain 会 drop 掉真摔轨=FN
	}
	for _, s := range sibs {
		w := p.HandoffWindowMs
		if s.Slow {
			w = p.SleepadWindowMs
		}
		if s.ArrivalDeltaMs <= 0 || s.ArrivalDeltaMs > w {
			continue // 矩形窗外：低端 Δ≤0 反向/共存（非先离后现）/ 高端超最大步行（当可能摔进盲区不注入）
		}
		if s.GainedReal < handoffGainThresh {
			continue // ② 去-ghost：GainedReal 不够 real → 不算重现（sleepad gain 靠 InBed 转移+streak K+此阈，不强制 vital——InBed 不必然有 HR/RR）
		}
		return handoffLConst // 满强度恒定
	}
	return 0
}

// DelayWindowFor 规则⑤：D（延迟裁决窗）随覆盖差放长，**锚静止消失门限 + 余量**（同一物理时钟，用户裁定）。
//   stillVanishMs = 静止消失门限（~8min，雷达降功率滤掉静止人/物的物理时标）。
//   coverage ∈[0,1]（1=全覆盖，0=大量盲区）。覆盖差 → 二义只能靠 neighbor → 放长（上界 静止门限+余量）。
// D 由 decide/lost-fall 层消费（消失后等过此窗仍无 hand-off/重现 → fire），非 hand-off 注入本身。
func DelayWindowFor(stillVanishMs int64, marginMs int64, coverage float64) int64 {
	if coverage < 0 {
		coverage = 0
	} else if coverage > 1 {
		coverage = 1
	}
	return stillVanishMs + int64(float64(marginMs)*(1-coverage))
}
