// target_state_aggregator.go — sensor 端 Target/RoomState 派生字段聚合器（card_display_projector Task 2）。
//
// 职责（per-cardID 累加 4 个字段）：
//   - Target.LastActiveTs                       — radar walk≥2m OR walk_duration≥6s + total_people==1
//   - Target.weak_biometric_signal              — 30min lazy 滑窗 weak/HR/RR/ApneaH alarm 累加 0-100
//   - Target.visitor_start_ts / has_visitor_today / today_max_visitor_min
//                                               — TotalPeople≥2 持续≥5min 算 visitor（仅 unit_type=Private）
//   - RoomState.StandingContinuousMin           — stand_duration≥55s 累+1 封顶 8（仅 TotalPeople==1）
//
// 架构（详 [[target_state_aggregator]] design doc）：
//   Aggregator = 纯 state holder（不 publish）；
//   StreamPublisher 60s ticker 主动 pull GetSnapshot 合并进 RoomState/Target publish；
//   零依赖反向（aggregator 不引用 zoneengine 包）—— 通过 ZoneEvent listener 单向 push。
//
// Escalation alarm（WeakBiometricSignal score 跨 80 提级 Critical）由 aggregator 自己直接 publishAlarm
// 到 iot:alarm:stream，不走 StreamPublisher（不同流，alarm pipeline 自身多源）。
//
// P2 scaffold：struct + lifecycle + 接口面；4 个累加器内部逻辑 stub（P3/P4 填）。

package service

import (
	"context"
	"sync"
	"time"

	"owl-common/card"
	"owl-common/observation"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// AggregatorPublisher 允许 aggregator 反向触发 escalation alarm（不依赖 zoneengine 包）。
// 实现方持有 iot:alarm:stream client。
type AggregatorPublisher interface {
	// PublishEscalationAlarm 在 weakBio score 跨 80 上升沿调用。
	// alarmType 当前固定 "WeakBiometricSignal"；level=AlarmLevelCrit；contributorDeviceAddr 可空。
	PublishEscalationAlarm(ctx context.Context, cardID, alarmType string, score int, contributorDeviceAddr string) error
}

// ZoneEventSnapshot aggregator 关心的最小 ZoneEvent 子集（解耦 zoneengine 类型）。
// StreamPublisher 把 ZoneEvent 翻译成这个 push 给 aggregator。
type ZoneEventSnapshot struct {
	CardID      string
	ZoneID      string
	TotalPeople int
	UpdatedAtMs int64
}

// MonitorFrame 单帧 radar/sleepad observation（aggregator 关心的统计字段子集）。
// 由 monitor consumer 翻译成此结构后 push。
type MonitorFrame struct {
	CardID                 string
	DeviceAddr             string
	TsMs                   int64
	WalkDistanceMeters     int    // FieldWalkDistance
	WalkDurationSec        int    // FieldWalkDuration
	StandDurationSec       int    // FieldStandDuration
	MultiPersonDurationSec int    // FieldMultiPersonDuration（备用，当前主信号源用 ZoneEvent.TotalPeople）
	Pose                   int    // FieldPose
	WeakBioSignalRaw       int    // FieldWeakBiometricSignal 0-3
}

// AlarmEventSnapshot iot:alarm:stream 单条事件的 aggregator 关心子集。
type AlarmEventSnapshot struct {
	CardID     string
	AlarmType  string  // "WeakBiometricSignal" / "HeartRateAlert" / "RespRateAlert" / "ApneaHypopnea"
	TsMs       int64
	Producer   string  // 跳过自家产的 escalation alarm 防 loop
	RawValue   int     // 仅 WeakBiometricSignal 携带（state×20: 0/20/40/60）
}

// (Snapshot 类型已收口；外部 caller 用 GetSnapshot multi-return；这里不再暴露 struct)

// TargetStateAggregator 主结构。
//
// 输入：3 条 push channel（monitor / alarm / zone event）；
// 输出：GetSnapshot pull（StreamPublisher 60s 用）+ PublishEscalationAlarm（即时）。
type TargetStateAggregator struct {
	publisher AggregatorPublisher
	logger    *zap.Logger
	redis     *redislib.Client // 订阅 monitor / alarm 流用；P2 scaffold 允许 nil（无订阅）

	mu     sync.RWMutex
	accums map[string]*cardAccumulator // key = cardID (INET CIDR text)

	// Push channels（main loop 单 reader 防 race）
	monitorCh chan MonitorFrame
	alarmCh   chan AlarmEventSnapshot
	zoneCh    chan ZoneEventSnapshot
}

// cardAccumulator per-cardID 累加状态。
type cardAccumulator struct {
	cardID   string
	unitID   string // /80 prefix；visitor 跨卡写用
	unitType int    // 仅 Private(1) 才计 visitor

	// 4 个 sub-accumulator
	lastActive lastActiveState
	standing   standingState
	visitor    visitorState
	weakBio    weakBioWindow

	// 来自 ZoneEvent 的缓存（gate 用）
	totalPeople int
	lastZoneTs  int64

	// publish 调度
	lastPublishTs int64
	dirty         bool
}

type lastActiveState struct {
	lastActiveTs int64 // 写到 Target.LastActiveTs；60s 节流
}

type standingState struct {
	// 当前累加值 0..8；坐/走/躺/多人时 reset 0
	continuousMin int
	// 上次 +1 的时刻；避免一分钟内多次跳号
	lastIncrementTs int64
}

type visitorState struct {
	multiSegmentStartTs int64  // 当前持续多人段起始 (0 = 不在段中)
	visitorStartTs      int64  // 当日首次跨 5min 阈值时刻 → Target.visitor_start_ts
	todayMaxMin         int    // → Target.today_max_visitor_min
	hasVisitorToday     bool   // → Target.has_visitor_today
	dayKey              string // "2026-05-18" (unit timezone)；用于午夜 lazy reset
}

type weakBioWindow struct {
	events []weakBioEvent // 时序队列，30min 外 lazy drop
	score  int            // 缓存当前 score（0..100）
	last80 bool           // 上次 score 是否 ≥80；用于检测上升沿（避免每帧重复 escalation）
}

type weakBioEvent struct {
	tsMs      int64
	alarmType string
	rawValue  int // 仅 WeakBiometricSignal 0/20/40/60
}

// New 构造。publisher 可空（测试 / 无 escalation 场景）。
func NewTargetStateAggregator(publisher AggregatorPublisher, redis *redislib.Client, logger *zap.Logger) *TargetStateAggregator {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &TargetStateAggregator{
		publisher: publisher,
		logger:    logger,
		redis:     redis,
		accums:    make(map[string]*cardAccumulator),
		monitorCh: make(chan MonitorFrame, 1024),
		alarmCh:   make(chan AlarmEventSnapshot, 256),
		zoneCh:    make(chan ZoneEventSnapshot, 256),
	}
}

// Run 主循环（单 reader 多 channel select）。
// P2 scaffold：handler stub 内为空，P3/P4 填业务逻辑。
func (a *TargetStateAggregator) Run(ctx context.Context) {
	a.logger.Info("target_state_aggregator: started")
	defer a.logger.Info("target_state_aggregator: stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case f := <-a.monitorCh:
			a.handleMonitorFrame(f)
		case e := <-a.alarmCh:
			a.handleAlarmEvent(ctx, e)
		case z := <-a.zoneCh:
			a.handleZoneEvent(z)
		}
	}
}

// PushMonitorFrame 由 monitor consumer 调，非阻塞（满队列时 drop）。
func (a *TargetStateAggregator) PushMonitorFrame(f MonitorFrame) {
	select {
	case a.monitorCh <- f:
	default:
		// 队列满，drop（这条 frame 的累加丢失；监控指标记 lag 后续加）
	}
}

// PushAlarmEvent 由 alarm consumer 调。
func (a *TargetStateAggregator) PushAlarmEvent(e AlarmEventSnapshot) {
	// 防 escalation loop：自家产的 alarm 不再喂回 aggregator
	if e.Producer == EscalationProducerTag {
		return
	}
	select {
	case a.alarmCh <- e:
	default:
	}
}

// OnZoneEvent 实现 zoneengine.ZoneEventListener（通过 wiring 注册）。
// 拿到 ZoneEvent 后立刻翻译成 ZoneEventSnapshot 入队，不阻塞 engine。
func (a *TargetStateAggregator) OnZoneEvent(cardID, zoneID string, totalPeople int, updatedAtMs int64) {
	select {
	case a.zoneCh <- ZoneEventSnapshot{
		CardID:      cardID,
		ZoneID:      zoneID,
		TotalPeople: totalPeople,
		UpdatedAtMs: updatedAtMs,
	}:
	default:
	}
}

// GetSnapshot StreamPublisher 60s tick pull 用（multi-return 实现 zoneengine.AggregatorPuller）。
// 若该卡无 accumulator entry 返回 ok=false。
func (a *TargetStateAggregator) GetSnapshot(cardID string) (target *card.TargetState, standingMin int, dirty bool, ok bool) {
	a.mu.RLock()
	acc, found := a.accums[cardID]
	a.mu.RUnlock()
	if !found {
		return nil, 0, false, false
	}
	target = &card.TargetState{
		UpdatedAt:           time.Now().UnixMilli(),
		LastActiveTs:        acc.lastActive.lastActiveTs,
		WeakBiometricSignal: acc.weakBio.score,
		VisitorStartTs:      acc.visitor.visitorStartTs,
		TodayMaxVisitorMin:  acc.visitor.todayMaxMin,
		HasVisitorToday:     acc.visitor.hasVisitorToday,
	}
	return target, acc.standing.continuousMin, acc.dirty, true
}

// ActiveCardIDs StreamPublisher tick 用：返回当前有 accumulator entry 的所有 cardID。
// 用于 tick 时遍历需要 publish 的卡。
func (a *TargetStateAggregator) ActiveCardIDs() []string {
	a.mu.RLock()
	out := make([]string, 0, len(a.accums))
	for cid := range a.accums {
		out = append(out, cid)
	}
	a.mu.RUnlock()
	return out
}

// MarkPublished StreamPublisher publish 完后回调，清 dirty + 记录 publish ts。
func (a *TargetStateAggregator) MarkPublished(cardID string, tsMs int64) {
	a.mu.Lock()
	if acc, ok := a.accums[cardID]; ok {
		acc.dirty = false
		acc.lastPublishTs = tsMs
	}
	a.mu.Unlock()
}

// ===================================================================
// Internal: handler stubs (P3/P4 fill)
// ===================================================================

// handleMonitorFrame P3 填：解析 walk/stand 累加 lastActive + standing。
func (a *TargetStateAggregator) handleMonitorFrame(f MonitorFrame) {
	if f.CardID == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	acc := a.getOrCreateLocked(f.CardID)
	_ = acc
	// TODO P3: lastActive 60s 节流 (walk_distance≥2m OR walk_duration≥6s + totalPeople==1)
	// TODO P3: standing 累加 (stand_duration≥55s + totalPeople==1, 封顶 8, 多人/坐走躺 reset 0)
}

// handleAlarmEvent P4 填：累加 weakBio 30min 滑窗 + 跨 80 escalation。
func (a *TargetStateAggregator) handleAlarmEvent(ctx context.Context, e AlarmEventSnapshot) {
	if e.CardID == "" {
		return
	}
	a.mu.Lock()
	acc := a.getOrCreateLocked(e.CardID)
	_ = acc
	a.mu.Unlock()
	// TODO P4: append to weakBio.events; drop ≥30min old; recompute score
	//          if score crosses 80 (rising edge): a.publisher.PublishEscalationAlarm(...)
	_ = ctx
}

// handleZoneEvent P4 填：更新 totalPeople cache + visitor 5min 段判定。
func (a *TargetStateAggregator) handleZoneEvent(z ZoneEventSnapshot) {
	if z.CardID == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	acc := a.getOrCreateLocked(z.CardID)
	acc.totalPeople = z.TotalPeople
	acc.lastZoneTs = z.UpdatedAtMs
	// TODO P4: visitor 段跟踪
	//   if totalPeople>=2: 若 multiSegmentStartTs==0 → 设为 z.UpdatedAtMs
	//     elapsed≥5min 时设 visitor_start_ts + has_visitor_today, todayMaxMin = max
	//   else: multiSegmentStartTs = 0
}

// getOrCreateLocked accumulator lazy 创建。需在 a.mu 写锁下调。
func (a *TargetStateAggregator) getOrCreateLocked(cardID string) *cardAccumulator {
	if acc, ok := a.accums[cardID]; ok {
		return acc
	}
	acc := &cardAccumulator{
		cardID: cardID,
		// unitID / unitType 由后续 P4 从 device_meta 查；此处先空
	}
	a.accums[cardID] = acc
	return acc
}

// EscalationProducerTag aggregator 自家产 alarm 用此 producer 字段，防 loop。
const EscalationProducerTag = "sensor:target-state-aggregator"

// 防 dead-import 编译 warning（observation 字段常量后续 P3 用）。
var _ = observation.FieldWalkDistance
