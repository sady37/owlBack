// target_state_aggregator.go — sensor 端 Target/RoomState 派生字段聚合器（card_display_projector Task 2A）。
//
// 分层原则（用户拍板 2026-05-18）：
//   - unit / room / bed / device 都是**物理实体**；card 是 UI/展示逻辑层，零物理存在
//   - sensor 只做**单实体内**判断（fall / standing / lastActive / weakBio）
//   - 跨实体合并（visitor 等）归 cardagg 职责（详 memory [[visitor_belongs_to_cardagg]]）
//   - sensor 不读 cards 表；本聚合器按"物理地址"（spatial_prefix INET CIDR）做 key
//
// 职责（per spatial_prefix，sensor 单实体判断，3 个字段）：
//   - Target.LastActiveTs                       — radar walk≥2m OR walk_duration≥6s + total_people==1
//   - Target.weak_biometric_signal              — 30min lazy 滑窗 weak/HR/RR/ApneaH alarm 累加 0-100
//   - RoomState.StandingContinuousMin           — stand_duration≥55s 累+1 封顶 8（仅 TotalPeople==1）
//
// 已挪出 sensor（2026-05-18 后归 cardagg `VisitorDeriver`）：
//   - Target.visitor_start_ts / has_visitor_today / today_max_visitor_min
//     原因：visitor 跨多 /88 room 合并；Private gate 需查 unit_type；sensor 不做跨实体合并
//
// 架构（详 [[target_state_aggregator]] design doc）：
//   Aggregator = 纯 state holder（不 publish）；
//   StreamPublisher 60s ticker 主动 pull GetSnapshot 合并进 RoomState/Target publish；
//   零依赖反向（aggregator 不引用 zoneengine 包）—— 通过 ZoneEvent listener 单向 push。
//
// Escalation alarm（WeakBiometricSignal score 跨 80 提级 Critical）由 aggregator 自己直接 publishAlarm
// 到 iot:alarm:stream，不走 StreamPublisher（不同流，alarm pipeline 自身多源）。
//
// P2 scaffold：struct + lifecycle + 接口面；3 个累加器内部逻辑 stub（P3/P4 填）。

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
	PublishEscalationAlarm(ctx context.Context, spatialPrefix, alarmType string, score int, contributorDeviceAddr string) error
}

// ZoneEventSnapshot aggregator 关心的最小 ZoneEvent 子集（解耦 zoneengine 类型）。
// StreamPublisher 把 ZoneEvent 翻译成这个 push 给 aggregator。
type ZoneEventSnapshot struct {
	SpatialPrefix string // INET CIDR (物理实体地址；可能是 /88 room 或 /96 bed)
	ZoneID        string
	TotalPeople   int
	UpdatedAtMs   int64
}

// MonitorFrame 单帧 radar/sleepad observation（aggregator 关心的统计字段子集）。
// 由 monitor consumer 翻译成此结构后 push。
type MonitorFrame struct {
	SpatialPrefix          string // INET CIDR (radar/sleepad 所在物理实体的地址，通常 /96 bed 或 /88 room)
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
	SpatialPrefix string // INET CIDR (告警归属的物理实体)
	AlarmType     string // "WeakBiometricSignal" / "HeartRateAlert" / "RespRateAlert" / "ApneaHypopnea"
	TsMs          int64
	Producer      string // 跳过自家产的 escalation alarm 防 loop
	RawValue      int    // 仅 WeakBiometricSignal 携带（state×20: 0/20/40/60）
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
	accums map[string]*spatialAccumulator // key = spatial_prefix INET CIDR (物理实体地址)

	// Push channels（main loop 单 reader 防 race）
	monitorCh chan MonitorFrame
	alarmCh   chan AlarmEventSnapshot
	zoneCh    chan ZoneEventSnapshot
}

// spatialAccumulator per spatial_prefix 累加状态（物理实体维度，非 card；单实体内判断）。
type spatialAccumulator struct {
	spatialPrefix string // INET CIDR；物理地址（/88 room / /96 bed / /128 device 之一）

	// 3 个 sub-accumulator（单实体内判断；visitor 跨 room 合并归 cardagg，不在此）
	lastActive lastActiveState
	standing   standingState
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
		accums:    make(map[string]*spatialAccumulator),
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
// spatialPrefix = ZoneEvent.CardID（v2 实现里同 INET CIDR）；从 aggregator 视角是物理实体地址。
func (a *TargetStateAggregator) OnZoneEvent(spatialPrefix, zoneID string, totalPeople int, updatedAtMs int64) {
	select {
	case a.zoneCh <- ZoneEventSnapshot{
		SpatialPrefix: spatialPrefix,
		ZoneID:        zoneID,
		TotalPeople:   totalPeople,
		UpdatedAtMs:   updatedAtMs,
	}:
	default:
	}
}

// GetSnapshot StreamPublisher 60s tick pull 用（multi-return 实现 zoneengine.AggregatorPuller）。
// 若该 spatial 实体无 accumulator entry 返回 ok=false。
func (a *TargetStateAggregator) GetSnapshot(spatialPrefix string) (target *card.TargetState, standingMin int, dirty bool, ok bool) {
	a.mu.RLock()
	acc, found := a.accums[spatialPrefix]
	a.mu.RUnlock()
	if !found {
		return nil, 0, false, false
	}
	// 仅 3 个 sensor 单实体内判断字段；visitor 已挪到 cardagg VisitorDeriver。
	target = &card.TargetState{
		UpdatedAt:           time.Now().UnixMilli(),
		LastActiveTs:        acc.lastActive.lastActiveTs,
		WeakBiometricSignal: acc.weakBio.score,
	}
	return target, acc.standing.continuousMin, acc.dirty, true
}

// ActiveSpatialPrefixes StreamPublisher tick 用：返回当前有 accumulator entry 的所有物理地址。
// 用于 tick 时遍历需要 publish 的实体。
func (a *TargetStateAggregator) ActiveSpatialPrefixes() []string {
	a.mu.RLock()
	out := make([]string, 0, len(a.accums))
	for sp := range a.accums {
		out = append(out, sp)
	}
	a.mu.RUnlock()
	return out
}

// MarkPublished StreamPublisher publish 完后回调，清 dirty + 记录 publish ts。
func (a *TargetStateAggregator) MarkPublished(spatialPrefix string, tsMs int64) {
	a.mu.Lock()
	if acc, ok := a.accums[spatialPrefix]; ok {
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
	if f.SpatialPrefix == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	acc := a.getOrCreateLocked(f.SpatialPrefix)
	_ = acc
	// TODO P3: lastActive 60s 节流 (walk_distance≥2m OR walk_duration≥6s + totalPeople==1)
	// TODO P3: standing 累加 (stand_duration≥55s + totalPeople==1, 封顶 8, 多人/坐走躺 reset 0)
}

// handleAlarmEvent P4 填：累加 weakBio 30min 滑窗 + 跨 80 escalation。
func (a *TargetStateAggregator) handleAlarmEvent(ctx context.Context, e AlarmEventSnapshot) {
	if e.SpatialPrefix == "" {
		return
	}
	a.mu.Lock()
	acc := a.getOrCreateLocked(e.SpatialPrefix)
	_ = acc
	a.mu.Unlock()
	// TODO P4: append to weakBio.events; drop ≥30min old; recompute score
	//          if score crosses 80 (rising edge): a.publisher.PublishEscalationAlarm(...)
	_ = ctx
}

// handleZoneEvent 更新 totalPeople cache（给 lastActive / standing gate 用）。
// 注：visitor 跨 room 合并已挪到 cardagg VisitorDeriver，本 handler 不再做 visitor 判定。
func (a *TargetStateAggregator) handleZoneEvent(z ZoneEventSnapshot) {
	if z.SpatialPrefix == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	acc := a.getOrCreateLocked(z.SpatialPrefix)
	acc.totalPeople = z.TotalPeople
	acc.lastZoneTs = z.UpdatedAtMs
}

// getOrCreateLocked accumulator lazy 创建。需在 a.mu 写锁下调。
func (a *TargetStateAggregator) getOrCreateLocked(spatialPrefix string) *spatialAccumulator {
	if acc, ok := a.accums[spatialPrefix]; ok {
		return acc
	}
	acc := &spatialAccumulator{
		spatialPrefix: spatialPrefix,
	}
	a.accums[spatialPrefix] = acc
	return acc
}

// EscalationProducerTag aggregator 自家产 alarm 用此 producer 字段，防 loop。
const EscalationProducerTag = "sensor:target-state-aggregator"

// 防 dead-import 编译 warning（observation 字段常量后续 P3 用）。
var _ = observation.FieldWalkDistance
