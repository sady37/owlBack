package belief

import "math"

// decide.go — 读出与裁决（用户 2026-06-17 裁定：状态转移单阈，不受 P^F 区间限制）。
//   P^F(SFallen) ≥ pFire(0.85) → 报；持续 ≥ T_hold 防瞬时噪声。
//   不再有 55% 三分带区间（≥55报/≤45不报/45-55两可 C_FN）+ Λ gate（高度不可判默认不报）——
//   裁决只看 SFallen 后验是否达阈，与 P^F 落在哪个区间无关、不被 Λ 可判性门控。
//   与 engine lost-fire 的 0.85 阈统一（present/lost 同一状态转移判据）。
//
// 推翻记录（诚实）：§26「55% 三分 + C_FN 仅两可窗 + Λ 作 gate」已废——实证 risk-time 被关在
// 0.45–0.55 窄窗、够不着被上游 pose/z 错报压低的真摔（d5f7-0617 实测 P^F 峰值 0.249 进不了窗）。
// risk-time 如何不受限地融入裁决 = 后续工单（cFN/Λ 暂留 forensic，不门控 fire）。
//
// 形态锚（铁律 [[fall_data_is_artificial_test]]：阈值/曲线全人为数据不可标定，留 oracle）。

// RiskContext 裁决期风险因子（复用 risk_evaluator 同源因子，§8 连续消费非离散档）。
type RiskContext struct {
	AloneContinuousMin float64 // 独居持续分钟（独处→C_FN↑，连续饱和）
	// risktime(夜间)**不在此**：它只缩短 floor 时长兜底阈(Observation.IsRiskTime→tFloorFor)，不进报警阈 C_FN。
	PeopleCount int  // 房内人数（>1 有人代发现→C_FN 折扣，但有下限不归零）
	Disabled    bool // 失能（难自救→C_FN↑）
}

// decideParams §8 代价/持续形态参数（标定锚，非权威值，留 oracle）。
type decideParams struct {
	cFP         float64 // 误报代价基线（护士跑空腿，归一=1）
	cFNBase     float64 // 漏报代价基线（有人在场时）
	aloneGain   float64 // 独居持续对 C_FN 的连续增益上限
	aloneSatMin float64 // 独居增益饱和分钟
	disMult     float64 // 失能倍数
	peopleFloor float64 // 多人折扣下限（不归零：人多仍可能没人注意到摔倒）
	tHoldMs     int64   // 持续窗（防瞬时噪声，§8 ~90s）
}

func defaultDecideParams() decideParams {
	return decideParams{
		cFP: 1.0, cFNBase: 2.0,
		aloneGain: 4.0, aloneSatMin: 30,
		disMult: 1.5, peopleFloor: 0.3,
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
	Fire           bool    // 持续 ≥ T_hold 后的最终触发
	InstFire       bool    // 本帧瞬时判据满足（未含持续）
	Band           string  // 落在哪档：report(≥55) / no(≤45) / tie(45-55可判) / indeterminate(高度不可判)
	PFallen        float64 // P^F_t
	PeopleCount    int     // N_r（人数单源 = TrackCensus.Nr()，已排 ghost；forensic + 拍法 A 守门校验）
	RescuableCount int     // 可救援数 = present-real ∧ S≠Bed（folded）；forensic，当前不门控 fire（doc/cfn-rescuable-design.md）
	CFN            float64 // C_FN(risk)（仅 tie 档参与）
	Margin         float64 // P^F − 报阈（诊断：>0 越确定该报）
	Lambda         float64 // Λ_t 似然比
	Identifiable   bool    // Λ_t > 阈 = 可判（§26 起 **参与裁决**：高度不可判默认不报）
	FireSinceMs    int64   // 瞬时条件起始（持续计时；0=未武装）
}

// fire 单阈（form-anchor，留 oracle）。状态转移：SFallen 后验达阈即报，不受 P^F 区间限制。
const (
	pFire             = 0.85 // P^F(SFallen) ≥ → 报；持续 T_hold 防瞬时
	lambdaInformative = 3.0  // Λ > 此 = 可判（仅 forensic 诊断标记，不再 gate 裁决）
)

// Decider 持有持续计时状态（跨帧）。
type Decider struct {
	p           decideParams
	fireSinceMs int64
}

func NewDecider() *Decider { return &Decider{p: defaultDecideParams()} }

// Step 一帧裁决（§26 55% 三分）。pFallen=P^F_t（filter.PFallen）；lambda=ComputeLambda。
func (d *Decider) Step(nowMs int64, pFallen, lambda float64, rc RiskContext) Decision {
	cfn := d.p.cFN(rc)                         // forensic：risk 融入裁决待后续工单，当前不门控 fire
	identifiable := lambda > lambdaInformative // forensic：Λ 诊断标记，不再 gate
	inst := pFallen >= pFire                   // 状态转移单阈：达阈即报，不受 P^F 区间限制、不 Λ gate
	band := "no"
	if inst {
		band = "report"
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
		Margin:       pFallen - pFire, // 诊断：距报阈的距离
		Lambda:       lambda,
		Identifiable: identifiable,
		FireSinceMs:  d.fireSinceMs,
	}
}

// ArgmaxIsBed 本 track 的 S 9 态边缘后验峰值是否落在 SBed（=躺床）。
//
//	可救援数（doc/cfn-rescuable-design.md）与 sleepad 吸纳的"radar 在床"判定**共用此判据**（单源，防两套 drift）。
func ArgmaxIsBed(sMarg []float64) bool {
	if len(sMarg) == 0 {
		return false
	}
	best, bi := sMarg[0], 0
	for i, v := range sMarg {
		if v > best {
			best, bi = v, i
		}
	}
	return State(bi) == SBed
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
