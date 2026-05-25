package zonealarm

import (
	"context"
	"net/netip"
	"sync"
	"time"

	"owl-common/card"
)

// StateSnapshot zonealarm Tick 时拿的 sensor 派生 state 快照（rooms + beds + 元数据）。
//
// 由 wiring 层从 StreamPublisher.SnapshotRoomStates/SnapshotBedStates 组装。
type StateSnapshot struct {
	// Rooms key=/88 CIDR；State 是深拷贝 + Kind (RoomTypeDefault/Bathroom)
	Rooms map[string]RoomEntry
	// Beds key=/96 CIDR；State 深拷贝
	Beds map[string]*card.BedState
}

// RoomEntry snapshot 单项（room/bathroom 一并；Kind 区分）。
type RoomEntry struct {
	State *card.RoomState
	Kind  int // card.RoomTypeDefault / RoomTypeBathroom
}

// Evaluator 跑 state-anchored rules 的纯函数 + fired gate 维护。
//
// fired gate keyed on (zoneCIDR, AlarmType)：rule 满足阈值首次 fire 即置 true；
// 任一 KeepCounting condition 不满足、device-level disabled 任一 → 清零
// （下次条件复成立从锚点重新计）。
//
// per-device 阈值/使能：resolver 注入；nil 时退化按 yaml 默认全启用（NopResolver 等价）。
// devLookup 注入：把 zone CIDR (/96 bed / /88 room) → device /128 用于 resolver 查 spatial_config。
type Evaluator struct {
	mu        sync.Mutex
	fired     map[FireKey]bool
	resolver  ThresholdResolver
	devLookup DeviceLookup
}

func NewEvaluator() *Evaluator {
	return &Evaluator{
		fired:    make(map[FireKey]bool),
		resolver: NopResolver{},
	}
}

// SetResolver wiring 层注入；nil 时退化为 NopResolver（按 yaml 默认全启用，无 per-device 覆盖）。
func (e *Evaluator) SetResolver(r ThresholdResolver) {
	if r == nil {
		r = NopResolver{}
	}
	e.mu.Lock()
	e.resolver = r
	e.mu.Unlock()
}

// SetDeviceLookup wiring 层注入；nil 时 resolver 拿不到 device addr，返 (fallback, true)。
func (e *Evaluator) SetDeviceLookup(l DeviceLookup) {
	e.mu.Lock()
	e.devLookup = l
	e.mu.Unlock()
}

// FireDecision Check 输出：上层 Supervisor 拿到后调 firer.Fire（锁外）。
type FireDecision struct {
	Rule     *Rule
	Event    FireEvent
}

// Check 跑 rules 列表对 snap 做一轮判定，返回需要 fire 的决定（锁外执行避免持锁 publish）。
//
// ctx 用于 resolver DB 查询；loc 为 timezone（per-card 注入留后续，传 nil 退化 UTC）。
func (e *Evaluator) Check(ctx context.Context, rules []Rule, snap StateSnapshot, nowMs int64, loc *time.Location) []FireDecision {
	if loc == nil {
		loc = time.UTC
	}
	var out []FireDecision
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := range rules {
		r := &rules[i]
		switch r.Anchor {
		case EntityBathroom:
			for cidr, entry := range snap.Rooms {
				if entry.Kind != card.RoomTypeBathroom {
					continue
				}
				if dec := e.evalRoomLocked(ctx, r, cidr, entry.State, snap, nowMs, loc); dec != nil {
					out = append(out, *dec)
				}
			}
		case EntityRoom:
			for cidr, entry := range snap.Rooms {
				if entry.Kind != card.RoomTypeDefault {
					continue
				}
				if dec := e.evalRoomLocked(ctx, r, cidr, entry.State, snap, nowMs, loc); dec != nil {
					out = append(out, *dec)
				}
			}
		case EntityBed:
			for cidr, bs := range snap.Beds {
				if dec := e.evalBedLocked(ctx, r, cidr, bs, snap, nowMs, loc); dec != nil {
					out = append(out, *dec)
				}
			}
		}
	}
	return out
}

// evalRoomLocked 处理 Anchor=Room/Bathroom 一条 rule + 一个 room entity。
// caller 必须已持锁。
func (e *Evaluator) evalRoomLocked(ctx context.Context, r *Rule, cidr string, rs *card.RoomState,
	snap StateSnapshot, nowMs int64, loc *time.Location) *FireDecision {

	if rs == nil {
		return nil
	}
	key := FireKey{ZoneID: cidr, AlarmType: r.AlarmType}

	if r.NightOnly != nil && !r.NightOnly.Active(time.UnixMilli(nowMs).In(loc)) {
		delete(e.fired, key)
		return nil
	}
	if !conditionsSatisfiedForRoom(r.KeepCounting, cidr, rs, snap) {
		delete(e.fired, key)
		return nil
	}

	elapsedSec := readRoomElapsedSec(rs, r.AnchorField, nowMs)
	if elapsedSec <= 0 {
		return nil
	}

	defaultSec := r.ThresholdDaySec
	if r.ThresholdNightSec > 0 && card.IsRiskTime(nowMs, loc) {
		defaultSec = r.ThresholdNightSec
	}
	thSec, enabled := e.resolveLocked(ctx, cidr, r.AlarmType, defaultSec)
	if !enabled {
		delete(e.fired, key)
		return nil
	}
	if e.fired[key] {
		return nil
	}
	if elapsedSec < thSec {
		return nil
	}

	e.fired[key] = true
	return &FireDecision{
		Rule: r,
		Event: FireEvent{
			Key:      key,
			Level:    r.Level,
			AnchorTs: nowMs - int64(elapsedSec)*1000, // synthetic anchor for downstream
			NowMs:    nowMs,
			Snapshot: rs,
		},
	}
}

// evalBedLocked 处理 Anchor=Bed 一条 rule + 一个 bed entity。caller 必须已持锁。
func (e *Evaluator) evalBedLocked(ctx context.Context, r *Rule, cidr string, bs *card.BedState,
	snap StateSnapshot, nowMs int64, loc *time.Location) *FireDecision {

	if bs == nil {
		return nil
	}
	key := FireKey{ZoneID: cidr, AlarmType: r.AlarmType}

	if r.NightOnly != nil && !r.NightOnly.Active(time.UnixMilli(nowMs).In(loc)) {
		delete(e.fired, key)
		return nil
	}
	if !conditionsSatisfiedForBed(r.KeepCounting, cidr, bs, snap) {
		delete(e.fired, key)
		return nil
	}

	anchorTs := readBedAnchor(bs, r.AnchorField)
	if anchorTs == 0 {
		return nil
	}

	defaultSec := r.ThresholdDaySec
	if r.ThresholdNightSec > 0 && card.IsRiskTime(nowMs, loc) {
		defaultSec = r.ThresholdNightSec
	}
	thSec, enabled := e.resolveLocked(ctx, cidr, r.AlarmType, defaultSec)
	if !enabled {
		delete(e.fired, key)
		return nil
	}
	if e.fired[key] {
		return nil
	}
	if (nowMs-anchorTs)/1000 < int64(thSec) {
		return nil
	}

	e.fired[key] = true
	return &FireDecision{
		Rule: r,
		Event: FireEvent{
			Key:      key,
			Level:    r.Level,
			AnchorTs: anchorTs,
			NowMs:    nowMs,
			Snapshot: bs,
		},
	}
}

// resolveLocked 查 per-device threshold + enabled。
// 找不到 devLookup 或拿不到 device addr → 用 fallback + 默认 enabled (避免设备未注册时静默不报)。
func (e *Evaluator) resolveLocked(ctx context.Context, zoneCIDR, alarmType string, fallbackSec int) (int, bool) {
	deviceAddr := ""
	if e.devLookup != nil {
		deviceAddr = e.devLookup.FindPrimaryDevice(zoneCIDR, alarmType)
	}
	if deviceAddr == "" {
		return fallbackSec, true
	}
	return e.resolver.Resolve(ctx, deviceAddr, alarmType, fallbackSec)
}

// readRoomElapsedSec 给定 AnchorField，从 RoomState 算 elapsed 秒数（统一比阈值 sec 单位）。
//
// 两种语义：
//   - AnchorAloneContinuousMin: rs.AloneContinuousMin (MIN 整数 sensor 算好) * 60
//   - AnchorLastExitTs:         (nowMs - rs.LastExitTs) / 1000 — TS 现算
//
// 返 <= 0 表示无效（无锚点 / 还没到 elapsed）。
func readRoomElapsedSec(rs *card.RoomState, field string, nowMs int64) int {
	switch field {
	case AnchorAloneContinuousMin:
		return rs.AloneContinuousMin * 60
	case AnchorLastExitTs:
		if rs.LastExitTs > 0 && nowMs > rs.LastExitTs {
			return int((nowMs - rs.LastExitTs) / 1000)
		}
	}
	return 0
}

// readBedAnchor 从 BedState 读 AnchorField 对应 ts。
// LeftBed 用 BedStatusTs（NotInBed 状态下 = 离床时刻；BedStatus==InBed 时返 0，因为 KeepCounting
// vacant 条件会先把 fired 清掉）。
func readBedAnchor(bs *card.BedState, field string) int64 {
	if field == AnchorBedStatusTs {
		return bs.BedStatusTs
	}
	return 0
}

// conditionsSatisfiedForRoom rule.Anchor=Room/Bathroom 时检查 KeepCounting 所有条件。
//
// Self 条件直接看 rs；peer Bed 条件 → 找同 /88 内所有 beds 全得满足（任一 occupied 即 fail）；
// peer Room/Bathroom 条件 → 当前无业务用例（room 之间无 peer 关系），返 true 兼容未来扩展。
func conditionsSatisfiedForRoom(conds []Condition, cidr string, rs *card.RoomState, snap StateSnapshot) bool {
	for _, c := range conds {
		if c.UsesSelf || c.Entity == EntityRoom || c.Entity == EntityBathroom {
			// Self 检查 — 当前 entity 必须满足
			if c.UsesSelf {
				if !roomStateMatches(rs, c.State) {
					return false
				}
				continue
			}
			// Peer Room/Bathroom（业务未用）— 兼容性 no-op
			continue
		}
		// peer Bed：同 /88 内所有 beds 必须满足（任一不满足即 fail）
		if c.Entity == EntityBed {
			for bedCIDR, bs := range snap.Beds {
				if deriveRoomCIDRFromBed(bedCIDR) != cidr {
					continue
				}
				if !bedStateMatches(bs, c.State) {
					return false
				}
			}
		}
	}
	return true
}

// conditionsSatisfiedForBed rule.Anchor=Bed 时检查 KeepCounting 所有条件。
//
// Self 看 bedState；peer Room 看同 /88 (deriveRoomCIDRFromBed) 对应 RoomEntry。
func conditionsSatisfiedForBed(conds []Condition, cidr string, bs *card.BedState, snap StateSnapshot) bool {
	for _, c := range conds {
		if c.UsesSelf || c.Entity == EntityBed {
			if !bedStateMatches(bs, c.State) {
				return false
			}
			continue
		}
		if c.Entity == EntityRoom || c.Entity == EntityBathroom {
			roomCIDR := deriveRoomCIDRFromBed(cidr)
			if roomCIDR == "" {
				continue
			}
			entry, ok := snap.Rooms[roomCIDR]
			if !ok {
				// peer 不存在 → 视作满足条件（房间未曾被 publish 等同 vacant）
				continue
			}
			if !roomStateMatches(entry.State, c.State) {
				return false
			}
		}
	}
	return true
}

// roomStateMatches 判 RoomState 是否处于 want 状态。
//
//	Vacant   = TotalPeople == 0
//	Occupied = TotalPeople > 0
//	Alone    = TotalPeople == 1
func roomStateMatches(rs *card.RoomState, want EntityState) bool {
	if rs == nil {
		return want == StateVacant // nil 视作空
	}
	switch want {
	case StateVacant:
		return rs.TotalPeople == 0
	case StateOccupied:
		return rs.TotalPeople > 0
	case StateAlone:
		return rs.TotalPeople == 1
	}
	return false
}

// bedStateMatches 判 BedState 是否处于 want 状态。
//
//	Vacant   = BedStatus == 1 (NotInBed)
//	Occupied = BedStatus == 0 (InBed)
//	Alone    无 bed 语义，当 Vacant 处理（无人 = 床上空）
func bedStateMatches(bs *card.BedState, want EntityState) bool {
	if bs == nil {
		return want == StateVacant
	}
	switch want {
	case StateVacant:
		return bs.BedStatus == 1
	case StateOccupied:
		return bs.BedStatus == 0
	case StateAlone:
		return bs.BedStatus == 1
	}
	return false
}

// deriveRoomCIDRFromBed bed /96 → room /88（截断 8 bit）。
// 与 zoneengine.deriveRoomZoneIDFromBed 同语义；避免跨 package 依赖在此重实现。
func deriveRoomCIDRFromBed(bedCIDR string) string {
	if bedCIDR == "" {
		return ""
	}
	p, err := netip.ParsePrefix(bedCIDR)
	if err != nil || p.Bits() < 88 {
		return ""
	}
	return netip.PrefixFrom(p.Addr(), 88).Masked().String()
}

// ClearFired 测试 / hot reload 时清 fire-once gate。
func (e *Evaluator) ClearFired() {
	e.mu.Lock()
	e.fired = make(map[FireKey]bool)
	e.mu.Unlock()
}

// FiredCount 当前 fired 标记数（监控 / 测试用）。
func (e *Evaluator) FiredCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.fired)
}
