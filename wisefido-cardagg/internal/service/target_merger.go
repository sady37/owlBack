// target_merger.go — per-entity TargetState → per-card 折叠聚合 holder。
//
// 背景（doc/card_display.md §4.3, §9）：
//   - sensor 按 spatial 实体发 target.state（msg.SubjectEntity = /88 room 或 /96 bed CIDR）
//   - 一张 card 底下可挂多个实体（整户 /80 卡下多个 /88 房 + /96 床；self-card 时 card_id == 实体）
//   - cardagg 把 owning card 名下所有实体的 TargetState max-merge 成单 card:state.target
//     （与 room.state 的 pickPriorityRoom、bed.state 的 LookupEntry 同源：实体存自己 prefix、
//      折到 owning card 展示）
//
// 合并字段（max-merge across owning card 名下实体）：
//   LastActiveTs / StandingContinuousMin / WeakBiometricSignal
//
// **Standing / WeakBio 的 snapshot staleness 过滤**：standing 是瞬时状态（人坐下后 sensor
// 必须 reset 0），任何 stale snapshot 都会直接造成 risk_evaluator false-positive → 误报 alarm。
// snapshot UpdatedAt 老于阈值（standing 2min / weakBio 30min）视为该实体 push 异常，对应字段
// 不参与 max。LastActiveTs 不做 staleness 过滤——它是"最后活动时刻"历史戳，人走/雷达离线后
// 仍应显示 "Active Xm ago"（无人 ≠ 无最后活动）。
//
// Visitor 三字段由 VisitorDeriver 按 owning card_id 注入 cardVisitor，写 target 时合入。
//
// 单 owner 不破：本 holder 是 nullable 状态持有者；最终唯一 Hash writer 仍是
// SensorStateProjector（target.state 分支）和 VisitorDeriver（午夜 reset）。

package service

import (
	"context"
	"net/netip"
	"sync"
	"time"

	"owl-common/card"
)

// StandingSnapshotStaleMs Standing 字段允许的最大 snapshot 静默间隔（2min）。
// 超过此窗口的 snapshot 视为该实体异常（进程卡住 / push 持续丢），其 standing 不参与 max。
// 跟 sensor 端 standing 每分钟心跳 push 契约耦合（FollowUp-4）：1 push 容错。
const StandingSnapshotStaleMs int64 = 2 * 60 * 1000

// WeakBioSnapshotStaleMs WeakBio 字段允许的最大 snapshot 静默间隔（30min，对齐 sensor
// aggregator weakBioWindowMs lazy 滑窗）。
//
// 设计意图（W3，[[target_state_weak_bio_signal_design]]）：
//   - sensor aggregator lazy drop 30min 外事件，空闲老人卡 30min 后 score 自然回 0
//   - 但 lazy 设计依赖下次 event 触发；空闲长时间 sensor 不 push → cardagg 卡老值
//   - cardagg 端 30min staleness 自治：UpdatedAt 老于 30min → 视为 stale 不参与 max
//   - 等效"如果 30min 内无新 vital alarm，FE 横条应回 None"
const WeakBioSnapshotStaleMs int64 = 30 * 60 * 1000

// TargetMerger 维护 per-entity TargetState snapshot + per-card visitor 注入。
// OnDeviceTarget 传入实体 TargetState，反查 owning card 后返回折叠聚合的 card.TargetState
// （合并该 card 名下所有实体的最新 snapshot）。
type TargetMerger struct {
	cache *SpatialCache

	mu sync.Mutex
	// 实体 CIDR string（/88 room 或 /96 bed）→ 最近 sensor 发的该实体 TargetState
	entitySnapshots map[string]*card.TargetState
	// owning card_id → VisitorDeriver 注入的 visitor 三字段（写 card.target 时合进去）
	cardVisitor map[string]visitorFields
}

// visitorFields VisitorDeriver 注入的字段（card.TargetState 子集）。
type visitorFields struct {
	visitorStartTs  int64
	hasVisitorToday bool
}

// NewTargetMerger 构造。cache 必填（实体→owning card 反查 + 列同 card 名下实体）。
func NewTargetMerger(cache *SpatialCache) *TargetMerger {
	return &TargetMerger{
		cache:           cache,
		entitySnapshots: make(map[string]*card.TargetState),
		cardVisitor:     make(map[string]visitorFields),
	}
}

// OnDeviceTarget 收到 sensor 派的实体 TargetState（subject = /88 room 或 /96 bed CIDR）；
// 经 entry 表反查 owning card_id，返回 (owning cardID, 折叠聚合后的 card.TargetState)。
// subject 非法 / 未注册 entry（schema 缺对应实体行）→ 返回 ("", nil)，caller drop。
func (m *TargetMerger) OnDeviceTarget(_ context.Context, subject string, ts *card.TargetState) (string, *card.TargetState) {
	if subject == "" || ts == nil {
		return "", nil
	}
	pfx, err := netip.ParsePrefix(subject)
	if err != nil {
		return "", nil
	}
	ent := m.cache.LookupEntry(pfx)
	if ent == nil {
		return "", nil
	}
	cardID := ent.CardID.String()

	m.mu.Lock()
	m.entitySnapshots[subject] = ts
	visitor := m.cardVisitor[cardID]
	m.mu.Unlock()

	return cardID, m.mergeForCard(cardID, visitor)
}

// ApplyVisitor 由 VisitorDeriver 调用：注入某 card 的 visitor 三字段（cardID = owning card_id）
// 到 owner view + 返回该 card 当前折叠聚合的 TargetState（写 Hash 用）。
func (m *TargetMerger) ApplyVisitor(_ context.Context, cardID string, v visitorFields) *card.TargetState {
	if cardID == "" {
		return nil
	}
	m.mu.Lock()
	m.cardVisitor[cardID] = v
	m.mu.Unlock()
	return m.mergeForCard(cardID, v)
}

// ForgetCard cardChange 删卡时调（清 visitor 累积）。
func (m *TargetMerger) ForgetCard(cardID string) {
	if cardID == "" {
		return
	}
	m.mu.Lock()
	delete(m.cardVisitor, cardID)
	m.mu.Unlock()
}

// ResetAllVisitor reset 事件触发：清所有卡的 visitor 累积（实体 snapshot 不动）。
func (m *TargetMerger) ResetAllVisitor() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.cardVisitor = make(map[string]visitorFields)
	m.mu.Unlock()
}

// mergeForCard 遍历 owning card 名下所有 room/bed 实体，max-merge 各实体最新 snapshot 三字段
// + 叠加 visitor，得到写 card:state.target 的最终 payload。
//
// UpdatedAt 取参与合并实体的最大 ts（不强制 now，让下游 anchor 计算可重现）；
// SpatialAnchor 跟最新 UpdatedAt 的实体走（target 物理位置跟最新 event）。
func (m *TargetMerger) mergeForCard(cardID string, v visitorFields) *card.TargetState {
	pfx, err := netip.ParsePrefix(cardID)
	if err != nil {
		return nil
	}
	out := &card.TargetState{
		VisitorStartTs:  v.visitorStartTs,
		HasVisitorToday: v.hasVisitorToday,
	}
	nowMs := time.Now().UnixMilli()

	m.mu.Lock()
	defer m.mu.Unlock()
	var anchorWinnerTs int64
	for _, e := range m.cache.EntryByCard(pfx) {
		if !e.IsRoom() && !e.IsBed() {
			continue
		}
		ts := m.entitySnapshots[e.Prefix.String()]
		if ts == nil {
			continue
		}
		// LastActiveTs：历史戳，无 staleness 过滤（无人/离线仍显示 "Active Xm ago"）。
		if ts.LastActiveTs > out.LastActiveTs {
			out.LastActiveTs = ts.LastActiveTs
		}
		if ts.UpdatedAt > 0 && nowMs-ts.UpdatedAt <= StandingSnapshotStaleMs && ts.StandingContinuousMin > out.StandingContinuousMin {
			out.StandingContinuousMin = ts.StandingContinuousMin
		}
		if ts.UpdatedAt > 0 && nowMs-ts.UpdatedAt <= WeakBioSnapshotStaleMs && ts.WeakBiometricSignal > out.WeakBiometricSignal {
			out.WeakBiometricSignal = ts.WeakBiometricSignal
		}
		if ts.UpdatedAt > out.UpdatedAt {
			out.UpdatedAt = ts.UpdatedAt
		}
		if ts.UpdatedAt > anchorWinnerTs && ts.SpatialAnchor != "" {
			anchorWinnerTs = ts.UpdatedAt
			out.SpatialAnchor = ts.SpatialAnchor
		}
	}
	return out
}

// VisitorFields VisitorDeriver 调 ApplyVisitor 时构造此结构。
type VisitorFields = visitorFields

// MakeVisitorFields VisitorDeriver 构造 visitor 字段（避免暴露 unexported struct）。
func MakeVisitorFields(startTs int64, today bool) VisitorFields {
	return visitorFields{
		visitorStartTs:  startTs,
		hasVisitorToday: today,
	}
}
