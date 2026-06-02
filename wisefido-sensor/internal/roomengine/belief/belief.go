package belief

// Belief 单实体（一房一人，v1）信念估计器。持有当前 b 与最后推进时刻。
// 非线程安全——调用方（shadow 旁路，每 device 串行）自行串行化。
type Belief struct {
	model  *Model
	b      Vector
	lastTs int64
}

// New 以 model.Prior 初始化。
func New(model *Model) *Belief {
	return &Belief{model: model, b: model.Prior, lastTs: 0}
}

// Vector 当前后验快照。
func (be *Belief) Vector() Vector { return be.b }

// Predict 时间转移 b←A·b（HMM forward 的 prediction）。nSteps 通常 1，缺证据期可多步推进。
func (be *Belief) Predict(nSteps int) {
	for n := 0; n < nSteps; n++ {
		var next Vector
		for j := 0; j < numStates; j++ {
			s := 0.0
			for i := 0; i < numStates; i++ {
				s += be.model.A[i][j] * be.b[i]
			}
			next[j] = s
		}
		next.normalize()
		be.b = next
	}
}

// Observe 观测更新 b←normalize(diag(w^Conf)·b)（correction）。
// Conf=0 / !Fresh → w^0=I 不更新（§2.3 命门，治本 9h）。
func (be *Belief) Observe(o Observation) {
	conf := o.effConf()
	if conf <= 0 {
		return
	}
	w := temper(rawLikelihood(o), conf)
	for s := 0; s < numStates; s++ {
		be.b[s] *= w[s]
	}
	be.b.normalize()
	if o.Ts > be.lastTs {
		be.lastTs = o.Ts
	}
}

// Step 一帧完整推进：先按 dtSec 时间转移，再吸收本帧所有观测。
// dtSec≤0（同帧重入）跳过 Predict。
func (be *Belief) Step(nowMs int64, obs []Observation) {
	if be.lastTs > 0 && nowMs > be.lastTs {
		// 每帧 1 步转移；缺证据的长 gap 也只推 1 步（A 收敛慢=保守，不激进抹掉信念）。
		be.Predict(1)
	}
	for _, o := range obs {
		be.Observe(o)
	}
	if nowMs > be.lastTs {
		be.lastTs = nowMs
	}
}

// Decision 决策枚举。belief≠报警，报警=belief×非对称代价（§2.4，FN cost≫FP）。
type Decision int

const (
	DecisionNone      Decision = iota // 不报
	DecisionFall                      // P(Fallen)>θ_fire，报 Fall
	DecisionUncertain                 // 信念摊平，升级人工/不瞎猜
)

// 决策阈值（非对称）。θ_fire 高（漏报代价远大于误报）；θ_uncertain 低=信念摊平触发线。
const (
	thFire      = 0.55
	thUncertain = 0.40
)

// Decide 当前瞬时决策（无状态快照）。
func (be *Belief) Decide() Decision {
	if be.b[SFallen] > thFire {
		return DecisionFall
	}
	if _, p := be.b.Max(); p < thUncertain {
		return DecisionUncertain
	}
	return DecisionNone
}

// confirmMs Fall 确认窗：P(Fallen)>θ_fire 须连续维持此时长才确认报警。
// 标定：真跌倒维持数分钟，cd2b 类"人返回"在 ~47s 内 P 崩回 θ 下 → 90s 窗把后者过滤掉。
// 语义对齐 gate-list 的 lost-fall wait window（但这里是对 belief 的去抖，非独立计时器）。
const confirmMs = 90_000

// Decider 报警级有状态决策：要求 P(Fallen)>θ_fire 连续维持 confirmMs 才确认 Fall。
// 期间一旦 P 跌回 θ_fire 下（如人返回的反证），计时清零——治本 cd2b 类"fire 后恢复"误报。
type Decider struct {
	fireSince int64 // 0 = 当前不在 θ_fire 之上
}

// Update 喂入当前 belief 与时刻，返回确认级决策。应每帧 Step 后调一次。
func (d *Decider) Update(b Vector, nowMs int64) Decision {
	if b[SFallen] > thFire {
		if d.fireSince == 0 {
			d.fireSince = nowMs
		}
		if nowMs-d.fireSince >= confirmMs {
			return DecisionFall
		}
		return DecisionNone // 维持中，未达确认窗
	}
	d.fireSince = 0 // 跌回 θ 下：恢复/未起，清零
	if _, p := b.Max(); p < thUncertain {
		return DecisionUncertain
	}
	return DecisionNone
}
