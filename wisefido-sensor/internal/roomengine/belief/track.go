package belief

import "math"

// track.go — DBN 完备性 P1：per-track「存在/真伪」隐链 T_t（belief_dbn_completeness.md §3①、§4）。
//
// Room 层 9 态全是房间级姿态，"这条 blip 是真人还是反射 / 在不在场"是**每条 track 自己的隐过程**，
// 不在 S 里。本会话所有 lost_track 修复（method-2 ghost-last / 门区 exit / ghost-skip）都是把这条
// 缺失隐链手补成 adapter 里的 if。P1 把它建成显式 HMM：T_t 自带转移阵 A_T，由 verdict/几何/事件/
// 缺测发射驱动。**P1 只读 shadow——推断 + log，不驱动 Room 层、不 fire**。
//
// 结构核心（替代的 gate）：
//   - method-2「失锁前全 ghost 不 fire」      = A_T: Ghost→Lost ≈ 0（靠状态记忆，非外部 if）
//   - 门区 exit 推断「门区+np=0→exit」         = absent 发射按 last-geom 条件 Real(门区)→JustLeft
//   - ghost-vanish→None vs real-lost→Lost 的判别 = 全靠消失前的 T 先验（Real 还是 Ghost）

// TState per-track 隐状态。
type TState int

const (
	TNone     TState = iota // 不存在 / 无 track
	TReal                   // 真人在场
	TGhost                  // ghost / 反射在场
	TJustLeft               // 真人刚从门离开（过渡，→收敛 None）
	TLost                   // 真人 track 在非门区丢失（候选倒地）← 喂 Room 层 Fallen 的前置
)

const numTStates = 5

var tStateLabel = [numTStates]string{"None", "Real", "Ghost", "JustLeft", "Lost"}

func (t TState) String() string { return tStateLabel[t] }

// TVector Δ(T) 上的概率向量，Σ=1。
type TVector [numTStates]float64

func (v *TVector) normalize() {
	sum := 0.0
	for _, p := range v {
		sum += p
	}
	if sum <= 0 {
		for i := range v {
			v[i] = 1.0 / numTStates
		}
		return
	}
	for i := range v {
		v[i] /= sum
	}
}

// Max 返回最大分量及其状态。
func (v TVector) Max() (TState, float64) {
	best, bestP := TNone, v[0]
	for i := 1; i < numTStates; i++ {
		if v[i] > bestP {
			best, bestP = TState(i), v[i]
		}
	}
	return best, bestP
}

// P 取某态后验。
func (v TVector) P(t TState) float64 { return v[t] }

// TObsKind track 层观测种类。
type TObsKind int

const (
	TObsPresent  TObsKind = iota // 本帧有 track 检出（带 ghostness + geom）
	TObsAbsent                   // 本帧无检出（缺测发射，带 last-geom）
	TObsExit                     // firmware ExitRoom 事件（强推 →JustLeft）
	TObsPeerLive                 // 同房（单床）对等雷达此刻见真人 → 本 track 更可能是重影/重复（压 Lost）
)

// TObservation track 层观测。Ghostness/Geom 仅对 TObsPresent；Geom（last）对 TObsAbsent。
type TObservation struct {
	Kind      TObsKind
	Ghostness float64 // [0,1]，仅 TObsPresent：0=纯真人 1=纯 ghost
	Geom      Geom    // TObsPresent=当前 geom；TObsAbsent=消失前 last geom
	Conf      float64
	Ts        int64
	Fresh     bool
}

func (o TObservation) effConf() float64 {
	if !o.Fresh {
		return 0
	}
	return o.Conf
}

// transitionPropensityT A_T 转移倾向（行未归一，init 归一）。白盒，每格可解释。
//
// 关键格（替代 gate）：
//
//   - Ghost→Lost = floor：ghost 闪灭绝不变"丢失候选倒地"（method-2 结构化）。ghost 消散走 →None。
//
//   - Lost 近吸收（83）：真人丢失不自愈；recapture→Real / 门事件→JustLeft 由观测主动反证才出。
//
//   - Real→Lost / Real→JustLeft 给中等值：真人持续在场为主，丢失/离开由缺测发射 + geom 决定去向。
//
//     →None →Real →Ghost →JustLeft →Lost
var transitionPropensityT = [numTStates][numTStates]float64{
	TNone:     {85, 7, 7, 0.5, 0.5}, // 多数维持；出生到 Real/Ghost 靠 present 发射
	TReal:     {0, 88, 1, 5, 6},     // 持续在场为主；离开/丢失靠缺测发射分流
	TGhost:    {33, 1, 63, 0, 0.1},  // ghost 消散到 None；→Lost=floor（结构核心）
	TJustLeft: {88, 6, 0, 6, 0},     // 收敛 None；可能被重新检出回 Real
	TLost:     {2, 7, 0, 11, 80},    // 近吸收；recapture→Real / 门→JustLeft 才出
}

// tlk 构造 T 层 likelihood：默认全 1（中性），覆盖 set 中状态权重，下限 likelihoodFloor。
func tlk(set map[TState]float64) TVector {
	var v TVector
	for i := range v {
		v[i] = 1.0
	}
	for s, w := range set {
		if w < likelihoodFloor {
			w = likelihoodFloor
		}
		v[s] = w
	}
	return v
}

// ttemper 用 Conf 退火 w^Conf；Conf=0→I 不更新（缺证据命门，同 Room 层）。
func ttemper(w TVector, conf float64) TVector {
	if conf <= 0 {
		return tlk(nil)
	}
	if conf >= 1 {
		return w
	}
	var out TVector
	for i := range w {
		out[i] = math.Pow(w[i], conf)
	}
	return out
}

// rawTLikelihood 由观测产 T 层 likelihood。geom/ghostness 作条件。
func rawTLikelihood(o TObservation) TVector {
	switch o.Kind {
	case TObsPresent:
		// 有检出 → 偏好"在场"态（Real/Ghost）；ghostness 在二者间连续分配。
		// g=0 全真人，g=1 全 ghost。不在场态（None/JustLeft/Lost）压低。
		g := o.Ghostness
		if g < 0 {
			g = 0
		} else if g > 1 {
			g = 1
		}
		realW := 1.0 + 4.0*(1.0-g) // [1,5]
		ghostW := 1.0 + 4.0*g      // [1,5]
		return tlk(map[TState]float64{
			TReal: realW, TGhost: ghostW,
			TNone: 0.2, TJustLeft: 0.2, TLost: 0.2,
		})

	case TObsAbsent:
		// 无检出 → 压低在场态（Real/Ghost）。**关键纪律：absent 不区分 Lost vs None**——
		// 二者都是"无检出"，P(absent|Lost)≈P(absent|None)≈P(absent|JustLeft)，发射对它们等价。
		// 谁该是 Lost 谁该是 None **只由消失前 T 先验经 A_T 决定**（Ghost 先验→A_T 通 None，
		// 不通 Lost；Real 先验→A_T 通 Lost/JustLeft，不通 None）= 缺失段①的结构判别。
		// last-geom 只用来在两个"真人离场"去向间分流：门区→JustLeft，开阔地板→Lost。
		switch o.Geom {
		case GeomInEnter:
			return tlk(map[TState]float64{
				TJustLeft: 3, TNone: 3, TLost: 0.4,
				TReal: 0.2, TGhost: 0.3,
			})
		case GeomInBed:
			// 床区消失（被遮挡/躺下）：偏 None，绝不 Lost（床上不是倒地候选）。
			return tlk(map[TState]float64{
				TNone: 3, TJustLeft: 1, TLost: 0.2,
				TReal: 0.3, TGhost: 0.3,
			})
		default: // OpenFloor / Toilet / Unknown
			return tlk(map[TState]float64{
				TLost: 3, TNone: 3, TJustLeft: 0.5,
				TReal: 0.2, TGhost: 0.3,
			})
		}

	case TObsExit:
		// firmware ExitRoom：真人从门离开 → 强推 JustLeft，压 Lost。
		return tlk(map[TState]float64{
			TJustLeft: 6, TNone: 2, TLost: 0.1, TReal: 0.3,
		})

	case TObsPeerLive:
		// 同房（单床）对等雷达此刻见真人 → 房里那个真人是别台的 track，本 track 更可能是它的
		// 重影/重复，绝非独自倒地 → 压 TLost，偏 Ghost/None。单床闸在 adapter 已把关（仅单床房发此观测）；
		// 双床房不发，避免把另一位老人当"对等"误压（漏报）。固件确认跌倒走 Room 层 ObsFirmwareFall，不受此影响。
		return tlk(map[TState]float64{
			TLost: 0.15, TGhost: 2, TNone: 2, TReal: 0.5,
		})
	}
	return tlk(nil)
}

// TrackBelief per-track T 层信念滤波。镜像 Room 层 Belief：Predict(A_T·b) + Observe(diag(w^conf)·b)。
// 非线程安全——调用方（shadow，per-room 串行）自行串行化。
type TrackBelief struct {
	a      [numTStates][numTStates]float64
	b      TVector
	lastTs int64
}

// NewTrackBelief 出生时初始化：偏 None，少量 Real/Ghost 待首帧 present 发射拉开。
func NewTrackBelief() *TrackBelief {
	tb := &TrackBelief{}
	tb.a = rowNormalizedT(transitionPropensityT)
	tb.b = TVector{TNone: 0.7, TReal: 0.2, TGhost: 0.1}
	tb.b.normalize()
	return tb
}

// Vector 当前 T 后验快照。
func (tb *TrackBelief) Vector() TVector { return tb.b }

// Predict 时间转移 b←A_T·b。
func (tb *TrackBelief) Predict(nSteps int) {
	for n := 0; n < nSteps; n++ {
		var next TVector
		for j := 0; j < numTStates; j++ {
			s := 0.0
			for i := 0; i < numTStates; i++ {
				s += tb.a[i][j] * tb.b[i]
			}
			next[j] = s
		}
		next.normalize()
		tb.b = next
	}
}

// Observe 观测更新 b←normalize(diag(w^conf)·b)。Conf=0→不更新。
func (tb *TrackBelief) Observe(o TObservation) {
	conf := o.effConf()
	if conf <= 0 {
		return
	}
	w := ttemper(rawTLikelihood(o), conf)
	for s := 0; s < numTStates; s++ {
		tb.b[s] *= w[s]
	}
	tb.b.normalize()
	if o.Ts > tb.lastTs {
		tb.lastTs = o.Ts
	}
}

// Step 一帧推进：先 1 步时间转移（lastTs 后），再吸收本帧观测。
func (tb *TrackBelief) Step(nowMs int64, obs []TObservation) {
	if tb.lastTs > 0 && nowMs > tb.lastTs {
		tb.Predict(1)
	}
	for _, o := range obs {
		tb.Observe(o)
	}
	if nowMs > tb.lastTs {
		tb.lastTs = nowMs
	}
}

func rowNormalizedT(p [numTStates][numTStates]float64) [numTStates][numTStates]float64 {
	var a [numTStates][numTStates]float64
	for i := 0; i < numTStates; i++ {
		sum := 0.0
		for j := 0; j < numTStates; j++ {
			sum += p[i][j]
		}
		if sum <= 0 {
			a[i][i] = 1.0
			continue
		}
		for j := 0; j < numTStates; j++ {
			a[i][j] = p[i][j] / sum
		}
	}
	return a
}
