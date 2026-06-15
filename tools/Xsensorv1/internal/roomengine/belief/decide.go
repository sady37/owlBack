package belief

import "math"

// decide.go — 读出与裁决（§26 用户裁定 55% 三分判据，资源稀缺前提；推翻 §8 全局 C_FN 兜底）。
//   P^F ≥ 55%            → 报（证据自足，C_FN 不需要）
//   P^F ≤ 45%            → 不报
//   45–55% 两可 且 可判    → C_FN 风险偏好打破平衡（C_FN **唯一作用窗口**）
//   高度不可判（Λ 无判别力）→ 默认不报（C_FN 不介入）   ← §26 知情设计决定（署名）
// 持续 ≥ T_hold 防瞬时噪声。
//
// 知情设计决定（用户 2026-06-15 署名）：高度不可判默认不报 = 知情接受「设备不足时漏掉一部分**高度
// 不可判**真摔，换告警可信度」。理由：资源稀缺→护理注意力稀缺→低置信告警烧穿注意力(alarm fatigue)
// →真摔也被忽略，比偶尔漏报更糟。仅高度不可判侧；可判侧（≥55%/≤45%）正常报/不报，不受影响。
//
// 推翻记录（诚实）：§8/§17「C_FN 全局兜底、不可判走同一不等式、Λ 不 gate」是资源充足逻辑，
// 已被 §26 收窄——C_FN 只在 45–55% 两可窗打破平衡；高度不可判 **Λ 现作 gate**（默认不报）。
//
// 形态锚（铁律 [[fall_data_is_artificial_test]]：阈值/曲线全人为数据不可标定，留 oracle）。

// RiskContext 裁决期风险因子（复用 risk_evaluator 同源因子，§8 连续消费非离散档）。
type RiskContext struct {
	AloneContinuousMin float64 // 独居持续分钟（独处→C_FN↑，连续饱和）
	Night              bool    // 风险时段（21–08，夜间无人巡视→C_FN↑）
	PeopleCount        int     // 房内人数（>1 有人代发现→C_FN 折扣，但有下限不归零）
	Disabled           bool    // 失能（难自救→C_FN↑）
}

// decideParams §8 代价/持续形态参数（标定锚，非权威值，留 oracle）。
type decideParams struct {
	cFP         float64 // 误报代价基线（护士跑空腿，归一=1）
	cFNBase     float64 // 漏报代价基线（有人在场时）
	aloneGain   float64 // 独居持续对 C_FN 的连续增益上限
	aloneSatMin float64 // 独居增益饱和分钟
	nightMult   float64 // 夜间倍数
	disMult     float64 // 失能倍数
	peopleFloor float64 // 多人折扣下限（不归零：人多仍可能没人注意到摔倒）
	tHoldMs     int64   // 持续窗（防瞬时噪声，§8 ~90s）
}

func defaultDecideParams() decideParams {
	return decideParams{
		cFP: 1.0, cFNBase: 2.0,
		aloneGain: 4.0, aloneSatMin: 30,
		nightMult: 1.5, disMult: 1.5, peopleFloor: 0.3,
		tHoldMs: 90_000,
	}
}

// cFN §8 连续代价函数 C_FN(risk)。形态=框架（连续、各因子单调、多人折扣有下限）；取值=标定。
func (p decideParams) cFN(rc RiskContext) float64 {
	c := p.cFNBase
	// 独居持续：连续增益，饱和到 aloneSatMin。
	alone := rc.AloneContinuousMin / p.aloneSatMin
	if alone > 1 {
		alone = 1
	}
	c *= 1 + p.aloneGain*alone
	if rc.Night {
		c *= p.nightMult
	}
	if rc.Disabled {
		c *= p.disMult
	}
	// 多人折扣（连续 1/N，floor 不归零）。PeopleCount≤1 = 独处，不折扣。
	if rc.PeopleCount > 1 {
		disc := 1.0 / float64(rc.PeopleCount)
		if disc < p.peopleFloor {
			disc = p.peopleFloor
		}
		c *= disc
	}
	return c
}

// Decision 一帧裁决结果 + forensic（probe/sdl）。
type Decision struct {
	Fire         bool    // 持续 ≥ T_hold 后的最终触发
	InstFire     bool    // 本帧瞬时判据满足（未含持续）
	Band         string  // 落在哪档：report(≥55) / no(≤45) / tie(45-55可判) / indeterminate(高度不可判)
	PFallen      float64 // P^F_t
	PeopleCount  int     // N_r（人数单源 = TrackCensus.Nr()，已排 ghost；forensic + 拍法 A 守门校验）
	CFN          float64 // C_FN(risk)（仅 tie 档参与）
	Margin       float64 // P^F − 报阈（诊断：>0 越确定该报）
	Lambda       float64 // Λ_t 似然比
	Identifiable bool    // Λ_t > 阈 = 可判（§26 起 **参与裁决**：高度不可判默认不报）
	FireSinceMs  int64   // 瞬时条件起始（持续计时；0=未武装）
}

// 55% 三分阈 + 可判阈（form-anchor，留 oracle）。
const (
	pFireHi           = 0.55 // P^F ≥ → 报（证据自足）
	pFireLo           = 0.45 // P^F ≤ → 不报
	lambdaInformative = 3.0  // Λ > 此 = 可判；否则高度不可判（§26 gate 默认不报）
)

// Decider 持有持续计时状态（跨帧）。
type Decider struct {
	p           decideParams
	fireSinceMs int64
}

func NewDecider() *Decider { return &Decider{p: defaultDecideParams()} }

// Step 一帧裁决（§26 55% 三分）。pFallen=P^F_t（filter.PFallen）；lambda=ComputeLambda。
func (d *Decider) Step(nowMs int64, pFallen, lambda float64, rc RiskContext) Decision {
	cfn := d.p.cFN(rc)
	identifiable := lambda > lambdaInformative
	var inst bool
	var band string
	switch {
	case pFallen >= pFireHi:
		inst, band = true, "report" // ≥55% 证据自足（C_FN 不需要）
	case pFallen <= pFireLo:
		inst, band = false, "no" // ≤45% 不报
	case !identifiable:
		inst, band = false, "indeterminate" // 45-55% 但高度不可判 → 默认不报（§26 守告警可信度）
	default:
		inst, band = cfn > d.p.cFP, "tie" // 45-55% 可判两可 → C_FN 打破平衡（唯一作用窗口）
	}

	if inst {
		if d.fireSinceMs == 0 {
			d.fireSinceMs = nowMs
		}
	} else {
		d.fireSinceMs = 0 // 瞬时条件断开→复位持续（防瞬时噪声）
	}
	fired := d.fireSinceMs != 0 && nowMs-d.fireSinceMs >= d.p.tHoldMs

	return Decision{
		Fire:         fired,
		InstFire:     inst,
		Band:         band,
		PFallen:      pFallen,
		PeopleCount:  rc.PeopleCount,
		CFN:          cfn,
		Margin:       pFallen - pFireHi, // 诊断：距报阈的距离
		Lambda:       lambda,
		Identifiable: identifiable,
		FireSinceMs:  d.fireSinceMs,
	}
}

// ComputeLambda 似然比 Λ_t = Σ_bmask ΨΦ(F,bmask) / Σ_bmask ΨΦ(AtBed,bmask)。
// 当前帧证据能否分开「床边摔」与「睡床上」：informative→≫1；全暗（logPsi+logPhi≡0）→1。
// §26 起 Λ **参与裁决**：Λ≤lambdaInformative=高度不可判 → 默认不报（守告警可信度，非纯诊断）。
func ComputeLambda(js *JointSpace, logPsi, logPhi JointVector) float64 {
	fBase := int(SFallen) * js.bmaskN
	bBase := int(SBed) * js.bmaskN
	fLogs := make([]float64, js.bmaskN)
	bLogs := make([]float64, js.bmaskN)
	for b := 0; b < js.bmaskN; b++ {
		fLogs[b] = logPsi[fBase+b] + logPhi[fBase+b]
		bLogs[b] = logPsi[bBase+b] + logPhi[bBase+b]
	}
	return math.Exp(LogSumExp(fLogs) - LogSumExp(bLogs))
}
