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
// **WeakBio 设计修订 2026-05-19**（[[target_state_weak_bio_signal_design]]）：
//   原"score 跨 80 触发 Critical escalation alarm"砍掉——WeakBio 是**长期状态/风险描述符**，
//   不是事件源。firmware 已经直发 HR/RR/ApneaH/WeakBio raw alarm 进 alarm pipeline；
//   sensor 累计 score 仅作"风险放大系数"+ FE 横条显示（Section3.down.right），
//   不再 publish escalation alarm。AggregatorPublisher 接口废止；aggregator 退化为纯 holder。

package service

import (
	"context"
	"net/netip"
	"sync"
	"time"

	"owl-common/alarm"
	"owl-common/card"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ZoneEventSnapshot aggregator 关心的最小 ZoneEvent 子集（解耦 zoneengine 类型）。
// StreamPublisher 把 ZoneEvent 翻译成这个 push 给 aggregator。
type ZoneEventSnapshot struct {
	SpatialPrefix string // INET CIDR (物理实体地址；可能是 /88 room 或 /96 bed)
	ZoneID        string
	TotalPeople   int
	UpdatedAtMs   int64
}

// EventFrame radar 分钟级 activity stat（aggregator 关心的统计字段子集）。
// 来源：iot:event:stream category=activity（详 [[radar_activity_stats_on_event_stream]]）。
// **不是** monitor 流——monitor 是 1Hz raw track XY/HR/RR，不进 aggregator。
// 由 activity consumer 翻译成此结构后 push。
type EventFrame struct {
	SpatialPrefix          string // INET CIDR (radar/sleepad 所在物理实体的地址，通常 /96 bed 或 /88 room)
	DeviceAddr             string
	TsMs                   int64
	WalkDistanceMeters     int // FieldWalkDistance (m, 1min)
	WalkDurationSec        int // FieldWalkDuration (s, 1min cap 900)
	StandDurationSec       int // FieldStandDuration (s, 1min cap 600)
	MultiPersonDurationSec int // FieldMultiPersonDuration (s, 1min)
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
// 输出：GetSnapshot pull（StreamPublisher 60s 用）。纯 state holder，不 publish。
type TargetStateAggregator struct {
	logger *zap.Logger
	redis  *redislib.Client // 订阅 monitor / alarm 流用；P2 scaffold 允许 nil（无订阅）

	mu     sync.RWMutex
	accums map[string]*spatialAccumulator // key = spatial_prefix INET CIDR (物理实体地址)

	// Push channels（main loop 单 reader 防 race）
	eventCh chan EventFrame
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
}

type weakBioEvent struct {
	tsMs      int64
	alarmType string
	rawValue  int // 仅 WeakBiometricSignal 0/20/40/60
}

// NewTargetStateAggregator 构造（纯 state holder，不 publish；2026-05-19 砍 escalation 路径后无 publisher 入参）。
func NewTargetStateAggregator(redis *redislib.Client, logger *zap.Logger) *TargetStateAggregator {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &TargetStateAggregator{
		logger:    logger,
		redis:     redis,
		accums:    make(map[string]*spatialAccumulator),
		eventCh:   make(chan EventFrame, 1024),
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
		case f := <-a.eventCh:
			a.handleEventFrame(f)
		case e := <-a.alarmCh:
			a.handleAlarmEvent(ctx, e)
		case z := <-a.zoneCh:
			a.handleZoneEvent(z)
		}
	}
}

// PushEventFrame 由 activity consumer 调，非阻塞（满队列时 drop）。
func (a *TargetStateAggregator) PushEventFrame(f EventFrame) {
	select {
	case a.eventCh <- f:
	default:
		// 队列满，drop（这条 frame 的累加丢失；监控指标记 lag 后续加）
	}
}

// PushEventFields primitive-form 入口，供外部 consumer 用（避免 caller 反向 import service.EventFrame）。
func (a *TargetStateAggregator) PushEventFields(spatialPrefix, deviceAddr string, tsMs int64,
	walkDistanceMeters, walkDurationSec, standDurationSec, multiPersonDurationSec int) {
	a.PushEventFrame(EventFrame{
		SpatialPrefix:          spatialPrefix,
		DeviceAddr:             deviceAddr,
		TsMs:                   tsMs,
		WalkDistanceMeters:     walkDistanceMeters,
		WalkDurationSec:        walkDurationSec,
		StandDurationSec:       standDurationSec,
		MultiPersonDurationSec: multiPersonDurationSec,
	})
}

// PushAlarmEvent 由 alarm consumer 调。
func (a *TargetStateAggregator) PushAlarmEvent(e AlarmEventSnapshot) {
	select {
	case a.alarmCh <- e:
	default:
	}
}

// PushAlarmFields primitive-form 入口，供外部 consumer 用（避免 caller 反向 import service.AlarmEventSnapshot
// 形成包环：service→consumer 已存在，consumer→service 通过此 interface-friendly API 解耦）。
func (a *TargetStateAggregator) PushAlarmFields(spatialPrefix, alarmType, producer string, tsMs int64, rawValue int) {
	a.PushAlarmEvent(AlarmEventSnapshot{
		SpatialPrefix: spatialPrefix,
		AlarmType:     alarmType,
		TsMs:          tsMs,
		Producer:      producer,
		RawValue:      rawValue,
	})
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

// ForgetDevice 清空 device 派生的 accumulator（"device offline = 内存重启"原则）。
//
// 实现：把 device addr 推到 /96 bed prefix 和 /88 room prefix（v2 cards.spatial_prefix
// 在这两层）→ 清 accums 对应 entry。多 device 共 spatial 罕见但理论存在（如 share room
// 多 radar），over-clear 视为可接受（offline 本来就丢数据）。
//
// 不参与 weakBio.events 细粒度清理（events 按 ts 自然 30min 滑窗 lazy drop，全量清更简单）。
func (a *TargetStateAggregator) ForgetDevice(deviceAddr string) {
	if a == nil || deviceAddr == "" {
		return
	}
	addr, err := netip.ParseAddr(deviceAddr)
	if err != nil {
		return
	}
	sp96 := netip.PrefixFrom(addr, 96).Masked().String()
	sp88 := netip.PrefixFrom(addr, 88).Masked().String()
	a.mu.Lock()
	delete(a.accums, sp96)
	delete(a.accums, sp88)
	a.mu.Unlock()
	a.logger.Info("aggregator forgot device accums",
		zap.String("device_addr", deviceAddr),
		zap.String("sp96", sp96),
		zap.String("sp88", sp88),
	)
}

// WeakBioScore 返回 spatial 实体当前 WeakBio 累加 score（0-100）。
// 无 entry 返回 0；roomengine fall verifier "WeakBio≥80 force real" 提级路径用。
//
// 实现 roomengine.WeakBioSource interface（解耦：sensor service ← roomengine 单向消费 score；
// 不引 service 类型，零 import cycle 风险）。
func (a *TargetStateAggregator) WeakBioScore(spatialPrefix string) int {
	if a == nil || spatialPrefix == "" {
		return 0
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	acc, found := a.accums[spatialPrefix]
	if !found {
		return 0
	}
	return acc.weakBio.score
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

// handleEventFrame 解析 radar 分钟级 activity stat → LastActiveTs + StandingContinuousMin。
//
// 输入来源：iot:event:stream category=activity（firmware 每分钟一发；详
// [[radar_activity_stats_on_event_stream]]）。**心跳契约** = firmware 分钟级 push 既有事实，
// 不另造心跳——只要 firmware 每分钟发，aggregator 就每分钟 dirty。
//
// 规则：
//   - LastActiveTs：WalkDistance≥2m OR WalkDuration≥6s → 触发活跃；60s 节流（防同一分钟重复刷）
//   - StandingContinuousMin：TotalPeople==1 + StandDurationSec≥55s + MultiPersonDuration==0
//     → 累 +1 封顶 8；不满足条件 reset 0
//
// 多人 / 站立<55s（人坐/走/躺）→ standing reset；让 cardagg side risk_evaluator 看到 0 不误报。
func (a *TargetStateAggregator) handleEventFrame(f EventFrame) {
	if f.SpatialPrefix == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	acc := a.getOrCreateLocked(f.SpatialPrefix)

	// LastActive: walk≥2m OR walk_duration≥6s → 活跃；60s 节流避免同一分钟双写
	if f.WalkDistanceMeters >= 2 || f.WalkDurationSec >= 6 {
		if f.TsMs-acc.lastActive.lastActiveTs >= 60_000 {
			acc.lastActive.lastActiveTs = f.TsMs
			acc.dirty = true
		}
	}

	// Standing: gate（totalPeople==1 + 单人持续站≥55s + 无多人事件）→ +1 cap 8；否则 reset 0
	if acc.totalPeople == 1 && f.StandDurationSec >= 55 && f.MultiPersonDurationSec == 0 {
		if f.TsMs-acc.standing.lastIncrementTs >= 60_000 {
			if acc.standing.continuousMin < 8 {
				acc.standing.continuousMin++
			}
			acc.standing.lastIncrementTs = f.TsMs
			acc.dirty = true
		}
	} else if acc.standing.continuousMin > 0 {
		acc.standing.continuousMin = 0
		acc.standing.lastIncrementTs = 0
		acc.dirty = true
	}
}

// handleAlarmEvent 累加 WeakBio 30min 滑窗 score（[[target_state_weak_bio_signal_design]]）。
//
// score = max(weakBio_raw) + 5×|HR| + 5×|RR| + 15×|ApneaH|，封顶 100。
// 30min 外的 event lazy drop（下次 event 到来时清；空闲累加器自然回 0）。
//
// 2026-05-19 砍 escalation：score 仅作"风险描述符"+ FE Section3.down.right 横条显示
// （30/60/80 三档配色），**不再独立触发 Critical alarm**。原 firmware 直发的 HR/RR/ApneaH/
// WeakBio raw alarm 仍正常进 alarm pipeline，score 只是聚合摘要供 FE / 风险评估读。
func (a *TargetStateAggregator) handleAlarmEvent(ctx context.Context, e AlarmEventSnapshot) {
	if e.SpatialPrefix == "" || !isWeakBioRelatedAlarm(e.AlarmType) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	acc := a.getOrCreateLocked(e.SpatialPrefix)
	cutoff := e.TsMs - weakBioWindowMs

	// lazy drop 窗外事件
	kept := acc.weakBio.events[:0]
	for _, ev := range acc.weakBio.events {
		if ev.tsMs >= cutoff {
			kept = append(kept, ev)
		}
	}
	acc.weakBio.events = kept

	// 加入本次 event
	acc.weakBio.events = append(acc.weakBio.events, weakBioEvent{
		tsMs:      e.TsMs,
		alarmType: e.AlarmType,
		rawValue:  e.RawValue,
	})

	// 重算 score 缓存
	acc.weakBio.score = computeWeakBioScore(acc.weakBio.events)
	acc.dirty = true
	_ = ctx
}

const weakBioWindowMs int64 = 30 * 60 * 1000 // 30min

// isWeakBioRelatedAlarm 4 种关联 alarm 进 weakBio 计算（其它一律忽略）。
// caller（AlarmConsumer）已把 .High/.Low 变体规整到 base name。
func isWeakBioRelatedAlarm(t string) bool {
	switch t {
	case alarm.WeakBiometricSignal, alarm.HeartRateAlert, alarm.RespRateAlert, alarm.ApneaHypopnea:
		return true
	}
	return false
}

// computeWeakBioScore 按 [[target_state_weak_bio_signal_design]] 修订权重：
//   score = max(weakBio_raw) + 5×|HR| + 5×|RR| + 15×|ApneaH|，封顶 100。
//
// 权重调整理由（2026-05-19）：
//   - HR/RR 单次异常常见（睡眠/翻身边缘漂移），降权 5 让单 1-2 条噪声不达 30 阈值
//   - ApneaH 单次仍有临床意义但不应单条触红，15 让 2 条以上才显示
//   - WeakBio raw 保持 max（firmware 已分严重等级，不累加）
//
// raw 0/20/40/60 由 firmware state×20 给出；同窗多源不去重（雷达+sleepad 各自计）。
func computeWeakBioScore(events []weakBioEvent) int {
	var maxRaw, hrCount, rrCount, apneaCount int
	for _, ev := range events {
		switch ev.alarmType {
		case alarm.WeakBiometricSignal:
			if ev.rawValue > maxRaw {
				maxRaw = ev.rawValue
			}
		case alarm.HeartRateAlert:
			hrCount++
		case alarm.RespRateAlert:
			rrCount++
		case alarm.ApneaHypopnea:
			apneaCount++
		}
	}
	score := maxRaw + 5*hrCount + 5*rrCount + 15*apneaCount
	if score > 100 {
		score = 100
	}
	return score
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

