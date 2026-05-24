package zoneengine

import (
	"net/netip"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Engine 全局 zone 状态机引擎。
//
// 一个 sensor 进程一个 Engine instance，服务所有 card。内部 map: StateKey → zoneInstance。
// 每个 zoneInstance 包含一个 Scorer + StateMachine + ZoneState 元数据。
//
// 线程模型：
//   - Apply (写) / GetState (读) 都加锁
//   - 子集不变量 reconcile 时跨 zoneInstance 需要按固定顺序加锁，避免死锁
//     约定：bed → room → bathroom 顺序加锁
//
// 输出：
//   - listeners []ZoneEventListener 翻转沿回调（RedisAdapter / sensor 派生消费 / metrics 等）
//   - feedbackListeners []FeedbackListener 负反馈事件回调（监控 / 训练样本采集）
type Engine struct {
	rulesStore        *RulesStore
	logger            *zap.Logger
	bedSizeLookup     BedSizeLookup // 注入：根据 bedZoneID 查床型 bucket（legacy bed FSM 用）
	bedModeLookup     BedModeLookup // 注入：根据 bedZoneID 查 BedMode（Bayesian 用）；nil → Facility
	useBedBayesian    bool          // 切到贝叶斯 bed FSM；默认 false 用 legacy Scorer/StateMachine
	listeners         []ZoneEventListener
	feedbackListeners []FeedbackListener

	mu     sync.Mutex
	states map[StateKey]*zoneInstance

	// pendingValidations: 翻转后短窗内的状态，由 Tick 检查 self-contradiction
	pendingValidations map[StateKey]*pendingFlip

	// lastInvariantRepairTs 上次周期巡检（subset_invariant 双向修复）时间戳 ms。
	// 0 表示从未跑过 → 首次 Tick 即触发。
	lastInvariantRepairTs int64
}

// BedModeLookup 按 /96 bed prefix 查 BedMode（Facility=hospital bed / Home=B2C 大床）。
// 实现：UnitPropertyLookup 包一层（bed → /80 unit prefix → unit_property → BedMode）。
type BedModeLookup interface {
	BedMode(bedZoneID string) BedMode
}

// BedSizeLookup bed zone 的床型查询接口；engine 用它决定 Scorer 的 bedBucket。
// 实现侧（如 sensor 主进程）注入查 DB / cache 的具体逻辑；测试侧注入 stub。
type BedSizeLookup interface {
	BedSizeBucket(bedZoneID string) string // "small" / "large"
}

// zoneInstance 单 (zoneType, zoneID) 的运行时实例。物理寻址，多源同 FSM。
//
// bed zone 在 engine.useBedBayesian=true 时使用 bedBayesian 字段，scorer/stateMachine 留 nil；
// 否则使用 legacy scorer/stateMachine 路径。
type zoneInstance struct {
	key          StateKey
	scorer       *Scorer
	stateMachine *StateMachine
	bedBayesian  *BedBayesianScorer // 仅 bed + useBedBayesian=true 时非 nil
	state        ZoneState
	recentTrace  []SignalEvidence // 最近 N 条 evidence，供 ZoneEvent.Trace
}

// pendingFlip 翻转后的"观察期"记录（self_contradiction 检测窗口）。
type pendingFlip struct {
	key          StateKey
	flippedAt    int64
	directionOcc bool // 翻转后 occupied=true
	triggerSrc   string
}

const traceWindowLen = 8 // 每 zone 保留最近 8 条 evidence 供翻转 trace

// NewEngine 创建 engine。bedSizeLookup 必填（即使所有床都默认 standard，也实现一个 stub 返回 "small"）。
func NewEngine(rules *Rules, bedSizeLookup BedSizeLookup, logger *zap.Logger) *Engine {
	if rules == nil {
		rules = DefaultRules()
	}
	return &Engine{
		rulesStore:         NewRulesStore(rules),
		logger:             logger,
		bedSizeLookup:      bedSizeLookup,
		states:             make(map[StateKey]*zoneInstance),
		pendingValidations: make(map[StateKey]*pendingFlip),
	}
}

// AddListener 注册 ZoneEvent 翻转沿监听者。
func (e *Engine) AddListener(l ZoneEventListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, l)
}

// SetUseBedBayesian 切 bed FSM 引擎。enabled=true → ZoneTypeBed 走 BedBayesianScorer；
// modeLookup nil → 所有 bed 默认 Facility。仅启动期调用一次。
func (e *Engine) SetUseBedBayesian(enabled bool, modeLookup BedModeLookup) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.useBedBayesian = enabled
	e.bedModeLookup = modeLookup
}

// AddFeedbackListener 注册 FeedbackEvent 监听者。
func (e *Engine) AddFeedbackListener(l FeedbackListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.feedbackListeners = append(e.feedbackListeners, l)
}

// ReloadRules 替换 rules 并传播到所有已存在 zoneInstance（P1-2 hot reload）。
//
// 直接替换 rulesStore 不够：Scorer / StateMachine 在 NewScorer / NewStateMachine 时
// 持的是 *ZoneRules / *StateMachineRules 指针（旧 Rules 结构体的成员地址），rulesStore
// 替换根 *Rules 不会影响这些已持指针。必须遍历每个 zoneInstance 显式调 UpdateRules。
//
// 运行时状态（score 累积 / 当前 status / timer）全部保留，仅切换规则表。
func (e *Engine) ReloadRules(r *Rules) {
	if r == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rulesStore.Replace(r)
	for key, z := range e.states {
		zoneRules := e.zoneRulesFor(r, key.ZoneType)
		z.scorer.UpdateRules(zoneRules)
		z.stateMachine.UpdateRules(&zoneRules.StateMachine)
	}
}

// GetState 读单 zone 当前状态（拷贝，外部修改不影响内部）。
func (e *Engine) GetState(key StateKey) (ZoneState, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if z, ok := e.states[key]; ok {
		return z.state, true
	}
	return ZoneState{}, false
}

// IsBedOccupied 查 spatial prefix（/96 bed CIDR text）的 bed FSM 是否在床（含 Leaving 过渡）。
// 无该 prefix entry → false（无信息 = 不假设 occupied）。
//
// SleepStageConsumer 用：sleepace 报 SleepStage 时若 bed FSM 已离床（Vacant），视作 OOB 异常
// → drop event + emit device_failure alarm（C 项 deferred 收口；详 [[cardagg_v1_to_v2_migration_audit]] S4）。
func (e *Engine) IsBedOccupied(spatialPrefix string) bool {
	return e.IsZoneOccupied(ZoneTypeBed, spatialPrefix)
}

// IsZoneOccupied 通用查询：(zoneType, spatial prefix) 对应的 FSM 是否在 IsPresent 态。
// 无该 entry → false。home-mode sleepad LeftBed 协同等用：判 room 还在不在用，决定是否 suppress
// sleepad 假阳 LeftBed。
func (e *Engine) IsZoneOccupied(zoneType ZoneType, spatialPrefix string) bool {
	if spatialPrefix == "" {
		return false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	z, ok := e.states[StateKey{ZoneType: zoneType, ZoneID: spatialPrefix}]
	if !ok {
		return false
	}
	return z.state.IsPresent()
}

// Apply 引擎主入口：接收一条 SignalEvidence，可能触发翻转沿。
//
// 流程：
//  1. 找到或创建 zoneInstance
//  2. count_change 走单独路径（直写 zone state.count，不经过 Scorer/StateMachine）
//  3. 其它 evidence → Scorer.Apply → StateMachine.Evaluate
//  4. 若翻转 → 装配 ZoneEvent → fire listeners → 登记 pendingValidation
//  5. 检查 subset_invariant（bed → room reconcile）
func (e *Engine) Apply(ev SignalEvidence) {
	if ev.Ts == 0 {
		ev.Ts = time.Now().UnixMilli()
	}
	now := ev.Ts

	e.mu.Lock()
	z := e.getOrCreate(ev.ZoneType, ev.ZoneID)
	rules := e.rulesStore.Get()

	// count_change 直写
	if ev.Kind == "count_change" {
		prev := z.state
		z.state.Count = ev.Count
		z.state.UpdatedAt = now
		z.state.LastSource = ev.Source
		newState := z.state
		listeners := append([]ZoneEventListener(nil), e.listeners...)
		e.mu.Unlock()
		e.emitEvent(listeners, ZoneEvent{
			ZoneType:   ev.ZoneType,
			ZoneID:     ev.ZoneID,
			Transition: "count_change",
			NewState:   newState,
			PrevState:  prev,
			Trace:      []SignalEvidence{ev},
			Ts:         now,
		})
		return
	}

	// Bayesian bed FSM 路径
	if z.bedBayesian != nil {
		z.appendTrace(ev)
		if !z.bedBayesian.IngestEvidence(ev) {
			e.mu.Unlock()
			return
		}
		prev := z.state
		z.state.LastSource = ev.Source // FE BedConfidence ladder 依赖 LastSource (sleepace=90/radar=60)
		bayesianEvent := e.applyBedBayesianLocked(z, now)
		listeners := append([]ZoneEventListener(nil), e.listeners...)
		subsetEvents := e.maybeReconcileSubset(ev.ZoneType, z.state, now, rules)
		e.mu.Unlock()
		if bayesianEvent != nil {
			bayesianEvent.PrevState = prev
			bayesianEvent.Trace = z.copyTrace()
			e.emitEvent(listeners, *bayesianEvent)
		}
		for _, sub := range subsetEvents {
			e.emitEvent(listeners, sub)
		}
		return
	}

	// Scorer
	newScore, accepted := z.scorer.Apply(ev, now)
	z.appendTrace(ev)
	if !accepted {
		e.mu.Unlock()
		return
	}

	// StateMachine
	transRules := e.zoneStateMachineRules(rules, ev.ZoneType)
	_ = transRules // already injected at scorer create; stateMachine has its own ref
	res := z.stateMachine.Evaluate(newScore, now)

	prev := z.state
	z.state.Score = newScore
	z.state.UpdatedAt = now
	z.state.LastSource = ev.Source

	var flippedEvent *ZoneEvent
	if res.Flipped {
		applyTransitionToState(&z.state, res, now)
		flippedEvent = &ZoneEvent{
			ZoneType:   ev.ZoneType,
			ZoneID:     ev.ZoneID,
			Transition: res.Transition,
			NewState:   z.state,
			PrevState:  prev,
			Trace:      z.copyTrace(),
			Ts:         now,
		}
		// 登记 self_contradiction 观察窗 —— 仅 Occupied 方向（包括 Vacant→Occupied 和 Returned）。
		//
		// 设计非对称（review fix P0-1）：
		//   Occupied 是正向维持态，翻转后需要 score 持续，否则 rollback。
		//   Vacant   是默认安全态，证据自然老化不视为矛盾（防"幽灵床位"）。
		//   Leaving  本身就是 self_contradiction 的"软形态"（timer 会自动确认或取消），不重复登记。
		if res.NewStatus == StatusOccupied {
			e.pendingValidations[z.key] = &pendingFlip{
				key:          z.key,
				flippedAt:    now,
				directionOcc: true,
				triggerSrc:   ev.Source,
			}
		}
	}

	listeners := append([]ZoneEventListener(nil), e.listeners...)
	subsetEvents := e.maybeReconcileSubset(ev.ZoneType, z.state, now, rules)
	e.mu.Unlock()

	if flippedEvent != nil {
		e.emitEvent(listeners, *flippedEvent)
	}
	for _, sub := range subsetEvents {
		e.emitEvent(listeners, sub)
	}
}

// Tick 周期推进。建议每 1s 调用一次。
//
// 作用：
//  1. 因 score 随时间衰减，可能跨过 exit_threshold 但没有新 evidence 触发 → 这里补翻转沿
//  2. self_contradiction 检测：翻转后 window_sec 内 score 未维持 → rollback + 发 FeedbackEvent
func (e *Engine) Tick(nowMs int64) {
	if nowMs == 0 {
		nowMs = time.Now().UnixMilli()
	}
	e.mu.Lock()
	rules := e.rulesStore.Get()
	var flippedEvents []ZoneEvent
	var feedbackEvents []FeedbackEvent

	for _, z := range e.states {
		if z.bedBayesian != nil {
			prev := z.state
			ev := e.applyBedBayesianLocked(z, nowMs)
			if ev != nil {
				ev.PrevState = prev
				ev.Trace = z.copyTrace()
				flippedEvents = append(flippedEvents, *ev)
			}
			continue
		}
		newScore := z.scorer.Score(nowMs)
		res := z.stateMachine.Evaluate(newScore, nowMs)
		if res.Flipped {
			prev := z.state
			z.state.Score = newScore
			z.state.UpdatedAt = nowMs
			applyTransitionToState(&z.state, res, nowMs)
			flippedEvents = append(flippedEvents, ZoneEvent{
				ZoneType:   z.key.ZoneType,
				ZoneID:     z.key.ZoneID,
				Transition: res.Transition,
				NewState:   z.state,
				PrevState:  prev,
				Trace:      z.copyTrace(),
				Ts:         nowMs,
			})
			// review fix P0-1：仅 Occupied 方向登记 self_contradiction（含 Vacant→Occupied + Returned）。
			if res.NewStatus == StatusOccupied {
				e.pendingValidations[z.key] = &pendingFlip{
					key:          z.key,
					flippedAt:    nowMs,
					directionOcc: true,
					triggerSrc:   "decay",
				}
			}
		} else {
			z.state.Score = newScore
			z.state.UpdatedAt = nowMs
		}
	}

	// self_contradiction 检查
	scRule := rules.Feedback.SelfContradiction
	if scRule.WindowSec > 0 {
		for key, pend := range e.pendingValidations {
			ageMs := nowMs - pend.flippedAt
			if ageMs > int64(scRule.WindowSec)*1000 {
				// 窗口过了，移除观察
				delete(e.pendingValidations, key)
				continue
			}
			z, ok := e.states[key]
			if !ok {
				delete(e.pendingValidations, key)
				continue
			}
			// review fix P0-1：仅检查 occupied 方向的 sustain。
			// pendingValidations 只会登记 directionOcc=true 的条目（Apply / Tick 入口已守卫），
			// 保留判断作为防御性兜底（万一有调用方绕开守卫直接 mutate map）。
			if !pend.directionOcc {
				delete(e.pendingValidations, key)
				continue
			}
			score := z.scorer.Score(nowMs)
			sustained := score >= scRule.RequireSustainAt
			if !sustained {
				// rollback：pend.directionOcc=true（守卫保证），翻回 Vacant。
				// 复用 applyTransitionToState 统一字段更新逻辑，避免与正常翻转路径漂移。
				z.stateMachine.Rollback(StatusVacant, nowMs)
				prev := z.state
				z.state.UpdatedAt = nowMs
				z.state.LastSource = "self_contradiction_rollback"
				applyTransitionToState(&z.state, TransitionResult{
					NewStatus:  StatusVacant,
					Transition: TransitionVacant,
				}, nowMs)
				flippedEvents = append(flippedEvents, ZoneEvent{
					ZoneType:   key.ZoneType,
					ZoneID:     key.ZoneID,
					Transition: TransitionVacant,
					NewState:   z.state,
					PrevState:  prev,
					Trace:      z.copyTrace(),
					Ts:         nowMs,
				})
				feedbackEvents = append(feedbackEvents, FeedbackEvent{
					ZoneType:    key.ZoneType,
					ZoneID:      key.ZoneID,
					Reason:      "self_contradiction",
					Affected:    []string{pend.triggerSrc},
					Penalty:     scRule.PenaltyWeight,
					DurationMin: scRule.PenaltyDurMin,
					Ts:          nowMs,
				})
				delete(e.pendingValidations, key)
			}
		}
	}

	// P1-1 修复：周期巡检 subset_invariant（双向修复，含 staleness 判定）
	if rules.Feedback.SubsetInvariant.Enabled && rules.Feedback.SubsetInvariant.RepairIntervalSec > 0 {
		intervalMs := int64(rules.Feedback.SubsetInvariant.RepairIntervalSec) * 1000
		if e.lastInvariantRepairTs == 0 || nowMs-e.lastInvariantRepairTs >= intervalMs {
			repairEvents := e.repairSubsetInvariant(nowMs, rules)
			flippedEvents = append(flippedEvents, repairEvents...)
			e.lastInvariantRepairTs = nowMs
		}
	}

	listeners := append([]ZoneEventListener(nil), e.listeners...)
	feedListeners := append([]FeedbackListener(nil), e.feedbackListeners...)
	e.mu.Unlock()

	for _, ev := range flippedEvents {
		e.emitEvent(listeners, ev)
	}
	for _, fb := range feedbackEvents {
		e.emitFeedback(feedListeners, fb)
	}
}

// repairSubsetInvariant 周期性双向修复 bed⊆room 子集不变量（P1-1）。caller 必须已持锁。
//
// 逻辑：扫所有 bed zone，对每个 bed.IsPresent + room.Vacant 不一致：
//   - bed.UpdatedAt 超过 stale_bed_threshold_sec → bed 是 stale，强制降 bed=Vacant（room 不动）
//   - 否则 → bed 可信，抬升 room=Occupied
//
// 返回收集到的 ZoneEvent 列表，由 Tick 在释放锁后统一 emit（避免在锁内调 listener 致死锁）。
//
// **两阶段实现**：先扫一遍 `e.states` 收集候选 bed keys 到 slice，再二次遍历处理。
// 原因：抬升 room 路径会通过 getOrCreate 写入 `e.states`；若与 `for range e.states` 同时进行
// 即触发 Go map iteration during write 的未定义行为（不 panic 但可能 skip/重复/乱序）。
func (e *Engine) repairSubsetInvariant(nowMs int64, rules *Rules) []ZoneEvent {
	cfg := rules.Feedback.SubsetInvariant
	if !cfg.Enabled || cfg.Mode != "lift_parent" {
		return nil
	}
	staleThresholdMs := int64(cfg.StaleBedThresholdSec) * 1000

	// 阶段 1：扫描收集需要处理的 bed keys（只读，安全）
	var candidates []StateKey
	for k, z := range e.states {
		if k.ZoneType != ZoneTypeBed {
			continue
		}
		if !z.state.IsPresent() {
			continue // bed 已 Vacant，无须处理
		}
		candidates = append(candidates, k)
	}

	// 阶段 2：处理候选项（getOrCreate 可写 e.states，已脱离 iteration）
	var events []ZoneEvent
	for _, bedKey := range candidates {
		z, ok := e.states[bedKey]
		if !ok {
			continue // 罕见：迭代之间被删
		}
		if !z.state.IsPresent() {
			continue // 罕见：阶段 1-2 之间被改
		}
		roomZoneID := deriveRoomZoneIDFromBed(bedKey.ZoneID)
		if roomZoneID == "" {
			continue
		}
		roomKey := StateKey{ZoneType: ZoneTypeRoom, ZoneID: roomZoneID}
		roomInst, ok := e.states[roomKey]
		if !ok {
			roomInst = e.getOrCreate(ZoneTypeRoom, roomZoneID)
		}
		if roomInst.state.IsPresent() {
			continue // 一致，无须修复
		}

		// 不一致：bed.IsPresent + room.Vacant
		// 注意：用 LastEvidenceTs() 不用 z.state.UpdatedAt —— 后者每秒被 Tick 刷新，
		// 永远显示新鲜。真正的 staleness 看是否真有 evidence 进来。
		var lastEvidenceTs int64
		if z.bedBayesian != nil {
			lastEvidenceTs = z.bedBayesian.LastEvidenceTs()
		} else {
			lastEvidenceTs = z.scorer.LastEvidenceTs()
		}
		bedAgeMs := nowMs - lastEvidenceTs
		stale := staleThresholdMs > 0 && bedAgeMs > staleThresholdMs

		if stale {
			// stale 路径：信 room，强制降 bed=Vacant
			prev := z.state
			if z.stateMachine != nil {
				z.stateMachine.ForceSet(StatusVacant, nowMs)
			}
			z.state.UpdatedAt = nowMs
			z.state.LastSource = "invariant_repair_stale_bed"
			applyTransitionToState(&z.state, TransitionResult{
				NewStatus:  StatusVacant,
				Transition: TransitionVacant,
			}, nowMs)
			// DEBUG: subset_invariant 强制 Vacant trace
			if e.logger != nil {
				e.logger.Info("zone bed flip (subset_invariant stale_bed)",
					zap.String("zone_id", bedKey.ZoneID),
					zap.Int64("bed_age_ms", bedAgeMs),
					zap.Int64("stale_threshold_ms", staleThresholdMs),
					zap.Int64("ts_ms", nowMs),
				)
			}
			events = append(events, ZoneEvent{
				ZoneType:   ZoneTypeBed,
				ZoneID:     bedKey.ZoneID,
				Transition: TransitionVacant,
				NewState:   z.state,
				PrevState:  prev,
				Trace:      []SignalEvidence{{Source: "invariant_repair", Kind: "stale_bed", Ts: nowMs}},
				Ts:         nowMs,
			})
		} else {
			// 非 stale：信 bed，抬升 room
			prev := roomInst.state
			roomInst.stateMachine.ForceSet(StatusOccupied, nowMs)
			roomInst.state.UpdatedAt = nowMs
			roomInst.state.LastSource = "invariant_repair_lift_room"
			applyTransitionToState(&roomInst.state, TransitionResult{
				NewStatus:  StatusOccupied,
				Transition: TransitionOccupied,
			}, nowMs)
			events = append(events, ZoneEvent{
				ZoneType:   ZoneTypeRoom,
				ZoneID:     roomZoneID,
				Transition: TransitionOccupied,
				NewState:   roomInst.state,
				PrevState:  prev,
				Trace:      []SignalEvidence{{Source: "invariant_repair", Kind: "lift_room", Ts: nowMs}},
				Ts:         nowMs,
			})
		}
	}
	return events
}

// getOrCreate 取或建 zoneInstance。caller 必须已持锁。
// 物理寻址：(zone_type, zone_id) 唯一标识一个 FSM；同一 /96 床上多个 device 喂同一 instance。
func (e *Engine) getOrCreate(zt ZoneType, zid string) *zoneInstance {
	key := StateKey{ZoneType: zt, ZoneID: zid}
	if z, ok := e.states[key]; ok {
		return z
	}
	rules := e.rulesStore.Get()
	zoneRules := e.zoneRulesFor(rules, zt)
	z := &zoneInstance{
		key: key,
		state: ZoneState{
			ZoneType: zt,
			ZoneID:   zid,
		},
	}
	if zt == ZoneTypeBed && e.useBedBayesian {
		z.bedBayesian = NewBedBayesianScorer()
		if e.bedModeLookup != nil {
			z.bedBayesian.SetMode(e.bedModeLookup.BedMode(zid))
		}
		// 注：InitPriorByHour 由 wiring 注入 UnitContextLookup 后单独调用，这里不依赖时区。
	} else {
		bedBucket := ""
		if zt == ZoneTypeBed && e.bedSizeLookup != nil {
			bedBucket = e.bedSizeLookup.BedSizeBucket(zid)
		}
		z.scorer = NewScorer(zoneRules, bedBucket)
		z.stateMachine = NewStateMachine(&zoneRules.StateMachine)
	}
	e.states[key] = z
	return z
}

// zoneRulesFor 按 zone_type 返回对应的 ZoneRules。
func (e *Engine) zoneRulesFor(r *Rules, zt ZoneType) *ZoneRules {
	switch zt {
	case ZoneTypeBed:
		return &r.Bed
	case ZoneTypeRoom:
		return &r.Room
	case ZoneTypeBathroom:
		return &r.Bathroom
	default:
		return &r.Room
	}
}

func (e *Engine) zoneStateMachineRules(r *Rules, zt ZoneType) *StateMachineRules {
	return &e.zoneRulesFor(r, zt).StateMachine
}

// maybeReconcileSubset bed⊆room：bed.IsPresent=true 但 room=Vacant → 抬升 room=Occupied。
// 仅 lift_parent mode 实现（user-confirmed）；drop_child mode TODO（保守，不推荐）。
//
// 注意 Leaving 也算 IsPresent=true（老人正在离床仍在房间），同样触发抬升。
//
// caller 必须已持锁；返回需要 emit 的事件（caller 在释放锁后 emit）。
func (e *Engine) maybeReconcileSubset(zt ZoneType, changed ZoneState, now int64, r *Rules) []ZoneEvent {
	if !r.Feedback.SubsetInvariant.Enabled {
		return nil
	}
	if r.Feedback.SubsetInvariant.Mode != "lift_parent" {
		return nil
	}
	// bed.IsPresent (Occupied 或 Leaving) 时触发抬升 room
	if zt != ZoneTypeBed || !changed.IsPresent() {
		return nil
	}
	roomZoneID := deriveRoomZoneIDFromBed(changed.ZoneID)
	if roomZoneID == "" {
		return nil
	}
	roomKey := StateKey{ZoneType: ZoneTypeRoom, ZoneID: roomZoneID}
	roomInst, ok := e.states[roomKey]
	if !ok {
		roomInst = e.getOrCreate(ZoneTypeRoom, roomZoneID)
	}
	if roomInst.state.IsPresent() {
		return nil // room 已 IsPresent (Occupied / Leaving)，无需调整
	}
	prev := roomInst.state
	roomInst.stateMachine.ForceSet(StatusOccupied, now)
	roomInst.state.Status = StatusOccupied
	roomInst.state.Occupied = true
	roomInst.state.SinceTs = now
	roomInst.state.UpdatedAt = now
	roomInst.state.LastSource = "subset_invariant_from_bed"
	roomInst.state.LastEnterTs = now
	roomInst.state.LeavingSince = 0
	if roomInst.state.Count < 1 {
		roomInst.state.Count = 1
	}
	return []ZoneEvent{{
		ZoneType:   ZoneTypeRoom,
		ZoneID:     roomZoneID,
		Transition: TransitionOccupied,
		NewState:   roomInst.state,
		PrevState:  prev,
		Trace:      []SignalEvidence{{Source: "subset_invariant", Kind: "enter", Ts: now}},
		Ts:         now,
	}}
}

// deriveRoomZoneIDFromBed 从 bed 的 /96 CIDR 派生 room 的 /88 CIDR。
// bed.ZoneID 形如 "fd00:0:3:111:3:101::/96" → room "fd00:0:3:111:3:101::/88"
func deriveRoomZoneIDFromBed(bedCIDR string) string {
	if bedCIDR == "" {
		return ""
	}
	p, err := netip.ParsePrefix(bedCIDR)
	if err != nil {
		return ""
	}
	if p.Bits() < 88 {
		return ""
	}
	return netip.PrefixFrom(p.Addr(), 88).Masked().String()
}

// applyBedBayesianLocked 评估 Bayesian bed FSM 并把 Decision 投射到 ZoneState；
// 若 Status 翻转返回 ZoneEvent（caller 补 PrevState/Trace 后 emit），否则 nil。caller 必须持锁。
func (e *Engine) applyBedBayesianLocked(z *zoneInstance, nowMs int64) *ZoneEvent {
	z.bedBayesian.Tick(nowMs)
	decision := z.bedBayesian.Decision(nowMs)

	prevStatus := z.state.Status
	var newStatus ZoneStatus
	var transition string
	switch decision {
	case BedDecisionInBed:
		newStatus = StatusOccupied
		transition = TransitionOccupied
	case BedDecisionLeftBed:
		newStatus = StatusVacant
		transition = TransitionVacant
	default:
		newStatus = prevStatus
	}

	z.state.Score = z.bedBayesian.Confidence()
	z.state.UpdatedAt = nowMs
	z.state.BayesianLogOdds = z.bedBayesian.LogOdds()
	z.state.BayesianProb = z.bedBayesian.Probability()
	z.state.BayesianGamma = z.bedBayesian.Gamma(nowMs)

	if newStatus == prevStatus {
		return nil
	}
	applyTransitionToState(&z.state, TransitionResult{
		NewStatus:  newStatus,
		Transition: transition,
	}, nowMs)

	return &ZoneEvent{
		ZoneType:   ZoneTypeBed,
		ZoneID:     z.key.ZoneID,
		Transition: transition,
		NewState:   z.state,
		Ts:         nowMs,
	}
}

// applyTransitionToState 把 StateMachine TransitionResult 投射到 ZoneState。
// 处理 4 种 transition：occupied / vacant / leaving / returned 各自的字段更新。
func applyTransitionToState(s *ZoneState, res TransitionResult, nowMs int64) {
	s.Status = res.NewStatus
	s.Occupied = (res.NewStatus != StatusVacant) // 兼容字段
	s.SinceTs = nowMs
	switch res.Transition {
	case TransitionOccupied, TransitionReturned:
		s.LastEnterTs = nowMs
		s.LeavingSince = 0
		if s.Count < 1 {
			s.Count = 1
		}
	case TransitionLeaving:
		s.LeavingSince = nowMs
		// Leaving 期间 Count 保留（仍 IsPresent），不归零
	case TransitionVacant:
		s.LastExitTs = nowMs
		s.LeavingSince = 0
		s.Count = 0
	}
}

func (e *Engine) emitEvent(listeners []ZoneEventListener, ev ZoneEvent) {
	for _, l := range listeners {
		l.OnZoneEvent(ev)
	}
}

func (e *Engine) emitFeedback(listeners []FeedbackListener, fb FeedbackEvent) {
	for _, l := range listeners {
		l.OnFeedback(fb)
	}
}

// appendTrace zoneInstance 内部循环缓冲，保留最近 N 条 evidence。
func (z *zoneInstance) appendTrace(ev SignalEvidence) {
	if len(z.recentTrace) >= traceWindowLen {
		z.recentTrace = z.recentTrace[1:]
	}
	z.recentTrace = append(z.recentTrace, ev)
}

func (z *zoneInstance) copyTrace() []SignalEvidence {
	out := make([]SignalEvidence, len(z.recentTrace))
	copy(out, z.recentTrace)
	return out
}

// StaticBedSizeLookup 固定返回单一 bucket 的 lookup（测试用，或所有床都默认 standard 的简化）。
type StaticBedSizeLookup struct{ Bucket string }

func (s StaticBedSizeLookup) BedSizeBucket(string) string {
	if s.Bucket == "" {
		return "small"
	}
	return s.Bucket
}
