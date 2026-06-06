package zoneengine

// bed_bayesian_scorer.go — bed FSM 的贝叶斯 in-bed 概率累加器。
//
// 模型：递归 log-odds + 簇级合并 + γ tempering + 非对称双阈值。
// 详 owlBack/doc/bed_bayesian_model.md (待写) 或 commit message。
//
// 关键参数：
//   - L_t ∈ [-5, +5]，P_t = sigmoid(L_t) ∈ [0.0067, 0.9933]
//   - LR_S (sleepad 簇)：facility ±2.94 (ln 19) / home ±2.20 (ln 9)
//   - LR_R (radar 簇)：正向 max(+1.45 pose, +1.39 InBed, +0.56 vital)；负向 -1.39 LeftBed
//   - γ schedule：sleepad LeftBed+无 vital 冲突时，[0-60s:1.0, 60-120s:0.5, ≥120s:0.0]
//   - 决策：P>0.70→InBed / P<0.50→LeftBed / 中间维持，维持区 2min 强制 LeftBed (跌床安全偏 LeftBed)
//   - prior: 21-07 local night L_0=+1.39 (P=0.8) / 白天 L_0=-0.85 (P=0.3)，仅初值

import (
	"math"
	"sync"
)

// BedMode 跟 unit_property 对齐。
type BedMode int

const (
	BedModeHome     BedMode = 0 // B2C 大床，10% miss rate (LR=9)
	BedModeFacility BedMode = 1 // B2B hospital bed (默认)，5% miss rate (LR=19)
)

// LR 常量（log-odds 增量；负向不取负号，使用时按需取反）。
const (
	lrSleepadFacility = 2.9444 // ln(19)，5% miss — explicit InBed/LeftBed event
	lrSleepadHome     = 2.1972 // ln(9)，10% miss — explicit InBed/LeftBed event
	lrRadarInBedEvt   = 1.3863 // ln(4)
	lrRadarLeftBedEvt = 1.3863 // ln(4)，使用时取负
	lrRadarPoseLying  = 1.4469 // ln(4.25) — firmware-gated (install_mod=full + bed area)，强证据
	lrRadarVital      = 0.5596 // ln(1.75)
	// vital sustain 没有 explicit InBed event 撑腰 = 弱证据。不能和 InBed event 同档。
	// LR=2 → ln(2)=0.69，即"略偏 in-bed"，不足以单独从 Vacant 推回 InBed。
	lrSleepadVitalOnly = 0.6931 // ln(2)
)

const (
	// L 累加边界 (你拍板的 cap)
	lCap   = +5.0
	lFloor = -5.0

	// 决策阈值（非对称，跌床安全偏 LeftBed）。
	// Home (B2C 大床) 翻身假阳风险更高 → 用 pInBedThresholdHome=0.75 更保守。
	pInBedThresholdFacility = 0.70
	pInBedThresholdHome     = 0.75
	pLeftBedThreshold       = 0.50

	// 时间窗
	// 事件 (InBed/LeftBed) 60s 记忆——firmware 状态翻转事件，跨多次 monitor cycle 仍有效。
	evidenceWindowMs = 60_000
	// vital (HR/RR/pose 等派生数据) 35s 窗——匹配 vital_source 30s freshness + 5s 缓冲。
	// 用 evidenceWindowMs 会让 sleepad 停发后还"幻觉"vital 活着，违反物理（无输入则 vital 应立即 stale）。
	vitalWindowMs         = 35_000
	minuteMs              = 60_000  // 1 contribution per minute per cluster
	gammaPhase1EndMs      = 60_000  // 0-60s: γ=1.0
	gammaPhase2EndMs      = 120_000 // 60-120s: γ=0.5；≥120s: γ=0.0
	maintainZoneTimeoutMs = 120_000 // 维持区 2min 强制 LeftBed
	priorNightHourStart   = 21
	priorNightHourEndExcl = 7
	priorLogOddsNight     = +1.39 // ln(0.8/0.2)
	priorLogOddsDay       = -0.85 // ln(0.3/0.7)

	// shadow：陈旧 LeftBed 不再永久否决 fresh InBed。无 fresh 贡献的分钟指数 leak。producer-first，不入生产 Decision。
	shadowDecayFactor     = 0.63    // floor → |L|<0.5 约 5min
	shadowStandbyBandL    = 0.5     // |shadowL|<此 → bed_status=8
	shadowRefreshWindowMs = 60_000  // sustain 刷新窗

	// 生产 L 衰减（治本：陈旧 InBed/LeftBed event 不再永久 sticky 把 L 钉死；无 fresh 贡献的
	// 分钟 L 指数 leak 回中性 → 陈旧 LeftBed 不再永久否决 fresh InBed）。sim_decay.py 真实数据验证。
	staleEventMs = 5 * 60_000 // 事件超 5min = 陈旧 → event 不再贡献（vital/pose 仍各自按窗活：卡 event 不卡 vital）
	leakFactor   = 0.55       // 无 fresh 贡献的分钟 L *= 0.55（floor → |L|<0.5 约 4min）
	standbyBandL = 0.5        // |L|<此 → Decision=Standby（bed_status=8 待机，床态不确定）

	// neverTs：uninitialized *LastTs sentinel；构造时填，保证 nowMs - neverTs > 任何 freshness 窗（含 nowMs=0）。
	neverTs = int64(-evidenceWindowMs - 1)
)

// BedDecision 决策输出。
type BedDecision int

const (
	BedDecisionMaintain BedDecision = iota // 维持上一态
	BedDecisionInBed
	BedDecisionLeftBed
	BedDecisionStandby // |L|<standbyBandL 中性带：床态不确定（→ bed_status=8）
)

func decisionName(d BedDecision) string {
	switch d {
	case BedDecisionInBed:
		return "InBed"
	case BedDecisionLeftBed:
		return "LeftBed"
	case BedDecisionStandby:
		return "Standby"
	default:
		return "Maintain"
	}
}

// clusterContribState 每个证据簇的 per-minute 贡献记账。
//
// lastMin == 0 时表示从未贡献过（实时 epoch ms / 60_000 永远 > 0）。
type clusterContribState struct {
	lastMin   int64
	lastValue float64
}

// BedBayesianScorer 单 /96 床贝叶斯打分器（线程安全）。
type BedBayesianScorer struct {
	mu   sync.Mutex
	mode BedMode

	L float64 // 当前 log-odds

	// Evidence freshness timestamps（neverTs = 无），便于 t=0 也能区分"从未收到"。
	sleepadInBedLastTs   int64
	sleepadLeftBedLastTs int64
	sleepadVitalLastTs   int64
	radarInBedLastTs     int64
	radarLeftBedLastTs   int64
	radarVitalLastTs     int64
	radarPoseLyingLastTs int64

	// 簇贡献记账
	sleepadContrib clusterContribState
	radarContrib   clusterContribState

	// radarCoversWeight MM covers 结构权重 [0,1]：这台 radar 盖不盖本床。radar 簇贡献乘此，
	// 候选(0.5)半权 / 确定(1.0)全权。与 γ(时间冲突)正交：covers=空间结构，γ=时间退让。
	// 默认 1.0；radar evidence 携 RadarCoversWeight>0 时更新。
	radarCoversWeight float64

	// γ 冲突 tracker
	conflictStartedTs int64 // 0 = 无冲突

	// Multi-source LeftBed sticky：sleepad 和 radar 在 60s 内都 fire LeftBed → 高置信"已离床"。
	// 设为非 0 后，全部 vital/pose 衍生证据被拒绝接收，直到任一源 fire InBed event 清零。
	// 物理依据：双源独立确认离床的联合概率 (0.95²)/(0.05²) ≈ 361 → ln=5.89，几乎不可能误判。
	multiSourceLeftBedTs int64

	// Hysteresis maintain 区超时计时
	maintainStartedTs int64       // 0 = 不在维持区
	lastDecision      BedDecision // 上一次有效决策 (InBed/LeftBed)

	// 生产 L 衰减：上次执行 leak 的分钟（无 fresh 贡献时 L *= leakFactor，每分钟最多一次）
	leakLastMin int64

	// shadow（producer-first，不入生产 Decision）：陈旧 event 指数衰减回中性 + 床区 track 刷新。
	shadowL              float64
	shadowLastMin        int64
	shadowRadarSign      int
	shadowRadarFreshTs   int64
	shadowSleepadSign    int
	shadowSleepadFreshTs int64
}

// NewBedBayesianScorer 默认 facility 模式 + L_0=0（中性）。
// 调 SetMode + InitPriorByHour 配对完成初始化。
func NewBedBayesianScorer() *BedBayesianScorer {
	return &BedBayesianScorer{
		mode:              BedModeFacility,
		L:                 0,
		radarCoversWeight: 1.0,                // 默认全权；MM covers 经 evidence 注入后打折
		lastDecision:      BedDecisionLeftBed, // 默认假定离床，首次 evidence 翻
		// lastMin = -1 sentinel：currentMin epoch ms / 60000 永远 ≥ 0，-1 保证首次评估必 fresh contribute
		sleepadContrib:       clusterContribState{lastMin: -1},
		radarContrib:         clusterContribState{lastMin: -1},
		sleepadInBedLastTs:   neverTs,
		sleepadLeftBedLastTs: neverTs,
		sleepadVitalLastTs:   neverTs,
		radarInBedLastTs:     neverTs,
		radarLeftBedLastTs:   neverTs,
		radarVitalLastTs:     neverTs,
		radarPoseLyingLastTs: neverTs,
		shadowLastMin:        -1,
		shadowRadarFreshTs:   neverTs,
		shadowSleepadFreshTs: neverTs,
	}
}

// SetMode 切 LR 强度档（Facility/Home）。
func (s *BedBayesianScorer) SetMode(mode BedMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
}

// InitPriorByHour 按 unit 本地小时设 L_0 prior。仅创建/重置时调一次。
// hour ∈ [0, 23]。21-06 视为夜间。
func (s *BedBayesianScorer) InitPriorByHour(localHour int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if localHour >= priorNightHourStart || localHour < priorNightHourEndExcl {
		s.L = priorLogOddsNight
	} else {
		s.L = priorLogOddsDay
	}
}

// LogOdds 当前 L 值（测试用）。
func (s *BedBayesianScorer) LogOdds() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.L
}

// Probability sigmoid(L) ∈ [0.0067, 0.9933]。
func (s *BedBayesianScorer) Probability() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return sigmoidLogit(s.L)
}

// Confidence 0-100 整数（bed_state.bed_confidence 直接用）。
func (s *BedBayesianScorer) Confidence() int {
	return int(math.Round(s.Probability() * 100))
}

// Gamma 当前冲突 γ tempering 值（1.0 = 无冲突 / 0.5 / 0.0）。debug 用。
func (s *BedBayesianScorer) Gamma(nowMs int64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.gammaLocked(nowMs)
}

// BedDebug bed_bayesian 决策点的全特征快照（trace + 未来持久化决策日志用）。
type BedDebug struct {
	L            float64 // 当前 log-odds
	P            float64 // sigmoid(L)
	Gamma        float64 // 时间冲突 tempering
	SleepadLR    float64 // sleepad 簇 LR（+ 在床 / - 离床 / 0 无）
	RadarLR      float64 // radar 簇 LR（未乘 γ/covers）
	CoversWeight float64 // MM covers 结构权重
	InBedThresh  float64 // 翻 InBed 的 P 阈（facility 0.70 / home 0.75）

	ShadowL        float64 // shadow log-odds（陈旧衰减）
	ShadowDecision string  // InBed / LeftBed / Standby / maintain
}

// Debug 返回决策点全特征快照。决策不翻转时也能看清"差在哪"（radar 单源 P 没过阈等）。
func (s *BedBayesianScorer) Debug(nowMs int64) BedDebug {
	s.mu.Lock()
	defer s.mu.Unlock()
	th := pInBedThresholdFacility
	if s.mode == BedModeHome {
		th = pInBedThresholdHome
	}
	return BedDebug{
		L: s.L, P: sigmoidLogit(s.L), Gamma: s.gammaLocked(nowMs),
		SleepadLR: s.sleepadClusterLRLocked(nowMs), RadarLR: s.radarClusterLRLocked(nowMs),
		CoversWeight: s.radarCoversWeight, InBedThresh: th,
		ShadowL: s.shadowL, ShadowDecision: s.shadowDecisionNameLocked(),
	}
}

// LastEvidenceTs 7 类证据时间戳的最大值；neverTs / 全空 → 0。subset_invariant stale_bed 用。
func (s *BedBayesianScorer) LastEvidenceTs() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	max := int64(0)
	for _, t := range []int64{
		s.sleepadInBedLastTs, s.sleepadLeftBedLastTs, s.sleepadVitalLastTs,
		s.radarInBedLastTs, s.radarLeftBedLastTs, s.radarVitalLastTs, s.radarPoseLyingLastTs,
	} {
		if t > max {
			max = t
		}
	}
	return max
}

// IngestEvidence 按 SignalEvidence.Source/Kind 派发到对应 OnXxx 方法。
// engine bed 路径统一入口；返回 true 表示这条 evidence 被消费。
//
// 路由表（与现 adapter_sleepace / adapter_radar / adapter_vital 发射对齐）：
//
//	sleepace + enter      → OnSleepadInBed
//	sleepace + leave      → OnSleepadLeftBed
//	sleepace + sustain    → OnSleepadVital  (SleepStage drive)
//	vital    + sustain    → OnSleepadVital  (MonitorVitalSource，仅 sleepad；radar 排除)
//	radar    + enter      → OnRadarInBed
//	radar    + leave      → OnRadarLeftBed
//	radar    + pose_lying → OnRadarPoseLying  (Stage 6 增）
//
// 未列组合返回 false，调用方丢弃。
func (s *BedBayesianScorer) IngestEvidence(ev SignalEvidence) bool {
	nowMs := ev.Ts
	switch ev.Source {
	case "sleepace":
		switch ev.Kind {
		case "enter":
			s.OnSleepadInBed(nowMs)
			return true
		case "leave":
			s.OnSleepadLeftBed(nowMs)
			return true
		case "sustain":
			s.OnSleepadVital(nowMs)
			return true
		}
	case "vital":
		if ev.Kind == "sustain" {
			s.OnSleepadVital(nowMs)
			return true
		}
	case "radar":
		if ev.RadarCoversWeight > 0 {
			s.mu.Lock()
			s.radarCoversWeight = ev.RadarCoversWeight
			s.mu.Unlock()
		}
		switch ev.Kind {
		case "enter":
			s.OnRadarInBed(nowMs)
			return true
		case "leave":
			s.OnRadarLeftBed(nowMs)
			return true
		case "pose_lying":
			s.OnRadarPoseLying(nowMs)
			return true
		}
	}
	return false
}

// === Evidence ingestion ===

// 物理原则：**自不否定，他可否定**。
//   同一设备的 explicit state event（InBed/LeftBed）权威；自家派生信号（vital/pose）不能否定它。
//   只有跨设备的对立证据才能 override（sleepad LeftBed vs radar lying，γ tempering 介入）。
// 推论：
//   - LeftBed event 后，同设备 vital 历史全废（vital 是同一压感数据流，物理上不可能 LeftBed 同时还有 vital）
//   - InBed event 后，同设备 LeftBed 历史废
//   - vital/pose 是衍生信号，被同设备对立 explicit event 压制时不能 revive

// OnSleepadInBed 事件到达 → 立即重算 + contribute。
// Physical: sleepad InBed 权威，自家 LeftBed 历史废；清 multiSourceLeftBed sticky（人回床上了）。
func (s *BedBayesianScorer) OnSleepadInBed(nowMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sleepadInBedLastTs = nowMs
	s.sleepadLeftBedLastTs = neverTs
	s.multiSourceLeftBedTs = 0
	s.shadowPulseSleepadLocked(+1, nowMs)
	s.evaluateAndContributeLocked(nowMs)
}

// OnSleepadLeftBed 事件到达 → 立即重算 + contribute。
// Physical: sleepad LeftBed 权威，自家 InBed + vital 历史全废。
// 若 radar 也最近 fire 过 LeftBed (60s 内) → 双源确认，设 multiSourceLeftBed sticky。
func (s *BedBayesianScorer) OnSleepadLeftBed(nowMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sleepadLeftBedLastTs = nowMs
	s.sleepadInBedLastTs = neverTs
	s.sleepadVitalLastTs = neverTs
	if nowMs-s.radarLeftBedLastTs <= evidenceWindowMs {
		s.multiSourceLeftBedTs = nowMs
	}
	s.shadowPulseSleepadLocked(-1, nowMs)
	s.evaluateAndContributeLocked(nowMs)
}

// OnSleepadVital sleepad monitor 任一压感衍生信号活跃。
// 自家 LeftBed 压制窗内 vital 拒绝接受（衍生不能否定同设备 explicit event）。
// multiSourceLeftBed sticky 中 vital 全部拒绝（无论哪源）。
func (s *BedBayesianScorer) OnSleepadVital(nowMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.multiSourceLeftBedTs > 0 {
		return
	}
	if s.sleepadLeftBedLastTs != neverTs {
		// 自家 LeftBed event sticky 状态下，同设备 vital 全程拒绝（firmware 自相矛盾保护）。
		return
	}
	s.sleepadVitalLastTs = nowMs
	if s.shadowSleepadSign > 0 {
		s.shadowSleepadFreshTs = nowMs
	}
	s.evaluateAndContributeLocked(nowMs)
}

// OnRadarInBed radar firmware bed-area InBed event。
// Physical: radar 自家 LeftBed 历史废；清 multiSourceLeftBed sticky。
func (s *BedBayesianScorer) OnRadarInBed(nowMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.radarInBedLastTs = nowMs
	s.radarLeftBedLastTs = neverTs
	s.multiSourceLeftBedTs = 0
	s.shadowPulseRadarLocked(+1, nowMs)
	s.evaluateAndContributeLocked(nowMs)
}

// OnRadarLeftBed radar firmware bed-area LeftBed event。
// Physical: radar 自家 InBed + vital + pose lying 历史全废。
// 若 sleepad 也最近 fire 过 LeftBed (60s 内) → 双源确认，设 multiSourceLeftBed sticky。
func (s *BedBayesianScorer) OnRadarLeftBed(nowMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.radarLeftBedLastTs = nowMs
	s.radarInBedLastTs = neverTs
	s.radarVitalLastTs = neverTs
	s.radarPoseLyingLastTs = neverTs
	if nowMs-s.sleepadLeftBedLastTs <= evidenceWindowMs {
		s.multiSourceLeftBedTs = nowMs
	}
	s.shadowPulseRadarLocked(-1, nowMs)
	s.evaluateAndContributeLocked(nowMs)
}

// OnRadarVital radar monitor HR/RR alive。自家 LeftBed 压制 + multiSource sticky 拒绝。
func (s *BedBayesianScorer) OnRadarVital(nowMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.multiSourceLeftBedTs > 0 {
		return
	}
	if s.radarLeftBedLastTs != neverTs {
		return
	}
	s.radarVitalLastTs = nowMs
	s.evaluateAndContributeLocked(nowMs)
}

// OnRadarPoseLying radar track pose=lying。自家 LeftBed sticky + multiSource sticky 拒绝。
func (s *BedBayesianScorer) OnRadarPoseLying(nowMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.multiSourceLeftBedTs > 0 {
		return
	}
	if s.radarLeftBedLastTs != neverTs {
		return
	}
	s.radarPoseLyingLastTs = nowMs
	s.evaluateAndContributeLocked(nowMs)
}

// Tick 周期推进（建议 1s）。处理 minute 翻页贡献 + γ 衰减重算。
func (s *BedBayesianScorer) Tick(nowMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evaluateAndContributeLocked(nowMs)
}

// === Decision ===

// Decision 返回当前决策（含 hysteresis + 维持区 2min force LeftBed）。
//
// 调用时 P 已经是 L 的快照；Decision 内部维护 maintainStartedTs，所以**应**在 Tick 之后调。
func (s *BedBayesianScorer) Decision(nowMs int64) BedDecision {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 中性带优先：|L|<standbyBandL → 待机（床态不确定），优先于 In/Left 阈值（sim_decay.py 验证）。
	// 陈旧 event leak 回中性后落此带，避免继续呈现陈旧 LeftBed/InBed；不动 lastDecision（退出带后由阈值/维持区接管）。
	if math.Abs(s.L) < standbyBandL {
		s.maintainStartedTs = 0
		return BedDecisionStandby
	}

	p := sigmoidLogit(s.L)

	pInBedThreshold := pInBedThresholdFacility
	if s.mode == BedModeHome {
		pInBedThreshold = pInBedThresholdHome
	}

	if p > pInBedThreshold {
		s.maintainStartedTs = 0
		s.lastDecision = BedDecisionInBed
		return BedDecisionInBed
	}
	if p < pLeftBedThreshold {
		s.maintainStartedTs = 0
		s.lastDecision = BedDecisionLeftBed
		return BedDecisionLeftBed
	}
	// 维持区
	if s.maintainStartedTs == 0 {
		s.maintainStartedTs = nowMs
	}
	if nowMs-s.maintainStartedTs >= maintainZoneTimeoutMs {
		s.maintainStartedTs = 0
		s.lastDecision = BedDecisionLeftBed
		return BedDecisionLeftBed
	}
	return s.lastDecision // 维持前态
}

// === 内部 (Locked 调用方已持锁) ===

func (s *BedBayesianScorer) evaluateAndContributeLocked(nowMs int64) {
	s.updateConflictLocked(nowMs)

	currentMin := nowMs / minuteMs
	sleepadLR := s.sleepadClusterLRLocked(nowMs)
	radarLR := s.radarClusterLRLocked(nowMs)
	gamma := s.gammaLocked(nowMs)

	// Sleepad 簇：γ 不应用
	s.contributeClusterLocked(currentMin, sleepadLR, &s.sleepadContrib)

	// Radar 簇：先乘 covers 结构权重（两向都打折——盖不盖这床决定证据可信度），
	// 再对正向乘 γ（时间冲突退让；负向 LeftBed event 不打 γ）。
	radarContrib := radarLR * s.radarCoversWeight
	if radarContrib > 0 {
		radarContrib *= gamma
	}
	s.contributeClusterLocked(currentMin, radarContrib, &s.radarContrib)

	// 治本：两簇均无 fresh 贡献（event 陈旧 + vital/pose 失活）时，L 按分钟指数 leak 回中性，
	// 替代旧的"陈旧 LeftBed 每分钟重复累加把 L 钉死 floor"的永久否决 bug。每分钟最多 leak 一次。
	if currentMin > s.leakLastMin {
		s.leakLastMin = currentMin
		if sleepadLR == 0 && radarLR == 0 {
			s.L = clampLogOdds(s.L * leakFactor)
		}
	}

	s.shadowTickLocked(nowMs)
}

// contributeClusterLocked 按 per-minute dedup 贡献。
// 同一分钟内若 cluster 值变化（如事件翻转 InBed↔LeftBed），贡献 delta 修正。
func (s *BedBayesianScorer) contributeClusterLocked(currentMin int64, clusterLR float64, st *clusterContribState) {
	if currentMin > st.lastMin {
		// 新分钟 → fresh contribute 当前值
		s.L = clampLogOdds(s.L + clusterLR)
		st.lastMin = currentMin
		st.lastValue = clusterLR
		return
	}
	// 同分钟，值变 → 贡献 delta（如 +2.94 翻到 −2.94 时补 −5.88）
	if clusterLR != st.lastValue {
		delta := clusterLR - st.lastValue
		s.L = clampLogOdds(s.L + delta)
		st.lastValue = clusterLR
	}
}

// sleepadClusterLRLocked sleepad 簇 max-merge 后 LR。
//
//	LeftBed event fresh AND no vital alive → -LR_S (负向)
//	InBed event fresh   OR  vital alive    → +LR_S (正向)
//	otherwise                              → 0
func (s *BedBayesianScorer) sleepadClusterLRLocked(nowMs int64) float64 {
	lr := lrSleepadFacility
	if s.mode == BedModeHome {
		lr = lrSleepadHome
	}
	// 物理优先级（user 2026-05-23 拍板）：冲击信号 > 微振动信号。
	// LeftBed/InBed 是大幅压力冲击事件，HR/RR 是衍生微振动。前者 sticky（直到对立 event 才清），
	// 后者只在窗内活；如此 firmware 在 LeftBed 后自相矛盾继续推 vital 也无法反咬。
	// 事件按年龄门控：超 staleEventMs 的 InBed/LeftBed event 不再贡献（治本陈旧 sticky 永久否决）。
	// vital 不受此门控，仍按自己的 35s 窗活——睡着的人 InBed event 陈旧后由 vital/pose 维持（卡 event 不卡 vital）。
	leftBedFresh := s.sleepadLeftBedLastTs != neverTs && nowMs-s.sleepadLeftBedLastTs <= staleEventMs
	inBedFresh := s.sleepadInBedLastTs != neverTs && nowMs-s.sleepadInBedLastTs <= staleEventMs
	vitalAlive := nowMs-s.sleepadVitalLastTs <= vitalWindowMs

	if leftBedFresh {
		return -lr
	}
	if inBedFresh {
		return lr
	}
	if vitalAlive {
		return lrSleepadVitalOnly
	}
	return 0
}

// radarClusterLRLocked radar 簇 max-merge 后 LR。
//
//	任一正向证据 (InBed event / pose lying / vital) 活跃 → max(+1.45, +1.39, +0.56) 取活跃中最大
//	否则 LeftBed event fresh → -1.39
//	都没                     → 0
func (s *BedBayesianScorer) radarClusterLRLocked(nowMs int64) float64 {
	// 同 sleepad 原则：radar InBed/LeftBed event 按 staleEventMs 年龄门控；pose/vital 衍生信号各自窗内活
	// （pose_lying 每帧 re-pulse，睡着的人 InBed event 陈旧后由 pose/vital 维持）。
	inBedFresh := s.radarInBedLastTs != neverTs && nowMs-s.radarInBedLastTs <= staleEventMs
	leftBedFresh := s.radarLeftBedLastTs != neverTs && nowMs-s.radarLeftBedLastTs <= staleEventMs
	vitalAlive := nowMs-s.radarVitalLastTs <= vitalWindowMs
	poseFresh := nowMs-s.radarPoseLyingLastTs <= vitalWindowMs

	if leftBedFresh {
		return -lrRadarLeftBedEvt
	}
	var positive float64
	if inBedFresh && lrRadarInBedEvt > positive {
		positive = lrRadarInBedEvt
	}
	if poseFresh && lrRadarPoseLying > positive {
		positive = lrRadarPoseLying
	}
	if vitalAlive && lrRadarVital > positive {
		positive = lrRadarVital
	}
	return positive
}

// 跨设备冲突检测：sleepad LeftBed sticky vs radar 正向证据 → conflict，γ tempering 折扣 radar 正向。
func (s *BedBayesianScorer) updateConflictLocked(nowMs int64) {
	sleepadLeftHappened := s.sleepadLeftBedLastTs != neverTs
	radarPos := s.radarInBedLastTs != neverTs ||
		nowMs-s.radarPoseLyingLastTs <= vitalWindowMs ||
		nowMs-s.radarVitalLastTs <= vitalWindowMs

	if sleepadLeftHappened && radarPos {
		if s.conflictStartedTs == 0 {
			s.conflictStartedTs = nowMs
		}
		return
	}
	s.conflictStartedTs = 0
}

func (s *BedBayesianScorer) gammaLocked(nowMs int64) float64 {
	if s.conflictStartedTs == 0 {
		return 1.0
	}
	dur := nowMs - s.conflictStartedTs
	switch {
	case dur < gammaPhase1EndMs:
		return 1.0
	case dur < gammaPhase2EndMs:
		return 0.5
	default:
		return 0.0
	}
}

// === shadow (producer-first，不入生产 Decision) ===

func (s *BedBayesianScorer) shadowPulseRadarLocked(sign int, nowMs int64) {
	mag := lrRadarInBedEvt
	s.shadowL = clampLogOdds(s.shadowL + float64(sign)*mag)
	s.shadowRadarSign = sign
	s.shadowRadarFreshTs = nowMs
}

func (s *BedBayesianScorer) shadowPulseSleepadLocked(sign int, nowMs int64) {
	mag := lrSleepadFacility
	if s.mode == BedModeHome {
		mag = lrSleepadHome
	}
	s.shadowL = clampLogOdds(s.shadowL + float64(sign)*mag)
	s.shadowSleepadSign = sign
	s.shadowSleepadFreshTs = nowMs
}

func (s *BedBayesianScorer) shadowTickLocked(nowMs int64) {
	currentMin := nowMs / minuteMs
	if currentMin <= s.shadowLastMin {
		return
	}
	s.shadowLastMin = currentMin
	rFresh := s.shadowRadarSign != 0 && nowMs-s.shadowRadarFreshTs <= shadowRefreshWindowMs
	sFresh := s.shadowSleepadSign != 0 && nowMs-s.shadowSleepadFreshTs <= shadowRefreshWindowMs
	if !rFresh && !sFresh {
		s.shadowL = clampLogOdds(s.shadowL * shadowDecayFactor)
	}
}

func (s *BedBayesianScorer) shadowDecisionNameLocked() string {
	if math.Abs(s.shadowL) < shadowStandbyBandL {
		return "Standby"
	}
	th := pInBedThresholdFacility
	if s.mode == BedModeHome {
		th = pInBedThresholdHome
	}
	p := sigmoidLogit(s.shadowL)
	if p > th {
		return "InBed"
	}
	if p < pLeftBedThreshold {
		return "LeftBed"
	}
	return "maintain"
}

// === helpers ===

func sigmoidLogit(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func clampLogOdds(x float64) float64 {
	if x < lFloor {
		return lFloor
	}
	if x > lCap {
		return lCap
	}
	return x
}
