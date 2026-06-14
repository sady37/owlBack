package belief

import "math"

// decide.go — §8 读出与裁决（期望损失主框架，§B）。
//   fire ⟺ P^F·C_FN(risk) > (1-P^F)·C_FP   持续 ≥ T_hold
// 不靠 argmax（P^F 不需赢，只需代价翻转，C §7）；不可判（Λ→1）**不写独立分支**——
// 同一不等式统一处理：证据两可→P^F 中等，高 C_FN（独处）翻转。Λ_t 纯诊断（probe 暴露），绝不作 gate。
//
// 形态锚（铁律 [[fall_data_is_artificial_test]]：跌倒数据全人为，C_FN 曲线取值无法标定，
// 只设保守形态——连续、各风险因子单调、多人折扣有下限不归零——+ 显式非权威，留 oracle）。

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
	InstFire     bool    // 本帧瞬时满足不等式（未含持续）
	PFallen      float64 // P^F_t
	CFN          float64 // C_FN(risk)
	Margin       float64 // P^F·C_FN − (1−P^F)·C_FP（>0=瞬时该发）
	Lambda       float64 // Λ_t 似然比（纯诊断：informative≫1 / 全暗→1）
	Identifiable bool    // Λ_t > 阈（诊断标志，**不参与 fire 决策**）
	FireSinceMs  int64   // 瞬时条件起始（持续计时；0=未武装）
}

// lambdaInformative Λ_t 可判诊断阈（仅 forensic 标注，非 gate）。
const lambdaInformative = 3.0

// Decider 持有持续计时状态（跨帧）。
type Decider struct {
	p           decideParams
	fireSinceMs int64
}

func NewDecider() *Decider { return &Decider{p: defaultDecideParams()} }

// Step 一帧裁决。pFallen=§8 P^F_t（filter.PFallen）；lambda=ComputeLambda（诊断）。
func (d *Decider) Step(nowMs int64, pFallen, lambda float64, rc RiskContext) Decision {
	cfn := d.p.cFN(rc)
	margin := pFallen*cfn - (1-pFallen)*d.p.cFP
	inst := margin > 0

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
		PFallen:      pFallen,
		CFN:          cfn,
		Margin:       margin,
		Lambda:       lambda,
		Identifiable: lambda > lambdaInformative,
		FireSinceMs:  d.fireSinceMs,
	}
}

// ComputeLambda §8 似然比 Λ_t = Σ_bmask ΨΦ(F,bmask) / Σ_bmask ΨΦ(AtBed,bmask)。
// 当前帧证据能否分开「床边摔」与「睡床上」：informative→≫1；全暗（logPsi+logPhi≡0）→1。
// 纯诊断量（A 阶段3 立场①：不参与 fire）。
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
