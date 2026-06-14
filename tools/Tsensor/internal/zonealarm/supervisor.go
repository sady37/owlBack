package zonealarm

import (
	"context"
	"time"

	"owl-common/card"

	"go.uber.org/zap"
)

// StateProvider zonealarm Tick 时拿 sensor 派生 state 快照的窄接口。
// 由 sensor StreamPublisher 实现。
type StateProvider interface {
	SnapshotRooms() map[string]RoomEntry
	SnapshotBeds() map[string]*card.BedState
}

// RiskRefresher zonealarm Tick 周期 refresh room RiskLevel 的窄接口。
// 由 StreamPublisher 实现：传 cidr+nowMs，内部用最新 AloneSinceTs 重算 RiskLevel，
// 变化则 publish 让 cardagg display 跟随（5s coalesce）。
type RiskRefresher interface {
	RefreshRoomRiskAndMaybePublish(ctx context.Context, roomCIDR string, nowMs int64) bool
}

// Supervisor pure-Tick zonealarm 控制器（state-anchored 重构后）。
//
// 与旧版区别：
//   - 不实现 ZoneEventListener；不订阅 ZoneEvent 流
//   - 无 pending FSM（fire-once gate 在 Evaluator）
//   - Tick 一份代码同时跑 ① RiskLevel refresh（5s coalesce）② Evaluator.Check（每 tick）
type Supervisor struct {
	rules     []Rule
	firer     AlarmFirer
	logger    *zap.Logger
	evaluator *Evaluator

	provider  StateProvider
	refresher RiskRefresher

	// 5s coalesce RiskLevel refresh（Tick 上层 1s 喂）
	riskRefreshEverySec int
	lastRefreshMs       int64

	// Timezone — 暂用 nil（UTC fallback）；per-card tz 注入留后续
	loc *time.Location

	timeout time.Duration
}

// NewSupervisor 实例化。firer=nil 时退化为 NopFirer（仅日志，不真发）。
//
// 注：必须随后调 SetProvider/SetRefresher 注入 sensor 派生 state 数据源，否则 Tick 无所作为。
func NewSupervisor(rules []Rule, firer AlarmFirer, logger *zap.Logger) *Supervisor {
	if firer == nil {
		firer = NopFirer{}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Supervisor{
		rules:               rules,
		firer:               firer,
		logger:              logger,
		evaluator:           NewEvaluator(),
		riskRefreshEverySec: 5,
		timeout:             2 * time.Second,
	}
}

// SetProvider main wiring 注入 state snapshot 源（StreamPublisher）。
func (s *Supervisor) SetProvider(p StateProvider) { s.provider = p }

// SetRefresher main wiring 注入 RiskLevel refresh 回调（同 StreamPublisher）。
func (s *Supervisor) SetRefresher(r RiskRefresher) { s.refresher = r }

// SetThresholdResolver main wiring 注入 per-device 阈值/使能查询（AlarmEnablementCache 适配）。
func (s *Supervisor) SetThresholdResolver(r ThresholdResolver) { s.evaluator.SetResolver(r) }

// SetDeviceLookup main wiring 注入 zone CIDR → device /128 查询（BedDeviceLookup 适配）。
func (s *Supervisor) SetDeviceLookup(l DeviceLookup) { s.evaluator.SetDeviceLookup(l) }

// SetTimezone per-card timezone 注入留后续；当前所有 rule 共享同一 loc。
func (s *Supervisor) SetTimezone(loc *time.Location) { s.loc = loc }

// ReloadRules 替换规则集；fired gate 不动（运行时连续性 — rule 改阈值不重置已 fire）。
func (s *Supervisor) ReloadRules(rules []Rule) {
	s.rules = rules
	s.logger.Info("zonealarm rules reloaded", zap.Int("count", len(rules)))
}

// Tick 上层 ticker（建议 1s）调一次。两件事：
//
//	1) RiskLevel refresh：5s coalesce，遍历所有 room/bathroom 让 StreamPublisher 用最新
//	   nowMs 重算 RiskLevel 并 publish（cardagg display 跟随 alone 时长升档）
//	2) Evaluator.Check：每 tick 跑，3 条 rule 对当前 snapshot 判 fire（fire 锁外调）
func (s *Supervisor) Tick(nowMs int64) {
	if s.provider == nil {
		return
	}
	snap := s.takeSnapshot()

	// 1) RiskLevel refresh（5s coalesce）
	if s.refresher != nil && nowMs-s.lastRefreshMs >= int64(s.riskRefreshEverySec)*1000 {
		s.lastRefreshMs = nowMs
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		for cidr := range snap.Rooms {
			s.refresher.RefreshRoomRiskAndMaybePublish(ctx, cidr, nowMs)
		}
		cancel()
	}

	// 2) Evaluator.Check — ctx 给 resolver 用（cache 命中 ns 级；miss 走 DB）
	checkCtx, checkCancel := context.WithTimeout(context.Background(), s.timeout)
	decisions := s.evaluator.Check(checkCtx, s.rules, snap, nowMs, s.loc)
	checkCancel()
	for _, d := range decisions {
		s.callFire(d)
	}
}

// takeSnapshot 从 StateProvider 抓 rooms+beds 组成 Evaluator 用的 snapshot。
func (s *Supervisor) takeSnapshot() StateSnapshot {
	return StateSnapshot{
		Rooms: s.provider.SnapshotRooms(),
		Beds:  s.provider.SnapshotBeds(),
	}
}

func (s *Supervisor) callFire(d FireDecision) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	if err := s.firer.Fire(ctx, d.Event); err != nil {
		s.logger.Warn("zonealarm fire failed",
			zap.String("alarm", d.Rule.AlarmType),
			zap.String("zone_id", d.Event.Key.ZoneID),
			zap.Error(err))
	}
}
