package service

// bed_presence_fusion.go — per-/96 bed 的 sleepad/radar InBed 双源状态跟踪。
//
// 用途：RoomState.TotalPeople dedup。radar firmware NumberPeople 偶尔漏数床上人
// （静卧低 RCS / 床边遮挡 / 极慢动 microdoppler），但 sleepad 压感对在床稳定。
// 反之 radar 看到的 bed-area track 在 NumberPeople 已计入。
//
// 推导公式（与 [[radar_activity_stats_on_event_stream]] presence 路径同源）：
//
//	Z = radar NumberPeople (room 总数)
//	X = sleepad InBed (per-bed bool, 0/1)
//	Y = radar bed-area InBed (per-bed bool, 0/1)
//	publishTotalPeople = Z + Σ(X for bed where X && !Y)
//
// 即只对"sleepad 看到 radar 漏数的床位"补偿，不影响其他床位。
//
// 不引入 zoneengine 内：BedZone FSM 把 sleepad/radar 两源融合成单一 enter/leave，
// 源信息丢失。这里旁路记账，供 stream_publisher 在 publish 前查询。
//
// 不进 alarm 决策路径（[[feedback_no_dynamic_threshold_modulation]]）：dedup 是 presence
// 计数 fusion，不影响 verdict/severity/routing/dedup/timing。

import (
	"net/netip"
	"sync"
	"time"
)

// BedPresenceFusion 跟踪 per-/96 bed 的 sleepad/radar InBed 状态 + 时间戳。
type BedPresenceFusion struct {
	mu      sync.RWMutex
	entries map[string]*bedPresenceEntry // bedCIDR (/96) → state
	now     func() int64                 // 测试用注入
}

type bedPresenceEntry struct {
	sleepadInBed bool
	radarInBed   bool
	sleepadAt    int64 // 最近 sleepad 更新 ms
	radarAt      int64 // 最近 radar 更新 ms
}

// bedPresenceFreshMs 10min 内的 sleepad/radar 信号视作有效；超时跳过 dedup。
// sleepad 心跳 sample 周期 2-10s（依 sleepace adapter scheduler），10min 足以覆盖任何
// vendor backoff 场景；radar InBed alarm 是 edge 事件无心跳，10min stale 后视同未触发。
const bedPresenceFreshMs = 10 * 60 * 1000

func NewBedPresenceFusion() *BedPresenceFusion {
	return &BedPresenceFusion{
		entries: make(map[string]*bedPresenceEntry),
		now:     func() int64 { return time.Now().UnixMilli() },
	}
}

func (f *BedPresenceFusion) getOrCreate(bedCIDR string) *bedPresenceEntry {
	e := f.entries[bedCIDR]
	if e == nil {
		e = &bedPresenceEntry{}
		f.entries[bedCIDR] = e
	}
	return e
}

// SetSleepad sleepad InBed/LeftBed 翻转时调；ts=0 时取 now。
func (f *BedPresenceFusion) SetSleepad(bedCIDR string, inBed bool, ts int64) {
	if bedCIDR == "" {
		return
	}
	if ts == 0 {
		ts = f.now()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	e := f.getOrCreate(bedCIDR)
	e.sleepadInBed = inBed
	e.sleepadAt = ts
}

// SetRadar radar InBed/LeftBed 翻转时调；ts=0 时取 now。
func (f *BedPresenceFusion) SetRadar(bedCIDR string, inBed bool, ts int64) {
	if bedCIDR == "" {
		return
	}
	if ts == 0 {
		ts = f.now()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	e := f.getOrCreate(bedCIDR)
	e.radarInBed = inBed
	e.radarAt = ts
}

// ExtraPeopleInRoom 在 /88 roomCIDR 下扫所有已记录的 /96 bed，累加
// "sleepad InBed && !radar InBed" 的床数（每床贡献 1 个 radar 漏数的人）。
// stale entries（sleepad 心跳超 10min）跳过，避免久未上报的 sleepad 长期挂"InBed"
// 导致 phantom +1。
func (f *BedPresenceFusion) ExtraPeopleInRoom(roomCIDR string) int {
	if roomCIDR == "" {
		return 0
	}
	roomPfx, err := netip.ParsePrefix(roomCIDR)
	if err != nil {
		return 0
	}
	now := f.now()
	f.mu.RLock()
	defer f.mu.RUnlock()
	extra := 0
	for bedCIDR, e := range f.entries {
		if !e.sleepadInBed {
			continue
		}
		if now-e.sleepadAt > bedPresenceFreshMs {
			continue
		}
		bedPfx, err := netip.ParsePrefix(bedCIDR)
		if err != nil {
			continue
		}
		if !roomPfx.Contains(bedPfx.Addr()) {
			continue
		}
		// radar fresh 时信 Y；stale 时视作 Y=0（保守，倾向补偿——避免 radar 漏数后无人发声）
		radarFresh := e.radarAt > 0 && now-e.radarAt <= bedPresenceFreshMs
		if radarFresh && e.radarInBed {
			continue
		}
		extra++
	}
	return extra
}
