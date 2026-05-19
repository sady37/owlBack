// target_merger.go — per-device TargetState → per-card max-merge holder。
//
// 背景（doc/card_display.md §4.3, §9）：
//   - sensor 按 /128 device 发 target.state（msg.SubjectEntity = device IPv6 canonical 字符串）
//   - 同 card（/88 room 或 /96 bed）下多 radar 各发各的 TargetState
//   - cardagg 需把 owning card 下所有 device 的 TargetState max-merge 写单 card:state.target
//
// 合并字段（max-merge across owning card 的 devices）：
//   LastActiveTs / StandingContinuousMin / WeakBiometricSignal
//
// **LastActiveTs 的 device-offline 过滤**：device 离线后 snapshot 永不更新，max-merge
// 还会让 "Active 24h ago" 一直显示——这是误导（不知道 device 是失联还是真的没活动）。
// merger 注入 `DeviceOnlineChecker`（cardagg DeviceStatusTracker），offline device 的
// snapshot **不参与 LastActiveTs max**。
//
// **StandingContinuousMin 的双重过滤**：
//   1) offline 过滤（同 LastActive）—— device 失联不参与
//   2) **snapshot UpdatedAt 短窗 staleness**（2min）—— device 在线但 snapshot 老于 2min
//      视为该 device push 异常 / 进程卡住，不参与 max
//
// 双重必要：standing 是瞬时状态（人坐下后 sensor 必须 reset 0），任何 stale snapshot
// 都会直接造成 risk_evaluator false-positive Attention/Risk → 误报 alarm。IsOnline
// 兜不住 "device 网络通但 sensor 进程卡住" 的边界，必须靠 UpdatedAt 短窗补防。
//
// **契约**：sensor handleMonitorFrame 实施时（P3-FollowUp-4）必须保证每分钟心跳 push
// 一次 standing 值（即使无变化），否则 2min 阈值会误判正常持续站立为 stale。
//
// WeakBio 暂保持原 max（独立 staleness 议题 follow-up）。
//
// Visitor 三字段不在此合并 —— 由 VisitorDeriver 直接写 /88 room card.target，
// 经由 ApplyVisitor() 注入到 TargetMerger 的 ownerView 后参与下次写入。
//
// 单 owner 不破：本 holder 是 nullable 状态持有者；最终唯一 Hash writer 仍是
// SensorStateProjector（target.state 分支）和 VisitorDeriver（visitor 字段）。

package service

import (
	"context"
	"net/netip"
	"sync"
	"time"

	"owl-common/card"
)

// StandingSnapshotStaleMs Standing 字段允许的最大 snapshot 静默间隔（2min）。
// 超过此窗口的 snapshot 视为该 device 异常（进程卡住 / push 持续丢），其 standing 不参与 max。
// 跟 sensor 端 standing 每分钟心跳 push 契约耦合（FollowUp-4）：1 push 容错。
const StandingSnapshotStaleMs int64 = 2 * 60 * 1000

// DeviceOnlineChecker main wiring 注入 closure，返回 device 是否在线。
// 避免 service → consumer 反向导入；签名接收 device addr canonical IPv6 string。
type DeviceOnlineChecker func(deviceAddr string) bool

// TargetMerger 维护 per-device TargetState snapshot + per-card visitor 注入，
// 给 SensorStateProjector target.state 分支用：传入 device-level TargetState 返回
// owning cardID + merged card.TargetState（合并该 card 下所有 device 的快照）。
type TargetMerger struct {
	metaCache *DeviceMetaCache
	online    DeviceOnlineChecker // nil 时退化为 "总在线"（兼容老测试 / 未 wire 阶段）

	mu sync.Mutex
	// deviceAddr canonical IPv6 string → 最近 sensor 发的 device-level TargetState
	deviceSnapshots map[string]*card.TargetState
	// cardID → VisitorDeriver 注入的 visitor 三字段（写 card.target 时合进去）
	cardVisitor map[string]visitorFields
}

// SetOnlineChecker main wiring 注入 device 在线判定 callback。
// 未注入时所有 device 视为在线，LastActive merge 不过滤。
func (m *TargetMerger) SetOnlineChecker(fn DeviceOnlineChecker) {
	if m == nil {
		return
	}
	m.online = fn
}

// isDeviceOnline 判断 device 是否在线。nil checker 时 fallback true（兼容未 wire 场景）。
func (m *TargetMerger) isDeviceOnline(deviceAddr string) bool {
	if m == nil || m.online == nil {
		return true
	}
	return m.online(deviceAddr)
}

// visitorFields VisitorDeriver 注入的字段（card.TargetState 子集）。
type visitorFields struct {
	visitorStartTs     int64
	todayMaxVisitorMin int
	hasVisitorToday    bool
}

// NewTargetMerger 构造。metaCache 必填（device→card 反查 + 列同 card devices）。
func NewTargetMerger(metaCache *DeviceMetaCache) *TargetMerger {
	return &TargetMerger{
		metaCache:       metaCache,
		deviceSnapshots: make(map[string]*card.TargetState),
		cardVisitor:     make(map[string]visitorFields),
	}
}

// OnDeviceTarget 收到 sensor 派的 per-device TargetState；返回 owning cardID +
// max-merge 后的 card.TargetState；cardID 空 → caller 跳过（unbound device）。
//
// deviceAddr 期望为 /128 IPv6 canonical 字符串（不是 CIDR）。
// 若解析失败或 device 未绑卡，返回 ("", nil)。
func (m *TargetMerger) OnDeviceTarget(ctx context.Context, deviceAddr string, ts *card.TargetState) (string, *card.TargetState) {
	if deviceAddr == "" || ts == nil {
		return "", nil
	}
	addr, err := netip.ParseAddr(deviceAddr)
	if err != nil {
		return "", nil
	}
	cardID := m.metaCache.LookupCardByDeviceAddr(ctx, addr)
	if cardID == "" {
		return "", nil
	}

	m.mu.Lock()
	m.deviceSnapshots[deviceAddr] = ts
	visitor := m.cardVisitor[cardID]
	m.mu.Unlock()

	merged := m.mergeForCard(ctx, cardID, visitor)
	return cardID, merged
}

// ApplyVisitor 由 VisitorDeriver 调用：注入某 /88 room card 的 visitor 三字段
// 到 owner view + 返回该 card 当前 merged TargetState（写 Hash 用）。
func (m *TargetMerger) ApplyVisitor(ctx context.Context, cardID string, v visitorFields) *card.TargetState {
	if cardID == "" {
		return nil
	}
	m.mu.Lock()
	m.cardVisitor[cardID] = v
	m.mu.Unlock()
	return m.mergeForCard(ctx, cardID, v)
}

// CurrentVisitor 读 VisitorDeriver 累积态（segment reset 用，避免重复触发午夜清理）。
func (m *TargetMerger) CurrentVisitor(cardID string) (int64, int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.cardVisitor[cardID]
	return v.visitorStartTs, v.todayMaxVisitorMin, v.hasVisitorToday
}

// ForgetDevice 解绑 / Cleanup 时清 device snapshot。
func (m *TargetMerger) ForgetDevice(deviceAddr string) {
	if deviceAddr == "" {
		return
	}
	m.mu.Lock()
	delete(m.deviceSnapshots, deviceAddr)
	m.mu.Unlock()
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

// ResetAllVisitor reset 事件触发：清所有卡的 visitor 累积（device snapshot 不动）。
func (m *TargetMerger) ResetAllVisitor() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.cardVisitor = make(map[string]visitorFields)
	m.mu.Unlock()
}

// mergeForCard 取 cardID 下所有 device 的最新 snapshot，max-merge 三字段 +
// 叠加 visitor 注入，得到写 card:state.target 的最终 payload。
//
// UpdatedAt 取参与合并字段的最大 ts（不强制 now，让下游 anchor 计算可重现）。
func (m *TargetMerger) mergeForCard(ctx context.Context, cardID string, v visitorFields) *card.TargetState {
	meta := m.metaCache.GetOrLoad(ctx, cardID)
	if meta == nil {
		return nil
	}

	out := &card.TargetState{
		VisitorStartTs:     v.visitorStartTs,
		TodayMaxVisitorMin: v.todayMaxVisitorMin,
		HasVisitorToday:    v.hasVisitorToday,
	}

	nowMs := time.Now().UnixMilli()

	m.mu.Lock()
	defer m.mu.Unlock()
	for addr := range meta.Devices {
		ts := m.deviceSnapshots[addr]
		if ts == nil {
			continue
		}
		online := m.isDeviceOnline(addr)
		// LastActiveTs：offline device 的 snapshot 不参与 max（避免"Active 24h ago"误显）。
		// 在线 device 的 LastActiveTs 即便很久（老人静坐看电视）也保留——是真实历史事实，
		// 配合 SceneState 让护士能正确解读。
		if online && ts.LastActiveTs > out.LastActiveTs {
			out.LastActiveTs = ts.LastActiveTs
		}
		// Standing：offline 过滤 + snapshot UpdatedAt 短窗 staleness（防 false Attention/Risk）。
		// 进入 risk_evaluator 的阈值字段必须严防 stale；契约要求 sensor 每分钟心跳 push。
		standingFresh := online && ts.UpdatedAt > 0 && nowMs-ts.UpdatedAt <= StandingSnapshotStaleMs
		if standingFresh && ts.StandingContinuousMin > out.StandingContinuousMin {
			out.StandingContinuousMin = ts.StandingContinuousMin
		}
		// WeakBio 暂不加过滤（独立议题 followup）
		if ts.WeakBiometricSignal > out.WeakBiometricSignal {
			out.WeakBiometricSignal = ts.WeakBiometricSignal
		}
		if ts.UpdatedAt > out.UpdatedAt {
			out.UpdatedAt = ts.UpdatedAt
		}
	}
	return out
}

// VisitorFields VisitorDeriver 调 ApplyVisitor 时构造此结构。
type VisitorFields = visitorFields

// MakeVisitorFields VisitorDeriver 构造 visitor 三字段（避免暴露 unexported struct）。
func MakeVisitorFields(startTs int64, maxMin int, today bool) VisitorFields {
	return visitorFields{
		visitorStartTs:     startTs,
		todayMaxVisitorMin: maxMin,
		hasVisitorToday:    today,
	}
}
