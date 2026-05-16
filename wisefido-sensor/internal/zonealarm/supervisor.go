package zonealarm

import (
	"context"
	"sync"
	"time"

	"wisefido-sensor/internal/zoneengine"

	"go.uber.org/zap"
)

// Supervisor 实现 zoneengine.ZoneEventListener，按 Rules 派生 alarm。
//
// 线程模型：
//   - OnZoneEvent 写 / Tick 写 共用一把锁
//   - AlarmFirer 调用在锁外 emit，避免 publish slow path 挡 ZoneEvent 流入
//
// **fire-once-until-zone-leave 语义（2026-05-15 用户拍板）**：
//   - 已 fire 过的 (cardID, alarmType) 在 fired map 中标记
//   - 标记存在期间不再 arm 同一规则（同一段 occupancy 期内最多 1 条 alarm_events）
//   - arm zone "真正离开" arm status 时清除标记，允许下次重新 arm
//     · arm=Occupied 规则（Stay）：仅 zone→Vacant 才清（Leaving 不算，可能软离开后回弹）
//     · arm=Vacant 规则（LeftBed/NightAbsence/BedNightAbsence）：zone→Occupied 或 zone→Leaving 都清
//       （Leaving IsPresent=true 已表示人回来了）
//   - 设计动机：alarm 是警报通知不是状态轮询；且 fire 之前 still_fall 等更紧急规则已先触发
type Supervisor struct {
	rules  []Rule
	firer  AlarmFirer
	logger *zap.Logger

	mu      sync.Mutex
	pending map[PendingKey]*Pending
	fired   map[PendingKey]bool // fire-once-until-zone-leave gate

	// fireCtxFactory 每次 fire / arm / cancel 操作生成新的 ctx 用 — 默认 2s timeout
	timeout time.Duration
}

// NewSupervisor 实例化。firer=nil 时退化为 NopFirer（仅日志，不真发）。
func NewSupervisor(rules []Rule, firer AlarmFirer, logger *zap.Logger) *Supervisor {
	if firer == nil {
		firer = NopFirer{}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Supervisor{
		rules:   rules,
		firer:   firer,
		logger:  logger,
		pending: make(map[PendingKey]*Pending),
		fired:   make(map[PendingKey]bool),
		timeout: 2 * time.Second,
	}
}

// ReloadRules 替换规则集（运行时 pending 保留 — 它们已 arm，新规则只影响后续）。
func (s *Supervisor) ReloadRules(rules []Rule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = rules
	s.logger.Info("zonealarm rules reloaded", zap.Int("count", len(rules)))
}

// OnZoneEvent satisfy zoneengine.ZoneEventListener。
//
// 单条 event 可能命中多条规则的 arm 或 cancel：先收集所有动作，释放锁后再调 firer。
func (s *Supervisor) OnZoneEvent(e zoneengine.ZoneEvent) {
	if e.CardID == "" {
		return
	}
	now := e.Ts
	if now == 0 {
		now = time.Now().UnixMilli()
	}

	s.mu.Lock()
	var arms, cancels, fires []Pending
	var cancelKeys []PendingKey
	for i := range s.rules {
		r := &s.rules[i]
		key := PendingKey{CardID: e.CardID, AlarmType: r.AlarmType}

		// fire-once-until-zone-leave: arm zone "真正离开" arm status → 清 fired 标记
		// （允许下一段 occupancy 重新 arm）
		if e.ZoneType == r.ArmZone && s.fired[key] && zoneLeftArmStatus(r.ArmStatus, e.NewState.Status) {
			delete(s.fired, key)
			s.logger.Debug("zonealarm fired flag cleared",
				zap.String("alarm", r.AlarmType),
				zap.String("cid", e.CardID),
				zap.String("new_status", e.NewState.Status.String()))
		}

		// arm 检查
		if matchesArm(r, e) {
			// fire-once-until-zone-leave: 已 fired 则跳过 arm（同段 occupancy 不重复报）
			if s.fired[key] {
				continue
			}
			if r.TimeWindow.Active(time.UnixMilli(now)) {
				p := s.armPendingLocked(r, e, now)
				if p != nil {
					if r.DurationSec == 0 {
						// 立即 fire
						fires = append(fires, *p)
						delete(s.pending, p.Key)
						s.fired[p.Key] = true // fire-once: 锁内标记
					} else {
						arms = append(arms, *p)
					}
				}
			}
		}
		// cancel 检查（每条规则可能有多个 cancel trigger，任一命中即 cancel）
		for j := range r.Cancels {
			if matchesCancel(&r.Cancels[j], e) {
				if p, ok := s.pending[key]; ok {
					cancels = append(cancels, *p)
					cancelKeys = append(cancelKeys, key)
					delete(s.pending, key)
				}
				// cancel 不清 fired 标记 — fired 仅在 arm zone "真正离开" arm status 时清
				// （cancel 可能由 peer-zone 触发，此时 arm zone 还在 arm status，不该解锁重 fire）
				break // 同 rule 多 cancel trigger 命中一次足够
			}
		}
	}
	s.mu.Unlock()

	// emit 阶段 — 锁外
	for _, p := range arms {
		s.callArm(p)
	}
	for i, p := range cancels {
		s.callCancel(cancelKeys[i], e)
		_ = p
	}
	for _, p := range fires {
		s.callFire(p)
	}
}

// Tick 周期检查 pending 是否到期。建议每 1s 调一次（与 zone engine Tick 同步）。
func (s *Supervisor) Tick(nowMs int64) {
	if nowMs == 0 {
		nowMs = time.Now().UnixMilli()
	}
	s.mu.Lock()
	var due []Pending
	for k, p := range s.pending {
		if nowMs >= p.DueAt {
			due = append(due, *p)
			delete(s.pending, k)
			s.fired[k] = true // fire-once: 锁内标记，避免与 callFire 之间 race re-arm
		}
	}
	s.mu.Unlock()

	for _, p := range due {
		s.callFire(p)
	}
}

// zoneLeftArmStatus 判断 zone NewState 是否表示"真正完成"离开 arm status（用于清 fired 标记）。
//
// 原则（2026-05-15 用户拍板）：Leaving 是 zone engine 软离开中间态，**未完成确认**，不参与
// zonealarm 决策。只有完成态（Vacant 或 Occupied）才视为转换完成。
//
//	arm=Occupied 的规则（Stay）：仅 NewState=Vacant 算真离开。
//	arm=Vacant 的规则（LeftBed / NightAbsence / BedNightAbsence）：仅 NewState=Occupied 算人真回来。
//	Leaving 在两种方向上都被忽略 — 保持现有 fired 标记。
func zoneLeftArmStatus(armStatus, newStatus zoneengine.ZoneStatus) bool {
	switch armStatus {
	case zoneengine.StatusOccupied:
		return newStatus == zoneengine.StatusVacant
	case zoneengine.StatusVacant:
		return newStatus == zoneengine.StatusOccupied
	}
	return false
}

// armPendingLocked 创建 pending 实例。若已有同 key 的 pending（之前的 arm 未 fire 也未 cancel），
// 用新 trigger event **重置** DueAt（refresh）—— 这是相对宽容策略：避免老 pending 用过时
// trigger 决策；如果业务侧需"first-arm-wins"语义后续可加 yaml flag。
//
// caller 必须已持锁。
func (s *Supervisor) armPendingLocked(r *Rule, e zoneengine.ZoneEvent, now int64) *Pending {
	key := PendingKey{CardID: e.CardID, AlarmType: r.AlarmType}
	p := &Pending{
		Key:     key,
		ArmedAt: now,
		DueAt:   now + int64(r.DurationSec)*1000,
		Level:   r.Level,
		Trigger: e,
	}
	s.pending[key] = p
	return p
}

// matchesArm — event 是否触发本条 rule 的 arm。
//
// 同 zone type；status 翻转到 rule.ArmStatus。
// count_change 不参与 arm（它不是状态翻转）。
func matchesArm(r *Rule, e zoneengine.ZoneEvent) bool {
	if e.ZoneType != r.ArmZone {
		return false
	}
	if e.Transition == zoneengine.TransitionCountChange {
		return false
	}
	return e.NewState.Status == r.ArmStatus
}

// matchesCancel — event 是否触发本条 cancel trigger。
//
//	Zone 必须匹配（不限 same-zone vs peer-zone — Stay 的 bed.InBed cancel 是 peer-zone）
//	StatusAny=true → 不检查 status
//	CountGtZero=true → NewState.Count > 0 才命中
func matchesCancel(c *CancelTrigger, e zoneengine.ZoneEvent) bool {
	if e.ZoneType != c.Zone {
		return false
	}
	if c.CountGtZero {
		return e.NewState.Count > 0
	}
	if c.StatusAny {
		return true
	}
	return e.NewState.Status == c.Status
}

// callArm / callCancel / callFire — emit 包装，统一 timeout 控制 + 错误日志。
func (s *Supervisor) callArm(p Pending) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	if err := s.firer.Arm(ctx, p); err != nil {
		s.logger.Warn("zonealarm arm failed",
			zap.String("alarm", p.Key.AlarmType),
			zap.String("cid", p.Key.CardID),
			zap.Error(err))
	}
}

func (s *Supervisor) callCancel(key PendingKey, reason zoneengine.ZoneEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	if err := s.firer.Cancel(ctx, key, reason); err != nil {
		s.logger.Warn("zonealarm cancel failed",
			zap.String("alarm", key.AlarmType),
			zap.String("cid", key.CardID),
			zap.Error(err))
	}
}

func (s *Supervisor) callFire(p Pending) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	if err := s.firer.Fire(ctx, p); err != nil {
		s.logger.Warn("zonealarm fire failed",
			zap.String("alarm", p.Key.AlarmType),
			zap.String("cid", p.Key.CardID),
			zap.Error(err))
	}
}

// PendingCount 当前 pending 数（监控 / 测试用）。
func (s *Supervisor) PendingCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pending)
}
